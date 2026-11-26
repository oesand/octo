package mc

import (
	"slices"
	"testing"
	"time"
)

func expectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic but none occurred")
		}
	}()
	fn()
}

func TestWithLruEviction_SetsEvictor(t *testing.T) {
	c := New(WithLruEviction(3))

	if c.usageEvictor == nil {
		t.Fatalf("usageEvictor must be set")
	}

	lru, ok := c.usageEvictor.(*lruEvictor)
	if !ok {
		t.Fatalf("expected *lruEvictor, got %T", c.usageEvictor)
	}

	if lru.capacity != 3 {
		t.Fatalf("unexpected capacity: got %d want %d", lru.capacity, 3)
	}
}

func TestWithLruEviction_PanicsOnInvalidCapacity(t *testing.T) {
	expectPanic(t, func() {
		New(WithLruEviction(0))
	})
}

func TestWithLruEviction_PanicsIfAlreadySet(t *testing.T) {
	expectPanic(t, func() {
		New(WithLruEviction(2), WithLruEviction(4))
	})
}

func TestWithLfuEviction_SetsEvictor(t *testing.T) {
	c := New(WithLfuEviction(5))

	if c.usageEvictor == nil {
		t.Fatalf("usageEvictor must be set")
	}

	lfu, ok := c.usageEvictor.(*lfuEvictor)
	if !ok {
		t.Fatalf("expected *lfuEvictor, got %T", c.usageEvictor)
	}

	if lfu.capacity != 5 {
		t.Fatalf("unexpected capacity: got %d want %d", lfu.capacity, 5)
	}
}

func TestWithLfuEviction_PanicsOnInvalidCapacity(t *testing.T) {
	expectPanic(t, func() {
		New(WithLfuEviction(0))
	})
}

func TestWithLfuEviction_PanicsIfAlreadySet(t *testing.T) {
	expectPanic(t, func() {
		New(WithLfuEviction(2), WithLfuEviction(3))
	})
}

func TestMemCache_LruEvictionIntegration(t *testing.T) {
	c := New(WithLruEviction(2))

	// Create three entries; none expire immediately
	if _, err := GetOrCreate[string](c, "a", time.Hour, func() (string, error) { return "va", nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := GetOrCreate[string](c, "b", time.Hour, func() (string, error) { return "vb", nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := GetOrCreate[string](c, "c", time.Hour, func() (string, error) { return "vc", nil }); err != nil {
		t.Fatal(err)
	}

	// At this point LRU should consider "a" the worst (oldest)
	JanitorPurge(c)

	if ok, _, _ := TryGet[string](c, "a"); ok {
		t.Fatalf("expected key a to be evicted by LRU")
	}
	if ok, _, v := TryGet[string](c, "b"); !ok || v != "vb" {
		t.Fatalf("expected key b to remain, got %v %v", ok, v)
	}
	if ok, _, v := TryGet[string](c, "c"); !ok || v != "vc" {
		t.Fatalf("expected key c to remain, got %v %v", ok, v)
	}
}

func TestMemCache_LfuEvictionIntegration(t *testing.T) {
	c := New(WithLfuEviction(2))

	// Create three entries
	if _, err := GetOrCreate[string](c, "a", time.Hour, func() (string, error) { return "va", nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := GetOrCreate[string](c, "b", time.Hour, func() (string, error) { return "vb", nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := GetOrCreate[string](c, "c", time.Hour, func() (string, error) { return "vc", nil }); err != nil {
		t.Fatal(err)
	}

	// Increase usage of "b" so it becomes more frequently used
	for i := 0; i < 3; i++ {
		if ok, _, _ := TryGet[string](c, "b"); !ok {
			t.Fatalf("expected key b to exist during bump")
		}
	}

	// Now LFU should evict the least frequently used key ("a")
	JanitorPurge(c)

	if ok, _, _ := TryGet[string](c, "a"); ok {
		t.Fatalf("expected key a to be evicted by LFU")
	}
	if ok, _, v := TryGet[string](c, "b"); !ok || v != "vb" {
		t.Fatalf("expected key b to remain, got %v %v", ok, v)
	}
	if ok, _, v := TryGet[string](c, "c"); !ok || v != "vc" {
		t.Fatalf("expected key c to remain, got %v %v", ok, v)
	}
}

func TestIterWorst_AllowsUpdatesDuringYield(t *testing.T) {
	for _, tc := range []struct {
		name string
		evc  usageEvictor
	}{
		{"LRU", newLruEvictor(2)},
		{"LFU", newLfuEvictor(2)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.evc.Used("a")
			tc.evc.Used("b")
			var got []string
			var updates []chan struct{}
			for key := range tc.evc.IterWorst() {
				got = append(got, key)
				done := make(chan struct{})
				updates = append(updates, done)
				go func() {
					tc.evc.Used(key)
					tc.evc.Remove(key)
					close(done)
				}()
				select {
				case <-done:
				case <-time.After(time.Second):
					t.Error("evictor update blocked during yield")
				}
				if t.Failed() {
					break
				}
			}
			for _, done := range updates {
				<-done
			}
			if !slices.Equal(got, []string{"a", "b"}) {
				t.Fatalf("unexpected snapshot order: %v", got)
			}
		})
	}
}

type blockingRemoveEvictor struct {
	usageEvictor
	removed chan struct{}
	resume  chan struct{}
}

func (evc *blockingRemoveEvictor) Remove(key string) {
	evc.usageEvictor.Remove(key)
	close(evc.removed)
	<-evc.resume
}

func TestJanitorPurge_KeepsRunningWhenEntryAddedDuringCleanup(t *testing.T) {
	cache := New(WithLruEviction(2))
	// Represent a running janitor without scheduling a background purge.
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	cache.janitor = ticker
	cache.entries.Store("old", &cacheEntry{expiredAt: time.Now().Add(-time.Hour)})
	cache.usageEvictor.Used("old")
	evc := &blockingRemoveEvictor{
		usageEvictor: cache.usageEvictor,
		removed:      make(chan struct{}),
		resume:       make(chan struct{}),
	}
	cache.usageEvictor = evc
	defer close(evc.resume)

	done := make(chan bool, 1)
	go func() { done <- cache.janitorPurge() }()
	select {
	case <-evc.removed:
	case <-time.After(time.Second):
		t.Fatal("purge did not remove the expired entry")
	}

	_, err := GetOrCreate(cache, "new", time.Hour, func() (string, error) {
		return "fresh", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	evc.resume <- struct{}{}

	select {
	case stopped := <-done:
		if stopped {
			t.Error("janitor stopped despite a newly added entry")
		}
	case <-time.After(time.Second):
		t.Fatal("purge did not complete")
	}
	cache.janitorMu.Lock()
	running := cache.janitor != nil
	cache.janitorMu.Unlock()
	if !running {
		t.Error("new entry was left without a janitor")
	}
}
