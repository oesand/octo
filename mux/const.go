package mux

import "net/http"

func IsValidMethod(method string) bool {
	return method == http.MethodGet ||
		method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodDelete ||
		method == http.MethodOptions ||
		method == http.MethodHead ||
		method == http.MethodConnect ||
		method == http.MethodPatch ||
		method == http.MethodTrace
}
