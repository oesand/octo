package generic

type Generic struct{}

type Linked[T any] struct{}

type Super[T any] struct {
	Link *Linked[T]
}

type Base[T any] struct {
	Super[T]
	Link *Linked[T]
}

type Struct[T any, T1 any] struct {
	Base[T1]
	Link *Linked[T]
}
