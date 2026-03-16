package mux

import (
	"iter"
	"net/http"
)

type Route interface {
	Method() string
	Pattern() string
	Handler() http.Handler
	Flags() iter.Seq[string]
}
