package mediator_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/oesand/octo"
	"github.com/oesand/octo/mediator"
)

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

func TestPublish_SingleHandler(t *testing.T) {
	container := octo.New()
	h1 := &EventHandlerX{}
	octo.InjectValue(container, h1)
	manager := mediator.Inject(container)

	ev := EventX{Name: "test"}
	err := mediator.Publish(manager, context.Background(), ev)
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
	manager := mediator.Inject(container)

	err := mediator.Publish(manager, context.Background(), EventX{"multi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !h1.Called.Load() || !h2.Called.Load() {
		t.Fatal("expected both handlers called")
	}
}

func TestPublish_MultipleHandlersCancelled(t *testing.T) {
	container := octo.New()
	h1, h2 := &BlockHandler{}, &EventHandlerX{}
	octo.InjectValue(container, h1)
	octo.InjectValue(container, h2)
	manager := mediator.Inject(container)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := mediator.Publish(manager, ctx, EventX{"multi"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error: %v", err)
	}

	if !h1.Called.Load() || !h2.Called.Load() {
		t.Fatal("expected both handlers called")
	}
}
