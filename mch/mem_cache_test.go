package mch

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestGetOrCreate_StoresAndReturnsValue(t *testing.T) {
	cache := &MemCache{}
	calls := 0

	val, err := GetOrCreate(cache, "k", 100*time.Millisecond, func() (int, error) {
		calls++
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
	if calls != 1 {
		t.Fatalf("expected provider to be called once, got %d", calls)
	}
	if len(cache.store) != 1 {
		t.Fatalf("expected 1 item in cache, got %d", len(cache.store))
	}
}

func TestGetOrCreate_UsesCachedValue(t *testing.T) {
	cache := &MemCache{}
	calls := 0
	provider := func() (int, error) {
		calls++
		return 123, nil
	}

	_, _ = GetOrCreate(cache, "a", time.Second, provider)
	val2, err := GetOrCreate(cache, "a", time.Second, provider)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != 123 {
		t.Fatalf("expected cached value 123, got %v", val2)
	}
	if calls != 1 {
		t.Fatalf("provider called %d times, expected once", calls)
	}
}

func TestGetOrCreate_ExpiredValueRecreated(t *testing.T) {
	cache := &MemCache{}
	calls := 0
	provider := func() (string, error) {
		calls++
		return "value", nil
	}

	_, _ = GetOrCreate(cache, "x", 10*time.Millisecond, provider)
	time.Sleep(15 * time.Millisecond)
	val, err := GetOrCreate(cache, "x", 10*time.Millisecond, provider)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Fatalf("expected value, got %v", val)
	}
	if calls != 2 {
		t.Fatalf("expected provider called twice, got %d", calls)
	}
}

func TestGetOrCreate_ProviderErrorNotCached(t *testing.T) {
	cache := &MemCache{}
	calls := 0
	provider := func() (int, error) {
		calls++
		return 0, errors.New("fail")
	}

	_, err := GetOrCreate(cache, "bad", time.Second, provider)
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if len(cache.store) != 0 {
		t.Fatalf("expected no items cached on provider error")
	}
}

func TestGetOrCreate_SmallDurationError(t *testing.T) {
	cache := &MemCache{}
	_, err := GetOrCreate(cache, "x", 1*time.Millisecond, func() (int, error) {
		return 1, nil
	})
	if err == nil {
		t.Fatal("expected duration error, got nil")
	}
}

func TestJanitorPurge_RemovesExpiredItems(t *testing.T) {
	cache := &MemCache{}
	now := time.Now().UnixMilli()
	cache.store = map[string]*cacheItem{
		"a": {expiredAt: now - 10, value: 1},
		"b": {expiredAt: now + 9999, value: 2},
	}

	stopped := cache.janitorPurge()
	if stopped {
		t.Fatal("expected janitor not to stop yet")
	}
	if len(cache.store) != 1 {
		t.Fatalf("expected 1 remaining item, got %d", len(cache.store))
	}
}

func TestJanitorPurge_StopsWhenEmpty(t *testing.T) {
	cache := &MemCache{}
	now := time.Now().UnixMilli()
	cache.store = map[string]*cacheItem{
		"x": {expiredAt: now - 10, value: 1},
	}

	stopped := cache.janitorPurge()
	if !stopped {
		t.Fatal("expected janitor to stop when empty")
	}
	if cache.janitor != nil {
		t.Fatal("expected janitor ticker to be nil after purge")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := &MemCache{}
	wg := sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			_, _ = GetOrCreate(cache, "k", 50*time.Millisecond, func() (int, error) {
				time.Sleep(1 * time.Millisecond)
				return 99, nil
			})
		}()
	}
	wg.Wait()

	if len(cache.store) != 1 {
		t.Fatalf("expected 1 cached item, got %d", len(cache.store))
	}
}

// Test that janitor is created and runs cleanup
func TestCheckIfRunJanitor_StartsJanitorAndCleansUp(t *testing.T) {
	cache := &MemCache{
		store: make(map[string]*cacheItem),
	}

	now := time.Now().UnixMilli()
	cache.store["a"] = &cacheItem{expiredAt: now - 1000, value: 1}
	cache.store["b"] = &cacheItem{expiredAt: now - 2000, value: 2}

	cache.checkIfRunJanitor(10 * time.Millisecond)

	if cache.janitor == nil {
		t.Fatal("expected janitor to be started")
	}

	// Wait for janitor to tick and clean up
	time.Sleep(25 * time.Millisecond)

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if len(cache.store) != 0 {
		t.Errorf("expected store to be cleaned, got %d items", len(cache.store))
	}

	if cache.janitor != nil {
		t.Errorf("expected janitor to stop after cleanup")
	}
}

// Test that second call does not start a new janitor
func TestCheckIfRunJanitor_NoDuplicateJanitor(t *testing.T) {
	cache := &MemCache{}
	cache.checkIfRunJanitor(1 * time.Hour)

	first := cache.janitor
	cache.checkIfRunJanitor(10 * time.Millisecond)

	if cache.janitor != first {
		t.Errorf("expected same janitor instance, got new one")
	}
}

// Test that janitor stops only when all items are deleted
func TestCheckIfRunJanitor_StopsOnlyWhenEmpty(t *testing.T) {
	cache := &MemCache{
		store: make(map[string]*cacheItem),
	}
	cache.store["a"] = &cacheItem{expiredAt: time.Now().Add(5 * time.Second).UnixMilli(), value: 1}
	cache.store["b"] = &cacheItem{expiredAt: time.Now().Add(5 * time.Second).UnixMilli(), value: 2}
	cache.mu = sync.Mutex{}

	cache.checkIfRunJanitor(10 * time.Millisecond)
	if cache.janitor == nil {
		t.Fatal("expected janitor to start")
	}

	// Add one expired item so janitor partially cleans
	time.Sleep(20 * time.Millisecond)
	cache.mu.Lock()
	cache.store["expired"] = &cacheItem{
		expiredAt: time.Now().Add(-time.Second).UnixMilli(),
		value:     "x",
	}
	cache.mu.Unlock()

	time.Sleep(30 * time.Millisecond)
	cache.mu.Lock()
	count := len(cache.store)
	cache.mu.Unlock()

	if count == 0 {
		t.Errorf("expected janitor to remain active while store not empty")
	}
}
