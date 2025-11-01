package backoff

import (
	"context"
	"errors"
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
	got, err := BackOff(ctx, func(ctx context.Context) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Fatalf("expected ok, got %s", got)
	}
}

func TestBackOff_RetrySuccess(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	b := &mockBehaviour{delays: []time.Duration{time.Millisecond}}

	got, err := BackOff(ctx, func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("fail")
		}
		return "done", nil
	}, WithMaxAttempts(5), WithDefaultBehaviour(b))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "done" {
		t.Fatalf("expected done, got %s", got)
	}
	if b.called < 2 {
		t.Fatalf("expected behaviour called at least twice, got %d", b.called)
	}
}

func TestBackOff_ExceedAttempts(t *testing.T) {
	ctx := context.Background()
	b := &mockBehaviour{delays: []time.Duration{time.Millisecond}}

	attempts := 0
	_, err := BackOff(ctx, func(ctx context.Context) (string, error) {
		attempts++
		return "", errors.New("fail")
	}, WithMaxAttempts(2), WithDefaultBehaviour(b))

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
	_, err := BackOff(ctx, func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 2 {
			return "", Wrap(errors.New("wrapped"), wrappedB)
		}
		return "ok", nil
	}, WithMaxAttempts(3))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wrappedB.called == 0 {
		t.Fatal("expected wrapped behaviour to be used")
	}
}

func TestBackOff_NoBehaviour(t *testing.T) {
	ctx := context.Background()
	_, err := BackOff(ctx, func(ctx context.Context) (int, error) {
		return 0, errors.New("fail")
	}, WithMaxAttempts(1))
	if err == nil {
		t.Fatal("expected error when no behaviour provided")
	}
}

func TestBackOff_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	b := &mockBehaviour{delays: []time.Duration{time.Hour}}
	cancel()

	_, err := BackOff(ctx, func(ctx context.Context) (int, error) {
		return 0, errors.New("fail")
	}, WithMaxAttempts(3), WithDefaultBehaviour(b))

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestWrap_ContainsStackTrace(t *testing.T) {
	err := errors.New("oops")
	wrapped := Wrap(err, &mockBehaviour{})
	bw, ok := wrapped.(*BackoffWrap)
	if !ok {
		t.Fatal("expected *BackoffWrap type")
	}
	if len(bw.StackTrace) == 0 {
		t.Fatal("expected stack trace recorded")
	}
}
