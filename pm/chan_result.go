package pm

import (
	"context"
	"iter"
)

type resultErr[T any] struct {
	res T
	err error
}

func (b *resultErr[T]) Result() (T, error) {
	return b.res, b.err
}

// ChanRes is a typed channel whose elements implement Result() (T, error).
// The channel transports results of asynchronous operations.
type ChanRes[T any] chan interface {
	Result() (T, error)
}

// Size returns the channel capacity.
func (ch ChanRes[T]) Size() int {
	return cap(ch)
}

// UnBuffered reports whether the channel has unlimited capacity.
func (ch ChanRes[T]) UnBuffered() bool {
	return ch.Size() == 0
}

// I return an iterator over the channel contents.
// - For buffered channels: it reads exactly Size() items.
// - For unbuffered channels: it reads until channel close.
// - Cancels early if ctx is cancelled.
// - Stops iteration if yield returns false.
func (ch ChanRes[T]) I(ctx context.Context) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		size := ch.Size()
		for i := 0; size == 0 || i < size; i++ {
			select {
			case <-ctx.Done():
				var zero T
				yield(zero, context.Cause(ctx))
				return
			case res, ok := <-ch:
				if !ok || !yield(res.Result()) {
					return
				}
			}
		}
	}
}

// Wait collects all values produced by I(ctx) into a slice.
// Stops early if an error occurs.
func (ch ChanRes[T]) Wait(ctx context.Context) ([]T, error) {
	arr := make([]T, 0, ch.Size())
	for val, err := range ch.I(ctx) {
		if err != nil {
			return arr, err
		}
		arr = append(arr, val)
	}
	return arr, nil
}

// Put sends a value and an error into the channel.
func (ch ChanRes[T]) Put(res T, err error) {
	ch <- &resultErr[T]{
		res: res,
		err: err,
	}
}

// Go runs an operation in a goroutine and forwards its result into the channel.
func (ch ChanRes[T]) Go(ctx context.Context, op func(context.Context) (T, error)) {
	go ch.Put(op(ctx))
}

// Close closes the underlying channel.
func (ch ChanRes[T]) Close() {
	close(ch)
}
