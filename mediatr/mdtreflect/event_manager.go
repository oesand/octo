package mdtreflect

import (
	"fmt"
	"github.com/oesand/octo"
	"github.com/oesand/octo/internal"
	"reflect"
	"slices"
	"sync"
)

// InjectManager injects a EventManager into the container if not already registered.
// It automatically registers all event types discovered in the container.
func InjectManager(container *octo.Container) *EventManager {
	manager := octo.TryResolve[*EventManager](container)
	if manager != nil {
		if manager.container == nil {
			panic("EventManager cannot be injected manually")
		}
		return manager
	}

	manager = &EventManager{
		container: container,
		events:    make(map[string]*eventDecl),
	}
	octo.InjectValue(container, manager)

	return manager
}

type eventDecl struct {
	aliases []string
	typ     reflect.Type
}

// EventManager manages event types and their aliases, and
// provides methods for marshalling, unmarshalling, and notifying.
type EventManager struct {
	mu           sync.RWMutex
	container    *octo.Container
	events       map[string]*eventDecl
	autoRegister sync.Once
}

func (manager *EventManager) ensureAutoRegisterEvents() {
	manager.autoRegister.Do(manager.doAutoRegisterEvents)
}

func (manager *EventManager) doAutoRegisterEvents() {
	eventTypes := notificationEventTypes(manager.container)
	for eventType := range eventTypes {
		manager.registerEvent(eventType)
	}
}

func (manager *EventManager) registerEvent(eventType reflect.Type) *eventDecl {
	if manager.events == nil {
		panic("EventManager cannot be injected manually")
	}

	absoluteName := AbsoluteTypeName(eventType)
	if decl, ok := manager.events[absoluteName]; ok {
		return decl
	}

	decl := &eventDecl{
		aliases: []string{absoluteName},
		typ:     eventType,
	}
	manager.events[absoluteName] = decl
	return decl
}

// AliasEvent registers one or more aliases for an event type T.
// Panics if aliases are empty, duplicate, or match the absolute type name.
func AliasEvent[T any](manager *EventManager, aliases ...string) {
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
