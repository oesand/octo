//go:build octogen
// +build octogen

package foo

import (
	"github.com/oesand/octo/octogen"
)

func IncludeAny() {
	octogen.Inject[*Struct]()
	octogen.Inject[*Other]()
}

func IncludeSecond() {
	octogen.Inject[*Struct]()
	octogen.Inject[*Other]()
}

func IncludeThird() {
	octogen.Inject[*Struct]()
	octogen.Inject[*Other]()
}
