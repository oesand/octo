package mc

import (
	"slices"
	"sync"
	"testing"
)

func TestUsed_Order(t *testing.T) {
	lru := newLruEvictor(10)

	lru.Used("a")
	lru.Used("b")
	lru.Used("c")

	// "a" is the least recently used,
	// "c" is the most recently used
	got := slices.Collect(lru.IterWorst())
	want := []string{"a", "b", "c"}

	if !slices.Equal(got, want) {
		t.Fatalf("order mismatch: got %v want %v", got, want)
	}
}

func TestUsed_MoveToFront(t *testing.T) {
	lru := newLruEvictor(10)

	lru.Used("a")
	lru.Used("b")
	lru.Used("c")

	// Mark "a" as recently used again
	lru.Used("a")

	// "b" becomes the oldest, "a" the newest
	got := slices.Collect(lru.IterWorst())
	want := []string{"b", "c", "a"}

	if !slices.Equal(got, want) {
		t.Fatalf("order mismatch: got %v want %v", got, want)
	}
}

func TestGetExcess(t *testing.T) {
	lru := newLruEvictor(2)

	if ex := lru.GetExcess(); ex != 0 {
		t.Fatalf("unexpected excess: %d", ex)
	}

	lru.Used("a")
	lru.Used("b")

	if ex := lru.GetExcess(); ex != 0 {
		t.Fatalf("unexpected excess: %d", ex)
	}

	lru.Used("c")

	if ex := lru.GetExcess(); ex != 1 {
		t.Fatalf("expected excess 1, got %d", ex)
	}
}

func TestIterWorst_StopEarly(t *testing.T) {
	lru := newLruEvictor(10)

	lru.Used("a")
	lru.Used("b")
	lru.Used("c")

	var got []string

	// Stop iteration after the first element
	lru.IterWorst()(func(v string) bool {
		got = append(got, v)
		return false
	})

	if !slices.Equal(got, []string{"a"}) {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestRemove(t *testing.T) {
	lru := newLruEvictor(10)

	lru.Used("a")
	lru.Used("b")
	lru.Used("c")

	lru.Remove("b")

	got := slices.Collect(lru.IterWorst())
	want := []string{"a", "c"}

	if !slices.Equal(got, want) {
		t.Fatalf("order mismatch: got %v want %v", got, want)
	}
}

func TestRemove_NonExisting(t *testing.T) {
	lru := newLruEvictor(10)

	lru.Used("a")

	// Removing a non-existing key must be a no-op
	lru.Remove("does-not-exist")

	got := slices.Collect(lru.IterWorst())
	want := []string{"a"}

	if !slices.Equal(got, want) {
		t.Fatalf("unexpected state after remove: %v", got)
	}
}

func TestConcurrentSafety(t *testing.T) {
	lru := newLruEvictor(100)
	wg := sync.WaitGroup{}

	// Concurrent access should not cause data races
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lru.Used(string(rune('a' + i)))
			lru.GetExcess()
		}(i)
	}

	wg.Wait()
}
