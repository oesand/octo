package parse

import (
	"fmt"
	"github.com/oesand/octo/internal/decl"
	"github.com/oesand/octo/internal/prim"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"strings"
)

const octogenModule = "github.com/oesand/octo/octogen"

func ParseInjects(currentModule string, dir string) ([]*decl.PackageDecl, []string, []error) {
	fileSet := token.NewFileSet()

	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedImports | packages.NeedDeps,
		BuildFlags: []string{
			"-tags", "octogen",
		},
		Fset: fileSet,
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, nil, []error{err}
	}

	if len(pkgs) == 0 {
		return nil, nil, []error{fmt.Errorf("no packages found")}
	}

	var mediatrWarns []string
	var errs []error
	var pkgDecls []*decl.PackageDecl

	var mediatrInjects []types.Object
	var funcsIncludeMediatr []*decl.FuncDecl
	for _, pkg := range pkgs {
		pkgPath := pkg.ID
		if !strings.HasPrefix(pkgPath, currentModule) {
			continue
		}

		var imports prim.Set[string]
		var funcs []*decl.FuncDecl

		pkgDecl := &decl.PackageDecl{
			Name:    filepath.Base(pkgPath),
			PkgPath: pkgPath,
			Path:    pkg.Dir,
		}

		// scan functions
		for _, file := range pkg.Syntax {
			// check if in file has `+build octogen` or 'go:build octogen' flag
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
				for _, obj := range file.Scope.Objects {
					spec, ok := obj.Decl.(*ast.TypeSpec)
					if !ok {
						continue
					}

					typName, ok := pkg.TypesInfo.ObjectOf(spec.Name).(*types.TypeName)
					if !ok {
						continue
					}

					named, ok := typName.Type().(*types.Named)
					if !ok {
						continue
					}

					stctTyp, ok := named.Underlying().(*types.Struct)
					if !ok {
						continue
					}

					// Check if struct implements the interface
					if !implementsMediatrHandlers(named) {
						continue
					}

					if named.TypeParams().Len() > 0 {
						mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, obj.Pos(), "found struct that implements mediatr but was skipped because generics not supported"))
						continue
					}

					funcObj := file.Scope.Lookup("New" + obj.Name)
					if funcObj != nil {
						funcDecl, ok := funcObj.Decl.(*ast.FuncDecl)
						if !ok {
							mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, funcObj.Pos(), "found New... declaration for struct, that implements mediatr but was skipped because it not function"))
							continue
						}

						funcType, ok := pkg.TypesInfo.ObjectOf(funcDecl.Name).(*types.Func)
						if !ok {
							mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, funcObj.Pos(), "found New... declaration for struct, that implements mediatr but was skipped because it not function"))
							continue
						}

						funcSig := funcType.Signature()

						if funcSig.TypeParams().Len() > 0 {
							mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, funcType.Pos(), "found New... function for struct, that implements mediatr but was skipped because generics not supported"))
							continue
						}

						if funcSig.Results().Len() != 1 {
							mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, funcType.Pos(), "found New... function for struct, that implements mediatr but was skipped because returned invalid count results"))
							continue
						}

						const invalidReturnedTypeText = "found New... function for struct, that implements mediatr but was skipped because returned not pointer to struct"

						returnedTyp := funcSig.Results().At(0).Type()

						pointer, ok := returnedTyp.(*types.Pointer)
						if !ok {
							mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, funcType.Pos(), invalidReturnedTypeText))
							continue
						}

						returnedTyp, ok = pointer.Elem().(*types.Named)
						if !ok {
							mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, funcType.Pos(), invalidReturnedTypeText))
							continue
						}

						if returnedTyp.Underlying() != stctTyp {
							mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, funcType.Pos(), invalidReturnedTypeText))
							continue
						}

						mediatrInjects = append(mediatrInjects, funcType)
					} else {
						mediatrInjects = append(mediatrInjects, typName)
					}
				}

				continue
			}

			// scan imports for find alias for module octogen for generation shortcuts
			var octogenAlias string
			for _, im := range file.Imports {
				path := im.Path.Value[1 : len(im.Path.Value)-1] // strip quotes
				var alias string
				if im.Name != nil {
					alias = im.Name.Name
				} else {
					alias = path[strings.LastIndex(path, "/")+1:]
				}
				if path == octogenModule {
					octogenAlias = alias
				}
			}

			if octogenAlias == "" {
				continue
			}

			// scan files with declarative functions for generate
			ast.Inspect(file, func(n ast.Node) bool {
				if err != nil {
					return false
				}

				fn, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}

				if fn.Type.TypeParams.NumFields() > 0 {
					errs = append(errs, locatedErr(pkg.Fset, fn.Pos(), "has generic arguments in declaration function, expect no generics"))
					return false
				}

				if fn.Type.Params.NumFields() > 0 {
					errs = append(errs, locatedErr(pkg.Fset, fn.Pos(), "has arguments in declaration function, expect no arguments"))
					return false
				}

				if fn.Type.Results.NumFields() > 0 {
					errs = append(errs, locatedErr(pkg.Fset, fn.Pos(), "has returns in declaration function, expect no returns"))
					return false
				}

				var includeMediatr bool
				var injects []decl.InjectedDecl
				ast.Inspect(fn.Body, func(nn ast.Node) bool {
					if err != nil {
						return false
					}
					call, ok := nn.(*ast.CallExpr)
					if !ok {
						return true
					}

					// look for "octogen... funcs"
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == octogenAlias {
							switch sel.Sel.Name {
							// look for "octogen.Inject(...)" inside
							case "Inject":
								if len(call.Args) == 0 {
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "function argument not passed"))
									return false
								}

								if len(call.Args) > 2 {
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "too many arguments, maximum two arguments"))
									return false
								}

								const injectTypeParamUnsupportedError = "call without generic support only link to function"

								funcExpr := call.Args[0]

								var funcIdent *ast.Ident
								switch et := funcExpr.(type) {
								case *ast.Ident:
									funcIdent = et
								case *ast.SelectorExpr:
									funcIdent = et.Sel
								default:
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), injectTypeParamUnsupportedError))
									return false
								}

								funcObj, ok := pkg.TypesInfo.ObjectOf(funcIdent).(*types.Func)
								if !ok {
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), injectTypeParamUnsupportedError))
									return false
								}

								funcSig := funcObj.Signature()

								if funcSig.TypeParams().Len() != 0 {
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "linked function has generic arguments, expect no generics"))
									return false
								}

								if funcSig.Results().Len() != 1 {
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "linked function should return only one result"))
									return false
								}

								_, returnLoc, err := parseStructLocale(funcSig.Results().At(0).Type())
								if err != nil {
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "linked function returning: %s", err))
									return false
								}

								imports.Add(returnLoc.Package)

								var key string
								if len(call.Args) > 1 {
									if bl, ok := call.Args[1].(*ast.BasicLit); ok && bl.Kind == token.STRING {
										key = bl.Value[1 : len(bl.Value)-1] // strip quotes
									} else {
										errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "unexpected second argument, support only string"))
										return false
									}
								}

								var params = make([]*decl.LocaleInfo, funcSig.Params().Len())
								for i := 0; i < funcSig.Params().Len(); i++ {
									prm := funcSig.Params().At(i)

									var prmLoc *decl.LocaleInfo
									prmLoc, err = parseFieldLocale(prm.Type())
									if err != nil {
										errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "linked function param (%s [%d]): %s", prm.Name(), i+1, err))
										return false
									}

									imports.Add(prmLoc.Package)

									params[i] = prmLoc
								}

								funcPackage := funcObj.Pkg().Path()
								imports.Add(funcPackage)

								injects = append(injects, &decl.InjectedFunc{
									Locale: &decl.LocaleInfo{
										Package: funcPackage,
										Name:    funcObj.Name(),
									},
									Params:    params,
									Return:    returnLoc,
									KeyOption: key,
								})

							case "ScanForMediatr":
								if len(call.Args) != 0 {
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "too many arguments, expect no arguments"))
									return false
								}

								if includeMediatr {
									errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "ScanForMediatr already defined in this function"))
									return false
								}

								includeMediatr = true
							}
						}
					}

					// look for "alias.Inject[T]" or "alias.InjectNamed[T]"
					if idx, ok := call.Fun.(*ast.IndexExpr); ok {
						if sel, ok := idx.X.(*ast.SelectorExpr); ok {
							if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == octogenAlias {
								kind := sel.Sel.Name
								if kind == "Inject" {
									typ := pkg.TypesInfo.Types[idx.Index].Type
									key := ""
									if len(call.Args) == 1 {
										if bl, ok := call.Args[0].(*ast.BasicLit); ok && bl.Kind == token.STRING {
											key = bl.Value[1 : len(bl.Value)-1] // strip quotes
										} else {
											errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "unexpected second argument, support only string"))
											return false
										}
									} else if len(call.Args) > 0 {
										errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "too many arguments, maximum one argument"))
										return false
									}

									stct, stctLoc, err := parseStructLocale(typ)
									if err != nil {
										errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "inject with generic parameter support only struct or pointer struct"))
										return false
									}

									imports.Add(stctLoc.Package)

									var fields = make([]*decl.InjectedStructField, stct.NumFields())
									for i := 0; i < stct.NumFields(); i++ {
										field := stct.Field(i)
										if !field.Exported() {
											continue
										}

										fieldTags := stct.Tag(i)
										var keyOption string
										if idx := strings.Index(fieldTags, `key:"`); idx >= 0 {
											rest := fieldTags[idx+5:]
											if end := strings.Index(rest, `"`); end > 0 {
												keyOption = rest[:end]
											}
										}

										fieldLoc, err := parseFieldLocale(field.Type())
										if err != nil {
											errs = append(errs, locatedErr(pkg.Fset, ident.Pos(), "linked struct field (%s [%d]): %s", field.Name(), i+1, err))
											return false
										}

										imports.Add(fieldLoc.Package)

										fields[i] = &decl.InjectedStructField{
											Name:      field.Name(),
											Locale:    fieldLoc,
											KeyOption: keyOption,
										}
									}

									injects = append(injects, &decl.InjectedStruct{
										Fields:    fields,
										Return:    stctLoc,
										KeyOption: key,
									})
								}
							}
						}
					}
					return true
				})

				if len(injects) > 0 || includeMediatr {
					funcDecl := &decl.FuncDecl{
						Pkg:     pkgDecl,
						Name:    fn.Name.Name,
						Injects: injects,
					}

					funcs = append(funcs, funcDecl)
					if includeMediatr {
						funcsIncludeMediatr = append(funcsIncludeMediatr, funcDecl)
					}
				}

				return true
			})
		}

		if len(errs) > 0 {
			continue
		}

		if len(funcs) > 0 {
			if imports.Has(pkgDecl.PkgPath) {
				imports.Del(pkgDecl.PkgPath)
			}

			pkgDecl.Imports = imports
			pkgDecl.Funcs = funcs

			pkgDecls = append(pkgDecls, pkgDecl)
		}
	}

	if len(errs) > 0 {
		return nil, nil, errs
	}

	if len(funcsIncludeMediatr) > 0 {
		var extraImports prim.Set[string]
		var parsedInjects []decl.InjectedDecl
		for _, obj := range mediatrInjects {
			switch ot := obj.(type) {
			case *types.TypeName:
				var stctImports prim.Set[string]
				stctTyp := ot.Type().Underlying().(*types.Struct)

				var failed bool
				var fields []*decl.InjectedStructField
				for i := 0; i < stctTyp.NumFields(); i++ {
					field := stctTyp.Field(i)
					if !field.Exported() {
						continue
					}

					fieldTags := stctTyp.Tag(i)
					var keyOption string
					if idx := strings.Index(fieldTags, `key:"`); idx >= 0 {
						rest := fieldTags[idx+5:]
						if end := strings.Index(rest, `"`); end > 0 {
							keyOption = rest[:end]
						}
					}

					fieldLoc, err := parseFieldLocale(field.Type())
					if err != nil {
						mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, field.Pos(), "mediatr struct field (%s [%d]): %s", field.Name(), i+1, err))
						failed = true
						break
					}

					stctImports.Add(fieldLoc.Package)

					fields = append(fields, &decl.InjectedStructField{
						Name:      field.Name(),
						Locale:    fieldLoc,
						KeyOption: keyOption,
					})
				}

				if failed {
					continue
				}

				stctPath := ot.Pkg().Path()
				stctImports.Add(stctPath)

				extraImports.CopyFrom(stctImports)
				parsedInjects = append(parsedInjects, &decl.InjectedStruct{
					Fields: fields,
					Return: &decl.LocaleInfo{
						Ptr:     true,
						Name:    ot.Name(),
						Package: stctPath,
					},
					Optional: true,
				})

			case *types.Func:
				var funcImports prim.Set[string]
				funcSig := ot.Signature()

				_, returnLoc, err := parseStructLocale(funcSig.Results().At(0).Type())
				if err != nil {
					mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, ot.Pos(), "mediatr function returning: %s", err))
					continue
				}

				var failed bool
				var params = make([]*decl.LocaleInfo, funcSig.Params().Len())
				for i := 0; i < funcSig.Params().Len(); i++ {
					prm := funcSig.Params().At(i)

					var prmLoc *decl.LocaleInfo
					prmLoc, err = parseFieldLocale(prm.Type())
					if err != nil {
						mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, prm.Pos(), "mediatr function param (%s [%d]): %s", prm.Name(), i+1, err))
						failed = true
						break
					}

					funcImports.Add(prmLoc.Package)
					params[i] = prmLoc
				}

				if failed {
					continue
				}

				funcPackage := ot.Pkg().Path()
				funcImports.Add(funcPackage)

				extraImports.CopyFrom(funcImports)
				parsedInjects = append(parsedInjects, &decl.InjectedFunc{
					Locale: &decl.LocaleInfo{
						Package: funcPackage,
						Name:    ot.Name(),
					},
					Params:   params,
					Return:   returnLoc,
					Optional: true,
				})

			default:
				mediatrWarns = append(mediatrWarns, locatedMsg(fileSet, obj.Pos(), "found unexpected type implements mediatr handler"))
			}
		}

		if len(parsedInjects) > 0 {
			for _, funcDecl := range funcsIncludeMediatr {
				funcDecl.Pkg.Imports.CopyFrom(extraImports)
				funcDecl.Injects = append(funcDecl.Injects, parsedInjects...)
			}
		}

		return pkgDecls, mediatrWarns, errs
	}

	return pkgDecls, nil, errs
}
