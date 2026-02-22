package generic

type Generic struct{}

type Linked[T any] struct{}

type Base[T any] struct{}

type Struct struct {
	Base[int]
}

func NewStruct(lnk *Linked[int]) *Struct {
	return new(Struct)
}
