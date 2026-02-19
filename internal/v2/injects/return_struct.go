package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/v2/typing"
)

func ReturnStruct(typ typing.Renderer, fields map[string]ResolveRenderer) ReturnRenderer {
	return &returnStructRenderer{
		Type:    typ,
		Entries: fields,
	}
}

type returnStructRenderer struct {
	Type    typing.Renderer
	Entries map[string]ResolveRenderer
}

func (r *returnStructRenderer) RenderReturn(ctx RenderContext, b *bytes.Buffer) {
	b.WriteString("\t\treturn ")
	b.WriteString(r.Type.Render(ctx, typing.CallOp))
	b.WriteRune('{')

	for name, renderer := range r.Entries {
		b.WriteString("\t\t\t")
		b.WriteString(name)
		b.WriteRune(':')
		b.WriteString(renderer.RenderResolve(ctx))
		b.WriteRune(',')
	}

	b.WriteString("\t\t}")
}
