package injects

import (
	"bytes"
	"fmt"
)

func Pkg(name, dir string, ctx RenderContext, funcs []*FuncRenderer) *PkgRenderer {
	return &PkgRenderer{
		name:  name,
		dir:   dir,
		ctx:   ctx,
		funcs: funcs,
	}
}

type PkgRenderer struct {
	name, dir string

	ctx   RenderContext
	funcs []*FuncRenderer
}

func (r *PkgRenderer) Name() string {
	return r.name
}

func (r *PkgRenderer) Dir() string {
	return r.dir
}

func (r *PkgRenderer) Render() []byte {
	var b bytes.Buffer

	b.WriteString("package ")
	b.WriteString(r.name)
	b.WriteString("\n\n")

	b.WriteString("import (\n")
	b.WriteString("\t\"github.com/oesand/octo\"\n")

	for alias, imp := range r.ctx.Imports() {
		b.WriteString(fmt.Sprintf("\t%s \"%s\"\n", alias, imp))
	}

	b.WriteString(")\n")

	for _, fn := range r.funcs {
		b.WriteRune('\n')
		fn.Render(r.ctx, &b)
	}

	b.WriteRune('\n')

	return b.Bytes()
}
