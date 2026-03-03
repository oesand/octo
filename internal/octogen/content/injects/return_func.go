package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/typing"
)

func ReturnFunc(decl typing.Renderer, entries []ResolveRenderer) ReturnRenderer {
	return &returnFuncRenderer{
		decl:    decl,
		entries: entries,
	}
}

type returnFuncRenderer struct {
	decl    typing.Renderer
	entries []ResolveRenderer
}

func (r *returnFuncRenderer) RenderReturn(ctx content.RenderContext, b *bytes.Buffer) {
	b.WriteString("\t\treturn ")
	b.WriteString(r.decl.Render(ctx, typing.DeclOp))
	b.WriteString("(\n")

	for _, renderer := range r.entries {
		b.WriteString("\t\t\t")
		renderer.RenderResolve(ctx, b)
		b.WriteString(",\n")
	}

	b.WriteString("\t\t)\n")
}
