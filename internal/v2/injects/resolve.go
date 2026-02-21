package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/v2/typing"
)

func Resolve(key string, typ typing.Renderer) ResolveRenderer {
	return &resolveRenderer{
		Key:  key,
		Type: typ,
	}
}

type resolveRenderer struct {
	Key  string
	Type typing.Renderer
}

func (r *resolveRenderer) RenderResolve(ctx RenderContext, b *bytes.Buffer) {
	if key := r.Key; key != "" {
		renderedType := r.Type.Render(ctx, typing.DeclOp)
		b.WriteString("octo.ResolveNamed[" + renderedType + "](container, \"" + key + "\")")
		return
	}

	if typ := r.Type; typ.Kind() == typing.SliceKind {
		renderedType := typ.Child().Render(ctx, typing.DeclOp)
		b.WriteString("octo.ResolveAll[" + renderedType + "](container)")
		return
	}

	renderedType := r.Type.Render(ctx, typing.DeclOp)
	b.WriteString("octo.Resolve[" + renderedType + "](container)")
}
