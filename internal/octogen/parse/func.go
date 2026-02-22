package parse

import (
	"errors"
	"fmt"
	"go/types"

	"github.com/oesand/octo/internal/octogen/injects"
	"github.com/oesand/octo/internal/octogen/typing"
	"github.com/oesand/octo/pm"
)

func parseInjectFunc(key string, funcObj *types.Func) (injects.InjectRenderer, []string, error) {
	funcSig := funcObj.Signature()

	if funcSig.Results().Len() != 1 {
		return nil, nil, errors.New("function should return only one result")
	}

	imports := pm.Set[string]{}
	resType := funcSig.Results().At(0)

	returned, err := parseType(imports, resType.Type())
	if err != nil {
		return nil, nil, fmt.Errorf("function return: %w", err)
	}

	generics := make([]typing.Renderer, funcSig.TypeParams().Len())
	for i := 0; i < funcSig.TypeParams().Len(); i++ {
		generic, err := parseType(imports, funcSig.TypeParams().At(i))
		if err != nil {
			return nil, nil, fmt.Errorf("function generic[%d]: %w", i, err)
		}
		generics[i] = generic
	}

	params := make([]injects.ResolveRenderer, funcSig.Params().Len())
	for i := 0; i < funcSig.Params().Len(); i++ {
		prm := funcSig.Params().At(i)
		param, err := parseType(imports, prm.Type())
		if err != nil {
			return nil, nil, fmt.Errorf("function param '%s': %w", prm.Name(), err)
		}
		params[i] = injects.Resolve("", param)
	}

	funcName := funcObj.Name()
	var funcPkg string
	if fp := funcObj.Pkg(); fp != nil {
		funcPkg = fp.Path()
		imports.Add(funcPkg)
	}

	funcDecl := typing.NewNamed(funcPkg, funcName, generics)

	return injects.Inject(key, returned, injects.ReturnFunc(funcDecl, params)), imports.Values(), nil
}
