package foo

type Other struct {
	Name string
}

type Struct struct {
	Name  string
	Age   int
	Other *Other
}
