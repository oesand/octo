package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
	"github.com/oesand/octo/internal/octogen/typing"
)

func Inject(line int, key string, returnType typing.Renderer, returnRender ReturnRenderer) InjectRenderer {
	return &injectRenderer{
		line:         line,
		key:          key,
		returnType:   returnType,
		returnRender: returnRender,
	}
}

type injectRenderer struct {
	line         int
	key          string
	returnType   typing.Renderer
	returnRender ReturnRenderer
}

func (r *injectRenderer) OriginalLine() int {
	return r.line
}

func (r *injectRenderer) RenderInject(ctx content.RenderContext, b *bytes.Buffer) {
	returningRenderer := r.returnType.Render(ctx, typing.DeclOp)

	switch r.key {
	case "":
		b.WriteString("\tocto.Inject(container, ")
	case "~":
		b.WriteString("\tocto.TryInject(container, ")
	default:
		b.WriteString("\tocto.InjectNamed(container, \"")
		b.WriteString(r.key)
		b.WriteString("\", ")
	}

	b.WriteString("func(container *octo.Container) ")
	b.WriteString(returningRenderer)
	b.WriteString(" {\n")

	r.returnRender.RenderReturn(ctx, b)

	b.WriteString("\t})\n")
}
