package pm

import (
	"reflect"
	"sort"
	"testing"
)

func TestSet_Add_Has_Del(t *testing.T) {
	var s Set[string]

	// Initially empty
	if s.Has("a") {
		t.Fatal("expected empty set to not have 'a'")
	}

	// Add elements
	s.Add("a", "b", "c")
	for _, k := range []string{"a", "b", "c"} {
		if !s.Has(k) {
			t.Fatalf("expected set to have key %q", k)
		}
	}

	// Delete an element
	s.Del("b")
	if s.Has("b") {
		t.Fatal("expected 'b' to be deleted")
	}

	// Delete nonexistent key should not panic
	s.Del("zzz")

	// Re-add after delete
	s.Add("b")
	if !s.Has("b") {
		t.Fatal("expected 'b' to be re-added")
	}
}

func TestSet_CopyFrom(t *testing.T) {
	src := Set[int]{1: {}, 2: {}, 3: {}}
	var dst Set[int]

	dst.CopyFrom(src)

	for k := range src {
		if !dst.Has(k) {
			t.Fatalf("expected dst to have copied key %v", k)
		}
	}

	// Modifying src should not affect dst
	delete(src, 1)
	if !dst.Has(1) {
		t.Fatal("expected dst to remain unchanged after src modified")
	}
}

func TestSet_Values(t *testing.T) {
	s := Set[int]{3: {}, 1: {}, 2: {}}
	values := s.Values()
	sort.Ints(values)

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(values, expected) {
		t.Fatalf("expected %v, got %v", expected, values)
	}
}

func TestSet_NilBehavior(t *testing.T) {
	var s Set[string]

	// Has should return false on nil set
	if s.Has("x") {
		t.Fatal("expected Has to be false for nil set")
	}

	// Del should not panic
	s.Del("x")

	// Values should return nil
	if v := s.Values(); v != nil {
		t.Fatalf("expected nil Values for nil set, got %v", v)
	}

	// Add should auto-init
	s.Add("x")
	if !s.Has("x") {
		t.Fatal("expected Add to initialize and add element")
	}
}
