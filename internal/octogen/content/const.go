package content

import (
	"bytes"
	"iter"
	"os"

	"github.com/oesand/octo/internal/octogen/typing"
)

const (
	OctogenModule    = "github.com/oesand/octo/octogen"
	PrimitivesModule = "github.com/oesand/octo/pm"
	OctoModule       = "github.com/oesand/octo"
	BuildTag         = "octogen"
)

type PackageRenderer interface {
	Name() string
	Path() string
	Dir() string
	Render() []byte
	WriteFile(name string, mode os.FileMode) error
}

type RenderContext interface {
	typing.Context
	Import(pkg string)
	Imports() iter.Seq2[string, string]
}

type FileBlockRenderer interface {
	OriginalLine() int
	RenderFileBlock(ctx RenderContext, b *bytes.Buffer)
}
