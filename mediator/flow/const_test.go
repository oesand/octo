package flow_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/oesand/octo"
	"github.com/oesand/octo/mediator/flow"
)

func TestTransactional_CallbackRunsDuringSaveState(t *testing.T) {
	var initialHandled atomic.Int32
	var nextHandled atomic.Int32
	var callbackCalled atomic.Int32

	container := octo.New()

	testFlow := flow.Declare(
		container,
		flow.Initial(
			flow.Do(func(ctx context.Context, state *TestState) error {
				initialHandled.Add(1)
				// register transactional callback
				flow.Transactional(ctx, func(ctx context.Context) error {
					callbackCalled.Add(1)
					return nil
				})
				state.Step = "next"
				return nil
			}),
		),
		flow.On("next", flow.Do(func(ctx context.Context, state *TestState) error {
			nextHandled.Add(1)
			state.Step = "finished"
			return nil
		})),
	)

	ctx := context.Background()
	uid := flow.NewUid()

	manager := &flow.MemoryManager{}
	octo.InjectValue(container, manager)
	state := &TestState{}
	_ = manager.Create(ctx, uid, state)

	err := testFlow.Handle(ctx, flow.TriggerEvent(uid, state.Flow()))
	if err != nil {
		t.Error(err)
	}

	if c := initialHandled.Load(); c != 1 {
		t.Errorf("initial handled %d times", c)
	}
	if c := nextHandled.Load(); c != 1 {
		t.Errorf("next handled %d times", c)
	}
	if c := callbackCalled.Load(); c != 1 {
		t.Errorf("transactional callback called %d times", c)
	}
}

func TestAbort_SetsAbortAndStopsFurtherSteps(t *testing.T) {
	var initialHandled atomic.Int32
	var nextHandled atomic.Int32

	container := octo.New()

	testFlow := flow.Declare(
		container,
		flow.Initial(
			flow.Do(func(ctx context.Context, state *TestState) error {
				initialHandled.Add(1)
				// abort further processing
				flow.Abort(ctx)
				state.Step = "next"
				return nil
			}),
		),
		flow.On("next", flow.Do(func(ctx context.Context, state *TestState) error {
			nextHandled.Add(1)
			state.Step = "finished"
			return nil
		})),
	)

	ctx := context.Background()
	uid := flow.NewUid()

	manager := &flow.MemoryManager{}
	octo.InjectValue(container, manager)
	state := &TestState{}
	_ = manager.Create(ctx, uid, state)

	err := testFlow.Handle(ctx, flow.TriggerEvent(uid, state.Flow()))
	if err != nil {
		t.Error(err)
	}

	if c := initialHandled.Load(); c != 1 {
		t.Errorf("initial handled %d times", c)
	}
	if c := nextHandled.Load(); c != 0 {
		t.Errorf("next handled %d times; expected 0 because of Abort", c)
	}
}
