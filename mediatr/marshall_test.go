package mediatr

import (
	"context"
	"encoding/json"
	"github.com/oesand/octo"
	"reflect"
	"testing"
)

func TestAliasEvent_Success(t *testing.T) {
	manager := Inject(octo.New())

	AliasEvent[EventX](manager, "x1", "x2")

	decl := manager.events["x1"]
	if decl == nil || len(decl.aliases) < 3 {
		t.Fatalf("expected aliases registered: %+v", decl.aliases)
	}
}

func TestAliasEvent_Panics(t *testing.T) {
	manager := Inject(octo.New())

	assertPanic := func(f func(), msg string) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic: %s", msg)
			}
		}()
		f()
	}

	assertPanic(func() { AliasEvent[EventX](manager) }, "empty aliases")
	assertPanic(func() { AliasEvent[EventX](manager, AbsoluteEventName(reflect.TypeFor[EventX]())) }, "alias = type name")
	AliasEvent[EventX](manager, "dup")
	assertPanic(func() { AliasEvent[EventX](manager, "dup") }, "duplicate alias")
}

func TestMarshallEvent_Success(t *testing.T) {
	manager := Inject(octo.New())

	AliasEvent[EventX](manager, "x1", "x2")

	ev := EventX{Name: "marshal"}
	buf, err := MarshallEvent(manager, ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wrapped marshallEvent
	if err = json.Unmarshal(buf, &wrapped); err != nil {
		t.Fatal(err)
	}

	if len(wrapped.Aliases) == 0 || len(wrapped.Event) == 0 {
		t.Fatal("expected non-empty wrapped data")
	}
}

func TestMarshallEvent_NotRegistered(t *testing.T) {
	manager := Inject(octo.New())

	_, err := MarshallEvent(manager, EventX{})
	if err == nil {
		t.Fatal("expected error for unregistered event")
	}
}

func TestUnmarshallAndPublish_MultipleHandlers(t *testing.T) {
	container := octo.New()
	h1, h2 := &EventHandlerX{}, &EventHandlerX{}
	octo.InjectValue(container, h1)
	octo.InjectValue(container, h2)

	manager := Inject(container)

	ev := EventX{Name: "multi"}

	// Marshal the event into wrapped JSON
	buf, err := MarshallEvent(manager, ev)
	if err != nil {
		t.Fatalf("MarshallEvent failed: %v", err)
	}

	// Unmarshal and publish — should call both handlers via Publish()
	err = UnmarshallAndPublish(manager, context.Background(), buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !h1.Called.Load() || !h2.Called.Load() {
		t.Fatal("expected both handlers to be called")
	}
}

func TestUnmarshallAndPublish_InvalidJSON(t *testing.T) {
	manager := &Manager{container: octo.New()}
	err := UnmarshallAndPublish(manager, context.Background(), []byte(`{invalid`))
	if err == nil {
		t.Fatal("expected JSON error")
	}
}

func TestUnmarshallAndPublish_EmptyWrapped(t *testing.T) {
	manager := Inject(octo.New())

	buf, _ := json.Marshal(&marshallEvent{})
	err := UnmarshallAndPublish(manager, context.Background(), buf)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestUnmarshallAndPublish_UnknownAlias(t *testing.T) {
	manager := Inject(octo.New())

	buf, _ := json.Marshal(&marshallEvent{Aliases: []string{"unknown"}, Event: []byte(`{}`)})
	err := UnmarshallAndPublish(manager, context.Background(), buf)
	if err == nil {
		t.Fatal("expected unknown alias error")
	}
}
