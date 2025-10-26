package mdtreflect

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type wrappedEvent struct {
	Aliases []string `json:"als"`
	Event   []byte   `json:"ev"`
}

// MarshallEvent serializes an event of type T with its aliases.
// Returns error if the event is nil or marshalling fails.
func MarshallEvent(manager *EventManager, event any) ([]byte, error) {
	if event == nil {
		return nil, errors.New("octo: event must not be nil")
	}

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	manager.ensureAutoRegisterEvents()

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

	wrapped := &wrappedEvent{
		Aliases: aliases,
		Event:   eventBuf,
	}

	return json.Marshal(wrapped)
}

// UnmarshallAndPublish deserializes a wrapped event and notifies
// all registered handlers for the event type.
// If EventManager.SkipIfNotFound is true, missing event types are ignored.
func UnmarshallAndPublish(manager *EventManager, ctx context.Context, buf []byte) error {
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

	manager.ensureAutoRegisterEvents()

	var eventType reflect.Type
	for _, alias := range wrapped.Aliases {
		if decl, ok := manager.events[alias]; ok {
			eventType = decl.typ
			break
		}
	}

	if eventType == nil {
		return errors.New("octo: not found event by aliases, skip")
	}

	eventVal := reflect.New(eventType)
	if err = json.Unmarshal(wrapped.Event, eventVal.Interface()); err != nil {
		return fmt.Errorf("octo: unable to unmarshall wrapped event: %s", err)
	}

	notifyEvents(manager.container, ctx, eventType, eventVal.Elem())
	return nil
}
