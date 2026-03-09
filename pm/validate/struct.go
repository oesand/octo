package validate

func Struct[T any](fields ...Validator[*T]) Validator[*T] {
	return &structValidator[T]{
		fields: fields,
	}
}

type structValidator[T any] struct {
	fields []Validator[*T]
}

func (s *structValidator[T]) Validate(value *T) []string {
	var errors []string
	for _, field := range s.fields {
		errors = append(errors, field.Validate(value)...)
	}
	return errors
}
