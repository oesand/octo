package decl

type PackageDecl struct {
	Path    string
	Imports []string
	Funcs   []FuncDecl
}

type FuncDecl struct {
	Name    string
	Injects []InjectedDecl
}

type InjectedDeclType string

const (
	InjectedStructType InjectedDeclType = "struct"
	InjectedFuncType   InjectedDeclType = "func"
)

type InjectedDecl interface {
	Type() InjectedDeclType
}

type LocaleInfo struct {
	Sliced  bool
	Ptr     bool
	Package string
	Name    string
}
