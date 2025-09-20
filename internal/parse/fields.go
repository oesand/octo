package parse

import (
	"fmt"
	"go/ast"
)

func parseFields(fieldsList []*ast.Field) ([]*FieldInfo, error) {

	for _, field := range fieldsList {
		for _, name := range field.Names {
			fmt.Println("  Field:", name.Name)
		}
	}
	return nil, nil
}
