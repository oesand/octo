package decl

type FuncDecl struct {
	Pkg     *PackageDecl
	Name    string
	Injects []InjectedDecl
}
