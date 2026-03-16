package validate_test

import (
	"regexp"
	"testing"

	"github.com/oesand/octo/pm/validate"
)

func TestRegex(t *testing.T) {
	v := validate.Regex(regexp.MustCompile(`^abc$`))

	if res := v.Validate("abcd"); res.IsValid() {
		t.Fatalf("expected invalid for \"abcd\"")
	} else if err := res.Error(); err != "mismatch expected pattern" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate("abc"); !res.IsValid() {
		t.Fatalf("expected valid for \"abc\", got %v", res)
	}
}

func TestRunesExactly(t *testing.T) {
	v := validate.RunesExactly(3)

	if res := v.Validate("ab"); res.IsValid() {
		t.Fatalf("expected invalid for \"ab\"")
	} else if err := res.Error(); err != "must have exactly 3 characters" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate("abc"); !res.IsValid() {
		t.Fatalf("expected valid for \"abc\", got %v", res)
	}
}

func TestMinRunes(t *testing.T) {
	v := validate.MinRunes(2)

	if res := v.Validate("a"); res.IsValid() {
		t.Fatalf("expected invalid for \"a\"")
	} else if err := res.Error(); err != "must have at least 2 characters" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate("ab"); !res.IsValid() {
		t.Fatalf("expected valid for \"ab\", got %v", res)
	}
}

func TestMaxRunes(t *testing.T) {
	v := validate.MaxRunes(2)

	if res := v.Validate("abc"); res.IsValid() {
		t.Fatalf("expected invalid for \"abc\"")
	} else if err := res.Error(); err != "must have at most 2 characters" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate("ab"); !res.IsValid() {
		t.Fatalf("expected valid for \"ab\", got %v", res)
	}
}
