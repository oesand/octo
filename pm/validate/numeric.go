package validate

import (
	"fmt"
)

// Min creates a condition that validates a numeric value is greater than or equal to the minimum.
func Min[Value NumericTypes](min Value) Validator[Value] {
	return &numericMinValidator[Value]{min: min}
}

type numericMinValidator[Value NumericTypes] struct {
	min Value
}

func (validator *numericMinValidator[Value]) Validate(value Value) Errors {
	if value < validator.min {
		return []string{fmt.Sprintf("must be greater than or equal to %v", validator.min)}
	}
	return nil
}

// Max creates a condition that validates a numeric value is less than or equal to the maximum.
func Max[Value NumericTypes](max Value) Validator[Value] {
	return &numericMaxValidator[Value]{max: max}
}

type numericMaxValidator[Value NumericTypes] struct {
	max Value
}

func (validator *numericMaxValidator[Value]) Validate(value Value) Errors {
	if value > validator.max {
		return []string{fmt.Sprintf("must be less than or equal to %v", validator.max)}
	}
	return nil
}
