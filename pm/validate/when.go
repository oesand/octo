package validate

func When[T any](condition func(T) bool, validators ...Validator[T]) Validator[T] {
	return &whenCondition[T]{
		condition:  condition,
		validators: validators,
	}
}

type whenCondition[T any] struct {
	condition  func(T) bool
	validators []Validator[T]
}

func (w *whenCondition[T]) Validate(v T) []string {
	if w.condition(v) {
		var errors []string
		for _, validator := range w.validators {
			errors = append(errors, validator.Validate(v)...)
		}
		return errors
	}
	return nil
}
