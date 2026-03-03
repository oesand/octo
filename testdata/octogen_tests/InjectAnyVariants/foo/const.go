package foo

import (
	"github.com/oesand/octo/testdata/octogen_tests/InjectAnyVariants/foo/embedded"
	"github.com/oesand/octo/testdata/octogen_tests/InjectAnyVariants/foo/fnc"
)

type Named struct{}
type Linked struct{}

type Struct struct {
	Linked       *Linked
	Named        *Named `key:"named"`
	NestedFunc   *fnc.Struct
	NestedStruct *embedded.Struct
}
