package injects

import (
	"bytes"
	"strings"

	"github.com/oesand/octo/internal/v2/typing"
)

func ResolveEmbedded(typ typing.Renderer, fields map[string]ResolveRenderer, depth int) ResolveRenderer {
	return &resolveEmbeddedRenderer{
		Type:   typ,
		Fields: fields,
		Depth:  depth,
	}
}

type resolveEmbeddedRenderer struct {
	Type   typing.Renderer
	Fields map[string]ResolveRenderer
	Depth  int
}

func (r *resolveEmbeddedRenderer) RenderResolve(ctx RenderContext, b *bytes.Buffer) {
	b.WriteString(r.Type.Render(ctx, typing.CallOp))
	b.WriteString("{\n")

	for name, renderer := range r.Fields {
		b.WriteString("\t\t\t")
		b.WriteString(strings.Repeat("\t", r.Depth))
		b.WriteString(name)
		b.WriteRune(':')
		renderer.RenderResolve(ctx, b)
		b.WriteString(",\n")
	}

	b.WriteString("\t\t")
	b.WriteString(strings.Repeat("\t", r.Depth))
	b.WriteRune('}')
}
