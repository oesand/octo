package parse

type InitType int

const (
	StructInit InitType = iota
	FuncInit
)

type FieldKind int

const (
	PointerField FieldKind = iota
	ArrayField
)

type ParsedFile struct {
	Depends []*Depend
}

type Depend struct {
	Name   string
	Kind   InitType
	Fields []*FieldInfo

	ContainerOption string
	KeyOption       string
}

type FieldInfo struct {
	KindType  FieldKind
	TypeAlias string

	KeyOption string
}
