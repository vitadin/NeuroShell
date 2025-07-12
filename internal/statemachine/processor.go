package statemachine

import (
	"fmt"
	"strings"

	"neuroshell/internal/parser"
	"neuroshell/pkg/neurotypes"
)

// ProcessCurrentState handles the logic for the current execution state.
func (sm *StateMachine) ProcessCurrentState() error {
	currentState := sm.getCurrentState()
	switch currentState {
	case neurotypes.StateReceived:
		return sm.processReceived()
	case neurotypes.StateInterpolating:
		return sm.processInterpolating()
	case neurotypes.StateParsing:
		return sm.processParsing()
	case neurotypes.StateResolving:
		return sm.processResolving()
	case neurotypes.StateExecuting:
		return sm.processExecuting()
	case neurotypes.StateScriptLoaded:
		return sm.processScriptLoaded()
	case neurotypes.StateScriptExecuting:
		return sm.processScriptExecuting()
	case neurotypes.StateTryResolving:
		return sm.processTryResolving()
	case neurotypes.StateTryExecuting:
		return sm.processTryExecuting()
	case neurotypes.StateTryCompleted:
		return sm.processTryCompleted()
	default:
		return fmt.Errorf("unknown state: %s", currentState.String())
	}
}

// processReceived handles the initial state when a command is received.
func (sm *StateMachine) processReceived() error {
	input := sm.getExecutionInput()

	if input == "" {
		return fmt.Errorf("empty input received")
	}

	// Entry point - processing logic is in DetermineNextState
	return nil
}

// processInterpolating handles variable and macro expansion.
func (sm *StateMachine) processInterpolating() error {
	input := sm.getExecutionInput()

	expanded, hasVariables, err := sm.interpolator.InterpolateCommandLine(input)
	if err != nil {
		return fmt.Errorf("variable expansion failed: %w", err)
	}

	if hasVariables {
		// Check recursion limit
		recursionDepth := sm.getRecursionDepth()
		if recursionDepth >= sm.config.RecursionLimit {
			return fmt.Errorf("variable expansion nested too deeply (limit: %d)", sm.config.RecursionLimit)
		}

		// Increment recursion depth and set up for recursive re-entry
		sm.incrementRecursionDepth()
		sm.setExecutionInput(expanded)
		sm.logger.Debug("Variable expansion", "depth", recursionDepth+1)
	}

	return nil
}

// processParsing handles command structure parsing.
func (sm *StateMachine) processParsing() error {
	input := sm.getExecutionInput()

	cmd := parser.ParseInput(input)
	if cmd == nil {
		return fmt.Errorf("failed to parse command: %s", input)
	}

	sm.setParsedCommand(cmd)
	return nil
}

// processResolving handles command resolution through the priority system.
func (sm *StateMachine) processResolving() error {
	parsedCmd := sm.getParsedCommand()
	if parsedCmd == nil {
		return fmt.Errorf("no parsed command available")
	}

	// Handle try command as a special case FIRST
	if parsedCmd.Name == "try" {
		// Create a special resolved command for try
		resolved := &neurotypes.StateMachineResolvedCommand{
			Name: "try",
			Type: neurotypes.CommandTypeTry,
		}
		sm.setResolvedCommand(resolved)
		return nil
	}

	// Priority-based resolution: builtin → stdlib → user
	resolved, err := sm.resolveCommand(parsedCmd.Name)
	if err != nil {
		return err // Return the original error from resolver
	}

	sm.setResolvedCommand(resolved)
	return nil
}

// processExecuting handles execution of builtin commands.
func (sm *StateMachine) processExecuting() error {
	resolved := sm.getResolvedCommand()
	parsedCmd := sm.getParsedCommand()
	input := sm.getExecutionInput()

	if resolved == nil || resolved.Type != neurotypes.CommandTypeBuiltin {
		return fmt.Errorf("no builtin command to execute")
	}

	// Output command line with %%> prefix if echo_commands is enabled
	if sm.config.EchoCommands {
		fmt.Printf("%%%%> %s\n", input)
	}

	// Execute the builtin command
	err := resolved.BuiltinCommand.Execute(parsedCmd.Options, parsedCmd.Message)
	if err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

// processTryResolving handles try command setup and target extraction.
func (sm *StateMachine) processTryResolving() error {
	parsedCmd := sm.getParsedCommand()
	if parsedCmd == nil {
		return fmt.Errorf("no parsed command available for try resolution")
	}

	// Extract target command from the try command message
	targetCommand := strings.TrimSpace(parsedCmd.Message)

	// Always store the target command (even if empty)
	sm.setTryTargetCommand(targetCommand)

	if targetCommand == "" {
		// Empty try command - set success variables and mark as completed
		_ = sm.context.SetSystemVariable("_status", "0")
		_ = sm.context.SetSystemVariable("_error", "")
		_ = sm.context.SetSystemVariable("_output", "")
		return nil
	}

	return nil
}

// processTryExecuting executes the try target command with error capture.
func (sm *StateMachine) processTryExecuting() error {
	targetCommand := sm.getTryTargetCommand()
	if targetCommand == "" {
		// Empty try command - set success variables
		_ = sm.context.SetSystemVariable("_status", "0")
		_ = sm.context.SetSystemVariable("_error", "")
		_ = sm.context.SetSystemVariable("_output", "")
		return nil
	}

	// Clear status variables before execution to ensure clean state
	_ = sm.context.SetSystemVariable("_status", "")
	_ = sm.context.SetSystemVariable("_error", "")

	// Get the current _output value before execution to capture command output
	previousOutput, _ := sm.context.GetVariable("_output")

	// Execute the target command using internal execution (no global state reset)
	err := sm.ExecuteInternal(targetCommand)

	// Get the _output value after execution to see what the command produced
	currentOutput, _ := sm.context.GetVariable("_output")

	// Check what status variables were set by the command after execution
	status, statusErr := sm.context.GetVariable("_status")
	errorVar, errorErr := sm.context.GetVariable("_error")

	if err != nil {
		// Command execution failed at the state machine level
		// Check if the command already set status variables (like bash does)
		if statusErr != nil || status == "" {
			// Command didn't set status variables, so it's a real execution failure
			_ = sm.context.SetSystemVariable("_status", "1")

			// Unwrap script executor error messages to get the original error
			errorMsg := err.Error()
			if strings.HasPrefix(errorMsg, "command execution failed") {
				// Extract the original error message after the colon and space
				if idx := strings.Index(errorMsg, ": "); idx != -1 {
					errorMsg = errorMsg[idx+2:]
				}
			}
			_ = sm.context.SetSystemVariable("_error", errorMsg)
		}
		// else: Command already set status/error variables (like bash), keep them
	} else {
		// Command executed successfully at the state machine level
		// Check if it set status variables (like bash command exit codes)
		if statusErr != nil || status == "" {
			// Command didn't set status variables, so it succeeded
			_ = sm.context.SetSystemVariable("_status", "0")
		}
		if errorErr != nil || errorVar == "" {
			// Command didn't set error variable, so clear it for success
			_ = sm.context.SetSystemVariable("_error", "")
		}
		// else: Command already set status/error variables, keep them
	}

	// Always update output if it changed
	if currentOutput != previousOutput {
		_ = sm.context.SetSystemVariable("_output", currentOutput)
	}

	// Try command never fails - it always captures errors
	return nil
}

// processTryCompleted handles the completion of a try command.
func (sm *StateMachine) processTryCompleted() error {
	// Clean up try-specific state
	sm.setTryTargetCommand("")

	// Try command execution is complete - the variables are already set
	return nil
}
