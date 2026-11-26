package ocp

import "github.com/oesand/octo"

func InjectNamed[T any](name string, provider octo.Provider[T]) Injector {
	return func(container *octo.Container) {
		octo.InjectNamed(container, name, provider)
	}
}
