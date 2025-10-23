package mc

import (
	"errors"
	"sync"
	"testing"
	"time"
)

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
