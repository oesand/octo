package content

import (
	"bytes"
	"iter"

	"github.com/oesand/octo/internal/octogen/typing"
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

type FileBlockRenderer interface {
	Key() string
	RenderFileBlock(ctx RenderContext, b *bytes.Buffer)
}
