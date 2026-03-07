package cond

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

// Condition is a condition that can be used to validate a value.
type Condition[T any] interface {
	Validate(T) error
}
