package backoff

import (
	"context"
	"github.com/oesand/octo/internal"
)

var settingsKey = internal.CtxKey{Key: "backoff/settings"}

func GetSettings(ctx context.Context) *BackOffSettings {
	return ctx.Value(settingsKey).(*BackOffSettings)
}

type BackOffSettings struct {
	attempt, maxAttempts int
}

func (c *BackOffSettings) Attempt() int {
	return c.attempt
}

func (c *BackOffSettings) MaxAttempts() int {
	return c.maxAttempts
}
