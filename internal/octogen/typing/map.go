package typing

func NewMap(key, child Renderer) Renderer {
	return &mapRenderer{
		key:   key,
		child: child,
	}
}

type mapRenderer struct {
	key, child Renderer
}

func (s *mapRenderer) Kind() Kind {
	return MapKind
}

func (s *mapRenderer) Child() Renderer {
	return s.child
}

func (s *mapRenderer) Render(ctx Context, _ Operation) string {
	return "map[" + s.key.Render(ctx, DeclOp) + "]" + s.child.Render(ctx, DeclOp)
}
