package mediatr

import (
	"context"
	"github.com/oesand/octo"
)

// RequestHandler is a generic interface for handling requests.
// It takes a request of type TRequest and returns a response of type TResponse (or an error).
//
// This follows the Mediator pattern where requests are decoupled from their handlers.
type RequestHandler[TRequest any, TResponse any] interface {
	// Request processes the input request and returns a response or error.
	Request(ctx context.Context, request TRequest) (TResponse, error)
}

// Send resolves a RequestHandler for the given request/response types from the container
// and calls its Request method. This is the entry point for executing a request.
func Send[TRequest any, TResponse any](
	container *octo.Container,
	ctx context.Context,
	request TRequest,
) (TResponse, error) {
	handler := octo.Resolve[RequestHandler[TRequest, TResponse]](container)
	return handler.Request(ctx, request)
}
