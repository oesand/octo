package backoff

import (
	"context"
	"errors"
	"runtime/debug"
	"time"
)

// DefaultMaxAttempts defines how many times BackOff will try the operation before giving up.
var DefaultMaxAttempts = 5

// DefaultBehaviour defines the default backoff strategy (nil by default).
// It must implement Behaviour, which determines the delay for each retry attempt.
var DefaultBehaviour Behaviour

// BackOff executes the given operation `op` with retry semantics and backoff behaviour.
//
// The function will:
//   - Execute `op` at least once.
//   - Retry up to `attempts` times (default 5) if errors occur.
//   - Wait according to the provided `Behaviour` (or `DefaultBehaviour`).
//   - Stop early if context is cancelled.
//
// The generic type T represents the successful result type of the operation.
func BackOff[T any](ctx context.Context, op func(context.Context) (T, error), options ...BackOffOption) (T, error) {
	opts := backOffOptions{
		attempts:  DefaultMaxAttempts,
		behaviour: DefaultBehaviour,
	}

	for _, opt := range options {
		opt(&opts)
	}

	var res T
	var err error

	var attempt int
	for {
		res, err = op(ctx)
		if err == nil {
			break
		}

		var behaviour Behaviour

		var wb *BackoffWrap
		if errors.As(err, &wb) {
			behaviour = wb.behaviour
		}

		if opts.attempts > 0 && attempt >= opts.attempts {
			break
		}

		if behaviour == nil {
			behaviour = opts.behaviour
		}

		if behaviour == nil {
			break
		}

		waitDuration := behaviour.Calculate(attempt)
		if waitDuration > 0 {
			select {
			case <-time.After(waitDuration):
				break
			case <-ctx.Done():
				return res, context.Cause(ctx)
			}
		} else {
			select {
			case <-ctx.Done():
				return res, context.Cause(ctx)
			default:
			}
		}

		attempt++
	}

	return res, err
}

// Wrap attaches a retry behaviour and stack trace to an error.
// Useful for marking errors as retryable with custom timing.
func Wrap(err error, behaviour Behaviour) error {
	return &BackoffWrap{
		error:      err,
		behaviour:  behaviour,
		StackTrace: debug.Stack(),
	}
}

// BackoffWrap wraps an error with an optional retry behaviour and stack trace.
type BackoffWrap struct {
	error
	StackTrace []byte
	behaviour  Behaviour
}
