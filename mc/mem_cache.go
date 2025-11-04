package mc

import (
	"errors"
	"github.com/oesand/octo"
	"github.com/oesand/octo/pm"
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

// MemCache represents a thread-safe in-memory cache with key-based locking.
// Each entry expires automatically after a given duration and is purged by a janitor goroutine.
type MemCache struct {
	entriesMu sync.RWMutex
	keyedMu   pm.KeyLock[string]
	entries   sync.Map

	checkMu sync.Mutex
	janitor *time.Ticker
}

type cacheEntry struct {
	expiredAt time.Time
	value     any
}

func (cache *MemCache) lockKey(key string) {
	cache.entriesMu.RLock()
	cache.keyedMu.Lock(key)
}

func (cache *MemCache) unlockKey(key string) {
	cache.keyedMu.Unlock(key)
	cache.entriesMu.RUnlock()
}

func (cache *MemCache) checkIfRunJanitor(d time.Duration) {
	cache.checkMu.Lock()
	defer cache.checkMu.Unlock()

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
	cache.entriesMu.Lock()
	defer cache.entriesMu.Unlock()

	now := time.Now()

	var canContinue bool
	var keys []string
	cache.entries.Range(func(k, e interface{}) bool {
		entry := e.(*cacheEntry)
		if now.Before(entry.expiredAt) {
			canContinue = true
			return true
		}

		keys = append(keys, k.(string))
		return true
	})

	for _, key := range keys {
		cache.entries.Delete(key)
	}

	if !canContinue {
		cache.janitor = nil
		return true
	}
	return false
}

// TryGet attempts to retrieve a value from the cache.
// Returns (found, ttl, value). If the key is expired or missing, found=false.
func TryGet[T any](cache *MemCache, key string) (bool, time.Duration, T) {
	cache.lockKey(key)
	defer cache.unlockKey(key)

	var nilVal T
	e, ok := cache.entries.Load(key)
	if !ok {
		return false, 0, nilVal
	}

	entry := e.(*cacheEntry)
	ttl := time.Until(entry.expiredAt)
	if ttl <= 0 {
		cache.entries.Delete(key)
		return false, 0, nilVal
	}

	return true, ttl, entry.value.(T)
}

// GetOrCreate retrieves the value for a key or generates it using the provider.
// If the key is expired or missing, provider() is called and stored for the specified duration.
func GetOrCreate[T any](cache *MemCache, key string, d time.Duration, provider func() (T, error)) (T, error) {
	var nilVal T
	if d < 5*time.Millisecond {
		return nilVal, errors.New("cache duration must be least 5 milliseconds")
	}

	cache.lockKey(key)
	defer cache.unlockKey(key)

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
		}
		return value, err
	}

	if !has {
		entry = &cacheEntry{}
		cache.entries.Store(key, entry)
	}

	entry.expiredAt = time.Now().Add(d)
	entry.value = value

	cache.checkIfRunJanitor(10 * time.Minute)

	return value, nil
}
