// +build octogen

package foo

import (
	"github.com/oesand/octo/octogen"
)

func IncludeOne() {
	octogen.Inject[*Struct]()
	octogen.Inject[*Named]("key1")
	octogen.Inject[*Other]()
	octogen.Inject[*NewestStruct]()

	octogen.Inject(NewStruct)
	octogen.Inject(NewStct)
	octogen.Inject(NewStct, "key2")
}
