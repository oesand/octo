package typing

import (
	"strings"
)

func NewNamed(pkg, name string, generics []Renderer) Renderer {
	return &namedRenderer{
		pkg:      pkg,
		name:     name,
		generics: generics,
	}
}

type namedRenderer struct {
	pkg      string
	name     string
	generics []Renderer
}

func (p *namedRenderer) Kind() Kind {
	return NamedKind
}

func (p *namedRenderer) Child() Renderer {
	return nil
}

func (p *namedRenderer) Render(ctx Context, _ Operation) string {
	var b strings.Builder

	if pkg := p.pkg; pkg != "" {
		alias := ctx.ImportAlias(pkg)
		if alias != "" {
			b.WriteString(alias)
			b.WriteRune('.')
		}
	}

	b.WriteString(p.name)

	if len(p.generics) > 0 {
		b.WriteRune('[')
		for i, generic := range p.generics {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(generic.Render(ctx, DeclOp))
		}
		b.WriteRune(']')
	}

	return b.String()
}
