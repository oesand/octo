package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/v2/typing"
)

func ResolveEmbedded(typ typing.Renderer, fields map[string]ResolveRenderer) ResolveRenderer {
	return &resolveEmbeddedRenderer{
		Type:   typ,
		Fields: fields,
	}
}

type resolveEmbeddedRenderer struct {
	Type   typing.Renderer
	Fields map[string]ResolveRenderer
}

func (r *resolveEmbeddedRenderer) RenderResolve(ctx RenderContext, b *bytes.Buffer) {
	b.WriteString(r.Type.Render(ctx, typing.CallOp))
	b.WriteString("{\n")

	for name, renderer := range r.Fields {
		b.WriteString("\t\t\t")
		b.WriteString(name)
		b.WriteRune(':')
		renderer.RenderResolve(ctx, b)
		b.WriteString(",\n")
	}

	b.WriteString("\t\t}")
}
