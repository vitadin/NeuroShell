// Package shell provides the interactive shell interface and input processing for NeuroShell.
// It integrates the command system with the ishell interactive environment and handles user input routing.
package shell

import (
	"strings"

	"github.com/abiosoft/ishell/v2"
	"neuroshell/internal/commands"
	_ "neuroshell/internal/commands/builtin" // Import for side effects (init functions)
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ProcessInput handles user input from the interactive shell and executes commands.
func ProcessInput(c *ishell.Context) {
	if len(c.RawArgs) == 0 {
		return
	}

	rawInput := strings.Join(c.RawArgs, " ")
	rawInput = strings.TrimSpace(rawInput)

	logger.Debug("Processing user input", "input", rawInput)

	// Parse input - interpolation will happen later at execution time
	cmd := parser.ParseInput(rawInput)
	logger.Debug("Parsed command", "name", cmd.Name, "message", cmd.Message, "options", cmd.Options)

	// Execute the command
	executeCommand(c, cmd)
}

// Global context instance to persist across commands
var globalCtx = context.New()

// InitializeServices sets up all required services for the NeuroShell environment.
func InitializeServices(testMode bool) error {
	// Set test mode on global context
	globalCtx.SetTestMode(testMode)

	// Register all pure services
	if err := services.GlobalRegistry.RegisterService(services.NewScriptService()); err != nil {
		return err
	}

	if err := services.GlobalRegistry.RegisterService(services.NewVariableService()); err != nil {
		return err
	}

	if err := services.GlobalRegistry.RegisterService(services.NewExecutorService()); err != nil {
		return err
	}

	if err := services.GlobalRegistry.RegisterService(services.NewInterpolationService()); err != nil {
		return err
	}

	// Initialize all services with the global context
	if err := services.GlobalRegistry.InitializeAll(globalCtx); err != nil {
		return err
	}

	logger.Info("All services initialized successfully")
	return nil
}

func executeCommand(c *ishell.Context, cmd *parser.Command) {
	logger.CommandExecution(cmd.Name, cmd.Options)

	// Get interpolation service
	interpolationService, err := services.GlobalRegistry.GetService("interpolation")
	if err != nil {
		logger.Error("Interpolation service not available", "error", err)
		c.Printf("Error: interpolation service not available: %s\n", err.Error())
		return
	}

	is := interpolationService.(*services.InterpolationService)

	// Interpolate command components using service
	interpolatedCmd, err := is.InterpolateCommand(cmd, globalCtx)
	if err != nil {
		logger.Error("Command interpolation failed", "command", cmd.Name, "error", err)
		c.Printf("Error: interpolation failed: %s\n", err.Error())
		return
	}

	logger.InterpolationStep(cmd.Message, interpolatedCmd.Message)

	// Prepare input for execution
	input := interpolatedCmd.Message
	if interpolatedCmd.Name == "bash" && interpolatedCmd.ParseMode == neurotypes.ParseModeRaw && interpolatedCmd.BracketContent != "" {
		input = interpolatedCmd.BracketContent
	}

	// Execute command with interpolated values
	err = commands.GlobalRegistry.Execute(interpolatedCmd.Name, interpolatedCmd.Options, input, globalCtx)
	if err != nil {
		logger.Error("Command execution failed", "command", interpolatedCmd.Name, "error", err)
		c.Printf("Error: %s\\n", err.Error())
		if cmd.Name != "help" {
			c.Println("Type \\help for available commands")
		}
	} else {
		logger.Debug("Command executed successfully", "command", interpolatedCmd.Name)
	}
}
