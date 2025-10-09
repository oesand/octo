package mediatr

import (
	"context"
	"github.com/oesand/octo"
)

// ShortRequest [T any] interface declares a "Returns(T)" method
// Implementations of ShortRequest must define a Returns() method
// with a single parameter — the return type they declare.
//
// Example:
//
//	type MyQuery struct{}
//	func (MyQuery) Returns(MyResponse) {}
//
// The type system then encodes: MyQuery → MyResponse
type ShortRequest[T any] interface {
	Returns(T)
}

// SendShort is a shorthand helper for sending strongly-typed requests that
// implement the ShortRequest[TResponse] interface.
func SendShort[TRequest ShortRequest[TResponse], TResponse any](
	container *octo.Container,
	ctx context.Context,
	request TRequest,
) (TResponse, error) {
	return Send[TRequest, TResponse](container, ctx, request)
}
