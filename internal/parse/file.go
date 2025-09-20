package parse

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func trimComment(text string) string {
	// remove comment markers
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	text = strings.TrimSpace(text)
	return text
}

func isInjectComment(text string) bool {
	return strings.HasPrefix(text, "@inject")
}

func ParseFile(fileSet *token.FileSet, absolutePath string) (*ParsedFile, error) {
	node, err := parser.ParseFile(fileSet, absolutePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, decl := range node.Decls {
		switch d := decl.(type) {

		// Functions
		case *ast.FuncDecl:
			var injectComment string

			// Check inline comments within function declaration range
			for _, c := range node.Comments {
				if c.Pos() >= d.Pos() && c.End() <= d.End() {
					for _, ci := range c.List {
						comment := trimComment(ci.Text)
						if isInjectComment(comment) {
							injectComment = comment
							break
						}
					}
				}
			}

			// TODO : Check if not struct method

			if injectComment != "" {
				fmt.Println("Function to inject:", d.Name.Name)
			}

		// Structs
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				var injectComment string

				// Check inline comments within struct declaration range
				for _, c := range node.Comments {
					if c.Pos() >= d.Pos() && c.End() <= d.End() {
						for _, ci := range c.List {
							comment := trimComment(ci.Text)
							if isInjectComment(comment) {
								injectComment = comment
								break
							}
						}
					}
				}

				if injectComment != "" {
					fmt.Println("Struct to inject:", typeSpec.Name.Name)
					_, err = parseFields(structType.Fields.List)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	return nil, err
}
