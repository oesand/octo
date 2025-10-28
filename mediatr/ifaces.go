package mediatr

import "context"

// EventHandler defines a contract for handling notifications of type TEvent.
// Unlike requests, notifications do not return responses; instead, they are "fire-and-forget".
type EventHandler[TEvent any] interface {
	Notification(ctx context.Context, event TEvent) error
}

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
