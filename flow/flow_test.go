package flow_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/oesand/octo"
	"github.com/oesand/octo/flow"
)

type TestState struct {
	Step string
}

func (t *TestState) GetStep() string {
	return t.Step
}

func (t *TestState) Finished() bool {
	return t.Step == "finished"
}

func (t *TestState) Flow() string {
	return "TestFlow"
}

func TestBasicFlow(t *testing.T) {
	var initialHandled atomic.Int32
	var nextHandled atomic.Int32

	container := octo.New()

	testFlow := flow.Declare(
		container,
		flow.Initial(
			flow.Do(func(ctx context.Context, state *TestState) error {
				initialHandled.Add(1)
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
		t.Errorf("initial handled %d times\n", c)
	}

	if c := nextHandled.Load(); c != 1 {
		t.Errorf("next handled %d times\n", c)
	}
}

type TestExternalEvent struct {
	uid, flag string
}

func (ev *TestExternalEvent) Uid() string {
	return ev.uid
}

func (*TestExternalEvent) Flow() string {
	return "TestFlow"
}

func TestFlow_OnEvent(t *testing.T) {
	var initialHandled atomic.Int32
	var nextHandled atomic.Int32

	container := octo.New()

	testFlow := flow.Declare(
		container,
		flow.Initial(
			flow.Do(func(ctx context.Context, state *TestState) error {
				initialHandled.Add(1)
				state.Step = "next"
				return nil
			}),
		),
		flow.OnEvent[*TestExternalEvent](flow.Do(func(ctx context.Context, state *TestState) error {
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
		t.Errorf("initial handled %d times\n", c)
	}

	err = testFlow.Handle(ctx, &TestExternalEvent{uid: uid})
	if err != nil {
		t.Error(err)
	}

	if c := nextHandled.Load(); c != 1 {
		t.Errorf("next handled %d times\n", c)
	}
}

func TestFlow_DoEvent(t *testing.T) {
	var initialHandled atomic.Int32
	var nextHandled atomic.Int32

	container := octo.New()
	flag := flow.NewUid()

	testFlow := flow.Declare(
		container,
		flow.Initial(
			flow.Do(func(ctx context.Context, state *TestState) error {
				initialHandled.Add(1)
				state.Step = "next"
				return nil
			}),
		),
		flow.When(func(state *TestState) bool {
			return state.GetStep() == "next"
		}, flow.DoEvent(func(ctx context.Context, state *TestState, event *TestExternalEvent) error {
			if event.flag == flag {
				nextHandled.Add(1)
				state.Step = "finished"
			}
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
		t.Errorf("initial handled %d times\n", c)
	}

	err = testFlow.Handle(ctx, &TestExternalEvent{
		uid:  uid,
		flag: flag,
	})
	if err != nil {
		t.Error(err)
	}

	if c := nextHandled.Load(); c != 1 {
		t.Errorf("next handled %d times\n", c)
	}
}
