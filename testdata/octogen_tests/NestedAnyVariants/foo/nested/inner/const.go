package inner

type Inf interface{}

type Struct struct{}

type Named struct {
	Inf   Inf
	SlInf []Inf
}
