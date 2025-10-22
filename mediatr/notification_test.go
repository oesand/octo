package mediatr

import (
	"context"
	"github.com/oesand/octo"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- Example notifications and handlers ---

type UserCreated struct {
	Username string
}

// Logs notification to a slice
type LoggingHandler struct {
	mu      sync.Mutex
	entries []string
}

func (h *LoggingHandler) Notification(ctx context.Context, n UserCreated) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, "log:"+n.Username)
}

// Stores notification into an audit slice
type AuditHandler struct {
	mu    sync.Mutex
	audit []string
}

func (h *AuditHandler) Notification(ctx context.Context, n UserCreated) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.audit = append(h.audit, "audit:"+n.Username)
}

// Blocks until context is cancelled
type BlockingHandler struct {
	called atomic.Bool
}

func (h *BlockingHandler) Notification(ctx context.Context, n UserCreated) {
	h.called.Store(true)
	<-ctx.Done()
}

// Simple handler that records whether it was called
type SecondHandler struct {
	called atomic.Bool
}

func (h *SecondHandler) Notification(ctx context.Context, n UserCreated) {
	h.called.Store(true)
}

// --- Tests ---

func TestNotify_SingleHandler(t *testing.T) {
	c := octo.New()
	handler := &LoggingHandler{}
	octo.InjectValue(c, handler)

	Publish[UserCreated](c, context.Background(), UserCreated{Username: "alice"})

	handler.mu.Lock()
	defer handler.mu.Unlock()
	if len(handler.entries) != 1 || handler.entries[0] != "log:alice" {
		t.Fatalf("expected [log:alice], got %#v", handler.entries)
	}
}

func TestNotify_MultipleHandlers(t *testing.T) {
	c := octo.New()
	log := &LoggingHandler{}
	audit := &AuditHandler{}

	octo.InjectValue(c, log)
	octo.InjectValue(c, audit)

	Publish[UserCreated](c, context.Background(), UserCreated{Username: "bob"})

	log.mu.Lock()
	audit.mu.Lock()
	defer log.mu.Unlock()
	defer audit.mu.Unlock()

	if len(log.entries) != 1 || log.entries[0] != "log:bob" {
		t.Errorf("expected log handler to get log:bob, got %#v", log.entries)
	}
	if len(audit.audit) != 1 || audit.audit[0] != "audit:bob" {
		t.Errorf("expected audit handler to get audit:bob, got %#v", audit.audit)
	}
}

func TestNotify_HandlerInject(t *testing.T) {
	c := octo.New()
	handler := &LoggingHandler{}
	octo.Inject(c, func(_ *octo.Container) *LoggingHandler {
		return handler
	})

	Publish[UserCreated](c, context.Background(), UserCreated{Username: "alice"})

	handler.mu.Lock()
	defer handler.mu.Unlock()
	if len(handler.entries) != 1 || handler.entries[0] != "log:alice" {
		t.Fatalf("expected [log:alice], got %#v", handler.entries)
	}
}

func TestNotify_ParallelHandling(t *testing.T) {
	c := octo.New()

	blocking := &BlockingHandler{}
	second := &SecondHandler{}

	octo.InjectValue(c, blocking)
	octo.InjectValue(c, second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go Publish[UserCreated](c, ctx, UserCreated{Username: "carol"})

	time.Sleep(10 * time.Millisecond)

	if !blocking.called.Load() {
		t.Error("expected blocking handler to be called")
	}
	if !second.called.Load() {
		t.Error("expected second handler NOT to be called after cancel")
	}
}
