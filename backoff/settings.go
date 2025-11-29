package backoff

import (
	"context"
	"github.com/oesand/octo/internal"
)

var settingsKey = internal.CtxKey{Key: "backoff/settings"}

// GetSettings retrieves BackOffSettings from the provided context.
// It performs a safe type assertion and returns nil if the value is
// not present or not of type *BackOffSettings.
func GetSettings(ctx context.Context) *BackOffSettings {
	settings, _ := ctx.Value(settingsKey).(*BackOffSettings)
	return settings
}

// BackOffSettings contains configuration and state for backoff attempts.
type BackOffSettings struct {
	attempt, maxAttempts int
}

// Attempt returns the current retry attempt number.
func (c *BackOffSettings) Attempt() int {
	return c.attempt
}

// MaxAttempts returns the maximum number of retry attempts allowed.
func (c *BackOffSettings) MaxAttempts() int {
	return c.maxAttempts
}
