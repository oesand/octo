package injects

import (
	"iter"
	"strconv"
	"strings"
)

func NewCtx(pkgPath string) RenderContext {
	return &renderCtx{
		pkgPath: pkgPath,
	}
}

type renderCtx struct {
	pkgPath       string
	imports       map[string]string // [alias]: import
	importAliases map[string]string // [import]: alias
}

func (r *renderCtx) Import(pkg string) {
	if pkg == "" || pkg == r.pkgPath || pkg == OctoModule {
		return
	}

	if len(r.imports) == 0 {
		r.imports = map[string]string{}
		r.importAliases = map[string]string{}
	} else if _, ok := r.importAliases[pkg]; ok {
		return
	}

	base := pkg[strings.LastIndexByte(pkg, '/')+1:]
	for i := 0; ; i++ {
		alias := base
		if i > 0 || alias == OctoAlias {
			alias = base + strconv.Itoa(i)
		}
		if _, ok := r.imports[alias]; ok {
			continue
		}

		r.imports[alias] = pkg
		r.importAliases[pkg] = alias
		break
	}
}

func (r *renderCtx) ImportAlias(pkg string) string {
	if pkg == "" || pkg == r.pkgPath {
		return ""
	}

	if pkg == OctoModule {
		return OctoAlias
	}

	if len(r.importAliases) == 0 {
		panic("no imports")
	}

	return r.importAliases[pkg]
}

func (r *renderCtx) Imports() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		if len(r.imports) == 0 {
			return
		}
		for alias, imp := range r.imports {
			if !yield(alias, imp) {
				break
			}
		}
	}
}
