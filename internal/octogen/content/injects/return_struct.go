package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/typing"
)

func ReturnStruct(typeRender typing.Renderer, fields []ResolveRenderer) ReturnRenderer {
	return &returnStructRenderer{
		typeRender: typeRender,
		fields:     fields,
	}
}

type returnStructRenderer struct {
	typeRender typing.Renderer
	fields     []ResolveRenderer
}

func (r *returnStructRenderer) RenderReturn(ctx content.RenderContext, b *bytes.Buffer) {
	b.WriteString("\t\treturn ")
	b.WriteString(r.typeRender.Render(ctx, typing.CallOp))
	b.WriteString("{\n")

	for _, renderer := range r.fields {
		b.WriteString("\t\t\t")
		renderer.RenderResolve(ctx, b)
		b.WriteString(",\n")
	}

	b.WriteString("\t\t}\n")
}
