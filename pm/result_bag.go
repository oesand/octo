package pm

import (
	"context"
	"sync"
)

// ResultBag aggregates results from multiple concurrent operations
// and provides synchronization via Wait().
//
// Typical usage:
//
//	var bag pm.ResultBag[int]
//	for _, job := range jobs {
//	    bag.Go(func() ([]int, error) {
//	        return []int{job.ID}, nil
//	    })
//	}
//
//	results, err := bag.Wait(context.Background())
//
// Features:
//   - Thread-safe Add / Put / Go coordination
//   - Waits for all added operations to complete
//   - Stops collecting results after the first error
//   - Reset() allows reusing the same bag safely
type ResultBag[T comparable] struct {
	mu      sync.Mutex
	waiters int

	items []T
	err   error
	done  chan struct{}
}

// Add increments the number of expected results.
// Must be called before launching asynchronous operations.
//
// Panics if delta <= 0.
func (b *ResultBag[T]) Add(delta int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if delta <= 0 {
		panic("pm: negative or zero delta")
	}

	b.add(delta)
}

func (b *ResultBag[T]) add(delta int) {
	if b.err != nil {
		return
	}

	if b.done == nil || b.waiters == 0 {
		b.done = make(chan struct{})
	}

	b.waiters += delta
}

// Put records a result or an error from an operation.
//
// If err is non-nil, all pending operations are canceled (waiters set to 0),
// items are cleared, and Wait() will return that error.
func (b *ResultBag[T]) Put(err error, values ...T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.waiters == 0 {
		return
	}

	if err != nil {
		b.waiters = 0
		b.items = nil
		b.err = err
		close(b.done)
	} else {
		b.waiters--
		b.items = append(b.items, values...)

		if b.waiters == 0 {
			close(b.done)
		}
	}
}

// Go starts a goroutine running op() and automatically tracks
// its result using Put().
//
// If the bag already has an error, Go does nothing.
func (b *ResultBag[T]) Go(op func() ([]T, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.err != nil {
		return
	}

	b.add(1)
	go func() {
		res, err := op()
		b.Put(err, res...)
	}()
}

// Wait blocks until all results are collected, an error occurs, or
// the given context is canceled.
//
// Returns all accumulated items and the first encountered error (if any).
func (b *ResultBag[T]) Wait(ctx context.Context) ([]T, error) {
	b.mu.Lock()
	done := b.done
	b.mu.Unlock()

	if done != nil {
		select {
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		case <-done:
			break
		}
	}

	b.mu.Lock()
	items, err := b.items, b.err
	b.mu.Unlock()
	return items, err
}

// Reset clears the bag for reuse.
// If the bag was mid-wait, it will close the done channel and cancel remaining waiters.
func (b *ResultBag[T]) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.done != nil && b.waiters > 0 {
		b.waiters = 0
		close(b.done)
		b.done = nil
	}

	b.items = nil
	b.err = nil
}
