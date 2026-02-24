package fnc

type Iface interface {
	Do()
}

type Linked struct{}

type Struct struct{}

func (*Struct) Do() {

}

func NewStruct(lnk Linked) Struct {
	return Struct{}
}

func NewPtrStruct(lnk *Linked) *Struct {
	return &Struct{}
}

func NewIface(lnk *Linked) Iface {
	return &Struct{}
}
