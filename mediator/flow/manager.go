package flow

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/oesand/octo/mediator"
)

// Manager defines storage and execution behavior for flows. It is
// responsible for creating states, retrieving and saving state and
// scheduling or triggering the next processing step.
type Manager interface {
	Create(ctx context.Context, uid string, state State) error
	GetState(ctx context.Context, uid string, state any) error
	SaveError(ctx context.Context, event Event, err error) error
	SaveState(ctx context.Context, uid string, state State, callbacks []TransactionCallback) error
}

var _ Manager = &MemoryManager{}

// MemoryManager is an in-memory implementation of Manager intended
// primarily for tests and examples. It keeps states in a map keyed
// by UID.
type MemoryManager struct {
	saved map[string]*stateEntry
}

type stateEntry struct {
	uid       string
	state     State
	prevStep  string
	recursion int
	error     error
}

func (m *MemoryManager) Create(ctx context.Context, uid string, state State) error {
	return m.SaveState(ctx, uid, state, nil)
}

func (m *MemoryManager) GetState(_ context.Context, uid string, state any) error {
	if m.saved == nil {
		return fmt.Errorf("flow: no state found for %s", uid)
	}
	entry, ok := m.saved[uid]
	if !ok {
		return fmt.Errorf("flow: no state found for %s", uid)
	}
	reflect.ValueOf(state).Elem().Set(reflect.ValueOf(entry.state))
	return nil
}

func (m *MemoryManager) SaveError(_ context.Context, event Event, err error) error {
	uid := event.Uid()
	if m.saved == nil {
		return fmt.Errorf("flow: no state found for %s", uid)
	}
	entry, ok := m.saved[uid]
	if !ok {
		return fmt.Errorf("flow: no state found for %s", uid)
	}
	entry.error = err
	return nil
}

func (m *MemoryManager) SaveState(ctx context.Context, uid string, state State, callbacks []TransactionCallback) error {
	if m.saved == nil {
		m.saved = make(map[string]*stateEntry)
	}

	for _, callback := range callbacks {
		err := callback(ctx)
		if err != nil {
			return err
		}
	}

	entry, ok := m.saved[uid]
	if !ok {
		if !state.Finished() {
			entry = &stateEntry{
				uid:       uid,
				state:     state,
				prevStep:  state.GetStep(),
				recursion: 0,
				error:     nil,
			}
			m.saved[uid] = entry
		}
	} else {
		if state.Finished() {
			delete(m.saved, uid)
		} else {
			if entry.prevStep == state.GetStep() {
				entry.recursion++
				if entry.recursion > 10 {
					return fmt.Errorf("flow: no step changed 10 times, detected potential recursion")
				}
			} else {
				entry.recursion = 0
				entry.prevStep = state.GetStep()
			}
			entry.state = state
			entry.error = nil
		}
	}
	return nil
}

func (m *MemoryManager) TriggerNext(ctx context.Context, manager *mediator.Manager) error {
	for uid, entry := range m.saved {
		return Trigger(manager, ctx, uid, entry.state.Flow())
	}
	return errors.New("flow: not found")
}
