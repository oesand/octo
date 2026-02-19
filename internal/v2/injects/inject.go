package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/v2/typing"
)

func Inject(key string, typ typing.Renderer, ret ReturnRenderer) InjectRenderer {
	return &injectRenderer{
		Key:    key,
		Type:   typ,
		Return: ret,
	}
}

type injectRenderer struct {
	Key    string
	Type   typing.Renderer
	Return ReturnRenderer
}

func (r *injectRenderer) RenderInject(ctx RenderContext, b *bytes.Buffer) {
	returningRenderer := r.Type.Render(ctx, typing.DeclOp)

	if r.Key != "" {
		b.WriteString("\tocto.InjectNamed(container, \"")
		b.WriteString(r.Key)
		b.WriteString("\", ")
	} else {
		b.WriteString("\tocto.Inject(container, ")
	}

	b.WriteString("func(container *octo.Container) ")
	b.WriteString(returningRenderer)
	b.WriteString(" {\n")

	r.Return.RenderReturn(ctx, b)

	b.WriteString("\t})\n")
}
