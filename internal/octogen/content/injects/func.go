package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
)

func Func(line int, name string, injects []InjectRenderer) content.FileBlockRenderer {
	return &funcRenderer{
		line:    line,
		name:    name,
		injects: injects,
	}
}

type funcRenderer struct {
	line    int
	name    string
	injects []InjectRenderer
}

func (r *funcRenderer) OriginalLine() int {
	return r.line
}

func (r *funcRenderer) RenderFileBlock(ctx content.RenderContext, b *bytes.Buffer) {
	b.WriteString("func ")
	b.WriteString(r.name)
	b.WriteString("(container *octo.Container) {\n")

	for _, i := range r.injects {
		i.RenderInject(ctx, b)
	}

	b.WriteString("}")
}
