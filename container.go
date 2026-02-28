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
	mu      sync.RWMutex
	injects []InjectDeclaration

	resolveCacheMu sync.RWMutex
	resolveCache   map[reflect.Type]InjectDeclaration
}

func containerOrDefault(container *Container) *Container {
	if container == nil {
		return &DefaultContainer
	}
	return container
}

func (c *Container) addValue(typ reflect.Type, name string, value any) {
	decl := newDeclValue(typ, name, value)
	c.injects = append(c.injects, decl)
}

func (c *Container) addProvider(typ reflect.Type, name string, provider func() any) {
	decl := newDeclLazy(typ, name, provider)
	c.injects = append(c.injects, decl)
}

func (c *Container) resolve(typ reflect.Type, name string) InjectDeclaration {
	if name == "" {
		c.resolveCacheMu.RLock()
		if len(c.resolveCache) > 0 {
			if decl, ok := c.resolveCache[typ]; ok {
				return decl
			}
		}
		c.resolveCacheMu.RUnlock()
	}

	for _, inject := range c.injects {
		if !inject.Type().AssignableTo(typ) {
			continue
		}

		if name != "" && inject.Name() != name {
			continue
		}

		if name == "" {
			c.resolveCacheMu.Lock()
			if c.resolveCache == nil {
				c.resolveCache = make(map[reflect.Type]InjectDeclaration)
			}

			c.resolveCache[typ] = inject
			c.resolveCacheMu.Unlock()
		}

		return inject
	}

	return nil
}

// TryInjectValue registers a concrete value into the container if not registered.
func TryInjectValue[T any](container *Container, value T) bool {
	return TryInjectNamedValue[T](container, "", value)
}

// TryInjectNamedValue registers a concrete value with a name into the container if not registered.
func TryInjectNamedValue[T any](container *Container, name string, value T) bool {
	typ := reflect.TypeFor[T]()
	ensureCanInjectType(typ)

	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	if container.resolve(typ, name) != nil {
		return false
	}

	container.addValue(typ, name, value)
	return true
}

// InjectValue registers a concrete value into the container.
func InjectValue[T any](container *Container, value T) {
	InjectNamedValue[T](container, "", value)
}

// InjectNamedValue registers a concrete value with a name for named resolution.
func InjectNamedValue[T any](container *Container, name string, value T) {
	typ := reflect.TypeFor[T]()
	ensureCanInjectType(typ)

	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	container.addValue(typ, name, value)
}

// TryInject registers a provider function to lazily resolve a type if not registered.
func TryInject[T any](container *Container, provider Provider[T]) bool {
	return TryInjectNamed(container, "", provider)
}

// TryInjectNamed registers a named provider function to lazily resolve a type if not registered.
func TryInjectNamed[T any](container *Container, name string, provider Provider[T]) bool {
	typ := reflect.TypeFor[T]()
	ensureCanInjectType(typ)

	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	if container.resolve(typ, name) != nil {
		return false
	}

	container.addProvider(typ, name, func() any {
		return provider(container)
	})
	return true
}

// Inject registers a provider function to lazily resolve a type.
func Inject[T any](container *Container, provider Provider[T]) {
	InjectNamed[T](container, "", provider)
}

// InjectNamed registers a named provider function to lazily resolve a type.
func InjectNamed[T any](container *Container, name string, provider Provider[T]) {
	typ := reflect.TypeFor[T]()
	ensureCanInjectType(typ)

	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	container.addProvider(typ, name, func() any {
		return provider(container)
	})
}

// Resolve returns the first registered instance of type T.
// Panics if not found.
func Resolve[T any](container *Container) T {
	return ResolveNamed[T](container, "")
}

// ResolveNamed returns the instance of type T with the specified name.
// Panics if not found.
func ResolveNamed[T any](container *Container, name string) T {
	return resolve[T](container, name, true)
}

// TryResolve attempts to return the first registered instance of type T.
// Returns zero value if not found.
func TryResolve[T any](container *Container) T {
	return TryResolveNamed[T](container, "")
}

// TryResolveNamed returns the instance of type T with the specified name.
// Returns zero value if not found.
func TryResolveNamed[T any](container *Container, name string) T {
	return resolve[T](container, name, false)
}

func resolve[T any](container *Container, name string, required bool) T {
	typ := reflect.TypeFor[T]()

	if typ.AssignableTo(containerPtrType) {
		var val any = container
		return val.(T)
	}

	container = containerOrDefault(container)
	container.mu.RLock()
	defer container.mu.RUnlock()

	decl := container.resolve(typ, name)

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

// ResolveInjections returns an iterator over all registered injects in the container.
func ResolveInjections(container *Container) iter.Seq[InjectDeclaration] {
	container = containerOrDefault(container)
	return func(yield func(InjectDeclaration) bool) {
		container.mu.RLock()
		defer container.mu.RUnlock()

		for _, service := range container.injects {
			if !yield(service) {
				break
			}
		}
	}
}

// ResolveAll returns an iterator over registered injects in the container
// if the service's type is assignable to T (implements interface or same type).
func ResolveAll[T any](container *Container) []T {
	injects := ResolveInjections(container)
	var result []T
	typ := reflect.TypeFor[T]()
	for decl := range injects {
		if decl.Type().AssignableTo(typ) {
			result = append(result, decl.Value().(T))
		}
	}
	return result
}

// CleanInjections removes all service declarations that match the selector function.
// Do not use octo.* functions inside selector, this may cause deadlocks
func CleanInjections(container *Container, selector func(decl InjectDeclaration) bool) {
	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	injects := make([]InjectDeclaration, 0, len(container.injects))
	for _, decl := range container.injects {
		if !selector(decl) {
			injects = append(injects, decl)
		}
	}
	container.injects = injects
	container.resolveCache = nil
}
