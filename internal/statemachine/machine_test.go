package statemachine

import (
	"testing"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// TestStateMachine_NewStateMachine tests state machine creation.
func TestStateMachine_NewStateMachine(t *testing.T) {
	ctx := context.New()
	config := neurotypes.DefaultStateMachineConfig()

	sm := NewStateMachine(ctx, config)

	if sm == nil {
		t.Fatal("Expected state machine to be created, got nil")
	}

	if sm.context != ctx {
		t.Error("Expected context to be set correctly")
	}

	if sm.config != config {
		t.Error("Expected config to be set correctly")
	}

	// Test initial state
	if sm.getCurrentState() != neurotypes.StateReceived {
		t.Errorf("Expected initial state to be StateReceived, got %s", sm.getCurrentState().String())
	}
}

// TestStateMachine_NewStateMachineWithDefaults tests creation with defaults.
func TestStateMachine_NewStateMachineWithDefaults(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	if sm == nil {
		t.Fatal("Expected state machine to be created, got nil")
	}

	// Test default configuration is applied
	defaultConfig := neurotypes.DefaultStateMachineConfig()
	if sm.config.RecursionLimit != defaultConfig.RecursionLimit {
		t.Error("Expected default recursion limit to be applied")
	}
}

// TestStateMachine_Execute tests basic command execution.
func TestStateMachine_Execute(t *testing.T) {
	ctx := context.New()
	sm := NewStateMachineWithDefaults(ctx)

	// Test simple command execution
	err := sm.Execute("\\set[test=value]")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test empty input
	err = sm.Execute("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}
