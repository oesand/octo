package validate

import (
	"strings"
)

type ValidationErrors []string

func (errs ValidationErrors) IsValid() bool {
	return len(errs) == 0
}

func (errs ValidationErrors) Error() string {
	return strings.Join(errs, "\n")
}

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
	Validate(T) ValidationErrors
}

type FuncValidator[T any] func(T) ValidationErrors

func (f FuncValidator[T]) Validate(v T) ValidationErrors {
	return f(v)
}
