package mediator

import (
	"context"
	"reflect"
)

// EventHandler defines a contract for handling notifications of type TEvent.
// Unlike requests, notifications do not return responses; instead, they are "fire-and-forget".
type EventHandler[TEvent any] interface {
	Notification(ctx context.Context, event TEvent) error
}

// Publish publishes a event to all registered NotificationHandlers.
// The event is sent to every matching handler until either:
//   - The context is cancelled, or
//   - All handlers have been executed.
func Publish[T any](
	manager *Manager,
	ctx context.Context,
	event T,
) error {
	manager.ensureInit()

	decls, has := manager.handlers[reflect.TypeFor[T]()]
	if !has {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan error, len(decls))
	for _, decl := range decls {
		handler := decl.Value().(EventHandler[T])
		go func() {
			results <- handler.Notification(ctx, event)
		}()
	}

	defer close(results)

	for i := 0; i < len(decls); i++ {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case err := <-results:
			if err != nil {
				return err
			}
		}
	}

	return nil
}
