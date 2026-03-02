package parse

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"

	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/content/injects"
)

func Parse(module, dir string) ([]*content.PkgRenderer, []string, []error) {
	parseCtx := newCtx(module, dir)

	pkgs, err := parseCtx.Packages()
	if err != nil {
		return nil, nil, []error{err}
	}

	var outputs []*content.PkgRenderer
	for pkg := range pkgs {
		pkgPath := pkg.ID
		renderCtx := content.NewCtx(pkgPath)

		var blocks []content.FileBlockRenderer
		for _, file := range pkg.Syntax {
			hasBuildFlag := parseCtx.HasBuildTag(file)

			if !hasBuildFlag {
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

					if name := lookOctogenCall(call.Fun, octogenAlias); name == "Inject" {
						var funcObj *types.Func
						var injectKey string
						{ // Extract type info from Inject(...)
							if len(call.Args) == 0 {
								parseCtx.AddErr(call.Pos(), "injecting function not passed")
								continue
							}
							if len(call.Args) > 2 {
								parseCtx.AddErr(call.Pos(), "too many arguments, maximum two arguments")
								continue
							}

							if len(call.Args) > 1 {
								if bl, ok := call.Args[1].(*ast.BasicLit); ok && bl.Kind == token.STRING {
									injectKey = bl.Value[1 : len(bl.Value)-1] // strip quotes
								} else {
									parseCtx.AddErr(call.Pos(), "unexpected second argument, support only string")
									continue
								}
							}

							var funcIdent *ast.Ident
							switch et := call.Args[0].(type) {
							case *ast.Ident:
								funcIdent = et
							case *ast.SelectorExpr:
								funcIdent = et.Sel
							default:
								parseCtx.AddErr(call.Pos(), "not supported injecting target")
								continue
							}

							funcObj, ok = pkg.TypesInfo.ObjectOf(funcIdent).(*types.Func)
							if !ok {
								parseCtx.AddErr(call.Pos(), "not supported injecting target")
								continue
							}
						}

						inject, injectImports, err := parseInjectFunc(injectKey, funcObj)
						if err != nil {
							parseCtx.AddError(call.Pos(), err)
						} else {
							declaredInjects = append(declaredInjects, inject)
							for _, injectPkg := range injectImports {
								renderCtx.Import(injectPkg)
							}
						}

						continue
					}

					if name, genericExp := lookOctogenGenericCall(call, octogenAlias); name == "Inject" {
						var structType types.Type
						var injectKey string
						{
							if len(call.Args) > 1 {
								parseCtx.AddErr(call.Pos(), "too many arguments, maximum one argument")
								continue
							}

							if len(call.Args) > 0 {
								if bl, ok := call.Args[0].(*ast.BasicLit); ok && bl.Kind == token.STRING {
									injectKey = bl.Value[1 : len(bl.Value)-1] // strip quotes
								} else {
									parseCtx.AddErr(call.Pos(), "unexpected name argument, support only string")
									continue
								}
							}

							structType = pkg.TypesInfo.TypeOf(genericExp)
							if structType == nil {
								parseCtx.AddErr(call.Pos(), "unknown injecting type")
								continue
							}
						}

						inject, injectImports, err := parseInjectStruct(injectKey, structType)
						if err != nil {
							parseCtx.AddError(call.Pos(), err)
						} else {
							declaredInjects = append(declaredInjects, inject)
							for _, injectPkg := range injectImports {
								renderCtx.Import(injectPkg)
							}
						}

						continue
					}
				}

				if parseCtx.NoErrs() && len(declaredInjects) > 0 {
					injectFuncName := injectFunc.Name.Name

					blocks = append(blocks, injects.Func(injectFuncName, declaredInjects))
				}
			}
		}

		if parseCtx.NoErrs() && len(blocks) > 0 {
			sort.Slice(blocks, func(i, j int) bool {
				return blocks[i].Key() < blocks[j].Key()
			})

			outputs = append(outputs, content.Pkg(pkgPath, pkg.Dir, renderCtx, blocks))
		}
	}

	return outputs, parseCtx.warns, parseCtx.errs
}

func lookOctogenCall(exp ast.Expr, octogenAlias string) string {
	if sel, ok := exp.(*ast.SelectorExpr); ok {
		if idn, ok := sel.X.(*ast.Ident); !ok || idn.Name != octogenAlias {
			return ""
		}

		return sel.Sel.Name
	}
	return ""
}

func lookOctogenGenericCall(exp *ast.CallExpr, octogenAlias string) (string, ast.Expr) {
	if idx, ok := exp.Fun.(*ast.IndexExpr); ok {
		funcName := lookOctogenCall(idx.X, octogenAlias)
		if funcName == "" {
			return "", nil
		}

		return funcName, idx.Index
	}
	return "", nil
}
