//go:build octogen
// +build octogen

package foo

import (
	"github.com/oesand/octo/octogen"
)

func IncludeAny() {
	octogen.Inject()
	octogen.Inject(Invalid)
	octogen.Inject(ValidFunc, "key", 123)
	octogen.Inject(ValidFunc, 123)
	octogen.Inject(ValidStruct)
	octogen.Inject(MultipleReturnsFunc)
	octogen.Inject(GenericFunc)
	octogen.Inject(InvalidParamFunc)
	octogen.Inject(InvalidReturnFunc)

	octogen.Inject[]()
	octogen.Inject[Invalid]()
	octogen.Inject[ValidInterface]()
	octogen.Inject[ValidStruct]("key", 123)
	octogen.Inject[ValidStruct](123)
	octogen.Inject[InvalidFieldStruct]()
}
