package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/typing"
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

func (r *injectRenderer) RenderInject(ctx content.RenderContext, b *bytes.Buffer) {
	returningRenderer := r.Type.Render(ctx, typing.DeclOp)

	switch r.Key {
	case "":
		b.WriteString("\tocto.Inject(container, ")
	case "~":
		b.WriteString("\tocto.TryInject(container, ")
	default:
		b.WriteString("\tocto.InjectNamed(container, \"")
		b.WriteString(r.Key)
		b.WriteString("\", ")
	}

	b.WriteString("func(container *octo.Container) ")
	b.WriteString(returningRenderer)
	b.WriteString(" {\n")

	r.Return.RenderReturn(ctx, b)

	b.WriteString("\t})\n")
}
