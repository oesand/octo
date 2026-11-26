package octo_test

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/oesand/octo"
)

func TestTryResolveReturnsZeroValue(t *testing.T) {
	c := octo.New()
	res := octo.TryResolve[*MyService](c)
	if res != nil {
		t.Fatal("expected nil since nothing was registered")
	}
}

func TestResolveInjectionsIteration(t *testing.T) {
	c := octo.New()
	service, other := &MyService{}, &OtherService{}
	octo.InjectValue(c, service)
	octo.InjectValue(c, other)
	remaining := map[any]bool{service: true, other: true}

	iter := octo.ResolveInjections(c)
	for decl := range iter {
		value := decl.Value()
		if !remaining[value] {
			t.Fatalf("unexpected or duplicate injection: %#v", value)
		}
		delete(remaining, value)
	}

	if len(remaining) != 0 {
		t.Fatalf("missing injections: %v", remaining)
	}
}

func TestResolveAllIteration(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, &MyService{})
	octo.InjectValue(c, &OtherService{})

	sl := octo.ResolveAll[*MyService](c)

	if c := len(sl); c != 1 {
		t.Fatalf("expected 1 injection, got %d", c)
	}
}

func TestResolveAllIterationByInterface(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, &MyService{})
	octo.InjectValue(c, &MyService{})
	octo.InjectValue(c, &OtherService{})

	sl := octo.ResolveAll[ServiceInterface](c)

	if c := len(sl); c != 2 {
		t.Fatalf("expected 2 injection, got %d", c)
	}
}

func TestResolvePanicsIfRequiredNotFound(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for missing required service")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected panic not string: %T", r)
		}
		if msg != "octo: fail to resolve type *octo_test.MyService" {
			t.Fatalf("unexpected panic error message: %s", msg)
		}
	}()

	c := octo.New()
	octo.Resolve[*MyService](c)
}

func TestResolveDoesNotPanicIfOptionalNotFound(t *testing.T) {
	c := octo.New()
	res := octo.TryResolve[*MyService](c)
	if res != nil {
		t.Fatal("expected nil for TryResolve on missing service")
	}
}

func TestResolveContainer(t *testing.T) {
	c := octo.New()
	res := octo.Resolve[*octo.Container](c)
	if res != c {
		t.Fatal("expected container not returned")
	}
}

func TestTryResolveContainer(t *testing.T) {
	c := octo.New()
	res := octo.TryResolve[*octo.Container](c)
	if res != c {
		t.Fatal("expected container not returned")
	}
}

func TestTryResolveReturnsZeroValueByInterface(t *testing.T) {
	c := octo.New()
	res := octo.TryResolve[ServiceInterface](c)
	if res != nil {
		t.Fatal("expected nil since nothing was registered")
	}
}

func TestResolveInjectionsIterationByInterface(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, &MyService{name: "foo"})
	octo.InjectValue(c, &MyService{name: "bar"})
	var count atomic.Int32

	iter := octo.ResolveInjections(c)
	for decl := range iter {
		switch count.Load() {
		case 0:
			if !octo.OfType[ServiceInterface](decl) {
				t.Fatalf("expected type ServiceInterface, got %T", decl.Value())
			} else {
				if decl.Value().(*MyService).name != "foo" {
					t.Fatalf("expected foo, got %s", decl.Value())
				}
			}
		case 1:
			if !octo.OfType[ServiceInterface](decl) {
				t.Fatalf("expected type ServiceInterface, got %T", decl.Value())
			} else {
				if decl.Value().(*MyService).name != "bar" {
					t.Fatalf("expected bar, got %s", decl.Value())
				}
			}
		case 2:
			t.Fatal("unexpected iter")
		}

		count.Add(1)
	}

	if c := count.Load(); c != 2 {
		t.Fatalf("expected 2 injection, got %d", c)
	}
}

func TestResolvePanicsIfRequiredNotFoundByInterface(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for missing required service")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected panic not string: %T", r)
		}
		if msg != "octo: fail to resolve type octo_test.ServiceInterface" {
			t.Fatalf("unexpected panic error message: %s", msg)
		}
	}()

	c := octo.New()
	octo.Resolve[ServiceInterface](c)
}

func TestResolveDoesNotPanicIfOptionalNotFoundByInterface(t *testing.T) {
	c := octo.New()
	res := octo.TryResolve[ServiceInterface](c)
	if res != nil {
		t.Fatal("expected nil for TryResolve on missing service")
	}
}

