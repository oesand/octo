package mch

import (
	"errors"
	"github.com/oesand/octo"
	"sync"
	"time"
)

// Inject injects a MemCache into the container if not already registered.
func Inject(container *octo.Container) {
	manager := octo.TryResolve[*MemCache](container)
	if manager != nil {
		return
	}

	manager = &MemCache{}
	octo.InjectValue(container, manager)
}

// ValueProvider defines a function that provides a value to be cached.
// It returns the generated value and an optional error.
// If the provider returns a non-nil error, the value will not be stored in cache.
type ValueProvider[T any] func() (T, error)

// MemCache is a lightweight, thread-safe in-memory cache with expiration.
//
// Each key maps to a value and an expiration timestamp. Expired items are
// automatically purged by a background janitor goroutine, which is started
// on the first successful insertion and stops when the cache becomes empty.
type MemCache struct {
	mu    sync.Mutex
	store map[string]*cacheItem

	janitor *time.Ticker
}

type cacheItem struct {
	expiredAt int64
	value     any
}

func (cache *MemCache) checkIfRunJanitor(d time.Duration) {
	if cache.janitor != nil {
		return
	}

	janitor := time.NewTicker(d)
	cache.janitor = janitor
	go func() {
		defer janitor.Stop()
		for {
			<-janitor.C
			if cache.janitorPurge() {
				break
			}
		}
	}()
}

func (cache *MemCache) janitorPurge() bool {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	now := time.Now().UnixMilli()
	for key, item := range cache.store {
		if item.expiredAt > now {
			continue
		}

		delete(cache.store, key)
	}

	if len(cache.store) == 0 {
		cache.janitor = nil
		return true
	}
	return false
}

// GetOrCreate returns a cached value for the given key or creates it using the provider.
//
// If a valid (non-expired) cached value exists, it is returned immediately.
// Otherwise, the provider function is called to obtain a new value.
//
// The provider's return value is stored in the cache for the given duration.
// If the duration is less than 5 milliseconds, an error is returned and the value
// is not stored.
//
// Example:
//
//	val, err := GetOrCreate(cache, "user:42", time.Minute, func() (User, error) {
//	    return loadUserFromDB(42)
//	})
func GetOrCreate[T any](cache *MemCache, key string, d time.Duration, provider ValueProvider[T]) (T, error) {
	if d < 5*time.Millisecond {
		var val T
		return val, errors.New("cache duration must be least 5 milliseconds")
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	if cache.store == nil {
		cache.store = make(map[string]*cacheItem)
	}

	now := time.Now().UnixMilli()
	item, ok := cache.store[key]
	if ok {
		if now >= item.expiredAt {
			delete(cache.store, key)
		} else {
			return item.value.(T), nil
		}
	}

	value, err := provider()
	if err != nil {
		return value, err
	}

	if item == nil {
		item = &cacheItem{}
	}

	item.expiredAt = now + d.Milliseconds()
	item.value = value
	cache.store[key] = item

	cache.checkIfRunJanitor(10 * time.Minute)

	return value, nil
}
