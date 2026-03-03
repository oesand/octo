package octo_test

import (
	"sync/atomic"
	"testing"

	"github.com/oesand/octo"
)

type ServiceInterface interface {
	Hello() string
	Name() string
}

type MyService struct {
	name string
}

func (s *MyService) Hello() string {
	return "hi"
}

func (s *MyService) Name() string {
	return s.name
}

type OtherService struct{}

func TestTryInjectValue_Struct(t *testing.T) {
	c := octo.New()

	ok := octo.TryInjectValue(c, &MyService{name: "struct"})
	if !ok {
		t.Fatalf("expected TryInjectValue to succeed for struct")
	}

	res := octo.Resolve[*MyService](c)
	if res == nil || res.name != "struct" {
		t.Fatalf("expected resolved struct 'struct', got %#v", res)
	}
}

func TestTryInjectValue_DuplicateStruct(t *testing.T) {
	c := octo.New()

	_ = octo.TryInjectValue(c, &MyService{name: "first"})
	ok := octo.TryInjectValue(c, &MyService{name: "second"})
	if ok {
		t.Fatalf("expected duplicate inject to fail")
	}

	res := octo.Resolve[*MyService](c)
	if res.name != "first" {
		t.Fatalf("expected first value, got %q", res.name)
	}
}

func TestTryInjectValue_Interface(t *testing.T) {
	c := octo.New()

	octo.InjectValue(c, &OtherService{})

	ok := octo.TryInjectValue(c, &MyService{name: "iface"})
	if !ok {
		t.Fatalf("expected TryInjectValue to succeed for interface")
	}

	res := octo.Resolve[ServiceInterface](c)

	if name := res.Name(); name != "iface" {
		t.Fatalf("expected iface value, got %q", name)
	}

	if res.Hello() != "hi" {
		t.Fatalf("expected resolved interface with Hello=hi, got %#v", res)
	}
}

func TestTryInject_InterfaceProvider(t *testing.T) {
	c := octo.New()

	octo.InjectValue(c, &OtherService{})

	ok := octo.TryInject(c, func(c *octo.Container) ServiceInterface {
		return &MyService{name: "provider"}
	})
	if !ok {
		t.Fatalf("expected TryInject with interface provider to succeed")
	}

	res := octo.Resolve[ServiceInterface](c)
	if res == nil || res.Hello() != "hi" {
		t.Fatalf("expected provider Hello=hi, got %#v", res)
	}
}

func TestTryInject_InterfaceDuplicate(t *testing.T) {
	c := octo.New()

	octo.InjectValue(c, &OtherService{})

	_ = octo.TryInject(c, func(c *octo.Container) ServiceInterface {
		return &MyService{name: "first"}
	})

	ok := octo.TryInject(c, func(c *octo.Container) ServiceInterface {
		return &MyService{name: "second"}
	})
	if ok {
		t.Fatalf("expected duplicate interface inject to fail")
	}

	res := octo.Resolve[ServiceInterface](c)
	if name := res.Name(); name != "first" {
		t.Fatalf("expected first value, got %q", name)
	}

	if got := res.Hello(); got != "hi" {
		t.Fatalf("expected first provider Hello=hi, got %q", got)
	}
}

// Test direct type
func TestInjectValueAndResolve(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, &OtherService{})
	octo.InjectValue(c, &MyService{})

	res := octo.Resolve[*MyService](c)
	if res == nil {
		t.Fatal("expected non-nil MyService")
	}
	if res.Hello() != "hi" {
		t.Fatalf("expected 'hi', got %s", res.Hello())
	}
}

func TestInjectNamedValueAndResolveNamed(t *testing.T) {
	c := octo.New()
	octo.InjectNamedValue(c, "foo", &MyService{})

	res := octo.ResolveNamed[*MyService](c, "foo")
	if res == nil {
		t.Fatal("expected non-nil MyService for named injection")
	}
}

func TestInjectProviderAndResolve(t *testing.T) {
	c := octo.New()
	octo.Inject[*MyService](c, func(c *octo.Container) *MyService { return &MyService{} })

	res := octo.Resolve[*MyService](c)
	if res == nil {
		t.Fatal("expected non-nil MyService from provider")
	}
}

func TestTryResolveReturnsZeroValue(t *testing.T) {
	c := octo.New()
	res := octo.TryResolve[*MyService](c)
	if res != nil {
		t.Fatal("expected nil since nothing was registered")
	}
}

func TestResolveInjectionsIteration(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, &MyService{})
	octo.InjectValue(c, &OtherService{})
	var count atomic.Int32

	iter := octo.ResolveInjections(c)
	for decl := range iter {
		switch count.Load() {
		case 0:
			if !octo.OfType[*MyService](decl) {
				t.Fatalf("expected type MyService, got %T", decl.Value())
			}
		case 1:
			if !octo.OfType[*OtherService](decl) {
				t.Fatalf("expected type OtherService, got %T", decl.Value())
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

// Test interface

func TestInjectValueAndResolveByInterface(t *testing.T) {
	c := octo.New()
	octo.InjectValue(c, &OtherService{})
	octo.InjectValue(c, &MyService{name: "foo"})
	octo.InjectValue(c, &MyService{name: "bar"})

	res := octo.Resolve[ServiceInterface](c)
	if res == nil {
		t.Fatal("expected non-nil MyService")
	}
	srv, ok := res.(*MyService)
	if !ok {
		t.Fatalf("expected MyService, got %T", res)
	}
	if srv.name != "foo" {
		t.Fatalf("unexpected name, got %s", srv.name)
	}
}

func TestInjectNamedValueAndResolveNamedByInterface(t *testing.T) {
	c := octo.New()
	octo.InjectNamedValue(c, "foo", &MyService{name: "foo"})
	octo.InjectNamedValue(c, "bar", &MyService{name: "bar"})
	octo.InjectNamedValue(c, "bar", &OtherService{})
	octo.InjectNamedValue(c, "invalid", &MyService{name: "invalid"})

	res := octo.ResolveNamed[ServiceInterface](c, "bar")
	if res == nil {
		t.Fatal("expected non-nil MyService for named injection")
	}
	srv, ok := res.(*MyService)
	if !ok {
		t.Fatalf("expected MyService, got %T", res)
	}
	if srv.name != "bar" {
		t.Fatalf("unexpected name, got %s", srv.name)
	}
}

func TestInjectProviderAndResolveByInterface(t *testing.T) {
	c := octo.New()
	octo.Inject(c, func(container *octo.Container) *MyService { return &MyService{} })
	octo.Inject(c, func(container *octo.Container) *OtherService { return &OtherService{} })

	res := octo.Resolve[ServiceInterface](c)
	if res == nil {
		t.Fatal("expected non-nil MyService")
	}
	if _, ok := res.(*MyService); !ok {
		t.Fatalf("expected MyService, got %T", res)
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

func TestPanicWhenInjectContainer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(string); !ok || err != "cannot inject Container" {
				t.Fatalf("got unexpected error message: %s \n", err)
			}
		} else {
			t.Fatal("expected a panic")
		}
	}()

	c := octo.New()
	octo.Inject(c, func(c *octo.Container) *octo.Container {
		return c
	})
}

func TestPanicWhenInjectValueContainer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(string); !ok || err != "cannot inject Container" {
				t.Fatalf("got unexpected error message: %s \n", err)
			}
		} else {
			t.Fatal("expected a panic")
		}
	}()

	c := octo.New()
	octo.InjectValue(c, c)
}
