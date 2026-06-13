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

// TransactionCallback is executed after a flow step has finished and before
// the new state is persisted.
//
// It is registered with Transactional and passed to the manager when saving state.
type TransactionCallback func(ctx context.Context) error

type triggerFlowCtx struct {
	transactionCallbacks []TransactionCallback
	abort                bool
}

// Transactional registers a callback to be executed as part of the current
// flow trigger execution.
//
// Registered callbacks are stored in the active flow context and later
// dispatched when the flow state is saved.
func Transactional(ctx context.Context, callback TransactionCallback) {
	flowCtx := ctx.Value(triggerFlowCtxKey).(*triggerFlowCtx)
	flowCtx.transactionCallbacks = append(flowCtx.transactionCallbacks, callback)
}

// Abort marks the active flow execution as aborted.
//
// When called from inside a step handler, the flow engine will stop processing
// any further steps after the current one.
func Abort(ctx context.Context) {
	flowCtx := ctx.Value(triggerFlowCtxKey).(*triggerFlowCtx)
	flowCtx.abort = true
}
