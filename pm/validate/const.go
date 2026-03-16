package validate

import (
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

type Errors []string

func (errs Errors) IsValid() bool {
	return len(errs) == 0
}

func (errs Errors) Error() string {
	return strings.Join(errs, "\n")
}

type Validator[T any] interface {
	Validate(T) Errors
}

type FuncValidator[T any] func(T) Errors

func (f FuncValidator[T]) Validate(v T) Errors {
	return f(v)
}
