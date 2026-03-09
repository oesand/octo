package validate

import (
	"fmt"

	"github.com/oesand/octo/pm"
)

func Field[Struct any, Field any](descriptor pm.FieldDescriptor[Struct, Field], validators ...Validator[Field]) Validator[*Struct] {
	return &fieldValidator[Struct, Field]{
		descriptor: descriptor,
		validators: validators,
	}
}

type fieldValidator[Struct any, Field any] struct {
	descriptor pm.FieldDescriptor[Struct, Field]
	validators []Validator[Field]
}

func (v *fieldValidator[Struct, Field]) Validate(parent *Struct) []string {
	var errors []string
	name := v.descriptor.Name
	value := v.descriptor.Value(parent)
	for _, validator := range v.validators {
		for _, err := range validator.Validate(value) {
			errors = append(errors, fmt.Sprintf("> '%s': %s", name, err))
		}
	}
	return errors
}
