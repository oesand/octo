package flow

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/oesand/octo"
	"github.com/oesand/octo/mediator"
	"github.com/oesand/octo/pm"
)

type Flow[TState State] interface {
	Name() string
	mediator.MassEventHandler
}

func Declare[TState State](container *octo.Container, steps ...Step[TState]) Flow[TState] {
	var events pm.Set[reflect.Type]
	for _, child := range steps {
		events.Add(child.EventTypes()...)
	}
	events.Add(reflect.TypeFor[*triggerEvent]())

	var zeroState TState
	return &flowDeclaration[TState]{
		name:      zeroState.Flow(),
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
	var ptr State = state
	err := manager.GetState(ctx, event.Uid(), &ptr)
	if err != nil {
		return err
	}
	state = ptr.(TState)

	prevStep := state.GetStep()

	for {
		handled, err := f.handleNext(ctx, state, event)
		if !handled {
			return nil
		}

		if err != nil {
			saveErr := manager.SaveError(ctx, event.Uid(), err)
			if saveErr != nil {
				return fmt.Errorf("flow: fail to save error '%w': %w", err, saveErr)
			}
			return err
		}

		err = manager.SaveState(ctx, event.Uid(), state)
		if err != nil {
			return err
		}

		if state.Finished() || prevStep == state.GetStep() {
			break
		}
		prevStep = state.GetStep()
	}
	return nil
}

func (f *flowDeclaration[TState]) handleNext(ctx context.Context, state TState, event Event) (bool, error) {
	for _, step := range f.steps {
		handled, err := step.Handle(ctx, f.container, state, event)
		if handled {
			return true, err
		}
	}
	return false, nil
}
