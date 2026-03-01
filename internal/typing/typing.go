package typing

import "reflect"

type Type[T any] struct{}

func (Type[T]) Type() reflect.Type {
	return reflect.TypeFor[T]()
}

func (t Type[T]) AbsoluteName() string {
	return AbsoluteTypeName(t.Type())
}

type TypeKey interface {
	Type() reflect.Type
	AbsoluteName() string
}
