package injects

import (
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

func (r *resolveRenderer) RenderResolve(ctx RenderContext) string {
	if key := r.Key; key != "" {
		renderedType := r.Type.Render(ctx, typing.DeclOp)
		return "octo.ResolveNamed[" + renderedType + "](container, \"" + key + "\")"
	}

	if typ := r.Type; typ.Kind() == typing.SliceKind {
		renderedType := typ.Child().Render(ctx, typing.DeclOp)
		return "octo.ResolveAll[" + renderedType + "](container)"
	}

	renderedType := r.Type.Render(ctx, typing.DeclOp)
	return "octo.Resolve[" + renderedType + "](container)"
}
