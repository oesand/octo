package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/v2/typing"
)

func ReturnFunc(decl typing.Renderer, resolves []ResolveRenderer) ReturnRenderer {
	return &returnFuncRenderer{
		Decl:    decl,
		Entries: resolves,
	}
}

type returnFuncRenderer struct {
	Decl     typing.Renderer
	Generics []typing.Renderer
	Entries  []ResolveRenderer
}

func (r *returnFuncRenderer) RenderReturn(ctx RenderContext, b *bytes.Buffer) {
	b.WriteString("\t\treturn ")
	b.WriteString(r.Decl.Render(ctx, typing.DeclOp))
	b.WriteString("(\n")

	for _, renderer := range r.Entries {
		b.WriteString("\t\t\t")
		renderer.RenderResolve(ctx, b)
		b.WriteString(",\n")
	}

	b.WriteString("\t\t)\n")
}
