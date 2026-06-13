package flow

import (
	"context"
	"reflect"
	"testing"

	"github.com/oesand/octo"
)

// TestRecursive_WithDoStep verifies that Recursive works correctly with Do steps.
func TestRecursive_WithDoStep(t *testing.T) {
	callCount := 0
	handler := func(ctx context.Context, state *mockState) error {
		callCount++
		return nil
	}

	// Create a Do step with Recursive option
	step := Do(handler, Recursive())

	// The step should be a StepExecutor that has Recursion() method
	executor := step.Match(&mockState{step: "test"}, nil)
	if executor == nil {
		t.Fatal("Do step should return an executor")
	}

	// Verify recursion is enabled
	if !executor.Recursion() {
		t.Error("Recursion() should return true for step with Recursive option")
	}

	// Execute the step
	ctx := context.Background()
	err := executor.Execute(ctx, &octo.Container{}, &mockState{step: "test"}, nil)
	if err != nil {
		t.Errorf("step execution failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("handler should have been called once, was called %d times", callCount)
	}
}

// TestRecursive_WithDoEventStep verifies that Recursive works correctly with DoEvent steps.
func TestRecursive_WithDoEventStep(t *testing.T) {
	callCount := 0
	handler := func(ctx context.Context, state *mockState, event *mockEvent) error {
		callCount++
		return nil
	}

	// Create a DoEvent step with Recursive option
	step := DoEvent(handler, Recursive())

	// Test matching with correct event type
	mockEvent := &mockEvent{}
	executor := step.Match(&mockState{}, mockEvent)
	if executor == nil {
		t.Fatal("DoEvent step should match the event")
	}

	// Verify recursion is enabled
	if !executor.Recursion() {
		t.Error("Recursion() should return true for DoEvent step with Recursive option")
	}

	// Execute the step
	ctx := context.Background()
	err := executor.Execute(ctx, &octo.Container{}, &mockState{}, mockEvent)
	if err != nil {
		t.Errorf("step execution failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("handler should have been called once, was called %d times", callCount)
	}
}

// TestRecursive_EventTypeFiltering verifies that Recursive works with event type filtering.
func TestRecursive_EventTypeFiltering(t *testing.T) {
	handler := func(ctx context.Context, state *mockState, event *mockEvent) error {
		return nil
	}

	step := OnEvent[*mockEvent](
		DoEvent(handler, Recursive()),
	)

	// Get the event types
	eventTypes := step.EventTypes()
	if len(eventTypes) != 1 {
		t.Errorf("OnEvent should have 1 event type, got %d", len(eventTypes))
	}

	expectedType := reflect.TypeOf((*mockEvent)(nil))
	if eventTypes[0] != expectedType {
		t.Errorf("event type should be %v, got %v", expectedType, eventTypes[0])
	}

	// Match with correct event
	mockEvent := &mockEvent{}
	executor := step.Match(&mockState{}, mockEvent)
	if executor == nil {
		t.Fatal("step should match the event")
	}

	// Verify recursion is enabled
	if !executor.Recursion() {
		t.Error("Recursion() should be true on matched event")
	}
}

func TestNameGetterState_CustomStruct(t *testing.T) {
	type CustomNameState struct {
		NameGetterState[CustomNameState]
	}
	var state CustomNameState
	if got := state.Flow(); got != "CustomName" {
		t.Fatalf("expected NameGetter, got %q", got)
	}
}

// ===== Mock Types for Testing =====

type mockState struct {
	step     string
	finished bool
	flow     string
}

func (m *mockState) GetStep() string {
	return m.step
}

func (m *mockState) Finished() bool {
	return m.finished
}

func (m *mockState) Flow() string {
	return m.flow
}

type mockEvent struct {
	uid  string
	flow string
	data string
}

func (m *mockEvent) Uid() string {
	if m.uid == "" {
		return "test-uid"
	}
	return m.uid
}

func (m *mockEvent) Flow() string {
	if m.flow == "" {
		return "test-flow"
	}
	return m.flow
}
