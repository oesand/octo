package mediatr

import (
	"context"
	"reflect"

	"github.com/oesand/octo"
)

type handleEvent func(context.Context, any) error
type handleRequest func(context.Context, any) (any, error)

var (
	ctxType   = reflect.TypeFor[context.Context]()
	errorType = reflect.TypeFor[error]()
)

func verifyEventHandler(decl octo.ServiceDeclaration) (reflect.Type, handleEvent, bool) {
	if method, ok := decl.Type().MethodByName("Notification"); ok &&
		method.Type.NumIn() == 3 && method.Type.In(1).AssignableTo(ctxType) &&
		method.Type.NumOut() == 1 && method.Type.Out(0).AssignableTo(errorType) {

		eventType := method.Type.In(2)
		handler := func(ctx context.Context, ev any) error {
			handler := decl.Value()
			results := method.Func.Call([]reflect.Value{
				reflect.ValueOf(handler),
				reflect.ValueOf(ctx),
				reflect.ValueOf(ev),
			})

			errVal := results[0].Interface()
			if errVal != nil {
				return errVal.(error)
			}
			return nil
		}

		return eventType, handler, true
	}
	return nil, nil, false
}

func verifyRequestHandler(decl octo.ServiceDeclaration) (reflect.Type, handleRequest, bool) {
	if method, ok := decl.Type().MethodByName("Request"); ok &&
		method.Type.NumIn() == 3 && method.Type.In(1).AssignableTo(ctxType) &&
		method.Type.NumOut() == 2 && method.Type.Out(1).AssignableTo(errorType) {

		requestType := method.Type.In(2)
		handler := func(ctx context.Context, req any) (any, error) {
			handler := decl.Value()
			results := method.Func.Call([]reflect.Value{
				reflect.ValueOf(handler),
				reflect.ValueOf(ctx),
				reflect.ValueOf(req),
			})

			errVal := results[1].Interface()
			if errVal != nil {
				return nil, errVal.(error)
			}

			return results[0].Interface(), nil
		}

		return requestType, handler, true
	}
	return nil, nil, false
}
