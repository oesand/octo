package decl

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
