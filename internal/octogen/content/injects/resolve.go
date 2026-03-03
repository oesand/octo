package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/typing"
)

func Resolve(key string, typeRender typing.Renderer) ResolveRenderer {
	return &resolveRenderer{
		key:        key,
		typeRender: typeRender,
	}
}

type resolveRenderer struct {
	key        string
	typeRender typing.Renderer
}

func (r *resolveRenderer) RenderResolve(ctx content.RenderContext, b *bytes.Buffer) {
	renderer := r.typeRender
	if key := r.key; key != "" {
		renderedType := renderer.Render(ctx, typing.DeclOp)
		b.WriteString("octo.ResolveNamed[" + renderedType + "](container, \"" + key + "\")")
		return
	}

	if renderer.Kind() == typing.SliceKind {
		renderedType := renderer.Child().Render(ctx, typing.DeclOp)
		b.WriteString("octo.ResolveAll[" + renderedType + "](container)")
		return
	}

	renderedType := renderer.Render(ctx, typing.DeclOp)
	b.WriteString("octo.Resolve[" + renderedType + "](container)")
}
