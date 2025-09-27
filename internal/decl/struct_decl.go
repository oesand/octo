package decl

type InjectedStruct struct {
	Locale LocaleInfo
	Fields []InjectedStructField

	KeyOption string
}

func (InjectedStruct) Type() InjectedDeclType {
	return InjectedStructType
}

type InjectedStructField struct {
	Name   string
	Locale LocaleInfo

	KeyOption string
}
