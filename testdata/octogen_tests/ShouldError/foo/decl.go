//go:build octogen
// +build octogen

package foo

import (
	"github.com/oesand/octo/octogen"
)

func IncludeAny() {
	octogen.Inject() // Empty param

	octogen.Inject[Inf]()
	octogen.Inject(NewInf)
	octogen.Inject(NormalStruct)
	octogen.Inject(NormalStruct{})
	octogen.Inject[NewStruct]()
	octogen.Inject[NewStruct()]()

	octogen.Inject(int)
	octogen.Inject[int]()
	octogen.Inject(NewInf())
	octogen.Inject[NormalStruct](1)
	octogen.Inject[NormalStruct](true)

	octogen.Inject[NormalStruct]("key1", "key2")
	octogen.Inject[NormalStruct]("key1", 1)
	octogen.Inject(NewStruct, "key1", "key2")
	octogen.Inject(NewStruct, "key1", 1)
}

func IncludeWithGenerics[T any]() {
}

func IncludeWithArguments(arg0 string) {
}

func IncludeWithReturn() string {
}
