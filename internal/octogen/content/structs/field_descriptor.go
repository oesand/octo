package structs

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/typing"
)

func Descriptor(line int, name string, structType typing.Renderer, fields map[string]typing.Renderer) content.FileBlockRenderer {
	return &fieldsDescriptor{
		line:       line,
		name:       name,
		structType: structType,
		fields:     fields,
	}
}

type fieldsDescriptor struct {
	line       int
	name       string
	structType typing.Renderer
	fields     map[string]typing.Renderer
}

func (f *fieldsDescriptor) OriginalLine() int {
	return f.line
}

func (f *fieldsDescriptor) RenderFileBlock(ctx content.RenderContext, b *bytes.Buffer) {
	b.WriteString("var ")
	b.WriteString(f.name)
	b.WriteString(" = struct {\n")

	pmAlias := ctx.ImportAlias(content.PrimitivesModule)
	structType := f.structType.Render(ctx, typing.DeclOp)

	names := slices.Collect(maps.Keys(f.fields))
	sort.Strings(names)

	var d bytes.Buffer
	for _, name := range names {
		fieldType := f.fields[name].Render(ctx, typing.DeclOp)

		fieldDesc := fmt.Sprintf("%s.FieldDescriptor[%s, %s]", pmAlias, structType, fieldType)

		b.WriteRune('\t')
		b.WriteString(name)
		b.WriteRune(' ')
		b.WriteString(fieldDesc)
		b.WriteRune('\n')

		d.WriteRune('\t')
		d.WriteString(name)
		d.WriteRune(':')
		d.WriteString(fieldDesc)
		d.WriteString("{\n")

		d.WriteString("\t\tName: \"")
		d.WriteString(name)
		d.WriteString("\",\n")

		d.WriteString("\t\tValue: func(s *")
		d.WriteString(structType)
		d.WriteString(") ")
		d.WriteString(fieldType)
		d.WriteString(" {\n")

		d.WriteString("\t\t\treturn s.")
		d.WriteString(name)
		d.WriteRune('\n')
		d.WriteString("\t\t},\n")
		d.WriteString("\t},\n")
	}

	b.WriteString("}{\n")
	_, _ = d.WriteTo(b)
	b.WriteRune('}')
}
