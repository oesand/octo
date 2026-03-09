package validate

// Struct returns a validator for pointer-to-struct values that runs the
// provided field validators and aggregates their errors.
func Struct[Struct any](fields ...Validator[*Struct]) Validator[*Struct] {
	return &structValidator[Struct]{
		fields: fields,
	}
}

type structValidator[Struct any] struct {
	fields []Validator[*Struct]
}

func (validator *structValidator[Struct]) Validate(value *Struct) ValidationErrors {
	var errors []string
	for _, field := range validator.fields {
		errors = append(errors, field.Validate(value)...)
	}
	return errors
}
