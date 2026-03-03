package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
)

func ResolveField(name string, child ResolveRenderer) ResolveRenderer {
	return &resolveFieldRenderer{
		name:  name,
		child: child,
	}
}

type resolveFieldRenderer struct {
	name  string
	child ResolveRenderer
}

func (r *resolveFieldRenderer) RenderResolve(ctx content.RenderContext, b *bytes.Buffer) {
	b.WriteString(r.name)
	b.WriteRune(':')
	r.child.RenderResolve(ctx, b)
}
