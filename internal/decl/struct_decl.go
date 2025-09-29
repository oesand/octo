package decl

type InjectedStruct struct {
	Fields []*InjectedStructField
	Return *LocaleInfo

	KeyOption string
}

func (*InjectedStruct) Type() InjectedDeclType {
	return InjectedStructType
}

type InjectedStructField struct {
	Name   string
	Locale *LocaleInfo

	KeyOption string
}
