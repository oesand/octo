package mediator

import (
	"context"
	"reflect"
	"sync"

	"github.com/oesand/octo"
)

// Inject injects a Manager into the container if not already registered.
//
// Options will be applied in any case.
func Inject(container *octo.Container) *Manager {
	manager := octo.TryResolve[*Manager](container)
	if manager != nil {
		if manager.container == nil {
			panic("Manager cannot be injected manually")
		}
	} else {
		manager = &Manager{
			container: container,
		}
		octo.InjectValue(container, manager)
	}
	return manager
}

type Manager struct {
	onceInit  sync.Once
	container *octo.Container
	handlers  map[reflect.Type][]octo.Declaration
}

func (m *Manager) ensureInit() {
	if m == nil {
		panic("Manager must not be nil")
	}
	if m.container == nil {
		panic("Manager cannot be injected manually, use Inject function")
	}
	m.onceInit.Do(m.doInit)
}

var (
	ctxType   = reflect.TypeFor[context.Context]()
	errorType = reflect.TypeFor[error]()
)

func (m *Manager) doInit() {
	if m.handlers == nil {
		m.handlers = make(map[reflect.Type][]octo.Declaration)
	}

	injects := octo.ResolveInjections(m.container)
	for decl := range injects {
		if method, ok := decl.Type().MethodByName("Notification"); ok &&
			method.Type.NumIn() == 3 && method.Type.In(1).AssignableTo(ctxType) &&
			method.Type.NumOut() == 1 && method.Type.Out(0).AssignableTo(errorType) {

			eventType := method.Type.In(2)
			m.handlers[eventType] = append(m.handlers[eventType], decl)
		}
	}
}
