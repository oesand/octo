package mediatr

import (
	"context"
	"github.com/oesand/octo"
)

// NotificationHandler defines a contract for handling notifications of type TNotification.
// Unlike requests, notifications do not return responses; instead, they are "fire-and-forget".
type NotificationHandler[TNotification any] interface {
	Notification(ctx context.Context, notification TNotification)
}

// Publish publishes a notification of type TNotification to all registered NotificationHandlers.
// The notification is sent to every matching handler until either:
//   - The context is cancelled, or
//   - All handlers have been executed.
func Publish[TNotification any](
	container *octo.Container,
	ctx context.Context,
	notification TNotification,
) {
	handlers := octo.ResolveOfType[NotificationHandler[TNotification]](container)
	for handler := range handlers {
		// stop if context was cancelled
		if ctx.Err() != nil {
			break
		}

		handler.Notification(ctx, notification)
	}
}
