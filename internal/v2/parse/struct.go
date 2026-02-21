package parse

import (
	"errors"
	"fmt"
	"go/types"
	"strings"

	"github.com/oesand/octo/internal/v2/injects"
	"github.com/oesand/octo/internal/v2/typing"
	"github.com/oesand/octo/pm"
)

func parseInjectStruct(key string, typ types.Type) (injects.InjectRenderer, []string, error) {
	isPtr, named, structType, err := splitStructType(typ)
	if err != nil {
		return nil, nil, err
	}

	var pkgs pm.Set[string]
	structRender, err := parseStructTypeRender(pkgs, named)
	if err != nil {
		return nil, nil, err
	}

	fields, err := parseStructFieldsRender(pkgs, structType)
	if err != nil {
		return nil, nil, err
	}

	if isPtr {
		structRender = typing.NewPointer(1, structRender)
	}

	return injects.Inject(key, structRender, injects.ReturnStruct(structRender, fields)), pkgs.Values(), nil
}

func parseStructTypeRender(pkgs pm.Set[string], named *types.Named) (typing.Renderer, error) {
	generics := make([]typing.Renderer, named.TypeParams().Len())
	for i := 0; i < named.TypeParams().Len(); i++ {
		generic, err := parseType(pkgs, named.TypeParams().At(i))
		if err != nil {
			return nil, fmt.Errorf("injecting struct generic[%d]: %w", i, err)
		}
		generics[i] = generic
	}

	structName := named.Obj().Name()
	var structPkg string
	if fp := named.Obj().Pkg(); fp != nil {
		structPkg = fp.Path()
		pkgs.Add(structPkg)
	}

	return typing.NewNamed(structPkg, structName, generics), nil
}

func parseStructFieldsRender(pkgs pm.Set[string], structType *types.Struct) (map[string]injects.ResolveRenderer, error) {
	fields := make(map[string]injects.ResolveRenderer, structType.NumFields())
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if !field.Exported() {
			continue
		}

		if field.Embedded() {
			// TODO: fix tabbing for embedded fields
			embeddedRenderer, err := parseEmbeddedFieldRenderer(pkgs, field.Type())
			if err != nil {
				// TODO: add prefix
				return nil, err
			}
			if embeddedRenderer != nil {
				fields[field.Name()] = embeddedRenderer
			}
			continue
		}

		fieldRender, err := parseType(pkgs, field.Type())
		if err != nil {
			return nil, fmt.Errorf("injecting struct field[%d]: %w", i, err)
		}

		fieldTags := structType.Tag(i)
		var resolveKey string
		if idx := strings.Index(fieldTags, `key:"`); idx >= 0 {
			rest := fieldTags[idx+5:]
			if end := strings.Index(rest, `"`); end > 0 {
				resolveKey = rest[:end]
			}
		}

		fields[field.Name()] = injects.Resolve(resolveKey, fieldRender)
	}

	return fields, nil
}

func parseEmbeddedFieldRenderer(pkgs pm.Set[string], typ types.Type) (injects.ResolveRenderer, error) {
	isPtr, named, structType, err := splitStructType(typ)
	if err != nil || isPtr {
		return nil, nil
	}

	structRender, err := parseStructTypeRender(pkgs, named)

	fields, err := parseStructFieldsRender(pkgs, structType)
	if err != nil {
		return nil, err
	}

	return injects.ResolveEmbedded(structRender, fields), nil
}

func splitStructType(typ types.Type) (bool, *types.Named, *types.Struct, error) {
	ptr, isPtr := typ.(*types.Pointer)
	if isPtr {
		typ = ptr.Elem()
	}

	const unexpectedTypeErr = "unexpected type, supported only struct, pointer struct"

	named, ok := typ.(*types.Named)
	if !ok {
		return false, nil, nil, errors.New(unexpectedTypeErr)
	}

	structType, ok := named.Underlying().(*types.Struct)
	if !ok {
		return false, nil, nil, errors.New(unexpectedTypeErr)
	}

	return isPtr, named, structType, nil
}
