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

type handleEvent func(ctx context.Context, event any) error

type Manager struct {
	onceInit  sync.Once
	container *octo.Container
	handlers  map[reflect.Type][]handleEvent
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
		m.handlers = make(map[reflect.Type][]handleEvent)
	}

	massHandlerType := reflect.TypeFor[MassEventHandler]()
	injects := octo.ResolveInjections(m.container)
	for decl := range injects {
		if method, ok := decl.Type().MethodByName("Notification"); ok &&
			method.Type.NumIn() == 3 && method.Type.In(1).AssignableTo(ctxType) &&
			method.Type.NumOut() == 1 && method.Type.Out(0).AssignableTo(errorType) {

			eventType := method.Type.In(2)
			m.handlers[eventType] = append(m.handlers[eventType], func(ctx context.Context, event any) error {
				handler := decl.Value()
				values := []reflect.Value{
					reflect.ValueOf(handler),
					reflect.ValueOf(ctx),
					reflect.ValueOf(event),
				}
				err := method.Func.Call(values)[0].Interface()
				if err != nil {
					return err.(error)
				}
				return nil
			})
			continue
		}
		if decl.Type().Implements(massHandlerType) {
			handler := decl.Value().(MassEventHandler)
			for _, eventType := range handler.EventTypes() {
				m.handlers[eventType] = append(m.handlers[eventType], func(ctx context.Context, event any) error {
					return handler.Handle(ctx, event)
				})
			}
		}
	}
}
