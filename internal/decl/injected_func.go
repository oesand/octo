package decl

type InjectedFunc struct {
	Locale *LocaleInfo
	Params []*LocaleInfo
	Return *LocaleInfo

	KeyOption string
}

func (*InjectedFunc) Type() InjectedDeclType {
	return InjectedFuncType
}
