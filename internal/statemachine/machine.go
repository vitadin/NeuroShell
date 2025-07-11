// Package statemachine implements the core state machine for NeuroShell command execution.
// It provides a unified execution pipeline that handles builtin commands, stdlib scripts,
// and user scripts through well-defined state transitions with integrated interpolation.
package statemachine

import (
	"fmt"

	"github.com/charmbracelet/log"
	"neuroshell/internal/context"
	"neuroshell/internal/data/embedded"
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
	"neuroshell/pkg/neurotypes"
)

// StateMachine implements the core state machine for NeuroShell command execution.
// It provides a unified execution pipeline that handles builtin commands, stdlib scripts,
// and user scripts through well-defined state transitions with integrated interpolation.
type StateMachine struct {
	// Direct reference to the context for state management
	context *context.NeuroContext
	// Integrated interpolation engine (not a separate service)
	interpolator *CoreInterpolator
	// Stdlib script loader for embedded scripts
	stdlibLoader *embedded.StdlibLoader
	// Configuration options
	config neurotypes.StateMachineConfig
	// State stack for recursive execution
	stateStack []neurotypes.StateSnapshot
	// Custom styled logger for State Machine operations
	logger *log.Logger

	// Internal execution state (Phase 1 implementation)
	// These will be moved to context in Phase 2
	currentState      neurotypes.State
	executionInput    string
	executionError    error
	recursionDepth    int
	parsedCommand     *parser.Command
	resolvedCommand   *neurotypes.StateMachineResolvedCommand
	scriptLines       []string
	currentScriptLine int
	// Try command execution state
	tryTargetCommand string // Command to execute in try mode
}

// NewStateMachine creates a new state machine with the given context and configuration.
func NewStateMachine(ctx *context.NeuroContext, config neurotypes.StateMachineConfig) *StateMachine {
	sm := &StateMachine{
		context:      ctx,
		interpolator: NewCoreInterpolator(ctx),
		stdlibLoader: embedded.NewStdlibLoader(),
		config:       config,
		stateStack:   make([]neurotypes.StateSnapshot, 0),
	}

	// Initialize custom styled logger
	sm.logger = logger.NewStyledLogger("StateMachine")

	return sm
}

// NewStateMachineWithDefaults creates a new state machine with default configuration.
func NewStateMachineWithDefaults(ctx *context.NeuroContext) *StateMachine {
	return NewStateMachine(ctx, neurotypes.DefaultStateMachineConfig())
}

// Execute is the main entry point for the state machine execution.
// It processes the input through the complete state machine pipeline until completion or error.
func (sm *StateMachine) Execute(input string) error {
	// Initialize execution state (full reset for main entry point)
	sm.initializeExecution(input)

	sm.logger.Debug("Execute command", "input", input)

	return sm.executeInternal()
}

// ExecuteInternal executes a command without resetting the global execution state.
// This is used for nested execution (e.g., by try commands, script lines).
func (sm *StateMachine) ExecuteInternal(input string) error {
	// Save current execution state
	snapshot := sm.saveExecutionState()

	// Set up new execution context without full reset
	sm.setState(neurotypes.StateReceived)
	sm.setExecutionInput(input)
	sm.setExecutionError(nil)
	sm.clearCommandData() // Clear command-specific data but keep recursion/script state

	sm.logger.Debug("ExecuteInternal command", "input", input)

	// Run the state machine
	err := sm.executeInternal()

	// Restore execution state
	sm.restoreExecutionState(snapshot)

	return err
}

// executeInternal contains the core state machine loop used by both Execute and ExecuteInternal.
func (sm *StateMachine) executeInternal() error {
	// Main state machine loop
	for {
		currentState := sm.getCurrentState()

		// Check for terminal states
		if currentState == neurotypes.StateCompleted || currentState == neurotypes.StateError {
			break
		}

		// Process current state
		if err := sm.ProcessCurrentState(); err != nil {
			sm.setExecutionError(err)
			sm.setState(neurotypes.StateError)
			break
		}

		// Determine and transition to next state
		nextState := sm.DetermineNextState()
		sm.setState(nextState)

		// Safety check to prevent infinite loops
		// Special case: StateScriptExecuting can legitimately stay in the same state
		// when processing multiple script lines
		if nextState == currentState && currentState != neurotypes.StateScriptExecuting {
			sm.logger.Error("Infinite loop detected", "state", currentState.String())
			return fmt.Errorf("state machine stuck in state: %s", currentState.String())
		}
	}

	return sm.getExecutionError()
}

// DetermineNextState determines the next state based on the current state and execution context.
// Each state has a single, predictable next state with no conditional logic.
func (sm *StateMachine) DetermineNextState() neurotypes.State {
	currentState := sm.getCurrentState()
	switch currentState {
	case neurotypes.StateReceived:
		// Check if we need variable interpolation
		input := sm.getExecutionInput()
		if sm.interpolator.HasVariables(input) {
			return neurotypes.StateInterpolating
		}
		return neurotypes.StateParsing
	case neurotypes.StateInterpolating:
		return neurotypes.StateReceived // Recursive re-entry with expanded input
	case neurotypes.StateParsing:
		return neurotypes.StateResolving
	case neurotypes.StateResolving:
		// Next state is determined by what was resolved
		resolved := sm.getResolvedCommand()
		if resolved == nil {
			return neurotypes.StateError
		}
		switch resolved.Type {
		case neurotypes.CommandTypeTry:
			return neurotypes.StateTryResolving
		case neurotypes.CommandTypeBuiltin:
			return neurotypes.StateExecuting
		case neurotypes.CommandTypeStdlib, neurotypes.CommandTypeUser:
			return neurotypes.StateScriptLoaded
		default:
			return neurotypes.StateError
		}
	case neurotypes.StateTryResolving:
		return neurotypes.StateTryExecuting
	case neurotypes.StateExecuting:
		return neurotypes.StateCompleted
	case neurotypes.StateTryExecuting:
		return neurotypes.StateTryCompleted
	case neurotypes.StateScriptLoaded:
		return neurotypes.StateScriptExecuting
	case neurotypes.StateScriptExecuting:
		// Check if more script lines to process
		if sm.hasMoreScriptLines() {
			return neurotypes.StateScriptExecuting // Stay in script execution to process next line
		}
		return neurotypes.StateCompleted
	case neurotypes.StateTryCompleted:
		return neurotypes.StateCompleted
	default:
		return neurotypes.StateError
	}
}

// initializeExecution sets up the initial state for execution.
func (sm *StateMachine) initializeExecution(input string) {
	// Reset execution state
	sm.setState(neurotypes.StateReceived)
	sm.setExecutionInput(input)
	sm.setExecutionError(nil)
	sm.resetRecursionDepth()
	sm.clearExecutionData()
}
