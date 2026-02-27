package mediatr

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/oesand/octo"
	"github.com/oesand/octo/backoff"
)

// Inject injects a Manager into the container if not already registered.
//
// Options will be applied in any case.
func Inject(container *octo.Container, options ...Option) *Manager {
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

	for _, opt := range options {
		opt(manager)
	}

	return manager
}

type Manager struct {
	mu       sync.RWMutex
	onceInit sync.Once

	container       *octo.Container
	events          map[string]*eventDecl
	requestHandlers map[reflect.Type]handleRequest
	eventHandlers   map[reflect.Type][]handleEvent

	useBackOff atomic.Pointer[backoffConf]
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

func (m *Manager) doInit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.requestHandlers == nil {
		m.requestHandlers = make(map[reflect.Type]handleRequest)
	}

	if m.eventHandlers == nil {
		m.eventHandlers = make(map[reflect.Type][]handleEvent)
	}

	if m.events == nil {
		m.events = make(map[string]*eventDecl)
	}

	injects := octo.ResolveInjections(m.container)
	for decl := range injects {
		if eventType, handler, ok := verifyEventHandler(decl); ok {
			handlers, _ := m.eventHandlers[eventType]
			m.eventHandlers[eventType] = append(handlers, handler)

			m.registerEvent(eventType)
			continue
		}

		if requestType, handler, ok := verifyRequestHandler(decl); ok {
			if _, ok = m.requestHandlers[requestType]; ok {
				continue
			}

			m.requestHandlers[requestType] = handler
		}
	}
}

// Publish publishes a event to all registered NotificationHandlers.
// The event is sent to every matching handler until either:
//   - The context is cancelled, or
//   - All handlers have been executed.
func Publish(
	manager *Manager,
	ctx context.Context,
	event any,
) error {
	manager.ensureInit()

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	handlers, has := manager.eventHandlers[reflect.TypeOf(event)]
	if !has {
		return nil
	}

	results := make(chan error, len(handlers))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	bc := manager.useBackOff.Load()
	for _, handler := range handlers {
		go func() {
			var err error
			if bc != nil {
				err = backoff.BackOff(ctx, func(ctx context.Context) error { return handler(ctx, event) }, bc.options...)
			} else {
				err = handler(ctx, event)
			}
			results <- err
		}()
	}

	defer close(results)

	for i := 0; i < len(handlers); i++ {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case err := <-results:
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Send resolves a RequestHandler for the given request/response types from the container
// and calls its Request method. This is the entry point for executing a request.
func Send[TRequest Request[TResponse], TResponse any](
	manager *Manager,
	ctx context.Context,
	request TRequest,
) (TResponse, error) {
	manager.ensureInit()

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	var nilVal TResponse
	handler, has := manager.requestHandlers[reflect.TypeOf(request)]
	if !has {
		return nilVal, errors.New("octo: handler not found")
	}

	var resp any
	var err error
	if bc := manager.useBackOff.Load(); bc != nil {
		err = backoff.BackOff(ctx, func(ctx context.Context) error {
			var err error
			resp, err = handler(ctx, request)
			return err
		}, bc.options...)
	} else {
		resp, err = handler(ctx, request)
	}

	if resp == nil {
		return nilVal, err
	}
	return resp.(TResponse), err
}
