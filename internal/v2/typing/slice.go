package typing

import "strconv"

func NewSlice(size int64, child Renderer) Renderer {
	return &sliceRenderer{
		size:  size,
		child: child,
	}
}

type sliceRenderer struct {
	size  int64
	child Renderer
}

func (s *sliceRenderer) Kind() Kind {
	return SliceKind
}

func (s *sliceRenderer) Child() Renderer {
	return s.child
}

func (s *sliceRenderer) Render(ctx Context, op Operation) string {
	var prefix string
	if s.size > 0 {
		prefix = "[" + strconv.FormatInt(s.size, 10) + "]"
	} else {
		prefix = "[]"
	}
	return prefix + s.child.Render(ctx, op)
}
