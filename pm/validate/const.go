package validate

import (
	"errors"
	"strings"
)

// NumericTypes represents types that support comparison operators
type NumericTypes interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

// BasicTypes represents types that are basic types
type BasicTypes interface {
	~string | ~bool | NumericTypes
}

type Validator[T any] interface {
	Validate(T) []string
}

type FuncValidator[T any] func(T) []string

func (f FuncValidator[T]) Validate(v T) []string {
	return f(v)
}

func CompactValidate[T any](validator Validator[T], value T) error {
	errs := validator.Validate(value)
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "\n"))
}
