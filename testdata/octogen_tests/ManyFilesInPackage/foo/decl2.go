//go:build octogen

package foo

import (
	og "github.com/oesand/octo/octogen"
)

func IncludeTwo() {
	og.Inject[*Struct]()
	og.Inject[*Named]("key1")
	og.Inject[*Other]()
	og.Inject[*NewestStruct]()

	og.Inject(NewStruct)
	og.Inject(NewStct)
	og.Inject(NewStct, "key2")
}
