package backoff_test

import (
	"context"
	"errors"
	"github.com/oesand/octo/backoff"
	"github.com/oesand/octo/errx"
	"testing"
	"time"
)

// mockBehaviour implements Behaviour for predictable delay testing.
type mockBehaviour struct {
	delays []time.Duration
	called int
}

func (m *mockBehaviour) Calculate(attempt int) time.Duration {
	m.called++
	if attempt < len(m.delays) {
		return m.delays[attempt]
	}
	return m.delays[len(m.delays)-1]
}

func TestBackOff_SuccessFirstTry(t *testing.T) {
	ctx := context.Background()
	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBackOff_RetrySuccess(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	b := &mockBehaviour{delays: []time.Duration{time.Millisecond}}

	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		bc := backoff.GetSettings(ctx)
		if bc == nil {
			t.Error("not found backoff context")
		}

		if bc.MaxAttempts() != 5 {
			t.Errorf("expected MaxAttempts=5, got %d", bc.MaxAttempts())
		}

		if bc.Attempt() != attempts {
			t.Errorf("expected Attempt=%d, got %d", attempts, bc.Attempt())
		}

		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	}, backoff.WithMaxAttempts(5), backoff.WithDefaultBehaviour(b))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.called < 2 {
		t.Fatalf("expected behaviour called at least twice, got %d", b.called)
	}
}

func TestBackOff_ExceedAttempts(t *testing.T) {
	ctx := context.Background()
	b := &mockBehaviour{delays: []time.Duration{time.Millisecond}}

	attempts := 0
	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		attempts++
		return errors.New("fail")
	}, backoff.WithMaxAttempts(2), backoff.WithDefaultBehaviour(b))

	if err == nil {
		t.Fatal("expected error after exceeding attempts")
	}
	if attempts != 3 { // initial + 2 retries
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestBackOff_WithWrappedBehaviour(t *testing.T) {
	ctx := context.Background()
	wrappedB := &mockBehaviour{delays: []time.Duration{time.Millisecond}}

	attempts := 0
	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			return backoff.Wrap(errors.New("wrapped"), wrappedB)
		}
		return nil
	}, backoff.WithMaxAttempts(3))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wrappedB.called == 0 || attempts != 2 {
		t.Fatalf("expected wrapped behaviour to be used, calls: %d", attempts)
	}
}

func TestBackOff_TestUnWrap(t *testing.T) {
	ctx := context.Background()

	type customError struct {
		error
	}

	attempts := 0
	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		attempts++
		return backoff.Wrap(&customError{errors.New("fail")}, backoff.Constant(time.Nanosecond))
	}, backoff.WithMaxAttempts(2))

	if err == nil {
		t.Fatal("expected error after exceeding attempts")
	}
	if attempts != 3 { // initial + 2 retries
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}

	var ce *customError
	if !errors.As(err, &ce) {
		t.Fatalf("expected customError be wrapped, got %T", err)
	}

	wrap, err := backoff.Catch(err)
	if err == nil {
		t.Fatal("expected return wrapped error")
	}
	if wrap == nil {
		t.Fatal("expected wrap be not nil")
	}

	if err.Error() != "fail" {
		t.Fatal("unexpected error")
	}

	if len(wrap.StackTrace) == 0 {
		t.Fatal("expected stack trace recorded")
	}
}

func TestBackOff_NoBehaviour(t *testing.T) {
	ctx := context.Background()
	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		return errors.New("fail")
	}, backoff.WithMaxAttempts(1))
	if err == nil {
		t.Fatal("expected error when no behaviour provided")
	}
}

func TestBackOff_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	b := &mockBehaviour{delays: []time.Duration{time.Hour}}
	cancel()

	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		return errors.New("fail")
	}, backoff.WithMaxAttempts(3), backoff.WithDefaultBehaviour(b))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// Test errx

func TestBackOff_RetrySuccessWithErrX(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	b := &mockBehaviour{delays: []time.Duration{time.Millisecond}}

	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		bc := backoff.GetSettings(ctx)
		if bc == nil {
			t.Error("not found backoff context")
		}

		if bc.MaxAttempts() != 5 {
			t.Errorf("expected MaxAttempts=5, got %d", bc.MaxAttempts())
		}

		if bc.Attempt() != attempts {
			t.Errorf("expected Attempt=%d, got %d", attempts, bc.Attempt())
		}

		attempts++
		if attempts < 3 {
			errx.Errorf("fail")
		}
		return nil
	}, backoff.WithMaxAttempts(5), backoff.WithDefaultBehaviour(b), backoff.WithErrX())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.called < 2 {
		t.Fatalf("expected behaviour called at least twice, got %d", b.called)
	}
}

func TestBackOff_ExceedAttemptsWithErrX(t *testing.T) {
	ctx := context.Background()
	b := &mockBehaviour{delays: []time.Duration{time.Millisecond}}

	attempts := 0
	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		attempts++
		errx.Errorf("fail")
		return nil
	}, backoff.WithMaxAttempts(2), backoff.WithDefaultBehaviour(b), backoff.WithErrX())

	if err == nil {
		t.Fatal("expected error after exceeding attempts")
	}
	if attempts != 3 { // initial + 2 retries
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}

	wrap, err := backoff.Catch(err)
	if err == nil {
		t.Fatal("expected return wrapped error")
	}
	if wrap == nil {
		t.Fatal("expected wrap be not nil")
	}

	if err.Error() != "fail" {
		t.Fatal("unexpected error")
	}

	if len(wrap.StackTrace) == 0 {
		t.Fatal("expected stack trace recorded")
	}
}

func TestBackOff_WithWrappedBehaviourWithErrX(t *testing.T) {
	ctx := context.Background()
	wrappedB := &mockBehaviour{delays: []time.Duration{time.Millisecond}}

	attempts := 0
	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			errx.Error(backoff.Wrap(errors.New("wrapped"), wrappedB))
		}
		return nil
	}, backoff.WithMaxAttempts(3), backoff.WithErrX())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wrappedB.called == 0 || attempts != 2 {
		t.Fatalf("expected wrapped behaviour to be used, calls: %d", attempts)
	}
}

func TestBackOff_TestUnWrapWithErrX(t *testing.T) {
	ctx := context.Background()

	type customError struct {
		error
	}

	attempts := 0
	err := backoff.BackOff(ctx, func(ctx context.Context) error {
		attempts++
		errx.Error(backoff.Wrap(&customError{errors.New("fail")}, backoff.Constant(time.Nanosecond)))
		return nil
	}, backoff.WithMaxAttempts(2), backoff.WithErrX())

	if err == nil {
		t.Fatal("expected error after exceeding attempts")
	}
	if attempts != 3 { // initial + 2 retries
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}

	var ce *customError
	if !errors.As(err, &ce) {
		t.Fatalf("expected customError be wrapped, got %T", err)
	}

	wrap, err := backoff.Catch(err)
	if err == nil {
		t.Fatal("expected return wrapped error")
	}
	if wrap == nil {
		t.Fatal("expected wrap be not nil")
	}

	if err.Error() != "fail" {
		t.Fatal("unexpected error")
	}

	if len(wrap.StackTrace) == 0 {
		t.Fatal("expected stack trace recorded")
	}
}
