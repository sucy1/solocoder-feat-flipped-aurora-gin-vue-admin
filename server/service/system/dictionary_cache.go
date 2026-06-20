package system

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type DictionaryCacheService struct{}

var DictionaryCacheServiceApp = new(DictionaryCacheService)

var (
	dictCache      = make(map[string]cachedDict)
	dictCacheMutex sync.RWMutex
	dictRefreshing int32
)

const (
	dictRefreshLockKey    = "gva:dict_cache:refresh_lock"
	dictRefreshLockTTL    = 30 * time.Second
	distributedLockMaxTry = 3
)

type cachedDict struct {
	Data     system.SysDictionary
	CachedAt time.Time
}

type distributedLock struct {
	heldByRedis bool
	heldByDB    bool
	sessionID   string
}

func acquireDistributedLock() (*distributedLock, error) {
	lock := &distributedLock{}
	if global.GVA_REDIS != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		sessionID := time.Now().Format("20060102150405.000000000")
		ok, err := global.GVA_REDIS.SetNX(ctx, dictRefreshLockKey, sessionID, dictRefreshLockTTL).Result()
		if err == nil && ok {
			lock.heldByRedis = true
			lock.sessionID = sessionID
			return lock, nil
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			return nil, err
		}
	}
	if global.GVA_DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var result int
		err := global.GVA_DB.WithContext(ctx).Raw("SELECT GET_LOCK(?, ?)", dictRefreshLockKey, 5).Scan(&result).Error
		if err == nil && result == 1 {
			lock.heldByDB = true
			return lock, nil
		}
	}
	return nil, errors.New("distributed lock acquisition failed")
}

func releaseDistributedLock(lock *distributedLock) {
	if lock == nil {
		return
	}
	if lock.heldByRedis && global.GVA_REDIS != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		val, err := global.GVA_REDIS.Get(ctx, dictRefreshLockKey).Result()
		if err == nil && val == lock.sessionID {
			global.GVA_REDIS.Del(ctx, dictRefreshLockKey)
		}
	}
	if lock.heldByDB && global.GVA_DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		global.GVA_DB.WithContext(ctx).Exec("SELECT RELEASE_LOCK(?)", dictRefreshLockKey)
	}
}

func (s *DictionaryCacheService) GetCachedDictionary(dictType string, id uint, status *bool) (system.SysDictionary, error) {
	dictCacheMutex.RLock()
	cached, exists := dictCache[dictType]
	dictCacheMutex.RUnlock()

	ttl := global.GVA_CONFIG.Dictionary.TTL
	if ttl == 0 {
		ttl = time.Hour
	}

	if exists && time.Since(cached.CachedAt) < ttl {
		return cached.Data, nil
	}

	result, err := DictionaryServiceApp.GetSysDictionary(dictType, id, status)
	if err != nil {
		if exists {
			return cached.Data, nil
		}
		return result, err
	}

	dictCacheMutex.Lock()
	dictCache[dictType] = cachedDict{Data: result, CachedAt: time.Now()}
	dictCacheMutex.Unlock()

	return result, nil
}

func (s *DictionaryCacheService) acquireLocalLock() bool {
	return atomic.CompareAndSwapInt32(&dictRefreshing, 0, 1)
}

func (s *DictionaryCacheService) releaseLocalLock() {
	atomic.StoreInt32(&dictRefreshing, 0)
}

func (s *DictionaryCacheService) RefreshCache() error {
	if !s.acquireLocalLock() {
		return errors.New("refresh already in progress")
	}
	defer s.releaseLocalLock()

	lock, err := acquireDistributedLock()
	if err != nil {
		return errors.New("refresh already in progress by another instance")
	}
	defer releaseDistributedLock(lock)

	var dicts []system.SysDictionary
	err = global.GVA_DB.Preload("SysDictionaryDetails", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ? and deleted_at is null", true).Order("sort")
	}).Find(&dicts).Error
	if err != nil {
		return err
	}

	newCache := make(map[string]cachedDict, len(dicts))
	now := time.Now()
	for _, d := range dicts {
		newCache[d.Type] = cachedDict{Data: d, CachedAt: now}
	}

	dictCacheMutex.Lock()
	dictCache = newCache
	dictCacheMutex.Unlock()

	return nil
}

func (s *DictionaryCacheService) IsRefreshing() bool {
	if atomic.LoadInt32(&dictRefreshing) == 1 {
		return true
	}
	if global.GVA_REDIS != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		exists, err := global.GVA_REDIS.Exists(ctx, dictRefreshLockKey).Result()
		if err == nil && exists > 0 {
			return true
		}
	}
	return false
}
