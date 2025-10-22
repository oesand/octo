package mediatr

import (
	"context"
	"github.com/oesand/octo"
	"sync"
)

// EventHandler defines a contract for handling notifications of type TEvent.
// Unlike requests, notifications do not return responses; instead, they are "fire-and-forget".
type EventHandler[TEvent any] interface {
	Notification(ctx context.Context, event TEvent)
}

// Publish publishes a notification of type TEvent to all registered NotificationHandlers.
// The notification is sent to every matching handler until either:
//   - The context is cancelled, or
//   - All handlers have been executed.
func Publish[TEvent any](
	container *octo.Container,
	ctx context.Context,
	event TEvent,
) {
	injects := octo.ResolveInjections(container)
	var wg sync.WaitGroup
	for decl := range injects {
		if !octo.DeclOfType[EventHandler[TEvent]](decl) {
			continue
		}

		// stop if context was cancelled
		if ctx.Err() != nil {
			break
		}

		handler := decl.Value().(EventHandler[TEvent])

		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.Notification(ctx, event)
		}()
	}
	wg.Wait()
}
