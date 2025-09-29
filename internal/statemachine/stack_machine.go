// Package statemachine implements the stack-based execution engine for NeuroShell.
// The StackMachine replaces the complex state-machine-driven model with a simpler
// stack-based approach while preserving all existing functionality.
package statemachine

import (
	"fmt"
	"neuroshell/internal/commands"
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
	"strings"

	"github.com/charmbracelet/log"
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
	// Silent handler for output suppression management
	silentHandler *SilentHandler
	// Configuration options
	config neurotypes.StateMachineConfig
	// Custom styled logger
	logger *log.Logger
	// Services
	stackService    *services.StackService
	variableService *services.VariableService
	errorService    *services.ErrorManagementService
}

// NewStackMachine creates a new stack-based execution engine.
func NewStackMachine(ctx *context.NeuroContext, config neurotypes.StateMachineConfig) *StackMachine {
	sm := &StackMachine{
		context:        ctx,
		stateProcessor: NewStateProcessor(ctx, config),
		tryHandler:     NewTryHandler(),
		silentHandler:  NewSilentHandler(),
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

	sm.errorService, err = services.GetGlobalErrorManagementService()
	if err != nil {
		sm.logger.Error("Failed to get error management service", "error", err)
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
	iterationCount := 0
	for !sm.stackService.IsEmpty() {
		iterationCount++

		// Debug: Check for potential infinite loops
		if iterationCount > 10000 {
			sm.logger.Error("POTENTIAL INFINITE LOOP: Too many iterations in processStack", "iterations", iterationCount)
			stackContents := sm.stackService.PeekStack()
			sm.logger.Error("Current stack contents", "stack", stackContents, "stackSize", len(stackContents))
			return fmt.Errorf("infinite loop detected in stack processing")
		}

		rawCommand, hasCommand := sm.stackService.PopCommand()
		if !hasCommand {
			break // Stack is empty
		}

		sm.logger.Debug("Processing stack command", "iteration", iterationCount, "command", rawCommand, "stackSize", sm.stackService.GetStackSize())

		// Process individual command through state pipeline
		err := sm.processCommand(rawCommand)
		if err != nil {
			sm.logger.Debug("Command error occurred", "command", rawCommand, "error", err, "inTryBlock", sm.tryHandler.IsInTryBlock())
			// Check if we're in a try block
			if sm.tryHandler.IsInTryBlock() {
				// Try block error capture using TryHandler
				sm.logger.Debug("Handling try block error", "command", rawCommand)
				sm.tryHandler.HandleTryError(err)
				sm.logger.Debug("Skipping to try block end", "command", rawCommand)
				sm.tryHandler.SkipToTryBlockEnd()
				sm.logger.Debug("Continuing after try block", "stackSize", sm.stackService.GetStackSize())
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

	// Check for silent boundary markers using SilentHandler
	if isMarker, silentID, isStart := sm.silentHandler.IsSilentBoundaryMarker(rawCommand); isMarker {
		if isStart {
			sm.silentHandler.EnterSilentBlock(silentID)
		} else {
			sm.silentHandler.ExitSilentBlock(silentID)
		}
		return nil
	}

	// Reset error state before processing command (moves current to last, resets current to success)
	// But only for commands that can change system state - not for read-only commands like \get
	shouldReset := sm.shouldResetErrorState(rawCommand)
	sm.logger.Debug("Error state reset decision", "command", rawCommand, "shouldReset", shouldReset)
	if sm.errorService != nil && shouldReset {
		sm.logger.Debug("Resetting error state before command", "command", rawCommand)
		if err := sm.errorService.ResetErrorState(); err != nil {
			sm.logger.Debug("Failed to reset error state", "error", err)
		}
	}

	// Output command line with %%> prefix if echo_commands is enabled and not in silent block
	if sm.config.EchoCommands && !sm.silentHandler.IsInSilentBlock() {
		fmt.Printf("%%%%> %q\n", rawCommand)
	}

	// Use the state processor to handle the command through the proven pipeline
	var err error
	var capturedOutput string
	if sm.silentHandler.IsInSilentBlock() {
		err = stringprocessing.WithSuppressedOutput(func() error {
			return sm.stateProcessor.ProcessCommand(rawCommand)
		})
		// No output capture in silent blocks
		capturedOutput = ""
	} else {
		// Capture output during command execution for ALL commands (including read-only ones)
		capturedOutput, err = stringprocessing.WithCapturedOutput(func() error {
			return sm.stateProcessor.ProcessCommand(rawCommand)
		})
	}

	// Set error state based on command execution result
	// But only for commands that can change system state - not for read-only commands like \get
	if sm.errorService != nil && sm.shouldResetErrorState(rawCommand) {
		sm.logger.Debug("Setting error state from command result", "command", rawCommand, "err", err)
		if setErr := sm.errorService.SetErrorStateFromCommandResult(err); setErr != nil {
			sm.logger.Debug("Failed to set error state", "error", setErr)
		}
	}

	// Capture output after command execution and display it to the user
	// Do this for ALL commands to track output history properly
	// This should never fail - worst case we store nothing but don't affect other components
	if sm.context != nil {
		sm.logger.Debug("Capturing command output", "command", rawCommand, "outputLength", len(capturedOutput))

		// Safely capture output - this should never panic or fail
		func() {
			defer func() {
				if r := recover(); r != nil {
					sm.logger.Debug("Output capture failed but continuing safely", "error", r)
				}
			}()
			sm.context.CaptureOutput(capturedOutput)
		}()

		// Display the captured output to the user (since we intercepted it)
		fmt.Print(capturedOutput)
	}

	return err
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

// shouldResetErrorState determines if error state should be reset before executing a command.
// Read-only commands like \get should not reset error state to preserve try block error capture.
func (sm *StackMachine) shouldResetErrorState(rawCommand string) bool {
	// Parse the command to get the command name
	cmd := strings.TrimSpace(rawCommand)
	if !strings.HasPrefix(cmd, "\\") {
		return true // Non-NeuroShell commands should reset error state
	}

	// Extract command name (everything after \ until first [ or space)
	cmdName := cmd[1:] // Remove leading \
	if idx := strings.IndexAny(cmdName, "[ "); idx != -1 {
		cmdName = cmdName[:idx]
	}

	// Get command from registry
	command, exists := commands.GetGlobalRegistry().Get(cmdName)
	if !exists {
		return true // Unknown commands reset error state
	}

	// Use context to check if command is read-only (considers both self-declaration and overrides)
	// Don't reset error state for read-only commands
	return !sm.context.IsCommandReadOnly(command)
}
