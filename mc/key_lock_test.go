package mc

import (
	"sync"
	"testing"
	"time"
)

// Test basic Lock/Unlock behavior and cleanup.
func TestKeyLock_Basic(t *testing.T) {
	var kl KeyLock[string]
	key := "alpha"

	kl.Lock(key)
	if len(kl.locks) != 1 {
		t.Fatalf("expected 1 lock entry, got %d", len(kl.locks))
	}

	kl.Unlock(key)
	if len(kl.locks) != 0 {
		t.Fatalf("expected no locks after unlock, got %d", len(kl.locks))
	}
}

// Test that concurrent goroutines with the same key block sequentially.
func TestKeyLock_SameKeyBlocks(t *testing.T) {
	var kl KeyLock[string]
	key := "beta"

	var order []string
	var mu sync.Mutex
	start := make(chan struct{})
	done := make(chan struct{}, 2)

	go func() {
		<-start
		kl.Lock(key)
		mu.Lock()
		order = append(order, "first")
		mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		kl.Unlock(key)
		done <- struct{}{}
	}()

	go func() {
		<-start
		time.Sleep(10 * time.Millisecond)
		kl.Lock(key)
		mu.Lock()
		order = append(order, "second")
		mu.Unlock()
		kl.Unlock(key)
		done <- struct{}{}
	}()

	close(start)
	<-done
	<-done

	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Fatalf("expected serialized order [first, second], got %+v", order)
	}
}

// Test that different keys can lock independently.
func TestKeyLock_DifferentKeys(t *testing.T) {
	var kl KeyLock[string]
	key1, key2 := "k1", "k2"

	var wg sync.WaitGroup
	var concurrent int
	var mu sync.Mutex
	start := make(chan struct{})
	maxConcurrent := 0

	work := func(k string) {
		defer wg.Done()
		<-start
		kl.Lock(k)
		mu.Lock()
		concurrent++
		if concurrent > maxConcurrent {
			maxConcurrent = concurrent
		}
		mu.Unlock()

		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		concurrent--
		mu.Unlock()
		kl.Unlock(k)
	}

	wg.Add(2)
	go work(key1)
	go work(key2)
	close(start)
	wg.Wait()

	if maxConcurrent < 2 {
		t.Fatalf("expected different keys to run concurrently, got maxConcurrent=%d", maxConcurrent)
	}
}

// Test panic when unlocking a key that was never locked.
func TestKeyLock_UnlockWithoutLockPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when unlocking uninitialized key")
		}
	}()
	var kl KeyLock[string]
	kl.Unlock("ghost")
}

// Test panic when unlocking key not found in map.
func TestKeyLock_UnlockUnknownKeyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when unlocking unknown key")
		}
	}()
	var kl KeyLock[string]
	kl.Lock("a")
	kl.Unlock("a")

	// key has been removed already
	kl.Unlock("a")
}

// Test that after multiple waiters finish, key is removed from map.
func TestKeyLock_CleanupAfterWaiters(t *testing.T) {
	var kl KeyLock[string]
	key := "gamma"
	var wg sync.WaitGroup

	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			kl.Lock(key)
			time.Sleep(30 * time.Millisecond)
			kl.Unlock(key)
			wg.Done()
		}()
	}

	wg.Wait()
	if len(kl.locks) != 0 {
		t.Fatalf("expected lock cleanup after last waiter, got %d entries", len(kl.locks))
	}
}

// Test that generic type works with int keys.
func TestKeyLock_IntKeys(t *testing.T) {
	var kl KeyLock[int]
	key := 123

	kl.Lock(key)
	if _, ok := kl.locks[key]; !ok {
		t.Fatalf("expected lock entry for int key %d", key)
	}
	kl.Unlock(key)

	if len(kl.locks) != 0 {
		t.Fatalf("expected cleanup for int key, got %d", len(kl.locks))
	}
}

// Stress test: 100 goroutines, 10 different keys, each updating 100 times.
func TestKeyLock_ConcurrentStress(t *testing.T) {
	const (
		goroutines = 100
		keys       = 10
		iterations = 100
	)

	var kl KeyLock[int]
	counters := make(map[int]int)
	var mu sync.Mutex // protect counters map

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				key := i % keys // pick key based on loop index
				kl.Lock(key)

				// critical section — modify per-key value safely
				mu.Lock()
				counters[key]++
				mu.Unlock()

				time.Sleep(time.Microsecond * 50) // simulate small work

				kl.Unlock(key)
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
		// continue
	case <-time.After(5 * time.Second):
		t.Fatal("timeout: possible deadlock")
	}

	// Verify: total count == goroutines * iterations
	expected := goroutines * iterations
	var total int
	for _, v := range counters {
		total += v
	}
	if total != expected {
		t.Fatalf("expected total=%d, got %d", expected, total)
	}

	// Verify: lock map should be cleaned up
	kl.mu.Lock()
	if len(kl.locks) != 0 {
		t.Fatalf("expected all locks removed, but have %d", len(kl.locks))
	}
	kl.mu.Unlock()
}
