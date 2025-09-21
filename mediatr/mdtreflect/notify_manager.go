package mdtreflect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oesand/octo"
	"github.com/oesand/octo/internal"
	"reflect"
	"slices"
	"sync"
)

type eventDecl struct {
	aliases []string
	typ     reflect.Type
}

// NotifyManager manages event types and their aliases, and
// provides methods for marshalling, unmarshalling, and notifying.
type NotifyManager struct {
	mu        sync.RWMutex
	container *octo.Container
	events    map[string]*eventDecl
}

// InjectNotifyManager injects a NotifyManager into the container if not already registered.
// It automatically registers all event types discovered in the container.
func InjectNotifyManager(container *octo.Container) *NotifyManager {
	manager := octo.TryResolve[*NotifyManager](container)
	if manager != nil {
		return manager
	}

	manager = &NotifyManager{
		container: container,
		events:    make(map[string]*eventDecl),
	}
	octo.InjectValue(container, manager)

	eventTypes := notificationEventTypes(container)
	for eventType := range eventTypes {
		absoluteName := AbsoluteTypeName(eventType)
		manager.events[absoluteName] = &eventDecl{
			aliases: []string{absoluteName},
			typ:     eventType,
		}
	}

	return manager
}

// AliasEvent registers one or more aliases for an event type T.
// Panics if aliases are empty, duplicate, or match the absolute type name.
func AliasEvent[T any](manager *NotifyManager, aliases ...string) {
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

	if manager.events == nil {
		manager.events = make(map[string]*eventDecl)
	}

	for _, alias := range aliases {
		if _, has := manager.events[alias]; has {
			panic(fmt.Sprintf("octo: alias '%s' already registered", alias))
		}
	}

	decl, has := manager.events[absoluteName]
	if !has {
		decl = &eventDecl{
			aliases: []string{absoluteName},
			typ:     typ,
		}
		manager.events[absoluteName] = decl
	}

	decl.aliases = append(decl.aliases, aliases...)
	for _, alias := range aliases {
		manager.events[alias] = decl
	}
}

type wrappedEvent struct {
	Aliases []string `json:"als"`
	Event   []byte   `json:"ev"`
}

// MarshallEvent serializes an event of type T with its aliases.
// Returns error if the event is nil or marshalling fails.
func MarshallEvent[T comparable](manager *NotifyManager, event T) ([]byte, error) {
	var nilVal T
	if event == nilVal {
		return nil, errors.New("octo: event must not be nil")
	}

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	typ := reflect.TypeFor[T]()
	absoluteName := AbsoluteTypeName(typ)

	decl, has := manager.events[absoluteName]
	if !has {
		panic(fmt.Sprintf("octo: event '%s' not registered", absoluteName))
	}

	aliases := decl.aliases

	eventBuf, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	wrapped := &wrappedEvent{
		Aliases: aliases,
		Event:   eventBuf,
	}

	return json.Marshal(wrapped)
}

// UnmarshallAndNotifyEvent deserializes a wrapped event and notifies
// all registered handlers for the event type.
// If skipIfNF is true, missing event types are ignored.
func UnmarshallAndNotifyEvent(manager *NotifyManager, ctx context.Context, buf []byte, skipIfNF bool) error {
	var wrapped wrappedEvent
	err := json.Unmarshal(buf, &wrapped)
	if err != nil {
		return fmt.Errorf("octo: unable to unmarshall wrapped event: %s", err)
	}

	if len(wrapped.Aliases) == 0 || len(wrapped.Event) == 0 {
		return errors.New("octo: invalid unmarshall wrapped event")
	}

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	var eventType reflect.Type
	for _, alias := range wrapped.Aliases {
		if decl, ok := manager.events[alias]; ok {
			eventType = decl.typ
			break
		}
	}

	if eventType == nil {
		if skipIfNF {
			return nil
		}
		return errors.New("octo: not found event by aliases, skip")
	}

	eventVal := reflect.New(eventType)
	if err = json.Unmarshal(wrapped.Event, eventVal.Interface()); err != nil {
		return fmt.Errorf("octo: unable to unmarshall wrapped event: %s", err)
	}

	notifyEvents(manager.container, ctx, eventType, eventVal.Elem())
	return nil
}
