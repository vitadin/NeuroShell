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
	case neurotypes.StateTryError:
		return sm.processTryError()
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

	// Handle try command - similar to interpolation pattern
	if parsedCmd.Name == "try" {
		targetCommand := strings.TrimSpace(parsedCmd.Message)
		if targetCommand == "" {
			// Empty try command - set success variables and complete
			_ = sm.context.SetSystemVariable("_status", "0")
			_ = sm.context.SetSystemVariable("_error", "")
			return nil
		}

		// Set try mode and new input (like interpolation does)
		sm.tryMode = true
		sm.setExecutionInput(targetCommand)

		return nil // DetermineNextState will handle transition back to StateReceived
	}

	// Priority-based resolution: builtin → stdlib → user
	resolved, err := sm.resolveCommand(parsedCmd.Name)
	if err != nil {
		return fmt.Errorf("command not found: %s", parsedCmd.Name)
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

// processTryError handles errors in try mode by capturing them as variables.
func (sm *StateMachine) processTryError() error {
	// Set error variables from the execution error
	err := sm.getExecutionError()

	if err != nil {
		_ = sm.context.SetSystemVariable("_status", "1")
		_ = sm.context.SetSystemVariable("_error", err.Error())
	} else {
		_ = sm.context.SetSystemVariable("_status", "0")
		_ = sm.context.SetSystemVariable("_error", "")
	}

	// Reset try mode completely when handling error
	sm.tryMode = false

	return nil
}
