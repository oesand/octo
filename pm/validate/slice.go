package validate

import (
	"fmt"
)

func Slice[Element any](validators ...Validator[Element]) Validator[[]Element] {
	return &sliceValidator[Element]{
		validators: validators,
	}
}

type sliceValidator[Element any] struct {
	validators []Validator[Element]
}

func (v *sliceValidator[Element]) Validate(slice []Element) []string {
	var errors []string
	for i, el := range slice {
		for _, validator := range v.validators {
			for _, err := range validator.Validate(el) {
				errors = append(errors, fmt.Sprintf("> [%d]: %s", i, err))
			}
		}
		if len(errors) > 0 {
			break
		}
	}

	return errors
}

func MinCount[Element any](minLength int) Validator[[]Element] {
	return &sliceMinValidator[Element]{minLength: minLength}
}

type sliceMinValidator[Element any] struct {
	minLength int
}

func (c *sliceMinValidator[Element]) Validate(slice []Element) []string {
	if len(slice) < c.minLength {
		return []string{fmt.Sprintf("count must be greater than or equal to %v", c.minLength)}
	}
	return nil
}

func MaxCount[Element any](maxLength int) Validator[[]Element] {
	return &sliceMaxValidator[Element]{maxLength: maxLength}
}

type sliceMaxValidator[Element any] struct {
	maxLength int
}

func (c *sliceMaxValidator[Element]) Validate(slice []Element) []string {
	if len(slice) > c.maxLength {
		return []string{fmt.Sprintf("count must be less than or equal to %v", c.maxLength)}
	}
	return nil
}
