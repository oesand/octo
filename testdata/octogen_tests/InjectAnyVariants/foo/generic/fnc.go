package generic

func NewStruct(lnk *Linked[int]) *Struct[int, *Generic] {
	return new(Struct[int, *Generic])
}
