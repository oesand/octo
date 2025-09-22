package mdtreflect

import (
	"github.com/oesand/octo"
	"reflect"
	"testing"
)

// ---------------- InjectManager Tests ----------------

func TestInjectManager_NewManagerCreated(t *testing.T) {
	c := octo.New()

	manager := InjectManager(c)
	if manager == nil {
		t.Fatal("expected manager, got nil")
	}
	if manager.container != c {
		t.Fatalf("expected manager.container to be %v, got %v", c, manager.container)
	}
	if manager.events == nil {
		t.Fatal("expected events map to be initialized")
	}

	// Second call should return the same manager (idempotent)
	m2 := InjectManager(c)
	if manager != m2 {
		t.Fatal("expected same manager instance on repeated calls")
	}
}

func TestInjectManager_ReturnsExistingManager(t *testing.T) {
	c := octo.New()

	// First inject
	m1 := InjectManager(c)
	if m1 == nil {
		t.Fatal("expected manager, got nil")
	}

	// Inject again → should return the same instance
	m2 := InjectManager(c)
	if m1 != m2 {
		t.Fatal("expected InjectManager to return the same manager if already registered")
	}
}

func TestInjectManager_PanicsOnManualInjection(t *testing.T) {
	c := octo.New()

	// Manually inject a manager without container (invalid state)
	octo.InjectValue(c, &EventManager{})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		if r != "EventManager cannot be injected manually" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()

	InjectManager(c)
}

// ---------------- RegisterEvent Tests ----------------

// Simple test event
type TestEvent struct {
	ID string
}

func TestRegisterEvent_PanicsIfEventsNil(t *testing.T) {
	manager := &EventManager{
		events: nil,
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		if r != "EventManager cannot be injected manually" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()

	manager.registerEvent(reflect.TypeOf(TestEvent{}))
}

func TestRegisterEvent_NewEventCreatesDecl(t *testing.T) {
	manager := &EventManager{
		events: make(map[string]*eventDecl),
	}

	typ := reflect.TypeOf(TestEvent{})
	abs := AbsoluteTypeName(typ)

	decl := manager.registerEvent(typ)

	if decl == nil {
		t.Fatal("expected decl, got nil")
	}
	if decl.typ != typ {
		t.Fatalf("expected decl.typ=%v, got %v", typ, decl.typ)
	}
	if len(decl.aliases) != 1 || decl.aliases[0] != abs {
		t.Fatalf("expected aliases [%s], got %v", abs, decl.aliases)
	}

	// Check that it was stored in events map
	if got, ok := manager.events[abs]; !ok || got != decl {
		t.Fatalf("expected decl stored in manager.events[%q]", abs)
	}
}

func TestRegisterEvent_ReusesExistingDecl(t *testing.T) {
	manager := &EventManager{
		events: make(map[string]*eventDecl),
	}

	typ := reflect.TypeOf(TestEvent{})
	abs := AbsoluteTypeName(typ)

	// First registration
	decl1 := manager.registerEvent(typ)
	// Second registration → should reuse
	decl2 := manager.registerEvent(typ)

	if decl1 != decl2 {
		t.Fatal("expected registerEvent to return the same decl for same type")
	}
	if _, ok := manager.events[abs]; !ok {
		t.Fatalf("expected decl still in events map under key %q", abs)
	}
}

// ---------------- AliasEvent Tests ----------------

type AliasEventTest struct {
	Name string
}

func TestAliasEvent_PanicsIfNoAliases(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got none")
		} else if r != "octo: aliases must not be empty" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()

	AliasEvent[AliasEventTest](manager) // no aliases
}

func TestAliasEvent_PanicsIfAliasEqualsAbsoluteName(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}
	typ := reflect.TypeOf(AliasEventTest{})
	abs := AbsoluteTypeName(typ)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got none")
		} else if r != "octo: alias cannot match type absolute name" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()

	AliasEvent[AliasEventTest](manager, abs) // alias == absolute name
}

func TestAliasEvent_PanicsIfAliasAlreadyRegistered(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}

	// First alias works
	AliasEvent[AliasEventTest](manager, "first")

	// Re-register with same alias should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got none")
		} else if r != "octo: alias 'first' already registered" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()

	AliasEvent[AliasEventTest](manager, "first")
}

func TestAliasEvent_AddsAliasesToDecl(t *testing.T) {
	manager := &EventManager{events: make(map[string]*eventDecl)}

	typ := reflect.TypeOf(AliasEventTest{})
	abs := AbsoluteTypeName(typ)

	// Add multiple aliases, with a duplicate
	AliasEvent[AliasEventTest](manager, "a1", "a2", "a1")

	decl, ok := manager.events[abs]
	if !ok {
		t.Fatalf("expected decl for %q", abs)
	}

	expectedAliases := map[string]bool{abs: true, "a1": true, "a2": true}
	for _, alias := range decl.aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias: %q", alias)
		}
	}

	// Ensure all aliases map to same decl
	for alias := range expectedAliases {
		if got := manager.events[alias]; got != decl {
			t.Errorf("alias %q not mapped to decl", alias)
		}
	}
}
