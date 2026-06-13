package flow

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/oesand/octo"
	"github.com/oesand/octo/internal"
	"github.com/oesand/octo/mediator"
)

type Flow[TState State] interface {
	Name() string
	mediator.MassEventHandler
}

func Declare[TState State](container *octo.Container, steps ...Step[TState]) Flow[TState] {
	if len(steps) == 0 {
		panic(errors.New("flow: must declare at least one step"))
	}

	var events internal.Set[reflect.Type]
	for _, child := range steps {
		events.Add(child.EventTypes()...)
	}
	events.Add(reflect.TypeFor[*triggerEvent]())

	var zeroState TState
	flowName := zeroState.Flow()
	if flowName == "" {
		panic(errors.New("flow: name cannot be empty"))
	}
	return &flowDeclaration[TState]{
		name:      flowName,
		steps:     steps,
		events:    events.Values(),
		container: container,
	}
}

type flowDeclaration[TState State] struct {
	name      string
	steps     []Step[TState]
	events    []reflect.Type
	container *octo.Container
}

func (f *flowDeclaration[TState]) Name() string {
	return f.name
}

func (f *flowDeclaration[TState]) EventTypes() []reflect.Type {
	return f.events
}

func (f *flowDeclaration[TState]) Handle(ctx context.Context, event any) error {
	flowEvent, ok := event.(Event)
	if !ok {
		return errors.New("flow: not a flow event")
	}
	return f.Execute(ctx, flowEvent)
}

func (f *flowDeclaration[TState]) Execute(ctx context.Context, event Event) error {
	if f.name != event.Flow() {
		return nil
	}

	manager := octo.Resolve[Manager](f.container)

	var state TState
	err := manager.GetState(ctx, event.Uid(), &state)
	if err != nil {
		return err
	}

	prevStep := state.GetStep()

	for {
		executor := f.findExecutor(state, event)
		if executor == nil {
			break
		}

		triggerCtx := new(triggerFlowCtx)
		ctx = context.WithValue(ctx, triggerFlowCtxKey, triggerCtx)

		err = executor.Execute(ctx, f.container, state, event)
		if err != nil {
			saveErr := manager.SaveError(ctx, event, err)
			if saveErr != nil {
				return fmt.Errorf("flow: fail to save error '%w': %w", err, saveErr)
			}
			return err
		}

		err = manager.SaveState(ctx, event.Uid(), state, triggerCtx.transactionCallbacks)
		if err != nil {
			return err
		}

		if state.Finished() || triggerCtx.abort {
			break
		}

		if !executor.Recursion() && prevStep == state.GetStep() {
			break
		}

		prevStep = state.GetStep()
	}
	return nil
}

func (f *flowDeclaration[TState]) findExecutor(state TState, event Event) StepExecutor[TState] {
	var executor StepExecutor[TState]
	for _, step := range f.steps {
		executor = step.Match(state, event)
		if executor != nil {
			break
		}
	}
	return executor
}
