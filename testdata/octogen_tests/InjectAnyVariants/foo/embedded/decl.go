package embedded

type Iface interface {
	Do()
}

type Linked struct{}

type Super struct {
	Link *Linked
}

type Base struct {
	Super
	Link *Linked
}

type Struct struct {
	Base
	Link *Linked
	If   Iface
}
