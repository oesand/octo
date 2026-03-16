package validate

// When returns a conditional validator that runs the provided validators only
// when `condition` evaluates to true for the validated value.
func When[T any](condition func(T) bool, validators ...Validator[T]) Validator[T] {
	return &whenConditionValidator[T]{
		condition:  condition,
		validators: validators,
	}
}

type whenConditionValidator[T any] struct {
	condition  func(T) bool
	validators []Validator[T]
}

func (cond *whenConditionValidator[T]) Validate(v T) Errors {
	if cond.condition(v) {
		var errors []string
		for _, validator := range cond.validators {
			errors = append(errors, validator.Validate(v)...)
		}
		return errors
	}
	return nil
}

// WhenNotNil returns a validator for pointer-to-struct values that runs
// the provided validators only when the pointer is non-nil. This is a
// convenience wrapper around `When` that checks `v != nil` before
// executing the nested validators, preventing nil dereferences in
// validators that assume a non-nil receiver.
func WhenNotNil[Struct any](validators ...Validator[*Struct]) Validator[*Struct] {
	return When(func(v *Struct) bool {
		return v != nil
	}, validators...)
}
