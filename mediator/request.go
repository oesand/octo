package mediator

import (
	"context"

	"github.com/oesand/octo"
)

// Request [T any] interface declares a "Returns(T)" method
// Implementations of Request must define a Returns() method
// with a single parameter — the return type they declare.
//
// Example:
//
//	type MyQuery struct{}
//	func (MyQuery) Returns(MyResponse) {}
//
// The type system then encodes: MyQuery → MyResponse
type Request[T any] interface {
	Returns(T)
}

// RequestHandler is a generic interface for handling requests.
// It takes a request of type TRequest and returns a response of type TResponse (or an error).
//
// This follows the Mediator pattern where requests are decoupled from their handlers.
type RequestHandler[TRequest Request[TResponse], TResponse any] interface {
	// Request processes the input request and returns a response or error.
	Request(ctx context.Context, request TRequest) (TResponse, error)
}

// Send resolves a RequestHandler for the given request/response types from the container
// and calls its Request method. This is the entry point for executing a request.
func Send[TRequest Request[TResponse], TResponse any](
	manager *Manager,
	ctx context.Context,
	request TRequest,
) (TResponse, error) {
	manager.ensureInit()
	handler := octo.Resolve[RequestHandler[TRequest, TResponse]](manager.container)
	return handler.Request(ctx, request)
}
