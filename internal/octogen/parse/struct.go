package parse

import (
	"errors"
	"fmt"
	"go/types"
	"strings"

	"github.com/oesand/octo/internal/octogen/content/injects"
	"github.com/oesand/octo/internal/octogen/typing"
	"github.com/oesand/octo/pm"
)

func parseInjectStruct(originalLine int, key string, typ types.Type) (injects.InjectRenderer, []string, error) {
	isPtr, named, structType, ok := splitStructType(typ)
	if !ok {
		return nil, nil, errors.New("unexpected type, supported only struct, pointer struct")
	}

	imports := pm.Set[string]{}
	structRender, err := parseStructTypeRender(imports, named)
	if err != nil {
		return nil, nil, err
	}

	fields, err := parseStructFieldsRender(imports, structType, 0)
	if err != nil {
		return nil, nil, err
	}

	if isPtr {
		structRender = typing.NewPointer(1, structRender)
	}

	return injects.Inject(originalLine, key, structRender, injects.ReturnStruct(structRender, fields)), imports.Values(), nil
}

func parseStructTypeRender(imports pm.Set[string], named *types.Named) (typing.Renderer, error) {
	generics := make([]typing.Renderer, named.TypeArgs().Len())
	for i := 0; i < named.TypeArgs().Len(); i++ {
		generic, err := parseType(imports, named.TypeArgs().At(i))
		if err != nil {
			return nil, fmt.Errorf("struct generic[%d]: %w", i, err)
		}
		generics[i] = generic
	}

	structName := named.Obj().Name()
	var structPkg string
	if fp := named.Obj().Pkg(); fp != nil {
		structPkg = fp.Path()
		imports.Add(structPkg)
	}

	return typing.NewNamed(structPkg, structName, generics), nil
}

func parseStructFieldsRender(imports pm.Set[string], structType *types.Struct, embeddedDepth int) ([]injects.ResolveRenderer, error) {
	fields := make([]injects.ResolveRenderer, 0, structType.NumFields())
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if !field.Exported() {
			continue
		}

		fieldName := field.Name()

		if field.Embedded() {
			embeddedRenderer, err := parseEmbeddedFieldRenderer(imports, field.Type(), embeddedDepth+1)
			if err != nil {
				return nil, fmt.Errorf("embedded field '%s': %w", fieldName, err)
			}
			if embeddedRenderer != nil {
				fields = append(fields, injects.ResolveField(fieldName, embeddedRenderer))
			}
			continue
		}

		fieldRender, err := parseType(imports, field.Type())
		if err != nil {
			return nil, fmt.Errorf("struct field '%s': %w", fieldName, err)
		}

		fieldTags := structType.Tag(i)
		var resolveKey string
		if idx := strings.Index(fieldTags, `key:"`); idx >= 0 {
			rest := fieldTags[idx+5:]
			if end := strings.Index(rest, `"`); end > 0 {
				resolveKey = rest[:end]
			}
		}

		fields = append(fields, injects.ResolveField(fieldName, injects.Resolve(resolveKey, fieldRender)))
	}

	return fields, nil
}

func parseEmbeddedFieldRenderer(imports pm.Set[string], typ types.Type, depth int) (injects.ResolveRenderer, error) {
	isPtr, named, structType, ok := splitStructType(typ)
	if !ok || isPtr {
		return nil, nil
	}

	structRender, err := parseStructTypeRender(imports, named)
	if err != nil {
		return nil, err
	}

	fields, err := parseStructFieldsRender(imports, structType, depth)
	if err != nil {
		return nil, err
	}

	return injects.ResolveEmbedded(structRender, fields, depth), nil
}
