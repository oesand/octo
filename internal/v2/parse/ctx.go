package parse

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"iter"
	"strings"

	"github.com/oesand/octo/internal/v2/injects"
	"golang.org/x/tools/go/packages"
)

func newCtx(module, dir string) *parseContext {
	fileSet := token.NewFileSet()

	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedImports | packages.NeedDeps,
		BuildFlags: []string{
			"-tags", injects.BuildTag,
		},
		Fset: fileSet,
		Dir:  dir,
	}

	return &parseContext{
		cfg:    cfg,
		fset:   fileSet,
		module: module,
	}
}

type parseContext struct {
	cfg    *packages.Config
	fset   *token.FileSet
	module string

	warns []string
	errs  []error
}

func (ctx *parseContext) Packages() (iter.Seq[*packages.Package], error) {
	pkgs, err := packages.Load(ctx.cfg, "./...")
	if err != nil {
		return nil, err
	}
	return func(yield func(*packages.Package) bool) {
		for _, pkg := range pkgs {
			pkgPath := pkg.ID
			if !strings.HasPrefix(pkgPath, ctx.module) {
				continue
			}

			if !yield(pkg) {
				return
			}
		}
	}, nil
}

// HasBuildTag check if in file has `+build octogen` or 'go:build octogen' flag
func (ctx *parseContext) HasBuildTag(file *ast.File) bool {
	for _, commentGroup := range file.Comments {
		for _, c := range commentGroup.List {
			text := strings.Trim(c.Text, "// ")

			if (strings.HasPrefix(text, "+build") || strings.HasPrefix(text, "go:build")) &&
				strings.Contains(text, injects.BuildTag) {
				return true
			}
		}
	}
	return false
}

func (ctx *parseContext) GetOctogenAlias(file *ast.File) string {
	for _, im := range file.Imports {
		path := im.Path.Value[1 : len(im.Path.Value)-1] // strip quotes
		var alias string
		if im.Name != nil {
			alias = im.Name.Name
		} else {
			alias = path[strings.LastIndex(path, "/")+1:]
		}
		if path == injects.OctogenModule {
			return alias
		}
	}
	return ""
}

func (ctx *parseContext) formatMsg(pos token.Pos, format string, a ...any) string {
	ps := ctx.fset.Position(pos)

	formatted := fmt.Sprintf(format, a...)
	return fmt.Sprintf("%s:%d: %s", ps.Filename, ps.Line, formatted)
}

func (ctx *parseContext) AddWarn(pos token.Pos, format string, a ...any) {
	ctx.warns = append(ctx.warns, ctx.formatMsg(pos, format, a...))
}

func (ctx *parseContext) AddErr(pos token.Pos, format string, a ...any) {
	ctx.errs = append(ctx.errs, errors.New(ctx.formatMsg(pos, format, a...)))
}

func (ctx *parseContext) NoErrs() bool {
	return len(ctx.errs) == 0
}
