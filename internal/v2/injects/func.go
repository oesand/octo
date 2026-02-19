package injects

import (
	"bytes"
)

type FuncRenderer struct {
	Name    string
	Injects []InjectRenderer
}

func (r *FuncRenderer) Render(ctx RenderContext, b *bytes.Buffer) string {
	b.WriteString("func ")
	b.WriteString(r.Name)
	b.WriteString("(container *octo.Container) {\n")

	for _, i := range r.Injects {
		i.RenderInject(ctx, b)
	}

	b.WriteString("}")

	return b.String()
}
