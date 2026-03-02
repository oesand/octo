package generic

type Generic struct{}

type Linked[T any] struct{}

type Super[T any] struct {
	Link *Linked[T]
}

type Struct[T any] struct {
	Super[T]
	Link *Linked[T]
}

type EmbeddedStruct[T any, T1 any] struct {
	Struct[T1]
	Link *Linked[T]
}
