package parse

type PackageDecl struct {
	Path  string
	Funcs []FuncDecl
}

type FuncDecl struct {
	Name    string
	Injects []InjectDecl
}

type InjectDeclType int

const (
	InjectDeclStruct InjectDeclType = iota
	InjectDeclFunc
)

type InjectDecl struct {
	KeyOption string
	Type      InjectDeclType
	Locale    LocaleInfo
	Fields    []InjectDeclField
}

type InjectDeclField struct {
	Name      string
	KeyOption string
	Locale    LocaleInfo
}

type LocaleInfo struct {
	PtrLevel int
	Package  string
	Name     string
}
