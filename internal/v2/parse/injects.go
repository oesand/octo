package parse

import (
	"errors"
	"fmt"
	"go/types"

	"github.com/oesand/octo/internal/v2/injects"
	"github.com/oesand/octo/internal/v2/typing"
	"github.com/oesand/octo/pm"
)

func parseInjectFunc(key string, funcObj *types.Func) (injects.InjectRenderer, []string, error) {
	funcSig := funcObj.Signature()

	if funcSig.Results().Len() != 1 {
		return nil, nil, errors.New("injecting function should return only one result")
	}

	var pkgs pm.Set[string]
	resType := funcSig.Results().At(0)

	returned, err := parseType(pkgs, resType.Type())
	if err != nil {
		return nil, nil, fmt.Errorf("injecting function return: %w", err)
	}

	generics := make([]typing.Renderer, funcSig.TypeParams().Len())
	for i := 0; i < funcSig.TypeParams().Len(); i++ {
		generic, err := parseType(pkgs, funcSig.TypeParams().At(i))
		if err != nil {
			return nil, nil, fmt.Errorf("injecting function generic[%d]: %w", i, err)
		}
		generics[i] = generic
	}

	params := make([]injects.ResolveRenderer, funcSig.Params().Len())
	for i := 0; i < funcSig.Params().Len(); i++ {
		prm := funcSig.Params().At(i)
		param, err := parseType(pkgs, prm.Type())
		if err != nil {
			return nil, nil, fmt.Errorf("injecting function param[%d]: %s", i, err)
		}
		params[i] = injects.Resolve("", param)
	}

	funcName := funcObj.Name()
	var funcPkg string
	if fp := funcObj.Pkg(); fp != nil {
		funcPkg = fp.Path()
		pkgs.Add(funcPkg)
	}

	funcDecl := typing.NewNamed(funcPkg, funcName, generics)

	return injects.Inject(key, returned, injects.ReturnFunc(funcDecl, params)), pkgs.Values(), nil
}

func parseType(pkgs pm.Set[string], typ types.Type) (typing.Renderer, error) {
	switch t := typ.(type) {
	case *types.Basic:
		return typing.NewNamed("", t.Name(), nil), nil

	case *types.Array:
		elemTyp, err := parseType(pkgs, t.Elem())
		if err != nil {
			return nil, fmt.Errorf("[%d] > %w", t.Len(), err)
		}
		return typing.NewSlice(t.Len(), elemTyp), nil

	case *types.Slice:
		elemTyp, err := parseType(pkgs, t.Elem())
		if err != nil {
			return nil, fmt.Errorf("[] > %w", err)
		}
		return typing.NewSlice(0, elemTyp), nil

	case *types.Map:
		keyTyp, err := parseType(pkgs, t.Key())
		if err != nil {
			return nil, fmt.Errorf("map[x] > %w", err)
		}

		elemTyp, err := parseType(pkgs, t.Elem())
		if err != nil {
			return nil, fmt.Errorf("map[]x > %w", err)
		}

		return typing.NewMap(keyTyp, elemTyp), nil

	case *types.Pointer:
		level := 1
		elem := t.Elem()
		for {
			if _, ok := elem.(*types.Pointer); !ok {
				break
			}

			elem = t.Elem()
			level++
		}

		elemTyp, err := parseType(pkgs, elem)
		if err != nil {
			return nil, fmt.Errorf("*%d > %w", level, err)
		}

		return typing.NewPointer(level, elemTyp), nil

	case *types.Named:
		name := t.Obj().Name()

		generics := make([]typing.Renderer, t.TypeParams().Len())
		for i := 0; i < t.TypeParams().Len(); i++ {
			typeParam := t.TypeParams().At(i)

			paramRenderer, err := parseType(pkgs, typeParam)
			if err != nil {
				return nil, fmt.Errorf("[%d] > %w", name, err)
			}

			generics[i] = paramRenderer
			if paramPkg := typeParam.Obj().Pkg(); paramPkg != nil {
				pkgs.Add(paramPkg.Path())
			}
		}

		var pkgPath string
		if pkg := t.Obj().Pkg(); pkg != nil {
			pkgPath = pkg.Path()
			pkgs.Add(pkgPath)
		}

		return typing.NewNamed(pkgPath, name, generics), nil
	}

	return nil, errors.New("unknown type")
}
