package content

import (
	"bytes"
	"iter"
	"os"

	"github.com/oesand/octo/internal/octogen/typing"
)

const (
	OctogenModule = "github.com/oesand/octo/octogen"
	BuildTag      = "octogen"
	OctoModule    = "github.com/oesand/octo"
	OctoAlias     = "octo"
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
