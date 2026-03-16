package validate

import "testing"

func TestOneOf_Ints(t *testing.T) {
	v := OneOf[int](1, 2, 3)

	if errs := v.Validate(2); !errs.IsValid() {
		t.Fatalf("expected value 2 to be valid, got errors: %v", errs)
	}

	if errs := v.Validate(4); errs.IsValid() {
		t.Fatalf("expected value 4 to be invalid")
	} else if err := errs.Error(); err != "must be in 1, 2, 3" {
		t.Fatalf("unexpected error %s", err)
	}
}

func TestOneOf_Strings(t *testing.T) {
	v := OneOf[string]("a", "b")

	if errs := v.Validate("a"); !errs.IsValid() {
		t.Fatalf("expected value 'a' to be valid, got errors: %v", errs)
	}

	if errs := v.Validate("z"); errs.IsValid() {
		t.Fatalf("expected value 'z' to be invalid")
	} else if err := errs.Error(); err != "must be in a, b" {
		t.Fatalf("unexpected error %s", err)
	}
}
