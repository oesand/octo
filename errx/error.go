package errx

import (
	"fmt"
	"runtime/debug"
)

// ErrWrap is a wrapper for errors that includes a stack trace
type ErrWrap struct {
	Err        error  // Original error
	StackTrace []byte // Stack trace captured at the point of creation
}

// Error implements the error interface for ErrWrap
func (e *ErrWrap) Error() string {
	return e.Err.Error()
}

// Errorf creates a formatted error, wraps it with a stack trace, and panics
func Errorf(format string, a ...any) {
	panic(&ErrWrap{
		Err:        fmt.Errorf(format, a...),
		StackTrace: debug.Stack(),
	})
}

// Error wraps an existing error with a stack trace and panics
func Error(err error) {
	panic(&ErrWrap{
		Err:        err,
		StackTrace: debug.Stack(),
	})
}
