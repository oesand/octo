package generic

func NewStruct(lnk *Linked[int]) *EmbeddedStruct[int, *Generic] {
	return new(EmbeddedStruct[int, *Generic])
}
