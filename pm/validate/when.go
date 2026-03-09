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

func (cond *whenConditionValidator[T]) Validate(v T) ValidationErrors {
	if cond.condition(v) {
		var errors []string
		for _, validator := range cond.validators {
			errors = append(errors, validator.Validate(v)...)
		}
		return errors
	}
	return nil
}
