package mdtreflect

import (
	"context"
	"github.com/oesand/octo"
	"reflect"
	"sync"
	"testing"
	"time"
)

// === Test event types ===
type EventA struct{ V string }
type EventB struct{ N int }

// === Handlers ===

// Valid handler for EventA
type HandlerA struct{ Called bool }

func (h *HandlerA) Notification(ctx context.Context, e EventA) {
	h.Called = true
}

// Valid handler for EventB
type HandlerB struct{ Called bool }

func (h *HandlerB) Notification(ctx context.Context, e EventB) {
	h.Called = true
}

// Custom handler to count calls
type CtxHandler struct {
	mu    *sync.Mutex
	calls *int
}

func (h *CtxHandler) Notification(ctx context.Context, e EventA) {
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.calls++
}

// === Tests ===

func TestNotificationEventTypes_NoHandlers(t *testing.T) {
	container := octo.New()
	manager := InjectManager(container)

	var got []reflect.Type
	for ev := range notificationEventTypes(manager.container) {
		got = append(got, ev)
	}

	if len(got) != 0 {
		t.Fatalf("expected no event types, got %v", got)
	}
}

func TestNotificationEventTypes_ValidHandler(t *testing.T) {
	container := octo.New()
	octo.InjectValue(container, &HandlerA{})
	manager := InjectManager(container)

	got := []reflect.Type{}
	for ev := range notificationEventTypes(manager.container) {
		got = append(got, ev)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 event type, got %d", len(got))
	}
	if got[0] != reflect.TypeOf(EventA{}) {
		t.Fatalf("expected EventA type, got %v", got[0])
	}
}

func TestNotificationEventTypes_DeduplicatesTypes(t *testing.T) {
	container := octo.New()
	octo.InjectValue(container, &HandlerA{})
	octo.InjectValue(container, &HandlerA{}) // duplicate
	manager := InjectManager(container)

	count := 0
	for _ = range notificationEventTypes(manager.container) {
		count++
	}

	if count != 1 {
		t.Fatalf("expected 1 unique event type, got %d", count)
	}
}

func TestNotificationEventTypes_MultipleTypes(t *testing.T) {
	container := octo.New()
	octo.InjectValue(container, &HandlerA{})
	octo.InjectValue(container, &HandlerB{})
	manager := InjectManager(container)

	got := map[reflect.Type]bool{}
	for ev := range notificationEventTypes(manager.container) {
		got[ev] = true
	}

	if !got[reflect.TypeOf(EventA{})] {
		t.Error("expected EventA type discovered")
	}
	if !got[reflect.TypeOf(EventB{})] {
		t.Error("expected EventB type discovered")
	}
}

func TestNotifyEvents_CallsMatchingHandler(t *testing.T) {
	container := octo.New()
	h := &HandlerA{}
	octo.InjectValue(container, h)

	ctx := context.Background()
	evVal := reflect.ValueOf(EventA{V: "x"})
	notifyEvents(container, ctx, evVal.Type(), evVal)

	time.Sleep(20 * time.Millisecond)

	if !h.Called {
		t.Fatal("expected handler to be called")
	}
}

func TestNotifyEvents_SkipsWrongType(t *testing.T) {
	container := octo.New()
	h := &HandlerA{}
	octo.InjectValue(container, h)

	ctx := context.Background()
	evVal := reflect.ValueOf(EventB{N: 42})
	notifyEvents(container, ctx, evVal.Type(), evVal)

	time.Sleep(20 * time.Millisecond)

	if h.Called {
		t.Fatal("expected handler not to be called")
	}
}

func TestNotifyEvents_MultipleHandlers(t *testing.T) {
	container := octo.New()
	h1 := &HandlerA{}
	h2 := &HandlerA{}
	octo.InjectValue(container, h1)
	octo.InjectValue(container, h2)

	ctx := context.Background()
	evVal := reflect.ValueOf(EventA{V: "multi"})
	notifyEvents(container, ctx, evVal.Type(), evVal)

	time.Sleep(20 * time.Millisecond)

	if !h1.Called || !h2.Called {
		t.Fatal("expected both handlers called")
	}
}

func TestNotifyEvents_StopsOnContextCancel(t *testing.T) {
	container := octo.New()

	var mu sync.Mutex
	calls := 0

	octo.InjectValue(container, &CtxHandler{mu: &mu, calls: &calls})
	octo.InjectValue(container, &CtxHandler{mu: &mu, calls: &calls})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	evVal := reflect.ValueOf(EventA{V: "stop"})
	notifyEvents(container, ctx, evVal.Type(), evVal)

	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if calls != 0 {
		t.Fatalf("expected no calls due to cancelled context, got %d", calls)
	}
}
