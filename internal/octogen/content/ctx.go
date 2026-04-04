package content

import (
	"fmt"
	"iter"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/oesand/octo/internal"
)

func NewCtx(pkgPath string) RenderContext {
	return &renderCtx{
		pkgPath: pkgPath,
	}
}

type renderCtx struct {
	pkgPath string

	aliases internal.Set[string]
	imports map[string]*importData
}

type importData struct {
	path           string
	alias          string
	redundantAlias bool
}

func (r *renderCtx) Import(pkg string) {
	if pkg == "" || pkg == r.pkgPath {
		return
	}

	if r.imports == nil {
		r.aliases = internal.Set[string]{}
		r.imports = map[string]*importData{}
	} else if _, ok := r.imports[pkg]; ok {
		return
	}

	base := pkg[strings.LastIndexByte(pkg, '/')+1:]
	for i := 0; ; i++ {
		alias := base
		if i > 0 {
			alias = base + strconv.Itoa(i)
		}
		if r.aliases.Has(alias) {
			continue
		}

		r.aliases.Add(alias)
		r.imports[pkg] = &importData{
			path:           pkg,
			alias:          alias,
			redundantAlias: i == 0,
		}
		break
	}
}

func (r *renderCtx) ImportAlias(pkg string) string {
	if pkg == "" || pkg == r.pkgPath {
		return ""
	}

	if len(r.imports) == 0 {
		panic("no imports")
	}

	data, ok := r.imports[pkg]
	if !ok {
		panic(fmt.Sprintf("no alias found for %s", pkg))
	}

	return data.alias
}

func (r *renderCtx) Imports() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		if len(r.imports) == 0 {
			return
		}

		paths := slices.Collect(maps.Keys(r.imports))
		sort.Strings(paths)

		for _, path := range paths {
			var alias string
			if data := r.imports[path]; !data.redundantAlias {
				alias = data.alias
			}

			if !yield(alias, path) {
				break
			}
		}
	}
}
