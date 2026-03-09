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
