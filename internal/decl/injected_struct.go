package decl

type InjectedStruct struct {
	Fields []*InjectedStructField
	Return *LocaleInfo

	Optional  bool
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
