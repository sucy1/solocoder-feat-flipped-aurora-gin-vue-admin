package system

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
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
	dictRefreshLockKey = "gva:dict_cache:refresh_lock"
	dictRefreshLockTTL = 30 * time.Second
)

type cachedDict struct {
	Data     system.SysDictionary
	CachedAt time.Time
}

type distLockRow struct {
	ID        uint   `gorm:"primarykey"`
	LockKey   string `gorm:"column:lock_key;uniqueIndex;size:128"`
	LockedBy  string `gorm:"column:locked_by;size:64"`
	LockedAt  time.Time
}

func (distLockRow) TableName() string {
	return "sys_distributed_lock"
}

func acquireDistributedLock() (*distLockRow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if global.GVA_REDIS != nil {
		sessionID := time.Now().Format("20060102150405.000000000")
		ok, err := global.GVA_REDIS.SetNX(ctx, dictRefreshLockKey, sessionID, dictRefreshLockTTL).Result()
		if err == nil && ok {
			return &distLockRow{LockKey: dictRefreshLockKey, LockedBy: sessionID}, nil
		}
	}

	if global.GVA_DB != nil {
		global.GVA_DB.WithContext(ctx).AutoMigrate(&distLockRow{})

		var existing distLockRow
		err := global.GVA_DB.WithContext(ctx).Where("lock_key = ?", dictRefreshLockKey).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sessionID := time.Now().Format("20060102150405.000000000")
			lockRow := distLockRow{
				LockKey:  dictRefreshLockKey,
				LockedBy: sessionID,
				LockedAt: time.Now(),
			}
			if err := global.GVA_DB.WithContext(ctx).Create(&lockRow).Error; err != nil {
				return nil, errors.New("distributed lock acquisition failed")
			}
			return &lockRow, nil
		}
		if err != nil {
			return nil, err
		}
		if time.Since(existing.LockedAt) > dictRefreshLockTTL {
			sessionID := time.Now().Format("20060102150405.000000000")
			result := global.GVA_DB.WithContext(ctx).
				Model(&distLockRow{}).
				Where("lock_key = ? AND locked_at = ?", dictRefreshLockKey, existing.LockedAt).
				Updates(map[string]interface{}{
					"locked_by": sessionID,
					"locked_at": time.Now(),
				})
			if result.Error != nil || result.RowsAffected == 0 {
				return nil, errors.New("distributed lock acquisition failed")
			}
			return &distLockRow{LockKey: dictRefreshLockKey, LockedBy: sessionID, LockedAt: time.Now()}, nil
		}
	}

	return nil, errors.New("distributed lock acquisition failed")
}

func releaseDistributedLock(lock *distLockRow) {
	if lock == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if global.GVA_REDIS != nil {
		val, err := global.GVA_REDIS.Get(ctx, dictRefreshLockKey).Result()
		if err == nil && val == lock.LockedBy {
			global.GVA_REDIS.Del(ctx, dictRefreshLockKey)
			return
		}
	}

	if global.GVA_DB != nil && lock.LockKey != "" {
		global.GVA_DB.WithContext(ctx).
			Where("lock_key = ? AND locked_by = ?", lock.LockKey, lock.LockedBy).
			Delete(&distLockRow{})
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
	if global.GVA_DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		var lock distLockRow
		err := global.GVA_DB.WithContext(ctx).Where("lock_key = ?", dictRefreshLockKey).First(&lock).Error
		if err == nil && time.Since(lock.LockedAt) <= dictRefreshLockTTL {
			return true
		}
	}
	return false
}
