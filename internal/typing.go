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

type ShadowType interface {
	Type() reflect.Type
	Zero() any
	Real() bool
}
