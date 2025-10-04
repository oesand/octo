package foo

type Inf interface{}

type NormalStruct struct{}

type StructWithInvalidField struct{
	Num int
	Str string
}

func NewStruct() NormalStruct {
	return &NormalStruct{}
}

func NewInf() Inf {
	return &NewestStruct{}
}

func FuncInvalidParam(num int) NormalStruct {
	return &NormalStruct{}
}

