package validate

// Struct returns a validator for pointer-to-struct values that runs the
// provided field validators and aggregates their errors.
func Struct[Struct any](validators ...Validator[*Struct]) Validator[*Struct] {
	return &structValidator[Struct]{
		validators: validators,
	}
}

type structValidator[Struct any] struct {
	validators []Validator[*Struct]
}

func (validator *structValidator[Struct]) Validate(value *Struct) Errors {
	var errors []string
	for _, v := range validator.validators {
		errors = append(errors, v.Validate(value)...)
	}
	return errors
}
