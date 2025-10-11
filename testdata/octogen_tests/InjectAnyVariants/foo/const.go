package foo

type Inf interface{}

type Struct struct{}

type Named struct {
	Oth   *Other
	Inf   Inf
	SlInf []Inf
}

type Other struct {
	Nm    *Named `key:"key1"`
	Inf   Inf
	SlInf []Inf
}

type NewestStruct struct{}

func NewStruct(
	i Inf,
	sl []Inf,
	o *Other,
	st Struct,
	nm *Named,
) *NewestStruct {
	return &NewestStruct{}
}

func NewStct(
	i Inf,
	sl []Inf,
	o Other,
	st *Struct,
	nm Named,
) NewestStruct {
	return NewestStruct{}
}

func NewInf(
	st *Struct,
) Inf {
	return *Other{}
}
