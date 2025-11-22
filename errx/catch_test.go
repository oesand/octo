package errx_test

import (
	"context"
	"errors"
	"github.com/oesand/octo/errx"
	"sync/atomic"
	"testing"
)

func Test_TryHandleErrorf(t *testing.T) {
	ctx := context.Background()
	wrap := errx.Try(ctx, func(ctx context.Context) {
		errx.Errorf("test err")
	})

	if wrap == nil {
		t.Fatal("wrap is nil")
	}
	if wrap.Error() != "test err" {
		t.Errorf("wrap.Error() = %q, want %q", wrap.Error(), "test err")
	}
}

func Test_TryHandleError(t *testing.T) {
	ctx := context.Background()
	wrap := errx.Try(ctx, func(ctx context.Context) {
		errx.Error(errors.New("test err"))
	})

	if wrap == nil {
		t.Fatal("wrap is nil")
	}
	if wrap.Error() != "test err" {
		t.Errorf("wrap.Error() = %q, want %q", wrap.Error(), "test err")
	}
}

func Test_TrySkipUnknownPanic(t *testing.T) {
	ctx := context.Background()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}

		if err := r.(error); err.Error() != "test err" {
			t.Errorf("wrap.Error() = %q, want %q", err.Error(), "test err")
		}
	}()

	errx.Try(ctx, func(ctx context.Context) {
		panic(errors.New("test err"))
	})
}

type customError struct{}

func (customError) Error() string {
	return "custom error"
}

func Test_TryCatch(t *testing.T) {
	ctx := context.Background()
	var handle atomic.Int32
	wrap := errx.Try(ctx, func(ctx context.Context) {
		errx.Error(customError{})
	}, errx.Catch(func(ctx context.Context, e customError) {
		handle.Add(1)
	}), errx.Catch(func(ctx context.Context, e customError) {
		handle.Add(1)
	}))

	if wrap == nil {
		t.Fatal("wrap is nil")
	}
	if _, ok := wrap.Unwrap().(customError); !ok {
		t.Errorf("wrap.Error() = %q, want %q", wrap.Error(), "custom error")
	}
	if handle.Load() != 1 {
		t.Errorf("handle.Load() = %d, want %d", handle.Load(), 1)
	}
}

func TestErrorf_Panic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got nil")
		}

		wrap, ok := r.(*errx.ErrWrap)
		if !ok {
			t.Fatalf("expected *ErrWrap, got %T", r)
		}

		if wrap.Error() != "formatted 42" {
			t.Errorf("expected 'formatted 42', got %q", wrap.Error())
		}
	}()

	errx.Errorf("formatted %d", 42)
}

func TestError_Panic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got nil")
		}

		wrap, ok := r.(*errx.ErrWrap)
		if !ok {
			t.Fatalf("expected *ErrWrap, got %T", r)
		}

		if wrap.Error() != "simple error" {
			t.Errorf("expected 'simple error', got %q", wrap.Error())
		}
	}()

	errx.Error(errors.New("simple error"))
}

func TestTry_Catch(t *testing.T) {
	ctx := context.Background()
	caught := false

	errx.Try(ctx, func(ctx context.Context) {
		errx.Errorf("panic inside Try")
	}, errx.Catch(func(ctx context.Context, e error) {
		caught = true
	}))

	if !caught {
		t.Error("expected error to be caught")
	}
}

func TestCatch_TypeMatching(t *testing.T) {
	ctx := context.Background()
	myErr := errors.New("specific error")
	caught := false

	errx.Try(ctx, func(ctx context.Context) {
		errx.Error(myErr)
	}, errx.Catch(func(ctx context.Context, e error) {
		if e.Error() == "specific error" {
			caught = true
		}
	}))

	if !caught {
		t.Error("expected specific error to be caught")
	}
}

func TestTry_RepanicNonErrWrap(t *testing.T) {
	defer func() {
		r := recover()
		if r != "non-ErrWrap panic" {
			t.Errorf("expected re-panic with 'non-ErrWrap panic', got %v", r)
		}
	}()

	errx.Try(context.Background(), func(ctx context.Context) {
		panic("non-ErrWrap panic")
	})
}

func TestErrWrap_StackTracePopulated(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got nil")
		}

		wrap, ok := r.(*errx.ErrWrap)
		if !ok {
			t.Fatalf("expected *ErrWrap, got %T", r)
		}

		if len(wrap.StackTrace) == 0 {
			t.Error("expected stack trace to be populated, got empty")
		}
	}()

	errx.Errorf("check stack trace")
}
