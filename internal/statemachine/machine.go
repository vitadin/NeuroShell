// Package statemachine provides a simple driver interface for the stack-based execution engine.
// The StateMachine delegates all execution to the StackMachine while preserving interfaces for extensibility.
package statemachine

import (
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"

	"github.com/charmbracelet/log"
)

// StateMachine provides a simple driver interface for the stack-based execution engine.
// It delegates all execution to the StackMachine while preserving the interface for future extensibility.
type StateMachine struct {
	// Direct reference to the context for state management
	context *context.NeuroContext
	// Stack-based execution engine that handles all command processing
	stackMachine *StackMachine
	// Configuration options (kept for interface compatibility)
	config neurotypes.StateMachineConfig
	// Custom styled logger for State Machine operations
	logger *log.Logger
}

// NewStateMachine creates a new state machine with the given context and configuration.
func NewStateMachine(ctx *context.NeuroContext, config neurotypes.StateMachineConfig) *StateMachine {
	sm := &StateMachine{
		context:      ctx,
		stackMachine: NewStackMachine(ctx, config),
		config:       config,
		logger:       logger.NewStyledLogger("StateMachine"),
	}

	return sm
}

// NewStateMachineWithDefaults creates a new state machine with default configuration.
func NewStateMachineWithDefaults(ctx *context.NeuroContext) *StateMachine {
	return NewStateMachine(ctx, neurotypes.DefaultStateMachineConfig())
}

// Execute is the main entry point for the state machine execution.
// It delegates all execution to the StackMachine while preserving the interface.
func (sm *StateMachine) Execute(input string) error {
	sm.logger.Debug("StateMachine Execute", "input", input)
	return sm.stackMachine.Execute(input)
}

// ExecuteInternal executes a command without resetting the global execution state.
// This is used for nested execution (e.g., by try commands, script lines).
func (sm *StateMachine) ExecuteInternal(input string) error {
	sm.logger.Debug("StateMachine ExecuteInternal", "input", input)
	return sm.stackMachine.ExecuteInternal(input)
}

// GetConfig returns the current configuration.
// This method is preserved for interface compatibility and future extensibility.
func (sm *StateMachine) GetConfig() neurotypes.StateMachineConfig {
	return sm.stackMachine.GetConfig()
}

// SetConfig updates the configuration.
// This method is preserved for interface compatibility and future extensibility.
func (sm *StateMachine) SetConfig(config neurotypes.StateMachineConfig) {
	sm.config = config
	sm.stackMachine.SetConfig(config)
}
