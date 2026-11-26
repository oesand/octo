package octo_test

import (
	"errors"
	"slices"
	"sync/atomic"
	"testing"

	"github.com/oesand/octo"
)

func TestCleanInjectionsRemovesSelected(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, &MyService{})
	octo.InjectValue(c, &OtherService{})

	// Remove MyService
	octo.CleanInjections(c, func(s octo.Declaration) bool {
		return octo.OfType[*MyService](s)
	})

	var count atomic.Int32
	iter := octo.ResolveInjections(c)
	for decl := range iter {
		switch count.Load() {
		case 0:
			if !octo.OfType[*OtherService](decl) {
				t.Fatalf("expected type OtherService, got %T", decl.Value())
			}
		case 1:
			t.Fatal("unexpected iter")
		}

		count.Add(1)
	}

	if c := count.Load(); c != 1 {
		t.Fatalf("expected 1 remaining service, got %d", c)
	}
}

func TestCleanInjectionsRemovesSelectedByInterface(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, &MyService{name: "foo"})
	octo.InjectValue(c, &OtherService{})
	octo.InjectValue(c, &MyService{name: "bar"})

	// Remove MyService
	octo.CleanInjections(c, func(s octo.Declaration) bool {
		return octo.OfType[ServiceInterface](s)
	})

	var count atomic.Int32
	iter := octo.ResolveInjections(c)
	for decl := range iter {
		switch count.Load() {
		case 0:
			if !octo.OfType[*OtherService](decl) {
				t.Fatalf("expected type OtherService, got %T", decl.Value())
			}
		case 1:
			t.Fatal("unexpected iter")
		}

		count.Add(1)
	}

	if c := count.Load(); c != 1 {
		t.Fatalf("expected 1 remaining service, got %d", c)
	}
}

func TestCleanInjectionsSelectorPanicPreservesRegistrations(t *testing.T) {
	c := octo.New()
	octo.InjectNamedValue(c, "number", 1)
	octo.InjectNamedValue(c, "text", "value")
	// Populate the caches before the interrupted cleanup.
	octo.Resolve[int](c)
	octo.Resolve[string](c)
	octo.Resolve[any](c)
	wantPanic := errors.New("selector failed")
	calls := 0
	func() {
		defer func() {
			if got := recover(); got != wantPanic {
				t.Errorf("expected selector panic %v, got %v", wantPanic, got)
			}
		}()
		octo.CleanInjections(c, func(octo.Declaration) bool {
			calls++
			if calls == 2 {
				panic(wantPanic)
			}
			return true
		})
	}()
	if got := octo.ResolveAll[int](c); !slices.Equal(got, []int{1}) {
		t.Errorf("number registration changed after panic: %v", got)
	}
	if got := octo.ResolveAll[string](c); !slices.Equal(got, []string{"value"}) {
		t.Errorf("text registration changed after panic: %v", got)
	}
	count := 0
	for range octo.ResolveInjections(c) {
		count++
	}
	if count != 2 {
		t.Errorf("expected both registrations after panic, got %d", count)
	}
	if got := octo.ResolveNamed[any](c, "number"); got != 1 {
		t.Errorf("expected number registration, got %#v", got)
	}
	if got := octo.ResolveNamed[any](c, "text"); got != "value" {
		t.Errorf("expected text registration, got %#v", got)
	}

	// A later successful cleanup must update both indexes and invalidate caches.
	octo.CleanInjections(c, func(d octo.Declaration) bool { return d.Name() == "number" })
	if got := octo.TryResolve[int](c); got != 0 {
		t.Errorf("removed number remains cached: %d", got)
	}
	if got := octo.TryResolveNamed[any](c, "number"); got != nil {
		t.Errorf("removed number remains in registration order: %#v", got)
	}
	if got := octo.Resolve[any](c); got != "value" {
		t.Errorf("expected remaining text registration, got %#v", got)
	}
}
