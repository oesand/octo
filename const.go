package octo

import "reflect"

// Provider is a function that returns an instance of type T given a container.
// It is used for lazy resolution of dependencies.
type Provider[T any] func(*Container) T

// ServiceDeclaration represents a registered service in the container.
// It exposes the service's name, its concrete value, and its type.
type ServiceDeclaration interface {
	// Name returns the optional name of the service.
	Name() string

	// Value returns the concrete instance of the service, if available.
	Value() any

	// Type returns the [reflect.Type] of the service.
	Type() reflect.Type
}

// DeclOfType checks if a ServiceDeclaration is compatible with type T.
// Returns true if the service's type is assignable to T (implements interface or same type).
func DeclOfType[T any](decl ServiceDeclaration) bool {
	return decl.Type().AssignableTo(reflect.TypeFor[T]())
}
