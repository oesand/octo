package internal

import (
	"reflect"
)

type Type[T any] struct{}

func (Type[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

func (Type[T]) Zero() any {
	var zero T
	return zero
}

func (t Type[T]) Real() bool {
	return t.Zero() != nil
}

func (t Type[T]) ConvertibleFrom(from ShadowType) bool {
	if t.Real() {
		_, ok := from.Zero().(T)
		return ok
	}
	return from.Type().AssignableTo(t.Type())
}

type ShadowType interface {
	Type() reflect.Type
	Zero() any
	Real() bool
	ConvertibleFrom(ShadowType) bool
}
