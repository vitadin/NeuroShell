package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/types"
)

type RunCommand struct{}

func (c *RunCommand) Name() string {
	return "run"
}

func (c *RunCommand) ParseMode() types.ParseMode {
	return types.ParseModeKeyValue
}

func (c *RunCommand) Description() string {
	return "Execute a .neuro script file"
}

func (c *RunCommand) Usage() string {
	return "\\run[file=\"script.neuro\"] or \\run script.neuro"
}

func (c *RunCommand) Execute(args map[string]string, input string, ctx types.Context) error {
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

	// Get filename from args or input
	filename := ""
	if fileArg, exists := args["file"]; exists && fileArg != "" {
		filename = fileArg
	} else if input != "" {
		filename = input
	} else {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Cast services to their concrete types
	ss := scriptService.(*services.ScriptService)
	vs := variableService.(*services.VariableService)
	es := executorService.(*services.ExecutorService)

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

		// Execute the command through command registry
		cmdArgs := cmd.Options
		cmdInput := cmd.Message

		// Special handling for bash command
		if cmd.Name == "bash" && cmd.ParseMode == types.ParseModeRaw && cmd.BracketContent != "" {
			cmdInput = cmd.BracketContent
		}

		// Execute through command registry
		err = commands.GlobalRegistry.Execute(cmd.Name, cmdArgs, cmdInput, ctx)
		if err != nil {
			// Mark execution error and return
			es.MarkExecutionError(ctx, err, cmd.String())
			return fmt.Errorf("script execution failed: %w", err)
		}

		// Mark command as executed
		es.MarkCommandExecuted(ctx)
	}

	// Phase 3: Mark successful completion
	es.MarkExecutionComplete(ctx)

	// Set success status in context variables
	vs.Set("_status", "0", ctx)
	vs.Set("_output", fmt.Sprintf("Script %s executed successfully", filename), ctx)

	return nil
}

func init() {
	commands.GlobalRegistry.Register(&RunCommand{})
}