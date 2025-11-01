package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Behaviour defines a strategy for calculating delay duration
// depending on the attempt number. It’s used for backoff or throttling.
type Behaviour interface {
	// Calculate returns the delay for the given attempt number (0-based).
	Calculate(attempt int) time.Duration
}

// Constant returns a Behaviour that always produces the same base delay
// duration for each attempt, with a small positive random jitter added
// to prevent synchronized retries (the "thundering herd" effect).
//
// The actual delay will be between 100% and 120% of the provided duration.
// For example, Constant(100*time.Millisecond) may yield values
// in the range [100ms, 120ms]
func Constant(duration time.Duration) Behaviour {
	return &constantBehaviour{
		Duration: duration,
	}
}

type constantBehaviour struct {
	Duration time.Duration
}

func (c *constantBehaviour) Calculate(_ int) time.Duration {
	// Add positive jitter (0–20%) to avoid synchronized retries
	const jitterFraction = 0.2
	if c.Duration <= 0 {
		return 0
	}

	// Random multiplier in range [1.0, 1.2]
	factor := 1 + rand.Float64()*jitterFraction
	return time.Duration(float64(c.Duration) * factor)
}

// Exponential returns a Behaviour that increases the delay exponentially
// with each attempt, starting from the given initial duration and capped
// at the specified maximum.
//
// The formula roughly follows: delay = min(initial * 2^attempt, max),
// plus a small random jitter (up to +initial) to reduce synchronization
// spikes between concurrent retries.
//
// Example:
//
//	b := Exponential(100*time.Millisecond, 5*time.Second)
//	b.Calculate(0) ≈ 100–200ms
//	b.Calculate(1) ≈ 200–300ms
//	b.Calculate(2) ≈ 400–500ms
//	b.Calculate(5) ≈ ~5s (capped)
func Exponential(initial, max time.Duration) Behaviour {
	return &exponentialBehaviour{
		Initial: initial,
		Max:     max,
	}
}

type exponentialBehaviour struct {
	Initial, Max time.Duration
}

func (b *exponentialBehaviour) Calculate(attempt int) time.Duration {
	backoff := min(float64(b.Initial)*math.Pow(2, float64(attempt)), float64(b.Max))

	// Small jitter to avoid sync spikes
	jitter := rand.Float64() * float64(b.Initial)
	return time.Duration(backoff + jitter)
}

// Linear returns a Behaviour that increases the delay linearly with each
// attempt, starting from the specified initial duration and increasing
// by a fixed step amount per attempt, up to the given maximum.
//
// The delay is calculated as:
//
//	delay = min(initial + step * attempt, max)
//
// A small random jitter (up to +step/2) is added to avoid synchronization
// spikes between concurrent retries.
//
// Example:
//
//	b := Linear(100*time.Millisecond, 200*time.Millisecond, 2*time.Second)
//	b.Calculate(0) ≈ 100–200ms
//	b.Calculate(1) ≈ 300–400ms
//	b.Calculate(2) ≈ 500–600ms
//	b.Calculate(5) ≈ ~2s (capped)
func Linear(initial, step, max time.Duration) Behaviour {
	return &linearBehaviour{
		Initial: initial,
		Step:    step,
		Max:     max,
	}
}

type linearBehaviour struct {
	Initial, Step, Max time.Duration
}

func (b *linearBehaviour) Calculate(attempt int) time.Duration {
	delay := min(b.Initial+b.Step*time.Duration(attempt), b.Max)

	// optional jitter
	jitter := time.Duration(rand.Int63n(int64(b.Step / 2)))
	return delay + jitter
}
