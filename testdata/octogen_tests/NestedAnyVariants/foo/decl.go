//go:build octogen
// +build octogen

package foo

import (
	"github.com/oesand/octo/octogen"
	"github.com/oesand/octo/testdata/octogen_tests/NestedAnyVariants/foo/nested"
	"github.com/oesand/octo/testdata/octogen_tests/NestedAnyVariants/foo/nested/inner"
)

func IncludeAny() {
	octogen.Inject[*inner.Struct]()
	octogen.Inject[*inner.Named]("key1")
	octogen.Inject[*nested.Other]()
	octogen.Inject[*nested.NewestStruct]()

	octogen.Inject(nested.NewStruct)
	octogen.Inject(nested.NewStct)
}
