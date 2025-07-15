// Package statemachine implements the stack-based execution engine for NeuroShell.
// The StackMachine replaces the complex state-machine-driven model with a simpler
// stack-based approach while preserving all existing functionality.
package statemachine

import (
	"fmt"
	"github.com/charmbracelet/log"
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// StackMachine implements the stack-based execution engine for NeuroShell.
// It processes commands from a stack in LIFO order, handling try blocks,
// error boundaries, and command execution through a unified pipeline.
type StackMachine struct {
	// Context for state management
	context *context.NeuroContext
	// State processor for individual command processing
	stateProcessor *StateProcessor
	// Try handler for error boundary management
	tryHandler *TryHandler
	// Configuration options
	config neurotypes.StateMachineConfig
	// Custom styled logger
	logger *log.Logger
	// Services
	stackService    *services.StackService
	variableService *services.VariableService
}

// NewStackMachine creates a new stack-based execution engine.
func NewStackMachine(ctx *context.NeuroContext, config neurotypes.StateMachineConfig) *StackMachine {
	sm := &StackMachine{
		context:        ctx,
		stateProcessor: NewStateProcessor(ctx, config),
		tryHandler:     NewTryHandler(),
		config:         config,
		logger:         logger.NewStyledLogger("StackMachine"),
	}

	// Initialize services
	var err error
	sm.stackService, err = services.GetGlobalStackService()
	if err != nil {
		sm.logger.Error("Failed to get stack service", "error", err)
	}

	sm.variableService, err = services.GetGlobalVariableService()
	if err != nil {
		sm.logger.Error("Failed to get variable service", "error", err)
	}

	return sm
}

// Execute is the main entry point for stack-based execution.
// It pushes the input command to the stack and processes all commands until the stack is empty.
func (sm *StackMachine) Execute(input string) error {
	// Check if required services are available
	if sm.stackService == nil {
		return fmt.Errorf("stack service not available")
	}

	// Update echo configuration based on _echo_commands variable
	sm.updateEchoConfig()

	// Push the input command to the stack
	sm.stackService.PushCommand(input)

	// Process the entire stack
	return sm.processStack()
}

// ExecuteInternal executes a command without affecting the global execution state.
// This is used for nested execution (e.g., by try commands, script lines).
func (sm *StackMachine) ExecuteInternal(input string) error {
	// Check if required services are available
	if sm.stackService == nil {
		return fmt.Errorf("stack service not available")
	}

	// Update echo configuration
	sm.updateEchoConfig()

	// Push the command to the stack
	sm.stackService.PushCommand(input)

	// Process the stack (this will handle any nested commands)
	return sm.processStack()
}

// processStack is the main stack processing loop.
// It pops commands from the stack and processes them until the stack is empty.
func (sm *StackMachine) processStack() error {
	for !sm.stackService.IsEmpty() {
		rawCommand, hasCommand := sm.stackService.PopCommand()
		if !hasCommand {
			break // Stack is empty
		}

		// Process individual command through state pipeline
		err := sm.processCommand(rawCommand)
		if err != nil {
			// Check if we're in a try block
			if sm.tryHandler.IsInTryBlock() {
				// Try block error capture using TryHandler
				sm.tryHandler.HandleTryError(err)
				sm.tryHandler.SkipToTryBlockEnd()
				continue // Continue processing after try block
			}
			// Normal error propagation
			return err
		}

		// Update echo configuration after each command in case _echo_command was modified
		sm.updateEchoConfig()
	}
	return nil
}

// processCommand processes a single command through the command processing pipeline.
// This preserves the existing proven pipeline: Interpolation → Parsing → Resolving → Execution
func (sm *StackMachine) processCommand(rawCommand string) error {
	sm.logger.Debug("Processing command", "command", rawCommand)

	// Check for error boundary markers using TryHandler
	if isMarker, tryID, isStart := sm.tryHandler.IsErrorBoundaryMarker(rawCommand); isMarker {
		if isStart {
			sm.tryHandler.EnterTryBlock(tryID)
		} else {
			sm.tryHandler.ExitTryBlock(tryID)
		}
		return nil
	}

	// Use the state processor to handle the command through the proven pipeline
	return sm.stateProcessor.ProcessCommand(rawCommand)
}

// updateEchoConfig updates the echo configuration based on the _echo_command variable.
func (sm *StackMachine) updateEchoConfig() {
	if sm.context == nil || sm.variableService == nil {
		return
	}

	echoCommandVar, err := sm.variableService.Get("_echo_command")
	if err != nil {
		return
	}

	// Check for truthy values
	switch echoCommandVar {
	case "true", "1", "yes":
		sm.config.EchoCommands = true
	case "false", "0", "no", "":
		sm.config.EchoCommands = false
	}

	// Propagate config changes to state processor
	if sm.stateProcessor != nil {
		sm.stateProcessor.SetConfig(sm.config)
	}
}

// GetConfig returns the current configuration.
func (sm *StackMachine) GetConfig() neurotypes.StateMachineConfig {
	return sm.config
}

// SetConfig updates the configuration.
func (sm *StackMachine) SetConfig(config neurotypes.StateMachineConfig) {
	sm.config = config
	if sm.stateProcessor != nil {
		sm.stateProcessor.SetConfig(config)
	}
}
