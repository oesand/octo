package errx

import (
	"fmt"
	"runtime/debug"
)

// ErrWrap is a wrapper for errors that includes a stack trace
type ErrWrap struct {
	error             // Original error
	StackTrace []byte // Stack trace captured at the point of creation
}

// Unwrap implements the error Wrap interface
func (e *ErrWrap) Unwrap() error {
	return e.error
}

// Errorf creates a formatted error, wraps it with a stack trace, and panics
func Errorf(format string, a ...any) {
	panic(&ErrWrap{
		error:      fmt.Errorf(format, a...),
		StackTrace: debug.Stack(),
	})
}

// Error wraps an existing error with a stack trace and panics
func Error(err error) {
	panic(&ErrWrap{
		error:      err,
		StackTrace: debug.Stack(),
	})
}
