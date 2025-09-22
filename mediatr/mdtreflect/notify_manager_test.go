package mdtreflect

import (
	"context"
	"encoding/json"
	"github.com/oesand/octo"
	"github.com/oesand/octo/mediatr"
	"reflect"
	"testing"
)

// Dummy event types
type EventX struct{ ID int }
type EventY struct{ Name string }

// Dummy handlers
type HandlerX struct{ Called bool }

func (h *HandlerX) Notification(ctx context.Context, e EventX) { h.Called = true }

type HandlerY struct{ Called bool }

func (h *HandlerY) Notification(ctx context.Context, e EventY) { h.Called = true }

// ---------------- InjectNotifyManager Tests ----------------

// Test singleton behavior: if manager already registered, returns the same instance
func TestInjectNotifyManager_Singleton(t *testing.T) {
	c := octo.New()

	manager1 := InjectNotifyManager(c)
	manager2 := InjectNotifyManager(c)

	if manager1 != manager2 {
		t.Errorf("expected singleton manager, got different instances")
	}
}

// Test new manager registration and automatic event discovery
func TestInjectNotifyManager_NewRegistration(t *testing.T) {
	c := octo.New()

	// Inject a dummy notification via mediatr
	h := &HandlerX{}
	mediatr.InjectNotification[EventX](c, func(c *octo.Container) mediatr.NotificationHandler[EventX] { return h })

	manager := InjectNotifyManager(c)
	if manager == nil {
		t.Fatal("expected non-nil manager")
	}

	absName := AbsoluteTypeName(reflect.TypeOf(EventX{}))
	decl, ok := manager.events[absName]
	if !ok {
		t.Fatalf("expected EventX to be automatically registered")
	}
	if decl.typ != reflect.TypeOf(EventX{}) {
		t.Errorf("expected decl.typ to match EventX")
	}
	if len(decl.aliases) != 1 || decl.aliases[0] != absName {
		t.Errorf("expected decl.aliases to contain absolute name")
	}
}

// Test multiple event types automatically registered
func TestInjectNotifyManager_MultipleEvents(t *testing.T) {
	c := octo.New()

	mediatr.InjectNotification[EventX](c, func(c *octo.Container) mediatr.NotificationHandler[EventX] { return &HandlerX{} })
	mediatr.InjectNotification[EventY](c, func(c *octo.Container) mediatr.NotificationHandler[EventY] { return &HandlerY{} })

	manager := InjectNotifyManager(c)

	absX := AbsoluteTypeName(reflect.TypeOf(EventX{}))
	absY := AbsoluteTypeName(reflect.TypeOf(EventY{}))

	if _, ok := manager.events[absX]; !ok {
		t.Errorf("expected EventX registered")
	}
	if _, ok := manager.events[absY]; !ok {
		t.Errorf("expected EventY registered")
	}
}

// Test that TryResolve returns an existing manager
func TestInjectNotifyManager_TryResolveExisting(t *testing.T) {
	c := octo.New()

	existing := &NotifyManager{
		container: c,
		events:    make(map[string]*eventDecl),
	}
	octo.InjectValue(c, existing)

	manager := InjectNotifyManager(c)
	if manager != existing {
		t.Errorf("expected existing manager to be returned")
	}
}

// Test that events registered via mediatr are automatically discovered
func TestInjectNotifyManager_WithMediatrInjection(t *testing.T) {
	c := octo.New()

	h := &HandlerX{}
	mediatr.InjectNotification[EventX](c, func(c *octo.Container) mediatr.NotificationHandler[EventX] { return h })

	manager := InjectNotifyManager(c)
	if manager == nil {
		t.Fatal("expected manager to be returned")
	}

	absName := AbsoluteTypeName(reflect.TypeOf(EventX{}))
	if _, ok := manager.events[absName]; !ok {
		t.Fatalf("expected EventX to be automatically registered via mediatr")
	}

	// Simulate notifying
	ctx := context.Background()
	evVal := reflect.ValueOf(EventX{ID: 42})
	notifyEvents(manager.container, ctx, reflect.TypeOf(EventX{}), evVal)

	if !h.Called {
		t.Errorf("expected handler to be called")
	}
}

