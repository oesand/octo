package nested

import "github.com/oesand/octo/internal/octogen_tests/NestedAnyVariants/foo/nested/inner"

type Other struct {
	Nm    *inner.Named `key:"key1"`
	Inf   inner.Inf
	SlInf []inner.Inf
}

type NewestStruct struct{}

func NewStruct(
	i inner.Inf,
	sl []inner.Inf,
	o *Other,
	st inner.Struct,
	nm *inner.Named,
) *NewestStruct {
	return &NewestStruct{}
}

func NewStct(
	i inner.Inf,
	sl []inner.Inf,
	o Other,
	st *inner.Struct,
	nm inner.Named,
) NewestStruct {
	return NewestStruct{}
}
