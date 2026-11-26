package mc

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestJanitor_StopsWithoutTick(t *testing.T) {
	var cache MemCache
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	cache.janitor = ticker
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		cache.runJanitor(ticker, stop)
		close(done)
	}()
	ticker.Stop()
	close(stop)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("janitor did not exit after stop signal")
	}
}

func TestJanitor_StaleTickerReleasesMutex(t *testing.T) {
	var cache MemCache
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	done := make(chan struct{})
	go func() {
		cache.runJanitor(ticker, nil)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stale janitor did not exit")
	}
	if !cache.janitorMu.TryLock() {
		t.Fatal("janitor exited with janitorMu locked")
	}
	cache.janitorMu.Unlock()
}

func TestJanitor_ManualPurgeSignalsStopAndAllowsRestart(t *testing.T) {
	cache := New(WithJanitorInterval(time.Hour))
	var previous <-chan struct{}
	for i := 0; i < 2; i++ {
		cache.checkIfRunJanitor()
		cache.janitorMu.Lock()
		stop := cache.janitorStop
		cache.janitorMu.Unlock()
		if stop == nil || stop == previous {
			t.Fatal("janitor did not start with a new stop channel")
		}
		JanitorPurge(cache)
		select {
		case <-stop:
		default:
			t.Fatal("manual purge did not signal janitor to stop")
		}
		previous = stop
	}
}

// Test basic GetOrCreate and TryGet logic.
func TestMemCache_Basic(t *testing.T) {
	var mc MemCache
	providerCalls := 0

	v, err := GetOrCreate(&mc, "key", time.Second, func() (string, error) {
		providerCalls++
		return "value", nil
	})
	if err != nil || v != "value" {
		t.Fatalf("unexpected result: %v, %v", v, err)
	}

	// Should use cached value.
	found, ttl, v2 := TryGet[string](&mc, "key")
	if !found || v2 != "value" || ttl <= 0 {
		t.Fatalf("expected found cached value, got %v (%v, ttl=%v)", v2, found, ttl)
	}

	// Provider should not have been called again.
	if providerCalls != 1 {
		t.Fatalf("expected 1 provider call, got %d", providerCalls)
	}
}

// Test expiration of keys.
func TestMemCache_Expiration(t *testing.T) {
	var mc MemCache
	_, err := GetOrCreate(&mc, "expiring", 10*time.Millisecond, func() (string, error) {
		return "soon", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)

	found, _, _ := TryGet[string](&mc, "expiring")
	if found {
		t.Fatal("expected expired value not to be found")
	}
}

// Test provider error removes key.
func TestMemCache_ProviderError(t *testing.T) {
	var mc MemCache
	_, err := GetOrCreate(&mc, "bad", time.Second, func() (string, error) {
		return "", errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected provider error")
	}

	found, _, _ := TryGet[string](&mc, "bad")
	if found {
		t.Fatal("expected failed key not cached")
	}
}

// Test janitor removes expired keys.
func TestMemCache_JanitorPurge(t *testing.T) {
	var mc MemCache

	// Add one expired, one active
	mc.entries.Store("a", &cacheEntry{expiredAt: time.Now().Add(-1 * time.Second), value: "old"})
	mc.entries.Store("b", &cacheEntry{expiredAt: time.Now().Add(time.Hour), value: "live"})

	stopped := mc.janitorPurge()
	if stopped {
		t.Fatal("expected janitor to continue (still live entry)")
	}

	mc.entries.Store("b", &cacheEntry{expiredAt: time.Now().Add(-1 * time.Second), value: "old"})
	stopped = mc.janitorPurge()
	if !stopped {
		t.Fatal("expected janitor to stop (no live entries)")
	}
}

func TestMemCache_JanitorPurgeReloadsEntryAfterLock(t *testing.T) {
	for _, replace := range []bool{false, true} {
		name := "removed"
		if replace {
			name = "replaced"
		}
		t.Run(name, func(t *testing.T) {
			var cache MemCache
			cache.entries.Store("key", &cacheEntry{expiredAt: time.Now().Add(-time.Hour)})
			cache.keyedMu.Lock("key")
			locked := true
			defer func() {
				if locked {
					cache.keyedMu.Unlock("key")
				}
			}()

			done := make(chan bool, 1)
			go func() { done <- cache.janitorPurge() }()

			// Wait until Range has read the old entry and is waiting for its key.
			deadline := time.Now().Add(time.Second)
			for {
				cache.keyedMu.mu.Lock()
				waiting := cache.keyedMu.locks["key"].wait > 1
				cache.keyedMu.mu.Unlock()
				if waiting {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("purge did not wait for the key")
				}
				time.Sleep(time.Millisecond)
			}

			cache.entries.Delete("key")
			if replace {
				cache.entries.Store("key", &cacheEntry{
					expiredAt: time.Now().Add(time.Hour), value: "fresh",
				})
			}
			cache.keyedMu.Unlock("key")
			locked = false

			select {
			case stopped := <-done:
				if stopped == replace {
					t.Errorf("unexpected janitor stop decision: %v", stopped)
				}
			case <-time.After(time.Second):
				t.Fatal("purge did not complete")
			}
			found, _, value := TryGet[string](&cache, "key")
			if found != replace || (replace && value != "fresh") {
				t.Fatalf("unexpected entry after purge: found=%v value=%q", found, value)
			}
		})
	}
}

// Stress test: many goroutines and keys concurrently.
func TestMemCache_ConcurrentStress(t *testing.T) {
	var mc MemCache
	const goroutines = 100
	const keys = 20
	const iters = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				key := string(rune('a' + (i % keys)))
				v, err := GetOrCreate(&mc, key, 100*time.Millisecond, func() (string, error) {
					return key, nil
				})
				if err != nil {
					t.Errorf("provider error: %v", err)
				}
				if v != key {
					t.Errorf("expected %q got %q", key, v)
				}
			}
		}(g)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout (possible deadlock)")
	}

	// Check cache has ≤ keys entries
	count := 0
	mc.entries.Range(func(_, _ any) bool { count++; return true })
	if count > keys {
		t.Fatalf("expected at most %d keys, got %d", keys, count)
	}
}

// Test invalid duration.
func TestMemCache_InvalidDuration(t *testing.T) {
	var mc MemCache
	_, err := GetOrCreate(&mc, "x", time.Millisecond, func() (string, error) {
		return "v", nil
	})
	if err == nil {
		t.Fatal("expected error for too short duration")
	}
}

func TestMemCache_Extend(t *testing.T) {
	var mc MemCache

	oldDuration := time.Now().Add(time.Second)
	newDuration := time.Now().Add(time.Hour)
	mc.entries.Store("b", &cacheEntry{expiredAt: oldDuration})

	ExtendUntil(&mc, "b", newDuration)

	ce, _ := mc.entries.Load("b")
	entry := ce.(*cacheEntry)
	if entry.expiredAt != newDuration {
		t.Errorf("expected %v, got %v", newDuration, entry.expiredAt)
	}
}

func TestMemCache_Forgot(t *testing.T) {
	var mc MemCache
	_, err := GetOrCreate(&mc, "x", time.Hour, func() (string, error) {
		return "value", nil
	})
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if !Forgot(&mc, "x") {
		t.Fatal("expected key should be removed")
	}

	if has, _, _ := TryGet[string](&mc, "x"); has {
		t.Fatal("key still in cache")
	}
}
