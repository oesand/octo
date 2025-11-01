package backoff

import (
	"math"
	"testing"
	"time"
)

// helper: ensure value in range
func assertInRange(t *testing.T, got, min, max time.Duration) {
	if got < min || got > max {
		t.Fatalf("expected in range [%v, %v], got %v", min, max, got)
	}
}

// ─────────────────────────────────────────────────────────────
// Constant Behaviour Tests
// ─────────────────────────────────────────────────────────────

func TestConstantBehaviour_Range(t *testing.T) {
	b := Constant(100 * time.Millisecond)
	min := 100 * time.Millisecond
	max := 120 * time.Millisecond // +20% jitter

	for i := 0; i < 20; i++ {
		got := b.Calculate(i)
		assertInRange(t, got, min, max)
	}
}

func TestConstantBehaviour_ZeroDuration(t *testing.T) {
	b := Constant(0)
	for i := 0; i < 5; i++ {
		if got := b.Calculate(i); got != 0 {
			t.Fatalf("expected 0, got %v", got)
		}
	}
}

func TestConstantBehaviour_JitterObserved(t *testing.T) {
	b := Constant(100 * time.Millisecond)

	minSeen := time.Hour
	maxSeen := time.Duration(0)

	for i := 0; i < 500; i++ {
		got := b.Calculate(i)
		if got < minSeen {
			minSeen = got
		}
		if got > maxSeen {
			maxSeen = got
		}
	}

	if minSeen == maxSeen {
		t.Fatalf("expected variation, got fixed value %v", minSeen)
	}
	assertInRange(t, minSeen, 100*time.Millisecond, 120*time.Millisecond)
	assertInRange(t, maxSeen, 100*time.Millisecond, 120*time.Millisecond)
}

// ─────────────────────────────────────────────────────────────
// Exponential Behaviour Tests
// ─────────────────────────────────────────────────────────────

func TestExponentialBehaviour_Growth(t *testing.T) {
	b := Exponential(100*time.Millisecond, 2*time.Second)

	for attempt := 0; attempt < 10; attempt++ {
		got := b.Calculate(attempt)

		// Compute expected base (without jitter)
		base := float64(100*time.Millisecond) * math.Pow(2, float64(attempt))
		if base > float64(2*time.Second) {
			base = float64(2 * time.Second)
		}

		// Jitter range: +initial
		min := time.Duration(base)
		max := time.Duration(base + float64(100*time.Millisecond))

		assertInRange(t, got, min, max)
	}
}

func TestExponentialBehaviour_Cap(t *testing.T) {
	b := Exponential(100*time.Millisecond, 500*time.Millisecond)

	for i := 0; i < 10; i++ {
		got := b.Calculate(i)
		if got > 600*time.Millisecond {
			t.Fatalf("expected capped at ~500ms, got %v", got)
		}
	}
}

func TestExponentialBehaviour_JitterRange(t *testing.T) {
	b := Exponential(200*time.Millisecond, 5*time.Second)

	got := b.Calculate(2)                                      // 200 * 2^2 = 800
	assertInRange(t, got, 800*time.Millisecond, 1*time.Second) // +200 jitter
}

// ─────────────────────────────────────────────────────────────
// Linear Behaviour Tests
// ─────────────────────────────────────────────────────────────

func TestLinearBehaviour_Growth(t *testing.T) {
	b := Linear(100*time.Millisecond, 200*time.Millisecond, 1*time.Second)

	for i := 1; i < 5; i++ {
		prev := b.Calculate(i - 1)
		got := b.Calculate(i)
		if got < prev {
			t.Fatalf("expected non-decreasing delay: got %v after %v", got, prev)
		}
	}
}

func TestLinearBehaviour_Cap(t *testing.T) {
	b := Linear(100*time.Millisecond, 300*time.Millisecond, 1*time.Second)

	for i := 0; i < 10; i++ {
		got := b.Calculate(i)
		if got > 1*time.Second+200*time.Millisecond {
			t.Fatalf("expected capped delay ≤ 1.2s, got %v", got)
		}
	}
}

func TestLinearBehaviour_JitterRange(t *testing.T) {
	b := Linear(100*time.Millisecond, 200*time.Millisecond, 1*time.Second)

	// attempt 2: delay = 100 + 200*2 = 500 ± step/2
	got := b.Calculate(2)
	assertInRange(t, got, 500*time.Millisecond, 600*time.Millisecond)
}

// ─────────────────────────────────────────────────────────────
// Benchmarks
// ─────────────────────────────────────────────────────────────

func BenchmarkConstantBehaviour(b *testing.B) {
	beh := Constant(100 * time.Millisecond)
	for i := 0; i < b.N; i++ {
		_ = beh.Calculate(i)
	}
}

func BenchmarkExponentialBehaviour(b *testing.B) {
	beh := Exponential(100*time.Millisecond, 5*time.Second)
	for i := 0; i < b.N; i++ {
		_ = beh.Calculate(i % 10)
	}
}

func BenchmarkLinearBehaviour(b *testing.B) {
	beh := Linear(100*time.Millisecond, 200*time.Millisecond, 5*time.Second)
	for i := 0; i < b.N; i++ {
		_ = beh.Calculate(i % 10)
	}
}
