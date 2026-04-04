package mux

import (
	"fmt"
	"iter"
	"net/http"
	"sort"
)

type Middleware func(http.Handler) http.Handler

type Mux struct {
	routes   map[string][]*muxRoute
	notFound http.Handler
}

type muxRoute struct {
	RoutePattern
	method  string
	handler http.Handler
	flags   []any
}

func (m *muxRoute) Method() string {
	return m.method
}

func (m *muxRoute) Pattern() string {
	return m.Original
}

func (m *muxRoute) Handler() http.Handler {
	return m.handler
}

func (m *muxRoute) Flags() []any {
	return m.flags
}

func New(routes ...Route) *Mux {
	mux := &Mux{
		routes: make(map[string][]*muxRoute),
	}

	for _, route := range routes {
		pattern := route.Pattern()
		compiledPattern, err := ParseRoutePattern(pattern)
		if err != nil {
			panic(fmt.Sprintf("cannot compile route pattern: %s", pattern))
		}
		method := route.Method()
		mux.routes[method] = append(mux.routes[method], &muxRoute{
			RoutePattern: *compiledPattern,
			handler:      route.Handler(),
			flags:        route.Flags(),
		})
	}

	for _, muxRoutes := range mux.routes {
		sort.Slice(muxRoutes, func(i, j int) bool {
			if muxRoutes[i].Depth == muxRoutes[j].Depth {
				return len(muxRoutes[i].ParamNames) < len(muxRoutes[j].ParamNames)
			}
			return muxRoutes[i].Depth > muxRoutes[j].Depth
		})
	}
	return mux
}

func (m *Mux) Routes() iter.Seq[Route] {
	return func(yield func(Route) bool) {
		for _, routes := range m.routes {
			for _, route := range routes {
				if !yield(route) {
					return
				}
			}
		}
	}
}

func (m *Mux) NotFound(handler http.Handler) {
	m.notFound = handler
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func (m *Mux) handlePath(w http.ResponseWriter, r *http.Request) {
	routes, ok := m.routes[r.Method]
	if !ok {
		for _, route := range routes {
			matched, params := route.Match(r.URL.Path)
			if !matched {
				continue
			}

			r.ParseForm()
		}
	}

	if nf := m.notFound; nf != nil {
		nf.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}
