package nested

import (
	"github.com/oesand/octo/testdata/octogen_tests/NestedAnyVariants/foo/nested/inner"
	"github.com/oesand/octo/mc"
	"github.com/oesand/octo/mediatr"
	"net"
)

type Other struct {
	Nm    *inner.Named `key:"key1"`
	Inf   inner.Inf
	SlInf []inner.Inf
	Mem   *mc.MemCache
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
	m *mc.MemCache
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
	manager *mediatr.Manager
) NewestStruct {
	return NewestStruct{}
}
