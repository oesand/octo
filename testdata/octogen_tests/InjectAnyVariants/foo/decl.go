//go:build octogen
// +build octogen

package foo

import (
	"github.com/oesand/octo/octogen"
	"github.com/oesand/octo/testdata/octogen_tests/InjectAnyVariants/foo/embedded"
	"github.com/oesand/octo/testdata/octogen_tests/InjectAnyVariants/foo/fnc"
	"github.com/oesand/octo/testdata/octogen_tests/InjectAnyVariants/foo/generic"
	"github.com/oesand/octo/testdata/octogen_tests/InjectAnyVariants/foo/stct"
)

func IncludeStruct() {
	octogen.Inject[*stct.Linked]()
	octogen.Inject[*stct.Named]("named")
	octogen.Inject[*stct.Struct]()
}

func IncludeFunc() {
	octogen.Inject[*fnc.Linked]()

	octogen.Inject(fnc.NewPtrStruct)
	octogen.Inject(fnc.NewStruct)
	octogen.Inject(fnc.NewIface)
	octogen.Inject(fnc.NewIface, "named")
}

func IncludeEmbedded() {
	octogen.Inject[*embedded.Linked]()
	octogen.Inject[*embedded.Struct]()
}

func IncludeGeneric() {
	octogen.Inject[*generic.Struct[int, *generic.Generic]]()
	octogen.Inject(generic.NewStruct)
}
