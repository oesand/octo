package flow

import (
	"context"
	"testing"
)

// TestTransactional_RegistersCallback verifies that Transactional registers a callback
// in the flow context.
func TestTransactional_RegistersCallback(t *testing.T) {
	// Create a flow context
	flowCtx := &triggerFlowCtx{
		transactionCallbacks: []TransactionCallback{},
		abort:                false,
	}

	// Create a context with the flow context value
	ctx := context.WithValue(context.Background(), triggerFlowCtxKey, flowCtx)

	// Create a test callback
	callbackCalled := false
	callback := func(ctx context.Context) error {
		callbackCalled = true
		return nil
	}

	// Register the callback
	Transactional(ctx, callback)

	// Verify the callback was registered
	if len(flowCtx.transactionCallbacks) != 1 {
		t.Errorf("expected 1 callback, got %d", len(flowCtx.transactionCallbacks))
	}

	// Execute the registered callback to verify it works
	if err := flowCtx.transactionCallbacks[0](ctx); err != nil {
		t.Errorf("callback execution failed: %v", err)
	}

	if !callbackCalled {
		t.Error("callback was not called")
	}
}

// TestAbort_SetsAbortFlag verifies that Abort sets the abort flag in the flow context.
func TestAbort_SetsAbortFlag(t *testing.T) {
	flowCtx := &triggerFlowCtx{
		transactionCallbacks: []TransactionCallback{},
		abort:                false,
	}

	ctx := context.WithValue(context.Background(), triggerFlowCtxKey, flowCtx)

	// Verify abort flag is initially false
	if flowCtx.abort {
		t.Error("expected abort flag to be false initially")
	}

	// Call Abort
	Abort(ctx)

	// Verify abort flag is now true
	if !flowCtx.abort {
		t.Error("expected abort flag to be true after Abort()")
	}
}
