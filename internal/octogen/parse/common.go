package parse

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/oesand/octo/internal/octogen/typing"
	"github.com/oesand/octo/pm"
)

func parseType(imports pm.Set[string], typ types.Type) (typing.Renderer, error) {
	switch t := typ.(type) {
	case *types.Basic:
		return typing.NewNamed("", t.Name(), nil), nil

	case *types.Array:
		elemTyp, err := parseType(imports, t.Elem())
		if err != nil {
			return nil, fmt.Errorf("[%d] > %w", t.Len(), err)
		}
		return typing.NewSlice(t.Len(), elemTyp), nil

	case *types.Slice:
		elemTyp, err := parseType(imports, t.Elem())
		if err != nil {
			return nil, fmt.Errorf("[] > %w", err)
		}
		return typing.NewSlice(0, elemTyp), nil

	case *types.Map:
		keyTyp, err := parseType(imports, t.Key())
		if err != nil {
			return nil, fmt.Errorf("map[x] > %w", err)
		}

		elemTyp, err := parseType(imports, t.Elem())
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

		elemTyp, err := parseType(imports, elem)
		if err != nil {
			return nil, fmt.Errorf("%s > %w", strings.Repeat("*", level), err)
		}

		return typing.NewPointer(level, elemTyp), nil

	case *types.Named:
		name := t.Obj().Name()

		generics := make([]typing.Renderer, t.TypeArgs().Len())
		for i := 0; i < t.TypeArgs().Len(); i++ {
			typeParam := t.TypeArgs().At(i)

			paramRenderer, err := parseType(imports, typeParam)
			if err != nil {
				return nil, fmt.Errorf("%s > %w", name, err)
			}

			generics[i] = paramRenderer
		}

		var pkgPath string
		if pkg := t.Obj().Pkg(); pkg != nil {
			pkgPath = pkg.Path()
			imports.Add(pkgPath)
		}

		return typing.NewNamed(pkgPath, name, generics), nil
	}

	return nil, fmt.Errorf("not supported type: %s", typ.String())
}

func splitStructType(typ types.Type) (bool, *types.Named, *types.Struct, bool) {
	ptr, isPtr := typ.(*types.Pointer)
	if isPtr {
		typ = ptr.Elem()
	}

	named, ok := typ.(*types.Named)
	if !ok {
		return false, nil, nil, false
	}

	structType, ok := named.Underlying().(*types.Struct)
	if !ok {
		return false, nil, nil, false
	}

	return isPtr, named, structType, true
}
