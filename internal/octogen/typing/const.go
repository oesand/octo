package typing

type Operation int

const (
	CallOp Operation = iota
	DeclOp
)

type Kind int

const (
	PointerKind Kind = iota
	SliceKind
	MapKind
	NamedKind
)

type Context interface {
	ImportAlias(pkg string) string
}

type Renderer interface {
	Kind() Kind
	Child() Renderer
	Render(ctx Context, operation Operation) string
}
