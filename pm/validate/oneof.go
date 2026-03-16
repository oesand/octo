package validate

import (
	"fmt"

	"github.com/oesand/octo/pm"
)

// OneOf returns a `Validator` that checks whether a value
// is one of the provided `values`.
func OneOf[Element comparable](values ...Element) Validator[Element] {
	var valuesString string
	for i, value := range values {
		if i > 0 {
			valuesString += fmt.Sprintf(", %v", value)
		} else {
			valuesString += fmt.Sprint(value)
		}
	}

	return &oneOfValidator[Element]{
		values:       pm.SetOf(values...),
		valuesString: valuesString,
	}
}

type oneOfValidator[Element comparable] struct {
	values       pm.Set[Element]
	valuesString string
}

func (validator *oneOfValidator[Element]) Validate(value Element) Errors {
	if !validator.values.Has(value) {
		return []string{fmt.Sprintf("must be in %s", validator.valuesString)}
	}
	return nil
}
