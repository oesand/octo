package octo

import (
	"fmt"
	"iter"
	"reflect"
	"sync"
)

// DefaultContainer is a global fallback container used when none is provided.
var DefaultContainer Container

// New creates a new empty container instance.
func New() *Container {
	return &Container{}
}

// Container stores service declarations and provides thread-safe access.
type Container struct {
	mu       sync.RWMutex
	services []ServiceDeclaration
}

func containerOrDefault(container *Container) *Container {
	if container == nil {
		return &DefaultContainer
	}
	return container
}

// InjectValue registers a concrete value into the container.
func InjectValue[T any](container *Container, value T) {
	InjectNamedValue[T](container, "", value)
}

// InjectNamedValue registers a concrete value with a name for named resolution.
func InjectNamedValue[T any](container *Container, name string, value T) {
	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	typ := reflect.TypeFor[T]()
	decl := newDeclValue(typ, name, value)
	container.services = append(container.services, decl)
}

// Inject registers a provider function to lazily resolve a type.
func Inject[T any](container *Container, provider Provider[T]) {
	InjectNamed[T](container, "", provider)
}

// InjectNamed registers a named provider function to lazily resolve a type.
func InjectNamed[T any](container *Container, name string, provider Provider[T]) {
	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	typ := reflect.TypeFor[T]()
	decl := newDeclLazy(typ, name, func() any {
		return provider(container)
	})
	container.services = append(container.services, decl)
}

// Resolve returns the first registered instance of type T.
// Panics if not found.
func Resolve[T any](container *Container) T {
	return ResolveNamed[T](container, "")
}

// TryResolve attempts to return the first registered instance of type T.
// Returns zero value if not found.
func TryResolve[T any](container *Container) T {
	return TryResolveNamed[T](container, "")
}

// ResolveNamed returns the instance of type T with the specified name.
// Panics if not found.
func ResolveNamed[T any](container *Container, name string) T {
	return resolve[T](container, name, true)
}

// TryResolveNamed returns the instance of type T with the specified name.
// Returns zero value if not found.
func TryResolveNamed[T any](container *Container, name string) T {
	return resolve[T](container, name, false)
}

func resolve[T any](container *Container, name string, required bool) T {
	container = containerOrDefault(container)
	container.mu.RLock()
	defer container.mu.RUnlock()

	var decl ServiceDeclaration

	typ := reflect.TypeFor[T]()
	for _, inject := range container.services {
		if !inject.Type().AssignableTo(typ) {
			continue
		}

		if name != "" && inject.Name() != name {
			continue
		}

		decl = inject
		break
	}

	if required && decl == nil {
		panic(fmt.Sprintf("octo: fail to resolve type %s", reflect.TypeFor[T]().String()))
	}

	var res T
	if decl != nil {
		if val := decl.Value(); val != nil {
			res = val.(T)
		}
	}

	return res
}

// ResolveInjections returns an iterator over all registered services in the container.
func ResolveInjections(container *Container) iter.Seq[ServiceDeclaration] {
	container = containerOrDefault(container)
	return func(yield func(ServiceDeclaration) bool) {
		container.mu.RLock()
		defer container.mu.RUnlock()

		for _, service := range container.services {
			if !yield(service) {
				break
			}
		}
	}
}

// ResolveAll returns an iterator over registered services in the container
// if the service's type is assignable to T (implements interface or same type).
func ResolveAll[T any](container *Container) []T {
	injects := ResolveInjections(container)
	var result []T
	for decl := range injects {
		if DeclOfType[T](decl) {
			result = append(result, decl.Value().(T))
		}
	}
	return result
}

// CleanInjections removes all service declarations that match the selector function.
// Do not use octo.* functions inside selector, this may cause deadlocks
func CleanInjections(container *Container, selector func(decl ServiceDeclaration) bool) {
	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	services := make([]ServiceDeclaration, 0, len(container.services))
	for _, decl := range container.services {
		if !selector(decl) {
			services = append(services, decl)
		}
	}
	container.services = services
}
