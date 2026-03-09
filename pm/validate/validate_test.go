package validate_test

import (
	"testing"

	"github.com/oesand/octo/pm"
	"github.com/oesand/octo/pm/validate"
)

func TestValidate(t *testing.T) {
	otherValidator := validate.Struct(
		validate.Field(OtherFields.Name, validate.MinRunes(10)),
	)

	structValidator := validate.Struct(
		validate.Field(StructFields.Name, validate.MaxRunes(20)),
		validate.Field(StructFields.Age, validate.Min(20)),
		validate.Field(StructFields.Other, otherValidator),
		validate.Field(StructFields.Others, validate.Slice(otherValidator)),
		validate.When(func(t *Struct) bool {
			return t.Name == "123"
		},
			validate.Field(StructFields.Age, validate.Min(20)),
		),
	)

	value := &Struct{
		Name: "123",
	}

	result := structValidator.Validate(value)

	t.Log(result)
}

type Other struct {
	Name string
}

type Struct struct {
	Name   string
	Age    int
	Other  *Other
	Others []*Other
}

var OtherFields = struct {
	Name pm.FieldDescriptor[Other, string]
}{
	Name: pm.FieldDescriptor[Other, string]{
		Name: "Name",
		Value: func(s *Other) string {
			return s.Name
		},
	},
}

var StructFields = struct {
	Age    pm.FieldDescriptor[Struct, int]
	Name   pm.FieldDescriptor[Struct, string]
	Other  pm.FieldDescriptor[Struct, *Other]
	Others pm.FieldDescriptor[Struct, []*Other]
}{
	Age: pm.FieldDescriptor[Struct, int]{
		Name: "Age",
		Value: func(s *Struct) int {
			return s.Age
		},
	},
	Name: pm.FieldDescriptor[Struct, string]{
		Name: "Name",
		Value: func(s *Struct) string {
			return s.Name
		},
	},
	Other: pm.FieldDescriptor[Struct, *Other]{
		Name: "Other",
		Value: func(s *Struct) *Other {
			return s.Other
		},
	},
	Others: pm.FieldDescriptor[Struct, []*Other]{
		Name: "Others",
		Value: func(s *Struct) []*Other {
			return s.Others
		},
	},
}
