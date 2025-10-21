//go:build octogen
// +build octogen

package foo

import (
	"github.com/oesand/octo/octogen"
)

func IncludeWithGenerics[T any]() {
}

func IncludeWithArguments(arg0 string) {
}

func IncludeWithReturn() string {
}

func IncludeAny() {
	octogen.Inject() // Empty param

	octogen.Inject[Inf]()
	octogen.Inject(NewGeneric)
	octogen.Inject(NormalStruct)
	octogen.Inject(NormalStruct{})
	octogen.Inject[NewStruct]()
	octogen.Inject[NewStruct()]()

	octogen.Inject(int)
	octogen.Inject[int]()
	octogen.Inject(NewStruct())
	octogen.Inject[NormalStruct](1)
	octogen.Inject[NormalStruct](true)

	octogen.Inject[NormalStruct]("key1", "key2")
	octogen.Inject[NormalStruct]("key1", 1)
	octogen.Inject(NewStruct, "key1", "key2")
	octogen.Inject(NewStruct, "key1", 1)
	octogen.Inject[GenericStruct]()
	octogen.Inject[GenericStruct[string]]()

	octogen.Inject(FuncInvalidReturn)
	octogen.Inject(FuncInvalidReturnCount)

	octogen.Inject[[]NormalStruct]()
	octogen.Inject[[]*NormalStruct]()
	octogen.Inject(FuncReturnPtrInf)
	octogen.Inject(FuncReturnSliceInf)
	octogen.Inject(FuncReturnSliceStct)
	octogen.Inject(FuncReturnSlicePtrStct)

	octogen.Inject[StructWithInvalidRef]()
	octogen.Inject[*StructWithInvalidRef]()
	octogen.Inject(FunctionWithInvalidReference)
}
