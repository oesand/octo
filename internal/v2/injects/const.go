package injects

import (
	"bytes"
	"iter"

	"github.com/oesand/octo/internal/v2/typing"
)

const (
	OctogenModule = "github.com/oesand/octo/octogen"
	BuildTag      = "octogen"
	OctoModule    = "github.com/oesand/octo"
	OctoAlias     = "octo"
)

type RenderContext interface {
	typing.Context
	Import(pkg string)
	Imports() iter.Seq2[string, string]
}

type InjectRenderer interface {
	RenderInject(ctx RenderContext, b *bytes.Buffer)
}

type ReturnRenderer interface {
	RenderReturn(ctx RenderContext, b *bytes.Buffer)
}

type ResolveRenderer interface {
	RenderResolve(ctx RenderContext) string
}
