package mediatr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oesand/octo/internal"
	"reflect"
	"slices"
)

type eventDecl struct {
	aliases []string
	typ     reflect.Type
}

type marshallEvent struct {
	Aliases []string `json:"als"`
	Event   []byte   `json:"ev"`
}

func (m *Manager) registerEvent(eventType reflect.Type) *eventDecl {
	absoluteName := AbsoluteTypeName(eventType)
	if decl, ok := m.events[absoluteName]; ok {
		return decl
	}

	decl := &eventDecl{
		aliases: []string{absoluteName},
		typ:     eventType,
	}
	m.events[absoluteName] = decl
	return decl
}

// AliasEvent registers one or more aliases for an event type T.
// Panics if aliases are empty, duplicate, or match the absolute type name.
func AliasEvent[T any](manager *Manager, aliases ...string) {
	if len(aliases) == 0 {
		panic("octo: aliases must not be empty")
	}

	typ := reflect.TypeFor[T]()
	absoluteName := AbsoluteTypeName(typ)

	if slices.Contains(aliases, absoluteName) {
		panic("octo: alias cannot match type absolute name")
	}

	aliases = internal.Unique(aliases)

	manager.mu.Lock()
	defer manager.mu.Unlock()

	if manager.container == nil {
		panic("Manager cannot be injected manually, use Inject function")
	}

	if manager.events == nil {
		manager.events = make(map[string]*eventDecl)
	}

	for _, alias := range aliases {
		if _, has := manager.events[alias]; has {
			panic(fmt.Sprintf("octo: alias '%s' already registered", alias))
		}
	}

	decl := manager.registerEvent(typ)
	decl.aliases = append(decl.aliases, aliases...)
	for _, alias := range aliases {
		manager.events[alias] = decl
	}
}

// MarshallEvent serializes an event of type T with its aliases.
// Supported only pointer events.
// Returns error if the event is nil or marshalling fails.
func MarshallEvent(manager *Manager, event any) ([]byte, error) {
	if event == nil {
		return nil, errors.New("octo: event must not be nil")
	}

	manager.ensureInit()

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	typ := reflect.TypeOf(event)
	absoluteName := AbsoluteTypeName(typ)

	decl, has := manager.events[absoluteName]
	if !has {
		return nil, fmt.Errorf("octo: event '%s' not registered", absoluteName)
	}

	aliases := decl.aliases

	eventBuf, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	wrapped := &marshallEvent{
		Aliases: aliases,
		Event:   eventBuf,
	}

	return json.Marshal(wrapped)
}

// UnmarshallAndPublish deserializes a wrapped event and notifies
// all registered handlers for the event type.
func UnmarshallAndPublish(manager *Manager, ctx context.Context, buf []byte) error {
	var marshalled marshallEvent
	err := json.Unmarshal(buf, &marshalled)
	if err != nil {
		return fmt.Errorf("octo: unable to unmarshall event: %s", err)
	}

	if len(marshalled.Aliases) == 0 || len(marshalled.Event) == 0 {
		return errors.New("octo: invalid unmarshall event")
	}

	manager.ensureInit()

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	var eventType reflect.Type
	for _, alias := range marshalled.Aliases {
		if decl, ok := manager.events[alias]; ok {
			eventType = decl.typ
			break
		}
	}

	if eventType == nil {
		return errors.New("octo: not found event by aliases, skip")
	}

	eventVal := reflect.New(eventType)
	if err = json.Unmarshal(marshalled.Event, eventVal.Interface()); err != nil {
		return fmt.Errorf("octo: unable to unmarshall event: %s", err)
	}

	return Publish(manager, ctx, eventVal.Elem().Interface())
}
