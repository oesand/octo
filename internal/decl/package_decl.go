package decl

type PackageDecl struct {
	Path  string
	Funcs []FuncDecl
}

type FuncDecl struct {
	Name    string
	Injects []InjectedDecl
}

type InjectedDeclType int

const (
	InjectedStructType InjectedDeclType = iota
	InjectedFuncType
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
