package octo

import (
	"fmt"
	"iter"
	"reflect"

	"github.com/oesand/octo/internal/typing"
)

func resolve[T any](container *Container, name string) Declaration {
	if container.injects == nil {
		return nil
	}

	if name == "" {
		container.resolveCacheMu.RLock()
		if len(container.resolveCache) > 0 {
			if decl, ok := container.resolveCache[typing.Type[T]{}]; ok {
				return decl
			}
		}
		container.resolveCacheMu.RUnlock()
	}

	resolveType := reflect.TypeFor[T]()
	isInterface := resolveType.Kind() == reflect.Interface

	var decl Declaration
	if !isInterface {
		var typeKey typing.Type[T]
		if group, ok := container.injects[typeKey]; ok && len(group) > 0 {
			for _, inject := range group {
				if name != "" && inject.Name() != name {
					continue
				}

				decl = inject
				break
			}
		}
	} else {
		for groupType, group := range container.injects {
			if len(group) == 0 || !groupType.Type().AssignableTo(resolveType) {
				continue
			}

			for _, inject := range group {
				if name != "" && inject.Name() != name {
					continue
				}

				decl = inject
				break
			}

			if decl != nil {
				break
			}
		}
	}

	if name == "" && decl != nil {
		container.resolveCacheMu.Lock()
		if container.resolveCache == nil {
			container.resolveCache = make(map[typing.TypeKey]Declaration)
		}

		container.resolveCache[typing.Type[T]{}] = decl
		container.resolveCacheMu.Unlock()
	}

	return decl
}

func resolveValue[T any](container *Container, name string, required bool) (result T) {
	container = containerOrDefault(container)

	var t T
	switch any(t).(type) {
	case *Container:
		return any(container).(T)
	}

	container.mu.RLock()
	defer container.mu.RUnlock()

	decl := resolve[T](container, name)

	if decl != nil {
		if val := decl.Value(); val != nil {
			result = val.(T)
		}
	} else if required {
		panic(fmt.Sprintf("octo: fail to resolve type %s", reflect.TypeFor[T]().String()))
	}

	return
}

// Resolve returns the first registered instance of type T.
// Panics if not found.
func Resolve[T any](container *Container) T {
	return ResolveNamed[T](container, "")
}

// ResolveNamed returns the instance of type T with the specified name.
// Panics if not found.
func ResolveNamed[T any](container *Container, name string) T {
	return resolveValue[T](container, name, true)
}

// TryResolve attempts to return the first registered instance of type T.
// Returns zero value if not found.
func TryResolve[T any](container *Container) T {
	return TryResolveNamed[T](container, "")
}

// TryResolveNamed returns the instance of type T with the specified name.
// Returns zero value if not found.
func TryResolveNamed[T any](container *Container, name string) T {
	return resolveValue[T](container, name, false)
}

// ResolveInjections returns an iterator over all registered injects in the container.
func ResolveInjections(container *Container) iter.Seq[Declaration] {
	container = containerOrDefault(container)
	return func(yield func(Declaration) bool) {
		container.mu.RLock()
		defer container.mu.RUnlock()

		for _, group := range container.injects {
			for _, inject := range group {
				if !yield(inject) {
					return
				}
			}
		}
	}
}

// ResolveAll returns slice of registered injects in the container
// if the service's type is assignable to T (implements interface or same type).
func ResolveAll[T any](container *Container) []T {
	container = containerOrDefault(container)
	container.mu.RLock()
	defer container.mu.RUnlock()

	var result []T
	if container.injects == nil {
		return result
	}

	resolveType := reflect.TypeFor[T]()
	isInterface := resolveType.Kind() == reflect.Interface
	if !isInterface {
		if group, ok := container.injects[typing.Type[T]{}]; ok {
			for _, inject := range group {
				result = append(result, inject.Value().(T))
			}
		}
	} else {
		for groupType, group := range container.injects {
			if groupType.Type().AssignableTo(resolveType) {
				for _, inject := range group {
					result = append(result, inject.Value().(T))
				}
			}
		}
	}
	return result
}

// CleanInjections removes all inject declarations that match the selector function.
// Do not use octo.* functions inside selector, this may cause deadlocks
func CleanInjections(container *Container, selector func(decl Declaration) bool) {
	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	for groupType, group := range container.injects {
		injects := make([]Declaration, 0, len(group))
		for _, inject := range group {
			if !selector(inject) {
				injects = append(injects, inject)
			}
		}

		if len(injects) == 0 {
			delete(container.injects, groupType)
		} else {
			container.injects[groupType] = injects
		}
	}

	container.resolveCache = nil
}
