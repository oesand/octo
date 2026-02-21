package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/v2/typing"
)

func ReturnStruct(typ typing.Renderer, fields map[string]ResolveRenderer) ReturnRenderer {
	return &returnStructRenderer{
		Type:   typ,
		Fields: fields,
	}
}

type returnStructRenderer struct {
	Type   typing.Renderer
	Fields map[string]ResolveRenderer
}

func (r *returnStructRenderer) RenderReturn(ctx RenderContext, b *bytes.Buffer) {
	b.WriteString("\t\treturn ")
	b.WriteString(r.Type.Render(ctx, typing.CallOp))
	b.WriteString("{\n")

	for name, renderer := range r.Fields {
		b.WriteString("\t\t\t")
		b.WriteString(name)
		b.WriteRune(':')
		renderer.RenderResolve(ctx, b)
		b.WriteString(",\n")
	}

	b.WriteString("\t\t}\n")
}
