//go:build octogen
// +build octogen

package foo

import (
	"github.com/oesand/octo/octogen"
)

func IncludeAny() {
	octogen.Inject[*Struct]()
	octogen.Inject[*Named]("key1")
	octogen.Inject[*Other]()
	octogen.Inject[*NewestStruct]()

	octogen.Inject(NewStruct)
	octogen.Inject(NewStct)
}
