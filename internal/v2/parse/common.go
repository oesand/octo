package parse

import (
	"errors"
	"fmt"
	"go/types"

	"github.com/oesand/octo/internal/v2/typing"
	"github.com/oesand/octo/pm"
)

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
