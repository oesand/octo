package mux

import "fmt"

func Routes(routes ...Route) []Route {
	return routes
}

func PrefixRoutes(prefix string, routes ...Route) []Route {
	if len(prefix) < 2 {
		panic("mux: router prefix must have at least two characters")
	}
	if prefix[0] != '/' {
		panic(fmt.Sprintf("mux: router prefix must starts with '/': %s", prefix))
	}

	var prefixRoutes []Route
	for _, route := range routes {
		pattern := prefix + route.Pattern()
		prefixRoute := Handle(route.Method(), pattern, route.Handler(), route.Flags()...)
		prefixRoutes = append(prefixRoutes, prefixRoute)
	}

	return prefixRoutes
}
