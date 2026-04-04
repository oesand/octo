package flow

import (
	"context"
	"reflect"

	"github.com/oesand/octo"
	"github.com/oesand/octo/mediator"
)

// State is the shared flow state interface required by flow steps.
// Implementations must report the current step name, whether the
// flow is finished and which flow the state belongs to.
type State interface {
	GetStep() string
	Finished() bool
	Flow() string
}

// Step represents a single flow step. A Step can declare which event
// types it is interested in via EventTypes and attempts to handle an
// incoming event via Handle. The generic parameter TState constrains
// the concrete state type used by the step.
type Step[TState State] interface {
	EventTypes() []reflect.Type
	Handle(context.Context, *octo.Container, TState, Event) (bool, error)
}

// Initial creates a step that matches the zero/initial step name and
// evaluates the provided child steps.
func Initial[TState State](children ...Step[TState]) Step[TState] {
	return On("", children...)
}

// On creates a conditional step that matches when the state's
// GetStep() equals the provided step name and then evaluates children.
func On[TState State](step string, children ...Step[TState]) Step[TState] {
	return When(func(state TState) bool {
		return state.GetStep() == step
	}, children...)
}

// When creates a conditional step based on the supplied rule. When the
// rule returns true the step's children are evaluated.
func When[TState State](rule func(TState) bool, children ...Step[TState]) Step[TState] {
	return &whenStep[TState]{
		rule:     rule,
		children: children,
	}
}

type whenStep[TState State] struct {
	rule     func(TState) bool
	children []Step[TState]
}

func (w *whenStep[TState]) EventTypes() []reflect.Type {
	return nil
}

func (w *whenStep[TState]) Handle(ctx context.Context, container *octo.Container, state TState, event Event) (bool, error) {
	if !w.rule(state) {
		return false, nil
	}

	for _, child := range w.children {
		handled, err := child.Handle(ctx, container, state, event)
		if handled {
			return true, err
		}
	}

	return false, nil
}

// OnEvent creates a step that triggers only for events of type TEvent.
// Child steps must only declare the same event type.
func OnEvent[TEvent Event, TState State](children ...Step[TState]) Step[TState] {
	eventType := reflect.TypeFor[TEvent]()
	for _, child := range children {
		events := child.EventTypes()
		for _, event := range events {
			if eventType != event {
				panic("flow: event condition cannot have child event rules other than its own")
			}
		}
	}
	return &onEventCondition[TEvent, TState]{
		eventType: eventType,
		children:  children,
	}
}

type onEventCondition[TEvent Event, TState State] struct {
	eventType reflect.Type
	children  []Step[TState]
}

func (w *onEventCondition[TEvent, TState]) EventTypes() []reflect.Type {
	return []reflect.Type{w.eventType}
}

func (w *onEventCondition[TEvent, TState]) Handle(ctx context.Context, container *octo.Container, state TState, event Event) (bool, error) {
	if event == nil {
		return false, nil
	}
	if _, ok := event.(TEvent); !ok {
		return false, nil
	}

	for _, child := range w.children {
		handled, err := child.Handle(ctx, container, state, event)
		if handled {
			return true, err
		}
	}

	return false, nil
}

// Do creates a step that invokes the provided handler for the current
// state. The handler runs synchronously and its returned error is
// propagated.
func Do[TState State](handle func(context.Context, TState) error) Step[TState] {
	return &handleStep[TState]{
		handler: handle,
	}
}

type handleStep[TState State] struct {
	handler func(context.Context, TState) error
}

func (*handleStep[TState]) EventTypes() []reflect.Type {
	return nil
}

func (w *handleStep[TState]) Handle(ctx context.Context, _ *octo.Container, state TState, _ Event) (bool, error) {
	return true, w.handler(ctx, state)
}

// DoEvent creates a step that handles an event of type TEvent using
// the supplied handler which receives the typed event.
func DoEvent[TEvent Event, TState State](handle func(context.Context, TState, TEvent) error) Step[TState] {
	return &handleEventStep[TEvent, TState]{
		handler: handle,
	}
}

type handleEventStep[TEvent Event, TState State] struct {
	handler func(context.Context, TState, TEvent) error
}

func (*handleEventStep[TEvent, TState]) EventTypes() []reflect.Type {
	return []reflect.Type{reflect.TypeFor[TEvent]()}
}

func (w *handleEventStep[TEvent, TState]) Handle(ctx context.Context, _ *octo.Container, state TState, event Event) (bool, error) {
	if event == nil {
		return false, nil
	}
	if ev, ok := event.(TEvent); ok {
		return true, w.handler(ctx, state, ev)
	}
	return false, nil
}

/*
func Send[TState State, TRequest mediator.Request[TResponse], TResponse any](
	request func(TState) TRequest,
	success func(context.Context, TState, TResponse) error,
	error func(context.Context, TState, error) error,
) Step[TState] {
	return &sendStep[TState, TRequest, TResponse]{
		request: request,
		success: success,
		error:   error,
	}
}

type sendStep[TState State, TRequest mediator.Request[TResponse], TResponse any] struct {
	request func(TState) TRequest
	success func(context.Context, TState, TResponse) error
	error   func(context.Context, TState, error) error
}

func (*sendStep[TState, TRequest, TResponse]) EventTypes() []reflect.Type {
	return nil
}

func (s *sendStep[TState, TRequest, TResponse]) Handle(ctx context.Context, container *octo.Container, state TState, _ Event) (bool, error) {
	manager := octo.Resolve[*mediator.Manager](container)
	request := s.request(state)

	response, err := mediator.Send(manager, ctx, request)
	if err != nil {
		err = s.error(ctx, state, err)
	} else {
		err = s.success(ctx, state, response)
	}

	return true, err
}

*/

// Publish creates a step that publishes an event constructed from the
// current state using the provided function. The event is published via
// the mediator manager resolved from the container.
func Publish[TState State, TEvent any](
	event func(TState) TEvent,
) Step[TState] {
	return &publishStep[TState, TEvent]{
		event: event,
	}
}

type publishStep[TState State, TEvent any] struct {
	event func(TState) TEvent
}

func (*publishStep[TState, TEvent]) EventTypes() []reflect.Type {
	return nil
}

func (s *publishStep[TState, TEvent]) Handle(ctx context.Context, container *octo.Container, state TState, _ Event) (bool, error) {
	manager := octo.Resolve[*mediator.Manager](container)
	event := s.event(state)
	return true, mediator.Publish(manager, ctx, event)
}
