package parse

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

func ParseInjects() []PackageDecl {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedImports | packages.NeedDeps,
		BuildFlags: []string{
			"-tags", "octogen",
		},
		Dir: ".", // current dir
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatal(err)
	}

	var pkgDecls []PackageDecl
	for _, pkg := range pkgs {
		imports := map[string]string{}
		var octogenAlias string
		for _, f := range pkg.Syntax {
			for _, im := range f.Imports {
				path := im.Path.Value[1 : len(im.Path.Value)-1] // strip quotes
				var alias string
				if im.Name != nil {
					alias = im.Name.Name
				} else {
					alias = path[strings.LastIndex(path, "/")+1:]
				}
				if path == "github.com/oesand/octo/octogen" {
					octogenAlias = alias
				}
				imports[path] = alias
			}
		}

		if octogenAlias == "" {
			continue
		}

		var funcs []FuncDecl

		// scan functions
		for _, file := range pkg.Syntax {
			var hasBuildFlag bool
			for _, commentGroup := range file.Comments {
				for _, c := range commentGroup.List {
					text := strings.Trim(c.Text, "// ")

					if (strings.HasPrefix(text, "+build") || strings.HasPrefix(text, "go:build")) &&
						strings.Contains(text, "octogen") {
						hasBuildFlag = true
						break
					}
				}

				if hasBuildFlag {
					break
				}
			}

			if !hasBuildFlag {
				continue
			}

			ast.Inspect(file, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}

				var injects []InjectDecl
				ast.Inspect(fn.Body, func(nn ast.Node) bool {
					call, ok := nn.(*ast.CallExpr)
					if !ok {
						return true
					}

					// look for "alias.Inject[T]" or "alias.InjectNamed[T]"
					if idx, ok := call.Fun.(*ast.IndexExpr); ok {
						if sel, ok := idx.X.(*ast.SelectorExpr); ok {
							if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == octogenAlias {
								kind := sel.Sel.Name
								if kind == "Inject" || kind == "InjectNamed" {
									typ := pkg.TypesInfo.Types[idx.Index].Type
									key := ""
									if kind == "InjectNamed" && len(call.Args) > 0 {
										if bl, ok := call.Args[0].(*ast.BasicLit); ok {
											key = bl.Value[1 : len(bl.Value)-1] // strip quotes
										}
									}

									typ, declLocale := parseTypeLocale(typ)
									if declLocale == nil {
										return true
									}

									fmt.Printf("Decl locale: %v \n", declLocale)

									var declType InjectDeclType
									var fields []InjectDeclField
									switch t := typ.(type) {
									// Look inner struct{...} declaration fields
									case *types.Struct:
										declType = InjectDeclStruct
										for i := 0; i < t.NumFields(); i++ {
											field := t.Field(i)
											if !field.Exported() {
												continue
											}

											fieldTags := t.Tag(i)
											var fieldKeyOption string
											if idx := strings.Index(fieldTags, `key:"`); idx >= 0 {
												rest := fieldTags[idx+5:]
												if end := strings.Index(rest, `"`); end > 0 {
													fieldKeyOption = rest[:end]
												}
											}

											_, fieldTypeLoc := parseTypeLocale(field.Type())

											fields = append(fields, InjectDeclField{
												Name:      field.Name(),
												KeyOption: fieldKeyOption,
												Locale:    *fieldTypeLoc,
											})
										}

										fmt.Printf("Fields: %v \n", fields)

									// Look function(...) declaration parameters
									case *types.Signature:
										declType = InjectDeclFunc
										for prm := range t.Params().Variables() {
											_, fieldTypeLoc := parseTypeLocale(prm.Type())

											fields = append(fields, InjectDeclField{
												Name:   prm.Name(),
												Locale: *fieldTypeLoc,
											})
										}

										fmt.Printf("Fields: %v \n", fields)

									default:
										return true
									}

									fmt.Printf("Decl locale: %v, Fields: %v \n", declLocale, fields)

									injects = append(injects, InjectDecl{
										KeyOption: key,
										Type:      declType,
										Fields:    fields,
										Locale:    *declLocale,
									})
								}
							}
						}
					}
					return true
				})

				if len(injects) > 0 {
					funcs = append(funcs, FuncDecl{
						Name:    fn.Name.Name,
						Injects: injects,
					})
				}

				return true
			})
		}

		if len(funcs) > 0 {
			pkgDecls = append(pkgDecls, PackageDecl{
				Path:  pkg.Dir,
				Funcs: funcs,
			})
		}
	}

	return pkgDecls
}
