package decl

import (
	"github.com/oesand/octo/internal/prim"
)

type PackageDecl struct {
	Name    string
	PkgPath string
	Path    string
	Imports prim.Set[string]
	Funcs   []*FuncDecl
}
