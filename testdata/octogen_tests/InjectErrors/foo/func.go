package foo

func ValidFunc(other *ValidStruct) ValidInterface {
	return nil
}

func MultipleReturnsFunc(other *ValidStruct) (ValidInterface, *ValidStruct) {
	return nil, nil
}

func GenericFunc[T any](other *ValidStruct) ValidInterface {
	return nil
}

func InvalidParamFunc(other Invalid) ValidInterface {
	return nil
}

func InvalidReturnFunc(other ValidInterface) Invalid {
	return nil
}
