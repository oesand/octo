package validate_test

import (
	"testing"

	"github.com/oesand/octo/pm/validate"
)

func TestMinInt(t *testing.T) {
	v := validate.Min(10)

	if res := v.Validate(9); res.IsValid() {
		t.Fatalf("expected invalid for 9")
	} else if err := res.Error(); err != "must be greater than or equal to 10" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate(10); !res.IsValid() {
		t.Fatalf("expected valid for 10, got %v", res)
	}

	if res := v.Validate(11); !res.IsValid() {
		t.Fatalf("expected valid for 11, got %v", res)
	}
}

func TestMaxInt64(t *testing.T) {
	v := validate.Max(int64(100))

	if res := v.Validate(int64(101)); res.IsValid() {
		t.Fatalf("expected invalid for 101")
	} else if err := res.Error(); err != "must be less than or equal to 100" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate(int64(100)); !res.IsValid() {
		t.Fatalf("expected valid for 100, got %v", res)
	}

	if res := v.Validate(int64(99)); !res.IsValid() {
		t.Fatalf("expected valid for 99, got %v", res)
	}
}

func TestMinFloat(t *testing.T) {
	v := validate.Min(3.14)

	if res := v.Validate(3.13); res.IsValid() {
		t.Fatalf("expected invalid for 3.13")
	} else if err := res.Error(); err != "must be greater than or equal to 3.14" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate(3.14); !res.IsValid() {
		t.Fatalf("expected valid for 3.14, got %v", res)
	}

	if res := v.Validate(3.15); !res.IsValid() {
		t.Fatalf("expected valid for 3.15, got %v", res)
	}
}

func TestMaxUint(t *testing.T) {
	v := validate.Max(uint(50))

	if res := v.Validate(uint(51)); res.IsValid() {
		t.Fatalf("expected invalid for 51")
	} else if err := res.Error(); err != "must be less than or equal to 50" {
		t.Fatalf("unexpected error %s", err)
	}

	if res := v.Validate(uint(50)); !res.IsValid() {
		t.Fatalf("expected valid for 50, got %v", res)
	}

	if res := v.Validate(uint(49)); !res.IsValid() {
		t.Fatalf("expected valid for 49, got %v", res)
	}
}
