package flow

import (
	"context"
	"reflect"
	"strings"

	"github.com/oesand/octo"
)

// State is the shared flow state interface required by flow steps.
// Implementations must report the current step name, whether the
// flow is finished and which flow the state belongs to.
type State interface {
	GetStep() string
	Finished() bool
	Flow() string
}

// NameGetterState is a helper state type that derives the flow name from
// the state type name by trimming the trailing "State" suffix.
type NameGetterState[TState any] struct{}

func (state *NameGetterState[TState]) Flow() string {
	stateType := reflect.TypeFor[TState]()
	if stateType.Kind() == reflect.Pointer {
		stateType = stateType.Elem()
	}
	return strings.TrimSuffix(stateType.Name(), "State")
}

// Step represents a single flow step. A Step can declare which event
// types it is interested in via EventTypes and attempts to handle an
// incoming event via Handle. The generic parameter TState constrains
// the concrete state type used by the step.
type Step[TState State] interface {
	EventTypes() []reflect.Type
	Match(TState, Event) StepExecutor[TState]
}

type StepExecutor[TState State] interface {
	Recursion() bool
	Execute(context.Context, *octo.Container, TState, Event) error
}

type stepExecutorOptions interface {
	SetRecursion(bool)
}

type StepOption func(stepExecutorOptions)

// Recursive is a StepOption that enables recursive execution for a step.
// When applied to a step via Do or DoEvent, it allows the step executor to
// trigger recursive evaluation of the flow after the step completes.
// This is useful for workflows where a step's execution may trigger additional
// flow processing that needs to re-evaluate step conditions.
func Recursive() StepOption {
	return func(step stepExecutorOptions) {
		step.SetRecursion(true)
	}
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

func (step *whenStep[TState]) EventTypes() []reflect.Type {
	return nil
}

func (step *whenStep[TState]) Match(state TState, event Event) StepExecutor[TState] {
	if !step.rule(state) {
		return nil
	}

	for _, child := range step.children {
		executor := child.Match(state, event)
		if executor != nil {
			return executor
		}
	}

	return nil
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

func (w *onEventCondition[TEvent, TState]) Match(state TState, event Event) StepExecutor[TState] {
	if event == nil {
		return nil
	}
	if _, ok := event.(TEvent); !ok {
		return nil
	}

	for _, child := range w.children {
		executor := child.Match(state, event)
		if executor != nil {
			return executor
		}
	}

	return nil
}

// Do creates a step that invokes the provided handler for the current
// state. The handler runs synchronously and its returned error is
// propagated.
func Do[TState State](handle func(context.Context, TState) error, options ...StepOption) Step[TState] {
	step := &handleStep[TState]{
		handler: handle,
	}

	for _, option := range options {
		option(step)
	}

	return step
}

type handleStep[TState State] struct {
	recursion bool
	handler   func(context.Context, TState) error
}

func (step *handleStep[TState]) SetRecursion(value bool) {
	step.recursion = value
}

func (step *handleStep[TState]) Recursion() bool {
	return step.recursion
}

func (*handleStep[TState]) EventTypes() []reflect.Type {
	return nil
}

func (step *handleStep[TState]) Match(_ TState, _ Event) StepExecutor[TState] {
	return step
}

func (step *handleStep[TState]) Execute(ctx context.Context, _ *octo.Container, state TState, _ Event) error {
	return step.handler(ctx, state)
}

// DoEvent creates a step that handles an event of type TEvent using
// the supplied handler which receives the typed event.
func DoEvent[TEvent Event, TState State](handle func(context.Context, TState, TEvent) error, options ...StepOption) Step[TState] {
	step := &handleEventStep[TEvent, TState]{
		handler: handle,
	}

	for _, option := range options {
		option(step)
	}

	return step
}

type handleEventStep[TEvent Event, TState State] struct {
	recursion bool
	handler   func(context.Context, TState, TEvent) error
}

func (step *handleEventStep[TEvent, TState]) SetRecursion(value bool) {
	step.recursion = value
}

func (*handleEventStep[TEvent, TState]) EventTypes() []reflect.Type {
	return []reflect.Type{reflect.TypeFor[TEvent]()}
}

func (step *handleEventStep[TEvent, TState]) Match(_ TState, event Event) StepExecutor[TState] {
	if event == nil {
		return nil
	}
	if _, ok := event.(TEvent); ok {
		return step
	}
	return nil
}

func (step *handleEventStep[TEvent, TState]) Recursion() bool {
	return step.recursion
}

func (step *handleEventStep[TEvent, TState]) Execute(ctx context.Context, _ *octo.Container, state TState, event Event) error {
	return step.handler(ctx, state, event.(TEvent))
}
