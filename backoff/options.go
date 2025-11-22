package backoff

import (
	"github.com/oesand/octo/pm"
	"reflect"
)

// Option represents a function that modifies retry configuration.
type Option func(*backOffOptions)

type backOffOptions struct {
	attempts   int
	behaviour  Behaviour
	catchErrX  bool
	abortTypes pm.Set[reflect.Type]
}

// WithMaxAttempts sets the maximum number of retry attempts.
func WithMaxAttempts(val int) Option {
	return func(o *backOffOptions) {
		o.attempts = val
	}
}

// WithDefaultBehaviour sets the fallback Behaviour used for all retries.
func WithDefaultBehaviour(val Behaviour) Option {
	return func(o *backOffOptions) {
		o.behaviour = val
	}
}

// WithErrX enables errx catch
func WithErrX() Option {
	return func(o *backOffOptions) {
		o.catchErrX = true
	}
}
