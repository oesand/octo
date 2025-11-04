package pm

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestResultBag_Success(t *testing.T) {
	var bag ResultBag[int]
	bag.Add(2)

	go func() {
		time.Sleep(10 * time.Millisecond)
		bag.Put(nil, 1)
		bag.Put(nil, 2)
	}()

	items, err := bag.Wait(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 || !(items[0] == 1 && items[1] == 2) {
		t.Fatalf("unexpected items: %v", items)
	}
}

func TestResultBag_ErrorStopsAll(t *testing.T) {
	var bag ResultBag[int]
	bag.Add(2)

	go func() {
		bag.Put(errors.New("fail"))
		bag.Put(nil, 2) // should be ignored
	}()

	items, err := bag.Wait(context.Background())
	if err == nil || err.Error() != "fail" {
		t.Fatalf("expected fail error, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no items after error, got %v", items)
	}
}

func TestResultBag_ContextCancel(t *testing.T) {
	var bag ResultBag[int]
	bag.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	items, err := bag.Wait(ctx)
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context timeout, got %v", err)
	}
	if time.Since(start) < 15*time.Millisecond {
		t.Fatalf("wait returned too early")
	}
	if len(items) != 0 {
		t.Fatalf("expected no items on cancel")
	}
}

func TestResultBag_ConcurrentGo(t *testing.T) {
	var bag ResultBag[int]

	const n = 10
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		i := i
		bag.Go(func() ([]int, error) {
			defer wg.Done()
			return []int{i}, nil
		})
	}

	wg.Wait()
	items, err := bag.Wait(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != n {
		t.Fatalf("expected %d items, got %d", n, len(items))
	}
}

func TestResultBag_Reset(t *testing.T) {
	var bag ResultBag[int]
	bag.Add(1)
	bag.Put(nil, 1)

	_, err := bag.Wait(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bag.Reset()
	bag.Add(1)
	go bag.Put(nil, 42)

	items, err := bag.Wait(context.Background())
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if len(items) != 1 || items[0] != 42 {
		t.Fatalf("unexpected items after reset: %v", items)
	}
}
