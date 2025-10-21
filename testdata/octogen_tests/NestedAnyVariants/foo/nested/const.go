package nested

import (
	"github.com/oesand/octo/testdata/octogen_tests/NestedAnyVariants/foo/nested/inner"
	"github.com/oesand/octo/mch"
	"github.com/oesand/octo/mediatr/mdtreflect"
	"net"
)

type Other struct {
	Nm    *inner.Named `key:"key1"`
	Inf   inner.Inf
	SlInf []inner.Inf
	Mem   *mch.MemCache
	d 	  *net.Dialer
	conn  net.Conn
}

type NewestStruct struct{}

func NewStruct(
	i inner.Inf,
	sl []inner.Inf,
	o *Other,
	st inner.Struct,
	nm *inner.Named,
	m *mch.MemCache
	d *net.Dialer
	conn net.Conn
) *NewestStruct {
	return &NewestStruct{}
}

func NewStct(
	i inner.Inf,
	sl []inner.Inf,
	o Other,
	st *inner.Struct,
	nm inner.Named,
	manager *mdtreflect.EventManager
) NewestStruct {
	return NewestStruct{}
}
