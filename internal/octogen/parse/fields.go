package parse

import (
	"errors"
	"fmt"
	"go/types"

	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/content/structs"
	"github.com/oesand/octo/internal/octogen/typing"
)

func parseStructFields(originalLine int, descName string, typ types.Type) (content.FileBlockRenderer, []string, error) {
	named, structType, ok := splitStructType(typ)
	if !ok {
		return nil, nil, errors.New("unexpected type, supported only struct")
	}

	imports := internal.Set[string]{}
	structRender, err := parseStructTypeRender(imports, named)
	if err != nil {
		return nil, nil, err
	}

	fields := make(map[string]typing.Renderer, structType.NumFields())
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if !field.Exported() || field.Embedded() {
			continue
		}

		fieldName := field.Name()
		fieldRender, err := parseType(imports, field.Type())
		if err != nil {
			return nil, nil, fmt.Errorf("struct field '%s': %w", fieldName, err)
		}

		fields[fieldName] = fieldRender
	}

	return structs.Descriptor(originalLine, descName, structRender, fields), imports.Values(), nil
}
