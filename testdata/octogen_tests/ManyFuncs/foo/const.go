package foo

type Inf interface{}

type Struct struct{}

type Other struct {
	Inf   Inf
	SlInf []Inf
	Str *Struct
}
