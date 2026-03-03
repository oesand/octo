package injects

import (
	"bytes"
	"strings"

	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/typing"
)

func ResolveEmbedded(typeRender typing.Renderer, fields []ResolveRenderer, depth int) ResolveRenderer {
	return &resolveEmbeddedRenderer{
		typeRender: typeRender,
		fields:     fields,
		depth:      depth,
	}
}

type resolveEmbeddedRenderer struct {
	typeRender typing.Renderer
	fields     []ResolveRenderer
	depth      int
}

func (r *resolveEmbeddedRenderer) RenderResolve(ctx content.RenderContext, b *bytes.Buffer) {
	b.WriteString(r.typeRender.Render(ctx, typing.CallOp))
	b.WriteString("{\n")

	for _, renderer := range r.fields {
		b.WriteString("\t\t\t")
		b.WriteString(strings.Repeat("\t", r.depth))
		renderer.RenderResolve(ctx, b)
		b.WriteString(",\n")
	}

	b.WriteString("\t\t")
	b.WriteString(strings.Repeat("\t", r.depth))
	b.WriteRune('}')
}
