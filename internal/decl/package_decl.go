package decl

import (
	"github.com/oesand/octo/pm"
)

type PackageDecl struct {
	Name    string
	PkgPath string
	Path    string
	Imports pm.Set[string]
	Funcs   []*FuncDecl
}
