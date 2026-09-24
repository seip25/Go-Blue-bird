package core

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

var (
	cacheStore = make(map[string]cacheEntry)
	cacheMutex sync.RWMutex
)

func CacheSet(key string, value interface{}, ttl time.Duration) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cacheStore[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func CacheGet(key string) (interface{}, bool) {
	cacheMutex.RLock()
	entry, found := cacheStore[key]
	cacheMutex.RUnlock()

	if !found {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		cacheMutex.Lock()
		delete(cacheStore, key)
		cacheMutex.Unlock()
		return nil, false
	}
	return entry.value, true
}

func CacheDelete(key string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	delete(cacheStore, key)
}

func CacheClear() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cacheStore = make(map[string]cacheEntry)
}
