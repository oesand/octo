package octo

import (
	"reflect"
	"sync"

	"github.com/oesand/octo/internal/typing"
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
	injects map[typing.TypeKey][]Declaration

	resolveCacheMu sync.RWMutex
	resolveCache   map[typing.TypeKey]Declaration
}

func containerOrDefault(container *Container) *Container {
	if container == nil {
		return &DefaultContainer
	}
	return container
}

func injectLazy[T any](container *Container, name string, provider Provider[T]) {
	var injection Declaration = &lazyInjection[T]{
		container: container,
		name:      name,
		provider:  provider,
	}

	if container.injects == nil {
		container.injects = make(map[typing.TypeKey][]Declaration)
	}

	var key typing.Type[T]
	container.injects[key] = append(container.injects[key], injection)
}

type lazyInjection[T any] struct {
	container *Container
	name      string

	provider Provider[T]
	doInit   sync.Once
	value    T
}

func (c *lazyInjection[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

func (c *lazyInjection[T]) Name() string {
	return c.name
}

func (c *lazyInjection[T]) Value() any {
	c.doInit.Do(func() {
		c.value = c.provider(c.container)
	})

	return c.value
}

func injectValue[T any](container *Container, name string, value T) {
	var injection Declaration = &valueInjection[T]{
		name:  name,
		value: value,
	}

	if container.injects == nil {
		container.injects = make(map[typing.TypeKey][]Declaration)
	}

	var key typing.Type[T]
	container.injects[key] = append(container.injects[key], injection)
}

type valueInjection[T any] struct {
	name  string
	value T
}

func (c *valueInjection[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

func (c *valueInjection[T]) Name() string {
	return c.name
}

func (c *valueInjection[T]) Value() any {
	return c.value
}
