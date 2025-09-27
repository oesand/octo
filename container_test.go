package octo

import (
	"sync/atomic"
	"testing"
)

type ServiceInterface interface {
	Hello() string
}

type MyService struct {
	name string
}

func (s *MyService) Hello() string { return "hi" }

type OtherService struct{}

// Test direct type
func TestInjectValueAndResolve(t *testing.T) {
	c := New()
	InjectValue(c, &MyService{})

	res := Resolve[*MyService](c)
	if res == nil {
		t.Fatal("expected non-nil MyService")
	}
	if res.Hello() != "hi" {
		t.Fatalf("expected 'hi', got %s", res.Hello())
	}
}

func TestInjectNamedValueAndResolveNamed(t *testing.T) {
	c := New()
	InjectNamedValue(c, "foo", &MyService{})

	res := ResolveNamed[*MyService](c, "foo")
	if res == nil {
		t.Fatal("expected non-nil MyService for named injection")
	}
}

func TestInjectProviderAndResolve(t *testing.T) {
	c := New()
	Inject[*MyService](c, func(c *Container) *MyService { return &MyService{} })

	res := Resolve[*MyService](c)
	if res == nil {
		t.Fatal("expected non-nil MyService from provider")
	}
}

func TestTryResolveReturnsZeroValue(t *testing.T) {
	c := New()
	res := TryResolve[*MyService](c)
	if res != nil {
		t.Fatal("expected nil since nothing was registered")
	}
}

func TestResolveInjectionsIteration(t *testing.T) {
	c := New()
	InjectValue(c, &MyService{})
	InjectValue(c, &OtherService{})
	var count atomic.Int32

	iter := ResolveInjections(c)
	for decl := range iter {
		switch count.Load() {
		case 0:
			if !DeclOfType[*MyService](decl) {
				t.Fatalf("expected type MyService, got %T", decl.Value())
			}
		case 1:
			if !DeclOfType[*OtherService](decl) {
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

func TestResolveOfTypeIteration(t *testing.T) {
	c := New()
	InjectValue(c, &MyService{})
	InjectValue(c, &OtherService{})

	sl := ResolveOfType[*MyService](c)

	if c := len(sl); c != 1 {
		t.Fatalf("expected 1 injection, got %d", c)
	}
}

func TestResolveOfTypeIterationByInterface(t *testing.T) {
	c := New()
	InjectValue(c, &MyService{})
	InjectValue(c, &MyService{})
	InjectValue(c, &OtherService{})

	sl := ResolveOfType[ServiceInterface](c)

	if c := len(sl); c != 2 {
		t.Fatalf("expected 2 injection, got %d", c)
	}
}

func TestCleanInjectionsRemovesSelected(t *testing.T) {
	c := New()
	InjectValue(c, &MyService{})
	InjectValue(c, &OtherService{})

	// Remove MyService
	CleanInjections(c, func(s ServiceDeclaration) bool {
		return DeclOfType[*MyService](s)
	})

	var count atomic.Int32
	iter := ResolveInjections(c)
	for decl := range iter {
		switch count.Load() {
		case 0:
			if !DeclOfType[*OtherService](decl) {
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
		if msg != "octo: fail to resolve type *octo.MyService" {
			t.Fatalf("unexpected panic error message: %s", msg)
		}
	}()

	c := New()
	Resolve[*MyService](c)
}

func TestResolveDoesNotPanicIfOptionalNotFound(t *testing.T) {
	c := New()
	res := TryResolve[*MyService](c)
	if res != nil {
		t.Fatal("expected nil for TryResolve on missing service")
	}
}

// Test interface

func TestInjectValueAndResolveByInterface(t *testing.T) {
	c := New()
	InjectValue(c, &OtherService{})
	InjectValue(c, &MyService{name: "foo"})
	InjectValue(c, &MyService{name: "bar"})

	res := Resolve[ServiceInterface](c)
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
	c := New()
	InjectNamedValue(c, "foo", &MyService{name: "foo"})
	InjectNamedValue(c, "bar", &MyService{name: "bar"})
	InjectNamedValue(c, "bar", &OtherService{})
	InjectNamedValue(c, "invalid", &MyService{name: "invalid"})

	res := ResolveNamed[ServiceInterface](c, "bar")
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
	c := New()
	Inject(c, func(container *Container) *MyService { return &MyService{} })
	Inject(c, func(container *Container) *OtherService { return &OtherService{} })

	res := Resolve[ServiceInterface](c)
	if res == nil {
		t.Fatal("expected non-nil MyService")
	}
	if _, ok := res.(*MyService); !ok {
		t.Fatalf("expected MyService, got %T", res)
	}
}

func TestTryResolveReturnsZeroValueByInterface(t *testing.T) {
	c := New()
	res := TryResolve[ServiceInterface](c)
	if res != nil {
		t.Fatal("expected nil since nothing was registered")
	}
}

func TestResolveInjectionsIterationByInterface(t *testing.T) {
	c := New()
	InjectValue(c, &MyService{name: "foo"})
	InjectValue(c, &MyService{name: "bar"})
	var count atomic.Int32

	iter := ResolveInjections(c)
	for decl := range iter {
		switch count.Load() {
		case 0:
			if !DeclOfType[ServiceInterface](decl) {
				t.Fatalf("expected type ServiceInterface, got %T", decl.Value())
			} else {
				if decl.Value().(*MyService).name != "foo" {
					t.Fatalf("expected foo, got %s", decl.Value())
				}
			}
		case 1:
			if !DeclOfType[ServiceInterface](decl) {
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
	c := New()
	InjectValue(c, &MyService{name: "foo"})
	InjectValue(c, &OtherService{})
	InjectValue(c, &MyService{name: "bar"})

	// Remove MyService
	CleanInjections(c, func(s ServiceDeclaration) bool {
		return DeclOfType[ServiceInterface](s)
	})

	var count atomic.Int32
	iter := ResolveInjections(c)
	for decl := range iter {
		switch count.Load() {
		case 0:
			if !DeclOfType[*OtherService](decl) {
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
		if msg != "octo: fail to resolve type octo.ServiceInterface" {
			t.Fatalf("unexpected panic error message: %s", msg)
		}
	}()

	c := New()
	Resolve[ServiceInterface](c)
}

func TestResolveDoesNotPanicIfOptionalNotFoundByInterface(t *testing.T) {
	c := New()
	res := TryResolve[ServiceInterface](c)
	if res != nil {
		t.Fatal("expected nil for TryResolve on missing service")
	}
}
