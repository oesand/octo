//go:build octogen
// +build octogen

package foo

import (
	"github.com/oesand/octo/internal/octogen_tests/NestedAnyVariants/foo/nested"
	"github.com/oesand/octo/internal/octogen_tests/NestedAnyVariants/foo/nested/inner"
	"github.com/oesand/octo/octogen"
)

func IncludeAny() {
	octogen.Inject[*inner.Struct]()
	octogen.Inject[*inner.Named]("key1")
	octogen.Inject[*nested.Other]()
	octogen.Inject[*nested.NewestStruct]()

	octogen.Inject(nested.NewStruct)
	octogen.Inject(nested.NewStct)
}