// ---------------- AliasEvent Tests ----------------

func TestAliasEvent_Normal(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	// Add a single alias
	AliasEvent[EventA](manager, "alias1")
	absName := AbsoluteTypeName(reflect.TypeOf(EventA{}))
	decl := manager.events[absName]
	if decl == nil {
		t.Fatal("expected eventDecl to exist")
	}
	if len(decl.aliases) != 2 { // absolute name + alias1
		t.Errorf("expected 2 aliases, got %v", len(decl.aliases))
	}

	// Add another alias
	AliasEvent[EventA](manager, "alias2")
	if len(decl.aliases) != 3 {
		t.Errorf("expected 3 aliases after adding alias2, got %v", len(decl.aliases))
	}
	if _, ok := manager.events["alias2"]; !ok {
		t.Errorf("expected alias2 to be registered in events map")
	}
}

func TestAliasEvent_EventDeclNotRegistered(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	// EventB not yet registered; should create eventDecl
	AliasEvent[EventB](manager, "b1")
	absName := AbsoluteTypeName(reflect.TypeOf(EventB{}))
	decl := manager.events[absName]
	if decl == nil {
		t.Fatal("expected eventDecl to exist")
	}
	if len(decl.aliases) != 2 {
		t.Errorf("expected 2 aliases (absolute + b1), got %v", len(decl.aliases))
	}
}

func TestAliasEvent_EmptyAliases_Panic(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	defer func() {
		r := recover()
		if r == nil || r.(string) != "octo: aliases must not be empty" {
			t.Errorf("expected panic for empty aliases, got %v", r)
		}
	}()

	AliasEvent[EventX](manager)
}

func TestAliasEvent_DuplicateAlias_Panic(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	AliasEvent[EventX](manager, "dup")
	defer func() {
		r := recover()
		if r == nil || r.(string) != "octo: alias 'dup' already registered" {
			t.Errorf("expected panic on duplicate alias, got %v", r)
		}
	}()

	AliasEvent[EventB](manager, "dup")
}

func TestAliasEvent_AliasMatchesAbsoluteName_Panic(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	absName := AbsoluteTypeName(reflect.TypeOf(EventA{}))
	defer func() {
		r := recover()
		if r == nil || r.(string) != "octo: alias cannot match type absolute name" {
			t.Errorf("expected panic on alias matching absolute name, got %v", r)
		}
	}()

	// Now try to add alias matching absolute name, should panic
	AliasEvent[EventA](manager, absName)
}

// ---------------- MarshallEvent Tests (reusing EventX and EventY) ----------------

func TestMarshallEvent_Success(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)
	AliasEvent[EventX](manager, "aliasX")

	ev := EventX{ID: 42}
	data, err := MarshallEvent[EventX](manager, ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty data")
	}

	// Unmarshal into wrappedEvent
	var wrapped wrappedEvent
	if err := json.Unmarshal(data, &wrapped); err != nil {
		t.Fatalf("unable to unmarshal wrapped event: %v", err)
	}

	// Check aliases
	expectedAliases := map[string]struct{}{
		AbsoluteTypeName(reflect.TypeOf(EventX{})): {},
		"aliasX": {},
	}
	if len(wrapped.Aliases) != len(expectedAliases) {
		t.Errorf("expected %d aliases, got %d", len(expectedAliases), len(wrapped.Aliases))
	}
	for _, a := range wrapped.Aliases {
		if _, ok := expectedAliases[a]; !ok {
			t.Errorf("unexpected alias %q in wrapped event", a)
		}
	}

	// Check Event content
	var evDecoded EventX
	if err := json.Unmarshal(wrapped.Event, &evDecoded); err != nil {
		t.Fatalf("unable to unmarshal EventX from wrapped event: %v", err)
	}
	if evDecoded.ID != ev.ID {
		t.Errorf("expected EventX.ID=%d, got %d", ev.ID, evDecoded.ID)
	}
}

