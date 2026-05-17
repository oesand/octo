package flow

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/oesand/octo/internal"
)

// NewUid returns a short unique identifier used to identify flow instances.
// It composes a small random prefix and the current millisecond timestamp.
func NewUid() string {
	sugar := make([]byte, 2)
	for i := 0; i < len(sugar); i++ {
		sugar[i] = byte(rand.Int63() & 0xff)
	}
	now := time.Now().UnixMilli()
	return fmt.Sprintf("%x_%x", sugar, now)
}

var triggerFlowCtxKey = internal.CtxKey{Key: "flow/triggerCtx"}

// TransactionCallback is a function type that represents a transactional callback
// within a flow. The callback receives a context and returns an error.
// Callbacks are typically used to perform side effects that should be executed
// as part of a transaction managed by the flow manager.
type TransactionCallback func(ctx context.Context) error

type triggerFlowCtx struct {
	transactionCallbacks []TransactionCallback
	abort                bool
}

// Transactional registers a TransactionCallback to be executed within the current
// flow's transaction. The callback will be invoked by the flow manager after the
// current flow event is processed. Multiple callbacks can be registered and will
// be executed in the order they were registered.
func Transactional(ctx context.Context, callback TransactionCallback) {
	flowCtx := ctx.Value(triggerFlowCtxKey).(*triggerFlowCtx)
	flowCtx.transactionCallbacks = append(flowCtx.transactionCallbacks, callback)
}

// Abort signals to the flow manager that the current flow should be aborted.
// Once called, the flow will not proceed with further event processing and
// any pending steps will be skipped. This is typically used for error handling
// or to terminate a flow based on certain conditions.
func Abort(ctx context.Context) {
	flowCtx := ctx.Value(triggerFlowCtxKey).(*triggerFlowCtx)
	flowCtx.abort = true
}
