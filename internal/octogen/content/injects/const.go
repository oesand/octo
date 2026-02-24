package injects

import (
	"bytes"

	"github.com/oesand/octo/internal/octogen/content"
)

type InjectRenderer interface {
	RenderInject(ctx content.RenderContext, b *bytes.Buffer)
}

type ReturnRenderer interface {
	RenderReturn(ctx content.RenderContext, b *bytes.Buffer)
}

type ResolveRenderer interface {
	RenderResolve(ctx content.RenderContext, b *bytes.Buffer)
}
