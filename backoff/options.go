package backoff

// BackOffOption represents a function that modifies retry configuration.
type BackOffOption func(*backOffOptions)

type backOffOptions struct {
	attempts  int
	behaviour Behaviour
}

// WithMaxAttempts sets the maximum number of retry attempts.
func WithMaxAttempts(val int) BackOffOption {
	return func(o *backOffOptions) {
		o.attempts = val
	}
}

// WithDefaultBehaviour sets the fallback Behaviour used for all retries.
func WithDefaultBehaviour(val Behaviour) BackOffOption {
	return func(o *backOffOptions) {
		o.behaviour = val
	}
}
