package system

import (
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

type cachedDict struct {
	Data     system.SysDictionary
	CachedAt time.Time
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

func (s *DictionaryCacheService) RefreshCache() error {
	if !atomic.CompareAndSwapInt32(&dictRefreshing, 0, 1) {
		return errors.New("refresh already in progress")
	}
	defer atomic.StoreInt32(&dictRefreshing, 0)

	var dicts []system.SysDictionary
	err := global.GVA_DB.Preload("SysDictionaryDetails", func(db *gorm.DB) *gorm.DB {
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
	return atomic.LoadInt32(&dictRefreshing) == 1
}
