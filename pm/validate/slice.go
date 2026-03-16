package validate

import (
	"fmt"
)

// Slice returns a validator that applies the provided element validators to
// each element of a slice. If any element produces validation errors the
// returned result will contain those errors prefixed with the element index.
func Slice[Element any](validators ...Validator[Element]) Validator[[]Element] {
	return &sliceValidator[Element]{
		validators: validators,
	}
}

type sliceValidator[Element any] struct {
	validators []Validator[Element]
}

func (validator *sliceValidator[Element]) Validate(slice []Element) Errors {
	var errors []string
	for i, el := range slice {
		for _, validator := range validator.validators {
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

// MinCount returns a validator that ensures a slice has at least `minLength`
// elements.
func MinCount[Element any](minLength int) Validator[[]Element] {
	return &sliceMinValidator[Element]{minLength: minLength}
}

type sliceMinValidator[Element any] struct {
	minLength int
}

func (validator *sliceMinValidator[Element]) Validate(slice []Element) Errors {
	if len(slice) < validator.minLength {
		return []string{fmt.Sprintf("count must be greater than or equal to %v", validator.minLength)}
	}
	return nil
}

// MaxCount returns a validator that ensures a slice has at most `maxLength`
// elements.
func MaxCount[Element any](maxLength int) Validator[[]Element] {
	return &sliceMaxValidator[Element]{maxLength: maxLength}
}

type sliceMaxValidator[Element any] struct {
	maxLength int
}

func (validator *sliceMaxValidator[Element]) Validate(slice []Element) Errors {
	if len(slice) > validator.maxLength {
		return []string{fmt.Sprintf("count must be less than or equal to %v", validator.maxLength)}
	}
	return nil
}
