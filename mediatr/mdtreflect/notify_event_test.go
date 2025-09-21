package mdtreflect

import (
	"context"
	"github.com/oesand/octo"
	"reflect"
	"sync/atomic"
	"testing"
)

// --- Sample notification types ---

type EventA struct{ ID int }
type EventB struct{ Name string }

// --- Handlers ---

type HandlerA struct {
	Called atomic.Bool
}

func (h *HandlerA) Notification(ctx context.Context, e EventA) {
	h.Called.Store(true)
}

type HandlerB struct {
	Called atomic.Bool
}

func (h *HandlerB) Notification(ctx context.Context, e EventB) {
	h.Called.Store(true)
}

// --- Tests ---

func TestNotificationEventTypes(t *testing.T) {
	c := octo.New()

	// Inject handlers
	octo.Inject(c, func(c *octo.Container) *HandlerA { return &HandlerA{} })
	octo.Inject(c, func(c *octo.Container) *HandlerB { return &HandlerB{} })

	types := make([]reflect.Type, 0)
	seq := notificationEventTypes(c)
	seq(func(t reflect.Type) bool {
		types = append(types, t)
		return true
	})

	if len(types) != 2 {
		t.Fatalf("expected 2 event types, got %d", len(types))
	}

	expected := []reflect.Type{reflect.TypeOf(EventA{}), reflect.TypeOf(EventB{})}
	for _, want := range expected {
		found := false
		for _, got := range types {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected type %v not found", want)
		}
	}
}

func TestNotifyEvents_CallsMatchingHandler(t *testing.T) {
	c := octo.New()
	hA := &HandlerA{}
	hB := &HandlerB{}

	octo.InjectNamedValue(c, "a", hA)
	octo.InjectNamedValue(c, "b", hB)

	ctx := context.Background()

	// Notify EventA
	evType := reflect.TypeOf(EventA{})
	evVal := reflect.ValueOf(EventA{ID: 1})
	notifyEvents(c, ctx, evType, evVal)

	if !hA.Called.Load() {
		t.Errorf("HandlerA was not called")
	}
	if hB.Called.Load() {
		t.Errorf("HandlerB should not be called")
	}

	// Reset and notify EventB
	hA.Called.Store(false)
	hB.Called.Store(false)
	evTypeB := reflect.TypeOf(EventB{})
	evValB := reflect.ValueOf(EventB{Name: "test"})
	notifyEvents(c, ctx, evTypeB, evValB)

	if hA.Called.Load() {
		t.Errorf("HandlerA should not be called")
	}
	if !hB.Called.Load() {
		t.Errorf("HandlerB was not called")
	}
}

func TestNotifyEvents_ContextCancelled(t *testing.T) {
	c := octo.New()
	h := &HandlerA{}
	octo.InjectNamedValue(c, "a", h)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	evType := reflect.TypeOf(EventA{})
	evVal := reflect.ValueOf(EventA{ID: 42})
	notifyEvents(c, ctx, evType, evVal)

	if h.Called.Load() {
		t.Errorf("Handler should not be called when context is canceled")
	}
}