func TestMarshallEvent_NilEvent(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	var nilEvent *EventX
	_, err := MarshallEvent[*EventX](manager, nilEvent)
	if err == nil {
		t.Errorf("expected error for nil event")
	}
}

func TestMarshallEvent_UnregisteredEvent_Panic(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic for unregistered event")
		}
	}()

	_, _ = MarshallEvent[EventY](manager, EventY{Name: "test"})
}

// ---------------- UnmarshallAndNotifyEvent ----------------

func TestUnmarshallAndNotifyEvent_Normal(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	h := &HandlerX{}
	mediatr.InjectNotification[EventX](c, func(c *octo.Container) mediatr.NotificationHandler[EventX] { return h })
	AliasEvent[EventX](manager, "x1")

	ev := EventX{ID: 10}
	data, _ := MarshallEvent[EventX](manager, ev)
	err := UnmarshallAndNotifyEvent(manager, context.Background(), data, false)
	if err != nil || !h.Called {
		t.Errorf("expected EventX handler to be called")
	}
}

func TestUnmarshallAndNotifyEvent_SkipIfNotFound(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	data, _ := json.Marshal(wrappedEvent{Aliases: []string{"unknown"}, Event: []byte(`{"ID":1}`)})
	if err := UnmarshallAndNotifyEvent(manager, context.Background(), data, true); err != nil {
		t.Errorf("expected no error when skipIfNF=true")
	}
}

func TestUnmarshallAndNotifyEvent_InvalidJSON(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	if err := UnmarshallAndNotifyEvent(manager, context.Background(), []byte("{invalid}"), false); err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}

func TestUnmarshallAndNotifyEvent_EmptyAliases(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	data, _ := json.Marshal(wrappedEvent{Aliases: []string{}, Event: []byte(`{"ID":1}`)})
	if err := UnmarshallAndNotifyEvent(manager, context.Background(), data, false); err == nil {
		t.Errorf("expected error for empty aliases")
	}
}

func TestUnmarshallAndNotifyEvent_EventNotFound(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	data, _ := json.Marshal(wrappedEvent{Aliases: []string{"nonexistent"}, Event: []byte(`{"ID":1}`)})
	if err := UnmarshallAndNotifyEvent(manager, context.Background(), data, false); err == nil {
		t.Errorf("expected error for event not found")
	}
}

func TestUnmarshallAndNotifyEvent_ContextCanceled(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	h := &HandlerX{}
	mediatr.InjectNotification[EventX](c, func(c *octo.Container) mediatr.NotificationHandler[EventX] { return h })
	AliasEvent[EventX](manager, "x1")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ev := EventX{ID: 42}
	data, _ := MarshallEvent[EventX](manager, ev)
	if err := UnmarshallAndNotifyEvent(manager, ctx, data, false); err != nil {
		t.Errorf("expected no error on canceled context, got %v", err)
	}
}

func TestUnmarshallAndNotifyEvent_MultipleHandlers(t *testing.T) {
	c := octo.New()
	manager := InjectNotifyManager(c)

	h1 := &HandlerX{}
	h2 := &HandlerX{}
	mediatr.InjectNotification[EventX](c, func(c *octo.Container) mediatr.NotificationHandler[EventX] { return h1 })
	mediatr.InjectNotification[EventX](c, func(c *octo.Container) mediatr.NotificationHandler[EventX] { return h2 })
	AliasEvent[EventX](manager, "x1")

	ev := EventX{ID: 50}
	data, _ := MarshallEvent[EventX](manager, ev)
	UnmarshallAndNotifyEvent(manager, context.Background(), data, false)

	if !h1.Called || !h2.Called {
		t.Errorf("expected all EventX handlers to be called")
	}
}
