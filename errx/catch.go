package errx

import (
	"context"
	"errors"
)

// Catcher is a callback function to handle an ErrWrap in a Try block
// Returns true if the error is handled, false otherwise
type Catcher func(ctx context.Context, wrap *ErrWrap) bool

// Try executes an operation and recovers from ErrWrap panics
// Optional catchers can handle specific error types
func Try(ctx context.Context, op func(context.Context), catchers ...Catcher) (wrap *ErrWrap) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && err != nil && errors.As(err, &wrap) {
				for _, cb := range catchers {
					if cb(ctx, wrap) {
						break
					}
				}
				return
			}
			panic(r)
		}
	}()

	op(ctx)
	return
}

// Catch creates a Catcher for a specific error type E
func Catch[E error](matcher func(context.Context, E)) Catcher {
	return func(ctx context.Context, wrap *ErrWrap) bool {
		var err E
		if errors.As(wrap, &err) {
			matcher(ctx, err)
			return true
		}
		return false
	}
}
