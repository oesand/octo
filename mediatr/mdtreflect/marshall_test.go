package mdtreflect

import (
	"context"
	"encoding/json"
	"github.com/oesand/octo"
	"reflect"
	"testing"
)

// --- Test events ---
type EventX struct{ Name string }

// --- Test handlers ---
type HandlerX struct {
	Called bool
	Last   EventX
}

func (h *HandlerX) Notification(ctx context.Context, e EventX) {
	h.Called = true
	h.Last = e
}

// --- Tests ---

func TestMarshallEvent_NilEvent_ReturnsError(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}
	manager.autoRegister.Do(func() {})

	var ptr *EventX
	_, err := MarshallEvent(manager, ptr)
	if err == nil {
		t.Fatal("expected error for nil event, got nil")
	}
}

func TestMarshallEvent_Success(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}
	manager.autoRegister.Do(func() {})

	// Register the type
	decl := manager.registerEvent(reflect.TypeOf(EventX{}))
	decl.aliases = append(decl.aliases, "ex")

	ev := EventX{Name: "alice"}

	b, err := MarshallEvent(manager, ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Unmarshal wrappedEvent
	var wrapped wrappedEvent
	if err := json.Unmarshal(b, &wrapped); err != nil {
		t.Fatalf("cannot unmarshal wrapped: %v", err)
	}

	// Check aliases
	found := false
	for _, a := range wrapped.Aliases {
		if a == "ex" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected alias 'ex' in wrappedEvent")
	}

	// Check inner event
	var inner EventX
	if err := json.Unmarshal(wrapped.Event, &inner); err != nil {
		t.Fatalf("cannot unmarshal inner event: %v", err)
	}
	if inner.Name != "alice" {
		t.Fatalf("expected Name=alice, got %q", inner.Name)
	}
}

func TestUnmarshallAndPublish_InvalidJSON_ReturnsError(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}
	manager.autoRegister.Do(func() {})

	err := UnmarshallAndPublish(manager, context.Background(), []byte("{bad json"), false)
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func TestUnmarshallAndPublish_EmptyFields_ReturnsError(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}
	manager.autoRegister.Do(func() {})

	buf, _ := json.Marshal(wrappedEvent{Aliases: nil, Event: nil})

	err := UnmarshallAndPublish(manager, context.Background(), buf, false)
	if err == nil {
		t.Fatal("expected error for empty wrapped event")
	}
}

func TestUnmarshallAndPublish_HandlerCalled(t *testing.T) {
	container := octo.New()
	h := &HandlerX{}
	octo.InjectValue(container, h)

	manager := InjectManager(container)
	manager.registerEvent(reflect.TypeOf(EventX{}))

	ev := EventX{Name: "bob"}
	data, err := MarshallEvent(manager, ev)
	if err != nil {
		t.Fatalf("marshall failed: %v", err)
	}

	ctx := context.Background()
	err = UnmarshallAndPublish(manager, ctx, data, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !h.Called {
		t.Fatal("expected handler to be called")
	}
	if h.Last.Name != "bob" {
		t.Fatalf("handler received wrong data: %v", h.Last)
	}
}

func TestUnmarshallAndPublish_ManualWrappedEvent_WithHandlerX(t *testing.T) {
	container := octo.New()
	handler := &HandlerX{}
	octo.InjectValue(container, handler)

	manager := InjectManager(container)
	manager.registerEvent(reflect.TypeOf(EventX{}))

	// Manually encode a wrappedEvent
	ev := EventX{Name: "manual-test"}
	evJSON, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	wrapped := wrappedEvent{
		Aliases: []string{"ManualAlias"}, // a custom alias
		Event:   evJSON,
	}
	wrappedJSON, err := json.Marshal(wrapped)
	if err != nil {
		t.Fatalf("failed to marshal wrappedEvent: %v", err)
	}

	// Register alias so UnmarshallAndPublish can find the event type
	AliasEvent[EventX](manager, "ManualAlias")

	// Call UnmarshallAndPublish
	ctx := context.Background()
	err = UnmarshallAndPublish(manager, ctx, wrappedJSON, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that handler was called
	if !handler.Called {
		t.Fatal("expected HandlerX to be called")
	}
	if handler.Last.Name != "manual-test" {
		t.Fatalf("handler received wrong data: %+v", handler.Last)
	}
}

func TestUnmarshallAndPublish_EventNotFound_SkipTrue(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}
	manager.autoRegister.Do(func() {})

	wrapped := wrappedEvent{
		Aliases: []string{"unknown"},
		Event:   []byte(`{"Name":"x"}`),
	}
	buf, _ := json.Marshal(wrapped)

	err := UnmarshallAndPublish(manager, context.Background(), buf, true)
	if err != nil {
		t.Fatalf("expected nil when skipIfNF=true, got %v", err)
	}
}

func TestUnmarshallAndPublish_EventNotFound_SkipFalse(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}
	manager.autoRegister.Do(func() {})

	wrapped := wrappedEvent{
		Aliases: []string{"unknown"},
		Event:   []byte(`{"Name":"x"}`),
	}
	buf, _ := json.Marshal(wrapped)

	err := UnmarshallAndPublish(manager, context.Background(), buf, false)
	if err == nil {
		t.Fatal("expected error when skipIfNF=false, got nil")
	}
}
