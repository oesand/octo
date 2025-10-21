package mdtreflect

import (
	"context"
	"github.com/oesand/octo"
	"github.com/oesand/octo/internal/prim"
	"iter"
	"reflect"
)

func notificationEventTypes(container *octo.Container) iter.Seq[reflect.Type] {
	ctxType := reflect.TypeFor[context.Context]()

	return func(yield func(reflect.Type) bool) {
		var seen prim.Set[reflect.Type]
		decls := octo.ResolveInjections(container)
		for decl := range decls {
			dval := reflect.ValueOf(decl.Value())

			method := dval.MethodByName("Notification")
			if !method.IsValid() {
				continue
			}

			methodType := method.Type()
			if methodType.NumIn() != 2 || !methodType.In(0).AssignableTo(ctxType) || methodType.NumOut() != 0 {
				continue
			}

			eventType := methodType.In(1)
			if !seen.Has(eventType) {
				if !yield(eventType) {
					return
				}

				seen.Add(eventType)
			}
		}
	}
}

func notifyEvents(container *octo.Container, ctx context.Context, evType reflect.Type, evVal reflect.Value) {
	ctxType := reflect.TypeFor[context.Context]()
	ctxValue := reflect.ValueOf(ctx)

	decls := octo.ResolveInjections(container)
	for decl := range decls {
		if ctx.Err() != nil {
			break
		}

		dval := reflect.ValueOf(decl.Value())

		method := dval.MethodByName("Notification")
		if !method.IsValid() {
			continue
		}

		methodType := method.Type()
		if methodType.NumIn() != 2 || !methodType.In(0).AssignableTo(ctxType) || methodType.NumOut() != 0 {
			continue
		}

		eventType := methodType.In(1)
		if eventType != evType {
			continue
		}

		go method.Call([]reflect.Value{ctxValue, evVal})
	}
}
