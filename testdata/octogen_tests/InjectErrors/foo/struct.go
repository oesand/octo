package foo

type ValidInterface interface{}

type ValidStruct struct {
	Other ValidInterface
}

type InvalidFieldStruct struct {
	Other Invalid
}
