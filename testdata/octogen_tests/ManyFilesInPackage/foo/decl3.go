// +build octogen

package foo

import (
	octog "github.com/oesand/octo/octogen"
)

func IncludeThree() {
	octog.Inject[*Struct]()
	octog.Inject[*Named]("key1")
	octog.Inject[*Other]()
	octog.Inject[*NewestStruct]()

	octog.Inject(NewStruct)
	octog.Inject(NewStct)
	octog.Inject(NewStct, "key2")
}
