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

func (c *numericMinValidator[Value]) Validate(value Value) []string {
	if value < c.min {
		return []string{fmt.Sprintf("must be greater than or equal to %v", c.min)}
	}
	return nil
}

// Max creates a condition that validates a numeric value is less than or equal to the maximum.
func Max[V NumericTypes](max V) Validator[V] {
	return &validateMax[V]{max: max}
}

type validateMax[V NumericTypes] struct {
	max V
}

func (c *validateMax[V]) Validate(value V) []string {
	if value > c.max {
		return []string{fmt.Sprintf("must be less than or equal to %v", c.max)}
	}
	return nil
}
