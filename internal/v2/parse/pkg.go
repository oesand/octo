package parse

import (
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"slices"

	"github.com/oesand/octo/internal/v2/injects"
)

func Parse(module, dir string) ([]*injects.PkgRenderer, []string, []error) {
	parseCtx := newCtx(module, dir)

	pkgs, err := parseCtx.Packages()
	if err != nil {
		return nil, nil, []error{err}
	}

	var outputs []*injects.PkgRenderer
	for pkg := range pkgs {
		pkgPath := pkg.ID
		renderCtx := injects.NewCtx(pkgPath)

		var funcs map[string]*injects.FuncRenderer
		for _, file := range pkg.Syntax {
			hasBuildFlag := parseCtx.HasBuildTag(file)

			if !hasBuildFlag {
				// todo : add parse mediatr
				continue
			}

			octogenAlias := parseCtx.GetOctogenAlias(file)
			if octogenAlias == "" {
				continue
			}

			for rootNode := range ast.Preorder(file) {
				injectFunc, ok := rootNode.(*ast.FuncDecl)
				if !ok {
					continue
				}

				if injectFunc.Type.TypeParams.NumFields() > 0 {
					parseCtx.AddWarn(injectFunc.Pos(), "expect no generics in declaration function")
				}

				if injectFunc.Type.Params.NumFields() > 0 {
					parseCtx.AddWarn(injectFunc.Pos(), "expect no arguments in declaration function")
				}

				if injectFunc.Type.Results.NumFields() > 0 {
					parseCtx.AddWarn(injectFunc.Pos(), "expect no returns in declaration function")
				}

				var declaredInjects []injects.InjectRenderer
				for bodyDecl := range ast.Preorder(injectFunc.Body) {
					call, ok := bodyDecl.(*ast.CallExpr)
					if !ok {
						continue
					}

					if name, ident := lookOctogenCall(call, octogenAlias); name != "" {
						switch name {
						case "Inject":
							{
								var funcObj *types.Func
								var injectKey string
								{ // Extract type info from Inject(...)
									if len(call.Args) == 0 {
										parseCtx.AddErr(ident.Pos(), "injecting function not passed")
										continue
									}
									if len(call.Args) > 2 {
										parseCtx.AddErr(ident.Pos(), "too many arguments, maximum two arguments")
										continue
									}

									var funcIdent *ast.Ident
									switch et := call.Args[0].(type) {
									case *ast.Ident:
										funcIdent = et
									case *ast.SelectorExpr:
										funcIdent = et.Sel
									default:
										parseCtx.AddErr(ident.Pos(), "not supported injecting target")
										continue
									}

									funcObj, ok = pkg.TypesInfo.ObjectOf(funcIdent).(*types.Func)
									if !ok {
										parseCtx.AddErr(ident.Pos(), "not supported injecting target")
										continue
									}

									if len(call.Args) > 1 {
										if bl, ok := call.Args[1].(*ast.BasicLit); ok && bl.Kind == token.STRING {
											injectKey = bl.Value[1 : len(bl.Value)-1] // strip quotes
										} else {
											parseCtx.AddErr(ident.Pos(), "unexpected second argument, support only string")
											continue
										}
									}
								}

								inject, injectPkgs := parseCtx.ParseInjectFunc(injectKey, funcObj)
								if parseCtx.NoErrs() && inject != nil {
									declaredInjects = append(declaredInjects, inject)
									for _, injectPkg := range injectPkgs {
										renderCtx.Import(injectPkg)
									}
								}
							}
						case "ScanForMediatr":
							{

							}
						}

						continue
					}

					//if name, ident := lookOctogenGenericCall(call, octogenAlias); name == "Inject" {
					//}
				}

				if parseCtx.NoErrs() && len(declaredInjects) > 0 {
					injectFuncName := injectFunc.Name.Name

					if funcs == nil {
						funcs = make(map[string]*injects.FuncRenderer, len(declaredInjects))
					}

					if fn, has := funcs[injectFuncName]; has {
						fn.Injects = append(fn.Injects, declaredInjects...)
					} else {
						funcs[injectFuncName] = &injects.FuncRenderer{
							Name:    injectFuncName,
							Injects: declaredInjects,
						}
					}
				}
			}
		}

		if parseCtx.NoErrs() && len(funcs) > 0 {
			outputs = append(outputs, injects.Pkg(pkg.Name, pkg.Dir, renderCtx, slices.Collect(maps.Values(funcs))))
		}
	}

	return outputs, parseCtx.warns, parseCtx.errs
}

func lookOctogenCall(exp *ast.CallExpr, octogenAlias string) (string, *ast.Ident) {
	if sel, ok := exp.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == octogenAlias {
			return sel.Sel.Name, ident
		}
	}
	return "", nil
}

func lookOctogenGenericCall(exp *ast.CallExpr, octogenAlias string) (string, *ast.Ident) {
	if idx, ok := exp.Fun.(*ast.IndexExpr); ok {
		if sel, ok := idx.X.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == octogenAlias {
				return sel.Sel.Name, ident
			}
		}
	}
	return "", nil
}
