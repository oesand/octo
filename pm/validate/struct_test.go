package validate_test

import (
	"testing"

	"github.com/oesand/octo/pm"
	"github.com/oesand/octo/pm/validate"
)

func TestStructAggregatesFieldErrors(t *testing.T) {
	type Parent struct {
		Age  int
		Name string
	}

	ageDesc := pm.FieldDescriptor[Parent, int]{
		Name:  "Age",
		Value: func(p *Parent) int { return p.Age },
	}

	nameDesc := pm.FieldDescriptor[Parent, string]{
		Name:  "Name",
		Value: func(p *Parent) string { return p.Name },
	}

	sv := validate.Struct(
		validate.Field(ageDesc, validate.Min(10)),
		validate.Field(nameDesc, validate.MinRunes(2)),
	)

	res := sv.Validate(&Parent{Age: 5, Name: "a"})
	if res.IsValid() {
		t.Fatalf("expected invalid")
	}

	want := "> 'Age': must be greater than or equal to 10\n> 'Name': must have at least 2 characters"
	if err := res.Error(); err != want {
		t.Fatalf("unexpected error %s", err)
	}
}
