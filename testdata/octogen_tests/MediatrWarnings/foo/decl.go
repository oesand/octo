//go:build octogen
// +build octogen

package foo

import "github.com/oesand/octo/octogen"

func IncludeAny() {
	octogen.ScanForMediatr()
}
