//go:build octogen
// +build octogen

package foo

import "github.com/oesand/octo/octogen"

func IncludeAny[T any](v int) int {
	octogen.Inject[Struct]()
}
