package errx

import (
	"errors"
)

// Catcher is a callback function to handle an ErrWrap in a Try block
// Returns true if the error is handled, false otherwise
type Catcher func(wrap *ErrWrap) bool

// Try executes an operation and recovers from ErrWrap panics
// Optional catchers can handle specific error types
func Try(op func(), catchers ...Catcher) (wrap *ErrWrap) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && err != nil && errors.As(err, &wrap) {
				for _, cb := range catchers {
					if cb(wrap) {
						break
					}
				}
				return
			}
			panic(r)
		}
	}()

	op()
	return
}

// Catch creates a Catcher for a specific error type E
func Catch[E error](matcher func(E)) Catcher {
	return func(wrap *ErrWrap) bool {
		var err E
		if errors.As(wrap, &err) {
			matcher(err)
			return true
		}
		return false
	}
}
