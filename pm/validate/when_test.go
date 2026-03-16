package validate_test

import (
	"testing"

	"github.com/oesand/octo/pm"
	"github.com/oesand/octo/pm/validate"
)

func TestWhenConditionRunsValidatorsOnlyWhenTrue(t *testing.T) {
	v := validate.When(func(i int) bool { return i%2 == 0 }, validate.Min(10))

	// condition true, validator runs and should fail for 8
	if res := v.Validate(8); res.IsValid() {
		t.Fatalf("expected invalid for 8")
	} else if err := res.Error(); err != "must be greater than or equal to 10" {
		t.Fatalf("unexpected error %s", err)
	}

	// condition false, validator should not run
	if res := v.Validate(9); !res.IsValid() {
		t.Fatalf("expected valid for 9, got %v", res)
	}
}

func TestWhenNotNil_SkipsNilAndValidatesPointer(t *testing.T) {
	v := validate.WhenNotNil(validate.FuncValidator[*int](func(p *int) validate.Errors {
		if *p < 1 {
			return validate.Errors{"must be at least 1"}
		}
		return nil
	}))

	if res := v.Validate(nil); !res.IsValid() {
		t.Fatalf("expected valid for nil input, got %v", res)
	}

	zero := 0
	if res := v.Validate(&zero); res.IsValid() {
		t.Fatalf("expected invalid for 0")
	} else if err := res.Error(); err != "must be at least 1" {
		t.Fatalf("unexpected error %s", err)
	}

	one := 1
	if res := v.Validate(&one); !res.IsValid() {
		t.Fatalf("expected valid for 1, got %v", res)
	}
}

func TestWhenNotNil_WithNestedStructAndField(t *testing.T) {
	type Inner struct{ N int }
	type Parent struct{ Child *Inner }

	childDesc := pm.FieldDescriptor[Parent, *Inner]{
		Name:  "Child",
		Value: func(p *Parent) *Inner { return p.Child },
	}

	nDesc := pm.FieldDescriptor[Inner, int]{
		Name:  "N",
		Value: func(i *Inner) int { return i.N },
	}

	innerValidator := validate.WhenNotNil(validate.Struct(
		validate.Field(nDesc, validate.Min(5)),
	))

	fieldValidator := validate.Field(childDesc, innerValidator)

	// child nil -> inner validators skipped
	if res := fieldValidator.Validate(&Parent{}); !res.IsValid() {
		t.Fatalf("expected valid when child is nil, got %v", res)
	}

	// child present but N < 5 -> prefixed error
	pLow := Parent{Child: &Inner{N: 2}}
	if res := fieldValidator.Validate(&pLow); res.IsValid() {
		t.Fatalf("expected invalid when Child.N < 5")
	} else if err := res.Error(); err != "> 'Child': > 'N': must be greater than or equal to 5" {
		t.Fatalf("unexpected error %s", err)
	}

	// child present and N >= 5 -> valid
	pOk := Parent{Child: &Inner{N: 6}}
	if res := fieldValidator.Validate(&pOk); !res.IsValid() {
		t.Fatalf("expected valid when Child.N >= 5, got %v", res)
	}
}
