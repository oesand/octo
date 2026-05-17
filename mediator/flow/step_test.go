package flow

import (
	"context"
	"reflect"
	"testing"

	"github.com/oesand/octo"
)

// TestRecursive_ReturnsStepOption verifies that Recursive returns a valid StepOption.
func TestRecursive_ReturnsStepOption(t *testing.T) {
	option := Recursive()

	// Verify it's a function that matches StepOption signature
	if option == nil {
		t.Error("Recursive() returned nil")
	}

	// Create a mock stepExecutorOptions to verify it can be called
	mock := &mockStepExecutorOptions{}
	option(mock)

	if mock.recursionSet == false {
		t.Error("option should set recursion to true")
	}
}

// TestRecursive_SetsRecursionTrue verifies that Recursive sets the recursion flag to true.
func TestRecursive_SetsRecursionTrue(t *testing.T) {
	mock := &mockStepExecutorOptions{
		recursionSet: false,
	}

	option := Recursive()
	option(mock)

	if !mock.recursionSet {
		t.Error("Recursive() should set recursion to true")
	}

	if mock.recursionValue != true {
		t.Error("Recursive() should set recursion value to true")
	}
}

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

// ===== Mock Types for Testing =====

type mockStepExecutorOptions struct {
	recursionSet   bool
	recursionValue bool
}

func (m *mockStepExecutorOptions) SetRecursion(value bool) {
	m.recursionSet = true
	m.recursionValue = value
}

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
