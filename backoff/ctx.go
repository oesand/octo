package backoff

import (
	"context"
	"github.com/oesand/octo/internal"
)

var ctxKey = internal.CtxKey{Key: "backoff/context"}

func GetContext(ctx context.Context) *BackOffContext {
	return ctx.Value(ctxKey).(*BackOffContext)
}

type BackOffContext struct {
	attempt, maxAttempts int
}

func (c *BackOffContext) Attempt() int {
	return c.attempt
}

func (c *BackOffContext) MaxAttempts() int {
	return c.maxAttempts
}
