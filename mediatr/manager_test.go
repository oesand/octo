package mediatr

import (
	"context"
	"errors"
	"github.com/oesand/octo"
	"sync/atomic"
	"testing"
	"time"
)

// --- Test Fixtures ---

type EventX struct {
	Name string
}

type EventHandlerX struct {
	Called atomic.Bool
}

func (h *EventHandlerX) Notification(ctx context.Context, e EventX) error {
	h.Called.Store(true)
	return nil
}

type BlockHandler struct {
	Called atomic.Bool
}

func (h *BlockHandler) Notification(ctx context.Context, e EventX) error {
	h.Called.Store(true)
	time.Sleep(1 * time.Second)
	return nil
}

type RequestX struct {
	Value int
}

func (RequestX) Returns(ResponseX) {}

type ResponseX struct {
	Result int
}

type RequestHandlerX struct {
	Called atomic.Bool
}

func (h *RequestHandlerX) Request(ctx context.Context, req RequestX) (ResponseX, error) {
	h.Called.Store(true)
	return ResponseX{Result: req.Value * 2}, nil
}

// --- Tests ---

func TestInject_Singleton(t *testing.T) {
	container := octo.New()
	m1 := Inject(container)
	m2 := Inject(container)

	if m1 != m2 {
		t.Fatal("expected same manager instance")
	}

	container = octo.New()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when injecting manually created Manager")
		}
	}()

	octo.InjectValue(container, &Manager{})

	Inject(container)
}

func TestPublish_SingleHandler(t *testing.T) {
	container := octo.New()
	h1 := &EventHandlerX{}
	octo.InjectValue(container, h1)
	manager := Inject(container)

	ev := EventX{Name: "test"}
	err := Publish(manager, context.Background(), ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !h1.Called.Load() {
		t.Fatal("expected both handlers called")
	}
}

func TestPublish_MultipleHandlers(t *testing.T) {
	container := octo.New()
	h1, h2 := &EventHandlerX{}, &EventHandlerX{}
	octo.InjectValue(container, h1)
	octo.InjectValue(container, h2)
	manager := Inject(container)

	err := Publish(manager, context.Background(), EventX{"multi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !h1.Called.Load() || !h2.Called.Load() {
		t.Fatal("expected both handlers called")
	}
}

func TestSend_RequestHandler(t *testing.T) {
	container := octo.New()
	h := &RequestHandlerX{}
	octo.InjectValue(container, h)
	manager := Inject(container)
	manager.doInit()

	resp, err := Send(manager, context.Background(), RequestX{Value: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Result != 6 {
		t.Fatalf("expected 6, got %d", resp.Result)
	}
	if !h.Called.Load() {
		t.Fatal("expected both handlers called")
	}
}

func TestPublish_MultipleHandlersCancelled(t *testing.T) {
	container := octo.New()
	h1, h2 := &BlockHandler{}, &EventHandlerX{}
	octo.InjectValue(container, h1)
	octo.InjectValue(container, h2)
	manager := Inject(container)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := Publish(manager, ctx, EventX{"multi"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error: %v", err)
	}

	if !h1.Called.Load() || !h2.Called.Load() {
		t.Fatal("expected both handlers called")
	}
}

func TestSend_NoHandler(t *testing.T) {
	container := octo.New()
	manager := Inject(container)
	manager.doInit()

	_, err := Send[RequestX, ResponseX](manager, context.Background(), RequestX{Value: 1})
	if err == nil {
		t.Fatal("expected error for missing handler")
	}
}
