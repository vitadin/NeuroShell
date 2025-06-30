package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// RunCommand implements the \run command for executing .neuro script files.
// It loads script files and executes their commands in sequence with variable interpolation.
type RunCommand struct{}

// Name returns the command name "run" for registration and lookup.
func (c *RunCommand) Name() string {
	return "run"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *RunCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the run command does.
func (c *RunCommand) Description() string {
	return "Execute a .neuro script file"
}

// Usage returns the syntax and usage examples for the run command.
func (c *RunCommand) Usage() string {
	return "\\run[file=\"script.neuro\"] or \\run script.neuro"
}

// Execute loads and runs a .neuro script file with full service orchestration.
// It handles variable interpolation, command execution, and error management.
func (c *RunCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	// Get all required services
	scriptService, err := services.GlobalRegistry.GetService("script")
	if err != nil {
		return fmt.Errorf("script service not available: %w", err)
	}

	variableService, err := services.GlobalRegistry.GetService("variable")
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	executorService, err := services.GlobalRegistry.GetService("executor")
	if err != nil {
		return fmt.Errorf("executor service not available: %w", err)
	}

	interpolationService, err := services.GlobalRegistry.GetService("interpolation")
	if err != nil {
		return fmt.Errorf("interpolation service not available: %w", err)
	}

	// Get filename from args or input
	filename := ""
	if fileArg, exists := args["file"]; exists && fileArg != "" {
		filename = fileArg
	} else if input != "" {
		filename = input
	} else {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Cast services to their concrete neurotypes
	ss := scriptService.(*services.ScriptService)
	vs := variableService.(*services.VariableService)
	es := executorService.(*services.ExecutorService)
	is := interpolationService.(*services.InterpolationService)

	// Phase 1: Load script file into execution queue
	if err := ss.LoadScript(filename, ctx); err != nil {
		return fmt.Errorf("failed to load script: %w", err)
	}

	// Phase 2: Execute all commands in the queue
	// Note: Variable interpolation now happens per-command in executeCommand
	for {
		// Get next command from queue
		cmd, err := es.GetNextCommand(ctx)
		if err != nil {
			return fmt.Errorf("failed to get next command: %w", err)
		}
		if cmd == nil {
			break // No more commands
		}

		// Interpolate command using service
		interpolatedCmd, err := is.InterpolateCommand(cmd, ctx)
		if err != nil {
			if markErr := es.MarkExecutionError(ctx, err, cmd.String()); markErr != nil {
				// Log the mark error but continue with the original error
				fmt.Printf("Warning: failed to mark execution error: %v\n", markErr)
			}
			return fmt.Errorf("interpolation failed: %w", err)
		}

		// Prepare input for execution
		cmdInput := interpolatedCmd.Message
		if interpolatedCmd.Name == "bash" && interpolatedCmd.ParseMode == neurotypes.ParseModeRaw && interpolatedCmd.BracketContent != "" {
			cmdInput = interpolatedCmd.BracketContent
		}

		// Execute command (RunCommand orchestrates execution)
		err = commands.GlobalRegistry.Execute(interpolatedCmd.Name, interpolatedCmd.Options, cmdInput, ctx)
		if err != nil {
			// Mark execution error and return
			if markErr := es.MarkExecutionError(ctx, err, cmd.String()); markErr != nil {
				// Log the mark error but continue with the original error
				fmt.Printf("Warning: failed to mark execution error: %v\n", markErr)
			}
			return fmt.Errorf("script execution failed: %w", err)
		}

		// Mark command as executed
		if err := es.MarkCommandExecuted(ctx); err != nil {
			fmt.Printf("Warning: failed to mark command as executed: %v\n", err)
		}
	}

	// Phase 3: Mark successful completion
	if err := es.MarkExecutionComplete(ctx); err != nil {
		fmt.Printf("Warning: failed to mark execution complete: %v\n", err)
	}

	// Set success status in context variables
	if err := vs.Set("_status", "0", ctx); err != nil {
		fmt.Printf("Warning: failed to set _status variable: %v\n", err)
	}
	if err := vs.Set("_output", fmt.Sprintf("Script %s executed successfully", filename), ctx); err != nil {
		fmt.Printf("Warning: failed to set _output variable: %v\n", err)
	}

	return nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&RunCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register run command: %v", err))
	}
}
