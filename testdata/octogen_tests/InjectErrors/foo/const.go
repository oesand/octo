package foo

type Inf interface{}

type NormalStruct struct{}

type GenericStruct[T any] struct{}

type StructWithInvalidField struct{
	Num int
	Str string
}

type StructWithInvalidRef struct{
	Ref InvalidRef
}

func NewStruct() NormalStruct {
	return &NormalStruct{}
}

func FuncInvalidParam(num int) NormalStruct {
	return &NormalStruct{}
}

func NewGeneric[T any]() *NormalStruct {
	return &NormalStruct{}
}

func FuncInvalidReturn() string {
	return ""
}

func FuncInvalidReturnCount() (*NormalStruct, error) {
	return NormalStruct{}, nil
}

func FuncReturnPtrInf() *Inf {
	return &NormalStruct{}
}

func FuncReturnSliceInf() []Inf {
	return &NormalStruct{}
}

func FuncReturnSliceStct() []NormalStruct {
	return []NormalStruct{}
}

func FuncReturnSlicePtrStct() []*NormalStruct {
	return nil
}

func FunctionWithInvalidReference(ref InvalidRef) *NormalStruct {}