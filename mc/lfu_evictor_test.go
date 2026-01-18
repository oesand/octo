package mc

import (
	"slices"
	"sync"
	"testing"
)

func TestLFU_BasicOrder(t *testing.T) {
	lfu := newLfuEvictor(10)

	lfu.Used("a") // a:1
	lfu.Used("b") // b:1
	lfu.Used("a") // a:2
	lfu.Used("c") // c:1

	// freq 1: b, c (in insertion order)
	// freq 2: a
	got := slices.Collect(lfu.IterWorst())
	want := []string{"b", "c", "a"}

	if !slices.Equal(got, want) {
		t.Fatalf("order mismatch: got %v want %v", got, want)
	}
}

func TestLFU_LRUWithinSameFrequency(t *testing.T) {
	lfu := newLfuEvictor(10)

	lfu.Used("a") // a:1
	lfu.Used("b") // b:1
	lfu.Used("c") // c:1

	// All have the same frequency (1),
	// eviction order must be LRU
	got := slices.Collect(lfu.IterWorst())
	want := []string{"a", "b", "c"}

	if !slices.Equal(got, want) {
		t.Fatalf("order mismatch: got %v want %v", got, want)
	}
}

func TestLFU_FrequencyBumpChangesOrder(t *testing.T) {
	lfu := newLfuEvictor(10)

	lfu.Used("a") // a:1
	lfu.Used("b") // b:1
	lfu.Used("c") // c:1

	lfu.Used("a") // a:2
	lfu.Used("b") // b:2

	// freq 1: c
	// freq 2: a, b (LRU order)
	got := slices.Collect(lfu.IterWorst())
	want := []string{"c", "a", "b"}

	if !slices.Equal(got, want) {
		t.Fatalf("order mismatch: got %v want %v", got, want)
	}
}

func TestLFU_GetExcess(t *testing.T) {
	lfu := newLfuEvictor(2)

	if ex := lfu.GetExcess(); ex != 0 {
		t.Fatalf("unexpected excess: %d", ex)
	}

	lfu.Used("a")
	lfu.Used("b")

	if ex := lfu.GetExcess(); ex != 0 {
		t.Fatalf("unexpected excess: %d", ex)
	}

	lfu.Used("c")

	if ex := lfu.GetExcess(); ex != 1 {
		t.Fatalf("expected excess 1, got %d", ex)
	}
}

func TestLFU_Remove(t *testing.T) {
	lfu := newLfuEvictor(10)

	lfu.Used("a") // a:1
	lfu.Used("b") // b:1
	lfu.Used("a") // a:2

	lfu.Remove("a")

	got := slices.Collect(lfu.IterWorst())
	want := []string{"b"}

	if !slices.Equal(got, want) {
		t.Fatalf("unexpected state: got %v want %v", got, want)
	}
}

func TestLFU_RemoveNonExisting(t *testing.T) {
	lfu := newLfuEvictor(10)

	lfu.Used("a")
	lfu.Remove("does-not-exist")

	got := slices.Collect(lfu.IterWorst())
	want := []string{"a"}

	if !slices.Equal(got, want) {
		t.Fatalf("unexpected state: got %v want %v", got, want)
	}
}

func TestLFU_IterWorst_StopEarly(t *testing.T) {
	lfu := newLfuEvictor(10)

	lfu.Used("a")
	lfu.Used("b")
	lfu.Used("a") // a:2

	var got []string

	lfu.IterWorst()(func(v string) bool {
		got = append(got, v)
		return false
	})

	// Only the worst element should be returned
	if !slices.Equal(got, []string{"b"}) {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestLFU_ConcurrentSafety(t *testing.T) {
	lfu := newLfuEvictor(100)

	wg := sync.WaitGroup{}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('a' + i))
			lfu.Used(key)
			lfu.Used(key)
			lfu.GetExcess()
		}(i)
	}

	wg.Wait()
}
