// Package orchestration provides workflow orchestration for NeuroShell operations.
// This package contains centralized logic for coordinating multiple services
// to accomplish complex operations like script execution.
package orchestration

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ExecuteScript orchestrates the complete script execution workflow.
// It loads a script file and executes all commands in sequence, handling
// interpolation, error tracking, and completion status.
//
// This function consolidates the script execution logic that was previously
// duplicated between batch mode and the \run command, ensuring consistent
// behavior and maintainability.
//
// Parameters:
//   - scriptPath: Path to the .neuro script file to execute
//   - ctx: Context for variable storage and execution state
//
// Returns:
//   - error: Any error that occurred during script execution
func ExecuteScript(scriptPath string, ctx neurotypes.Context) error {
	logger.Debug("Starting script execution", "script", scriptPath)

	// Phase 1: Get required services from global registry
	scriptService, err := services.GlobalRegistry.GetService("script")
	if err != nil {
		return fmt.Errorf("script service not available: %w", err)
	}

	executorService, err := services.GlobalRegistry.GetService("executor")
	if err != nil {
		return fmt.Errorf("executor service not available: %w", err)
	}

	interpolationService, err := services.GlobalRegistry.GetService("interpolation")
	if err != nil {
		return fmt.Errorf("interpolation service not available: %w", err)
	}

	variableService, err := services.GlobalRegistry.GetService("variable")
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Cast services to their concrete types for type safety
	ss := scriptService.(*services.ScriptService)
	es := executorService.(*services.ExecutorService)
	is := interpolationService.(*services.InterpolationService)
	vs := variableService.(*services.VariableService)

	// Phase 2: Load script file into execution queue
	if err := ss.LoadScript(scriptPath, ctx); err != nil {
		return fmt.Errorf("failed to load script: %w", err)
	}

	logger.Debug("Script loaded successfully", "script", scriptPath)

	// Phase 3: Execute all commands in the queue
	commandCount := 0
	for {
		// Get next command from queue
		cmd, err := es.GetNextCommand(ctx)
		if err != nil {
			return fmt.Errorf("failed to get next command: %w", err)
		}
		if cmd == nil {
			break // No more commands
		}

		commandCount++
		logger.Debug("Executing command", "number", commandCount, "command", cmd.Name, "message", cmd.Message)

		// Interpolate command using interpolation service
		interpolatedCmd, err := is.InterpolateCommand(cmd, ctx)
		if err != nil {
			// Mark execution error for tracking
			if markErr := es.MarkExecutionError(ctx, err, cmd.String()); markErr != nil {
				logger.Error("Failed to mark execution error", "error", markErr)
			}
			return fmt.Errorf("interpolation failed for command %d: %w", commandCount, err)
		}

		// Prepare input for execution (handle bash command special case)
		cmdInput := interpolatedCmd.Message
		if interpolatedCmd.Name == "bash" && interpolatedCmd.ParseMode == parser.ParseModeRaw && interpolatedCmd.BracketContent != "" {
			cmdInput = interpolatedCmd.BracketContent
		}

		// Execute command through the global command registry
		err = commands.GlobalRegistry.Execute(interpolatedCmd.Name, interpolatedCmd.Options, cmdInput, ctx)
		if err != nil {
			// Mark execution error and return
			if markErr := es.MarkExecutionError(ctx, err, cmd.String()); markErr != nil {
				logger.Error("Failed to mark execution error", "error", markErr)
			}
			return fmt.Errorf("command execution failed for command %d (%s): %w", commandCount, interpolatedCmd.Name, err)
		}

		// Mark command as successfully executed
		if err := es.MarkCommandExecuted(ctx); err != nil {
			logger.Error("Failed to mark command as executed", "error", err)
		}

		logger.Debug("Command executed successfully", "number", commandCount, "command", interpolatedCmd.Name)
	}

	// Phase 4: Mark successful completion
	if err := es.MarkExecutionComplete(ctx); err != nil {
		logger.Error("Failed to mark execution complete", "error", err)
	}

	// Phase 5: Set success status variables for caller access
	if err := vs.Set("_status", "0", ctx); err != nil {
		logger.Error("Failed to set _status variable", "error", err)
	}

	successMessage := fmt.Sprintf("Script %s executed successfully (%d commands)", scriptPath, commandCount)
	if err := vs.Set("_output", successMessage, ctx); err != nil {
		logger.Error("Failed to set _output variable", "error", err)
	}

	logger.Info("Script execution completed successfully", "script", scriptPath, "commands_executed", commandCount)
	return nil
}
