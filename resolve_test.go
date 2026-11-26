package octo_test

import (
	"slices"
	"testing"
	"time"

	"github.com/oesand/octo"
)

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
