package typing

import "strings"

func NewPointer(level int, child Renderer) Renderer {
	return &pointerRenderer{
		level: level,
		child: child,
	}
}

type pointerRenderer struct {
	level int
	child Renderer
}

func (p *pointerRenderer) Kind() Kind {
	return PointerKind
}

func (p *pointerRenderer) Child() Renderer {
	return p.child
}

func (p *pointerRenderer) Render(ctx Context, operation Operation) string {
	var prefix string
	switch operation {
	case DeclOp:
		prefix = strings.Repeat("*", p.level)
	case CallOp:
		prefix = strings.Repeat("&", p.level)
	}

	return prefix + p.child.Render(ctx, operation)
}