func TestResolveProviderCanRegisterDependency(t *testing.T) {
	for name, resolve := range map[string]func(*octo.Container) int{
		"Resolve":             func(c *octo.Container) int { return octo.Resolve[int](c) },
		"TryResolve":          func(c *octo.Container) int { return octo.TryResolve[int](c) },
		"ResolveAllConcrete":  func(c *octo.Container) int { return octo.ResolveAll[int](c)[0] },
		"ResolveAllInterface": func(c *octo.Container) int { return octo.ResolveAll[any](c)[0].(int) },
	} {
		t.Run(name, func(t *testing.T) {
			c := octo.New()
			octo.Inject(c, func(c *octo.Container) int {
				octo.InjectValue(c, "dependency")
				return len(octo.Resolve[string](c))
			})
			done := make(chan int, 1)
			go func() { done <- resolve(c) }()
			select {
			case got := <-done:
				if got != len("dependency") {
					t.Fatalf("unexpected provider result: %d", got)
				}
			case <-time.After(time.Second):
				t.Fatal("provider blocked while registering a dependency")
			}
		})
	}
}

func TestResolveInjectionsUsesSnapshot(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, 1)
	octo.InjectValue(c, 2)
	done := make(chan []int, 1)
	go func() {
		var values []int
		for decl := range octo.ResolveInjections(c) {
			if len(values) == 0 {
				octo.CleanInjections(c, func(octo.Declaration) bool { return true })
				octo.InjectValue(c, 3)
			}
			values = append(values, decl.Value().(int))
		}
		done <- values
	}()
	select {
	case got := <-done:
		if !slices.Equal(got, []int{1, 2}) {
			t.Fatalf("expected snapshot [1 2], got %v", got)
		}
		if got := octo.Resolve[int](c); got != 3 {
			t.Fatalf("expected replacement registration 3, got %d", got)
		}
	case <-time.After(time.Second):
		t.Fatal("iterator blocked while changing registrations")
	}
}

func TestResolveInterfaceRegistrationOrder(t *testing.T) {
	for _, name := range []string{"", "target"} {
		t.Run("name="+name, func(t *testing.T) {
			c := octo.New()
			// An earlier registration of the same concrete type must not
			// give its later named registration priority over another type.
			octo.InjectNamedValue(c, "other", 0)
			octo.InjectNamed(c, name, func(*octo.Container) string { return "first" })
			octo.InjectNamedValue(c, name, 1)
			if name == "" {
				octo.CleanInjections(c, func(d octo.Declaration) bool { return d.Name() == "other" })
			}
			for i := 0; i < 100; i++ {
				// Also check after cache invalidation.
				octo.CleanInjections(c, func(octo.Declaration) bool { return false })
				if got := octo.ResolveNamed[any](c, name); got != "first" {
					t.Fatalf("expected first registration, got %#v", got)
				}
			}
			octo.CleanInjections(c, func(d octo.Declaration) bool { return octo.OfType[string](d) })
			octo.InjectNamedValue(c, name, "last")
			if got := octo.ResolveNamed[any](c, name); got != 1 {
				t.Fatalf("expected next remaining registration, got %#v", got)
			}
		})
	}
}

func TestLazyInjectionRepeatsPanic(t *testing.T) {
	c := octo.New()
	want := errors.New("constructor failed")
	var calls atomic.Int32
	octo.Inject(c, func(*octo.Container) int {
		calls.Add(1)
		panic(want)
	})
	if calls.Load() != 0 {
		t.Fatal("provider ran before resolution")
	}
	// Check the initial panic, then concurrent calls after initialization failed.
	resolve := func() {
		defer func() {
			if got := recover(); got != want {
				t.Errorf("expected panic %v, got %v", want, got)
			}
		}()
		octo.Resolve[int](c)
	}
	resolve()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(resolve)
	}
	wg.Wait()
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected one provider call, got %d", got)
	}
}

func TestLazyInjectionConcurrentValue(t *testing.T) {
	c := octo.New()
	want := &MyService{name: "singleton"}
	var calls atomic.Int32
	octo.Inject(c, func(*octo.Container) *MyService {
		calls.Add(1)
		return want
	})
	if calls.Load() != 0 {
		t.Fatal("provider ran before resolution")
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if got := octo.Resolve[*MyService](c); got != want {
				t.Errorf("expected shared instance %p, got %p", want, got)
			}
		})
	}
	wg.Wait()
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected one provider call, got %d", got)
	}
}
