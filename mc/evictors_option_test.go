package mc

import (
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
