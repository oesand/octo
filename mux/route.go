package mux

import (
	"fmt"
	"net/http"
)

type Route interface {
	Method() string
	Pattern() string
	Handler() http.Handler
	Flags() []any
}

func Get(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodGet, pattern, handler, flags...)
}

func Post(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodPost, pattern, handler, flags...)
}

func Put(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodPut, pattern, handler, flags...)
}

func Delete(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodDelete, pattern, handler, flags...)
}

func Options(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodOptions, pattern, handler, flags...)
}

func Head(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodHead, pattern, handler, flags...)
}

func Connect(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodConnect, pattern, handler, flags...)
}

func Patch(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodPatch, pattern, handler, flags...)
}

func Trace(pattern string, handler http.HandlerFunc, flags ...any) Route {
	return Handle(http.MethodTrace, pattern, handler, flags...)
}

func Handle(method, pattern string, handler http.Handler, flags ...any) Route {
	if pattern == "" {
		panic("mux: route pattern must have at least one character")
	}
	if pattern[0] != '/' {
		panic(fmt.Sprintf("mux: route pattern must starts with '/': %s", pattern))
	}

	if !IsValidMethod(method) {
		panic(fmt.Sprintf("mux: invalid http method: %s", pattern))
	}
	if handler == nil {
		panic(fmt.Sprintf("plow: nil handler: %s", pattern))
	}

	return &prefaceRoute{
		method:  method,
		pattern: pattern,
		handler: handler,
		flags:   flags,
	}
}

type prefaceRoute struct {
	method, pattern string
	handler         http.Handler
	flags           []any
}

func (r *prefaceRoute) Method() string {
	return r.method
}

func (r *prefaceRoute) Pattern() string {
	return r.pattern
}

func (r *prefaceRoute) Handler() http.Handler {
	return r.handler
}

func (r *prefaceRoute) Flags() []any {
	return r.flags
}
