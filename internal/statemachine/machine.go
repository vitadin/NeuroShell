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
	tryMode           bool // Track when we're in try mode
	tryCompleted      bool // Track when a try command completed successfully
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
	// Initialize execution state in context
	sm.initializeExecution(input)

	sm.logger.Info("Execute command", "input", input)

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
			// In try mode, errors go to StateTryError instead of StateError
			if sm.tryMode {
				sm.setState(neurotypes.StateTryError)
				// Continue to next iteration to process StateTryError
				continue
			}
			sm.setState(neurotypes.StateError)
			break
		}

		// Determine and transition to next state
		nextState := sm.DetermineNextState()
		sm.setState(nextState)

		// Safety check to prevent infinite loops
		if nextState == currentState {
			sm.logger.Error("Infinite loop detected", "state", currentState.String())
			return fmt.Errorf("state machine stuck in state: %s", currentState.String())
		}
	}

	// Handle successful completion of try command
	if sm.getCurrentState() == neurotypes.StateCompleted && sm.tryCompleted {
		sm.tryCompleted = false
		// Set success variables for try command
		_ = sm.context.SetSystemVariable("_status", "0")
		_ = sm.context.SetSystemVariable("_error", "")
	}

	// Return any execution error
	// In try mode, StateTryError should return nil (success)
	if sm.getCurrentState() == neurotypes.StateTryError {
		return nil
	}
	return sm.getExecutionError()
}

// DetermineNextState determines the next state based on the current state and execution context.
func (sm *StateMachine) DetermineNextState() neurotypes.State {
	currentState := sm.getCurrentState()
	switch currentState {
	case neurotypes.StateReceived:
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
		// Handle try command recursive re-entry (like interpolation)
		if sm.tryMode {
			// Set flag to track that we're executing a try command
			sm.tryCompleted = true
			sm.tryMode = false
			return neurotypes.StateReceived // Recursive re-entry with try target
		}

		// Determine if builtin command or script
		if sm.getResolvedBuiltinCommand() != nil {
			return neurotypes.StateExecuting
		}
		if sm.getScriptContent() != "" {
			return neurotypes.StateScriptLoaded
		}
		return neurotypes.StateError
	case neurotypes.StateExecuting:
		return neurotypes.StateCompleted
	case neurotypes.StateScriptLoaded:
		return neurotypes.StateScriptExecuting
	case neurotypes.StateScriptExecuting:
		// Check if more script lines to process
		if sm.hasMoreScriptLines() {
			return neurotypes.StateReceived // Recursive: next script line re-enters
		}
		return neurotypes.StateCompleted
	case neurotypes.StateTryError:
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
