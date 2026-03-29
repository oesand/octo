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

type MassEventHandler interface {
	EventTypes() []reflect.Type
	Handle(ctx context.Context, event any) error
}

// Publish publishes a event to all registered NotificationHandlers.
// The event is sent to every matching handler until either:
//   - The context is canceled, or
//   - All handlers have been executed.
func Publish(
	manager *Manager,
	ctx context.Context,
	event any,
) error {
	manager.ensureInit()

	handlers, has := manager.handlers[reflect.TypeOf(event)]
	if !has {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan error, len(handlers))
	for _, handle := range handlers {
		go func() {
			results <- handle(ctx, event)
		}()
	}

	defer close(results)

	for i := 0; i < len(handlers); i++ {
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
