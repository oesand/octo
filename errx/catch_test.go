package errx_test

import (
	"errors"
	"github.com/oesand/octo/errx"
	"sync/atomic"
	"testing"
)

func Test_TryHandleErrorf(t *testing.T) {
	wrap := errx.Try(func() {
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
	wrap := errx.Try(func() {
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
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}

		if err := r.(error); err.Error() != "test err" {
			t.Errorf("wrap.Error() = %q, want %q", err.Error(), "test err")
		}
	}()

	errx.Try(func() {
		panic(errors.New("test err"))
	})
}

type customError struct{}

func (customError) Error() string {
	return "custom error"
}

func Test_TryCatch(t *testing.T) {
	var handle atomic.Int32
	wrap := errx.Try(func() {
		errx.Error(customError{})
	}, errx.Catch(func(e customError) {
		handle.Add(1)
	}), errx.Catch(func(e customError) {
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
	caught := false

	errx.Try(func() {
		errx.Errorf("panic inside Try")
	}, errx.Catch(func(e error) {
		caught = true
	}))

	if !caught {
		t.Error("expected error to be caught")
	}
}

func TestCatch_TypeMatching(t *testing.T) {
	myErr := errors.New("specific error")
	caught := false

	errx.Try(func() {
		errx.Error(myErr)
	}, errx.Catch(func(e error) {
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

	errx.Try(func() {
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
