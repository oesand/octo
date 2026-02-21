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
			return nil, nil, fmt.Errorf("injecting function param[%d]: %w", i, err)
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
