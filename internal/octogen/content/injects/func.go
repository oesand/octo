package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
)

func Func(name string, injects []InjectRenderer) content.FileBlockRenderer {
	return &funcRenderer{
		Name:    name,
		Injects: injects,
	}
}

type funcRenderer struct {
	Name    string
	Injects []InjectRenderer
}

func (r *funcRenderer) Key() string {
	return "func:" + r.Name
}

func (r *funcRenderer) RenderFileBlock(ctx content.RenderContext, b *bytes.Buffer) {
	b.WriteString("func ")
	b.WriteString(r.Name)
	b.WriteString("(container *octo.Container) {\n")

	for _, i := range r.Injects {
		i.RenderInject(ctx, b)
	}

	b.WriteString("}")
}
