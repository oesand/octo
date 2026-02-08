package mc

import (
	"errors"
	"github.com/oesand/octo"
	"github.com/oesand/octo/pm"
	"sync"
	"time"
)

const DefaultJanitorInterval = 5 * time.Minute

// Inject injects a MemCache into the container if not already registered.
func Inject(container *octo.Container, options ...Option) *MemCache {
	cache := octo.TryResolve[*MemCache](container)
	if cache == nil {
		cache = new(MemCache)
		octo.InjectValue(container, cache)
	}

	for _, option := range options {
		option(cache)
	}

	return cache
}

func New(options ...Option) *MemCache {
	cache := new(MemCache)

	for _, option := range options {
		option(cache)
	}

	return cache
}

// MemCache represents a thread-safe in-memory cache with key-based locking.
// Each entry expires automatically after a given duration and is purged by a janitor goroutine.
type MemCache struct {
	keyedMu pm.KeyLock[string]
	entries sync.Map

	usageEvictor usageEvictor

	janitorMu       sync.Mutex
	janitorInterval time.Duration
	janitor         *time.Ticker
}

type cacheEntry struct {
	expiredAt time.Time
	value     any
}

func (cache *MemCache) checkIfRunJanitor() {
	cache.janitorMu.Lock()
	defer cache.janitorMu.Unlock()

	if cache.janitor != nil {
		return
	}

	interval := cache.janitorInterval
	if interval == 0 {
		interval = DefaultJanitorInterval
	}
	janitor := time.NewTicker(interval)
	cache.janitor = janitor
	go func() {
		for {
			<-janitor.C
			cache.janitorMu.Lock()
			if cache.janitor == nil {
				break
			}
			cache.janitorMu.Unlock()
			if cache.janitorPurge() {
				break
			}
		}
	}()
}

func (cache *MemCache) removeFromEvictor(key string) {
	if evc := cache.usageEvictor; evc != nil {
		evc.Remove(key)
	}
}

func (cache *MemCache) pushUsedToEvictor(key string) {
	if evc := cache.usageEvictor; evc != nil {
		evc.Used(key)
	}
}

func (cache *MemCache) janitorPurge() bool {
	now := time.Now()

	var remainEntries bool
	var cleanKeys pm.Set[string]

	cache.entries.Range(func(k, e interface{}) bool {
		key := k.(string)
		cache.keyedMu.Lock(key)

		entry := e.(*cacheEntry)
		if now.Before(entry.expiredAt) {
			cache.keyedMu.Unlock(key)
			remainEntries = true
			return true
		}

		cleanKeys.Add(key)
		return true
	})

	if remainEntries {
		if evc := cache.usageEvictor; evc != nil {
			excess := evc.GetExcess() - len(cleanKeys)
			if excess > 0 {
				i := 0
				for key := range evc.IterWorst() {
					if cleanKeys.Has(key) {
						continue
					}

					cache.keyedMu.Lock(key)
					cleanKeys.Add(key)
					i++

					if i >= excess {
						break
					}
				}
			}
		}
	}

	for key := range cleanKeys {
		cache.entries.Delete(key)
		cache.removeFromEvictor(key)
		cache.keyedMu.Unlock(key)
	}

	if !remainEntries {
		cache.janitorMu.Lock()
		if cache.janitor != nil {
			cache.janitor.Stop()
			cache.janitor = nil
		}
		cache.janitorMu.Unlock()
		return true
	}
	return false
}

// TryGet attempts to retrieve a value from the cache.
// Returns (found, ttl, value). If the key is expired or missing, found=false.
func TryGet[T any](cache *MemCache, key string) (bool, time.Duration, T) {
	cache.keyedMu.Lock(key)
	defer cache.keyedMu.Unlock(key)

	var nilVal T
	e, ok := cache.entries.Load(key)
	if !ok {
		return false, 0, nilVal
	}

	entry := e.(*cacheEntry)
	ttl := time.Until(entry.expiredAt)
	if ttl <= 0 {
		cache.entries.Delete(key)
		cache.removeFromEvictor(key)
		return false, 0, nilVal
	}
	cache.pushUsedToEvictor(key)

	return true, ttl, entry.value.(T)
}

// GetOrCreate retrieves the value for a key or generates it using the provider.
// If the key is expired or missing, provider() is called and stored for the specified duration.
func GetOrCreate[T any](cache *MemCache, key string, d time.Duration, provider func() (T, error)) (T, error) {
	var nilVal T
	if d < 5*time.Millisecond {
		return nilVal, errors.New("cache duration must be least 5 milliseconds")
	}

	cache.keyedMu.Lock(key)
	defer cache.keyedMu.Unlock(key)

	var entry *cacheEntry
	e, has := cache.entries.Load(key)
	if has {
		entry = e.(*cacheEntry)
		if time.Now().Before(entry.expiredAt) {
			return entry.value.(T), nil
		}
	}

	value, err := provider()
	if err != nil {
		if has {
			cache.entries.Delete(key)
			cache.removeFromEvictor(key)
		}
		return value, err
	}

	if !has {
		entry = &cacheEntry{}
		cache.entries.Store(key, entry)
	}

	entry.expiredAt = time.Now().Add(d)
	entry.value = value

	cache.checkIfRunJanitor()
	cache.pushUsedToEvictor(key)

	return value, nil
}

// ExtendUntil changes expiration time of cache entry if exists
func ExtendUntil(cache *MemCache, key string, expiredAt time.Time) bool {
	cache.keyedMu.Lock(key)
	defer cache.keyedMu.Unlock(key)

	var entry *cacheEntry
	ce, has := cache.entries.Load(key)
	if !has {
		return false
	}

	entry = ce.(*cacheEntry)
	entry.expiredAt = expiredAt
	return true
}

// Forgot removes cache entry if exists
func Forgot(cache *MemCache, key string) bool {
	cache.keyedMu.Lock(key)
	defer cache.keyedMu.Unlock(key)

	_, has := cache.entries.LoadAndDelete(key)
	cache.removeFromEvictor(key)
	return has
}

// JanitorPurge force run janitor purge.
//
// It is recommended to use it only in tests for immediate launch.
func JanitorPurge(cache *MemCache) {
	cache.janitorPurge()
}
