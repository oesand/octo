package pm

import (
	"context"
	"errors"
	"testing"
)

// Basic buffered channel Put + Wait.
func TestWaitBufferedSuccess(t *testing.T) {
	ch := make(ChanRes[int], 3)

	ch.Put(1, nil)
	ch.Put(2, nil)
	ch.Put(3, nil)

	ctx := context.Background()
	values, err := ch.Wait(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []int{1, 2, 3}
	for i := range expected {
		if values[i] != expected[i] {
			t.Fatalf("expected %v got %v", expected, values)
		}
	}
}

// Error propagation through Wait().
func TestWaitPropagatesError(t *testing.T) {
	ch := make(ChanRes[int], 2)

	ch.Put(10, nil)
	ch.Put(0, errors.New("boom"))

	ctx := context.Background()
	values, err := ch.Wait(ctx)

	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected error 'boom', got: %v", err)
	}

	// first item should be collected before the error
	if len(values) != 1 || values[0] != 10 {
		t.Fatalf("values not collected correctly: %v", values)
	}
}

// Iterator on buffered channel should read exactly cap(ch) items.
func TestIteratorBuffered(t *testing.T) {
	ch := make(ChanRes[string], 2)

	ch.Put("A", nil)
	ch.Put("B", nil)

	ctx := context.Background()

	collected := []string{}

	for val, err := range ch.I(ctx) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		collected = append(collected, val)
	}

	if len(collected) != 2 || collected[0] != "A" || collected[1] != "B" {
		t.Fatalf("iterator collected wrong values: %v", collected)
	}
}

// UnBuffered channel: iterator reads until close.
func TestIteratorUnbuffered(t *testing.T) {
	ch := make(ChanRes[int])

	go func() {
		ch.Put(5, nil)
		ch.Put(7, nil)
		ch.Close()
	}()

	ctx := context.Background()

	var res []int
	for val, err := range ch.I(ctx) {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		res = append(res, val)
	}

	if len(res) != 2 || res[0] != 5 || res[1] != 7 {
		t.Fatalf("unexpected values: %v", res)
	}
}

// Test cancellation while iterating.
func TestWaitCanceled(t *testing.T) {
	ch := make(ChanRes[int], 3)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	values, err := ch.Wait(ctx)

	if err == nil {
		t.Fatalf("expected cancellation error")
	}
	if len(values) != 0 {
		t.Fatalf("expected no values, got %v", values)
	}
}

// Test Go() helper safely.
func TestGoHelper(t *testing.T) {
	ch := make(ChanRes[int], 1) // buffered channel

	ch.Go(context.Background(), func(ctx context.Context) (int, error) {
		return 99, nil
	})

	ctx := context.Background()
	values, err := ch.Wait(ctx) // Wait will block until goroutine sends value

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 1 || values[0] != 99 {
		t.Fatalf("expected [99], got %v", values)
	}
}

// Confirm unbuffered detection.
func TestUnbuffered(t *testing.T) {
	ch := make(ChanRes[int])

	if !ch.UnBuffered() {
		t.Fatalf("expected channel to be unbuffered")
	}

	ch2 := make(ChanRes[int], 5)
	if ch2.UnBuffered() {
		t.Fatalf("expected buffered channel")
	}
}

// Closing early should stop iterator cleanly.
func TestCloseEarly(t *testing.T) {
	ch := make(ChanRes[int], 5)

	ch.Put(1, nil)
	ch.Put(2, nil)
	ch.Close() // safe: stop iterator

	ctx := context.Background()

	values, err := ch.Wait(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %v", values)
	}
}

// Iterator should stop if yield returns false.
func TestIteratorYieldStop(t *testing.T) {
	ch := make(ChanRes[int], 3)
	ch.Put(10, nil)
	ch.Put(20, nil)
	ch.Put(30, nil)

	ctx := context.Background()

	collected := []int{}
	for val, err := range ch.I(ctx) {
		if err != nil {
			t.Fatal(err)
		}
		collected = append(collected, val)
		break // stop iteration early after first element
	}

	if len(collected) != 1 || collected[0] != 10 {
		t.Fatalf("unexpected collected: %v", collected)
	}
}

// Test multiple Go() goroutines safely.
func TestMultipleGoRoutines(t *testing.T) {
	ch := make(ChanRes[int], 3)

	ch.Go(context.Background(), func(ctx context.Context) (int, error) { return 1, nil })
	ch.Go(context.Background(), func(ctx context.Context) (int, error) { return 2, nil })
	ch.Go(context.Background(), func(ctx context.Context) (int, error) { return 3, nil })

	ctx := context.Background()
	values, err := ch.Wait(ctx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %v", values)
	}
}
