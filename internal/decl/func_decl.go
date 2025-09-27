package decl

type InjectedFunc struct {
	Package string
	Name    string
	Params  []LocaleInfo
	Return  LocaleInfo

	KeyOption string
}

func (InjectedFunc) Type() InjectedDeclType {
	return InjectedFuncType
}
