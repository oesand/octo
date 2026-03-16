package validate_test

import (
	"testing"

	"github.com/oesand/octo/pm/validate"
)

func TestMinCount(t *testing.T) {
	v := validate.MinCount[string](2)

	if res := v.Validate([]string{"a"}); res.IsValid() {
		t.Fatalf("expected invalid for count 1")
	} else if err := res.Error(); err != "count must be greater than or equal to 2" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate([]string{"a", "b"}); !res.IsValid() {
		t.Fatalf("expected valid for count 2, got %v", res)
	}
}

func TestMaxCount(t *testing.T) {
	v := validate.MaxCount[string](2)

	if res := v.Validate([]string{"a", "b", "c"}); res.IsValid() {
		t.Fatalf("expected invalid for count 3")
	} else if err := res.Error(); err != "count must be less than or equal to 2" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate([]string{"a", "b"}); !res.IsValid() {
		t.Fatalf("expected valid for count 2, got %v", res)
	}
}

func TestSlice_ElementValidatorStopsAtFirstError(t *testing.T) {
	// element validator returns error for value "bad"
	elemValidator := validate.FuncValidator[string](func(s string) validate.Errors {
		if s == "bad" {
			return []string{"element invalid"}
		}
		return nil
	})

	v := validate.Slice(elemValidator)

	res := v.Validate([]string{"ok", "bad", "x"})
	if res.IsValid() {
		t.Fatalf("expected invalid for element 'bad'")
	}

	if got := res.Error(); got != "> [1]: element invalid" {
		t.Fatalf("unexpected error; want %q, got %q", "> [1]: element invalid", got)
	}
}

func TestSlice_ElementMultipleErrorsAreAggregated(t *testing.T) {
	// two validators that both trigger for the same element
	v1 := validate.FuncValidator[string](func(s string) validate.Errors {
		if s == "x" {
			return []string{"err1"}
		}
		return nil
	})
	v2 := validate.FuncValidator[string](func(s string) validate.Errors {
		if s == "x" {
			return []string{"err2"}
		}
		return nil
	})

	v := validate.Slice(v1, v2)

	res := v.Validate([]string{"x"})
	if res.IsValid() {
		t.Fatalf("expected invalid for element 'x'")
	}

	want := "> [0]: err1\n> [0]: err2"
	if got := res.Error(); got != want {
		t.Fatalf("unexpected aggregated errors; want %q, got %q", want, got)
	}
}
