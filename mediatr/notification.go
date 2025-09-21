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

// InjectNotification registers a NotificationHandler for the given TNotification type
// in the provided container. Multiple handlers of the same notification type can be registered.
func InjectNotification[TNotification any](
	container *octo.Container,
	provider octo.Provider[NotificationHandler[TNotification]],
) {
	octo.Inject(container, provider)
}

// Notify publishes a notification of type TNotification to all registered NotificationHandlers.
// The notification is sent to every matching handler until either:
//   - The context is cancelled, or
//   - All handlers have been executed.
func Notify[TNotification any](
	container *octo.Container,
	ctx context.Context,
	notification TNotification,
) {
	decls := octo.ResolveInjections(container)
	for decl := range decls {
		// stop if context was cancelled
		if ctx.Err() != nil {
			break
		}

		// filter only handlers for this notification type
		if !octo.DeclOfType[NotificationHandler[TNotification]](decl) {
			continue
		}

		handler := decl.Value().(NotificationHandler[TNotification])
		handler.Notification(ctx, notification)
	}
}
