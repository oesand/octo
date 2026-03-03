package octo

import "reflect"

// Provider is a function that returns an instance of type T given a container.
// It is used for lazy resolution of dependencies.
type Provider[T any] func(*Container) T

// Declaration represents info of a registered injection in the container.
type Declaration interface {
	// Name returns the optional name of the service.
	Name() string

	// Type returns the [reflect.Type] of the service.
	// Not recommended, use if generics are not applicable in your case
	Type() reflect.Type

	// Value returns the concrete instance of the injection, if available.
	Value() any
}

// OfType checks if a ServiceDeclaration is compatible with type T.
// Returns true if the service's type is assignable to T (implements interface or same type).
func OfType[T any](decl Declaration) bool {
	if _, ok := decl.(*lazyInjection[T]); ok {
		return true
	}

	if _, ok := decl.(*valueInjection[T]); ok {
		return true
	}

	expectType := reflect.TypeFor[T]()
	return expectType.Kind() == reflect.Interface &&
		decl.Type().AssignableTo(expectType)
}
