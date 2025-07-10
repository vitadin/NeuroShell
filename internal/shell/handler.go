// Package shell provides the interactive shell interface and input processing for NeuroShell.
// It integrates the command system with the ishell interactive environment and handles user input routing.
package shell

import (
	"fmt"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"neuroshell/internal/commands"
	_ "neuroshell/internal/commands/assert"  // Import assert commands (init functions)
	_ "neuroshell/internal/commands/builtin" // Import for side effects (init functions)
	_ "neuroshell/internal/commands/model"   // Import model commands (init functions)
	_ "neuroshell/internal/commands/render"  // Import render commands (init functions)
	_ "neuroshell/internal/commands/session" // Import session commands (init functions)
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
)

// ProcessInput handles user input from the interactive shell and executes commands.
func ProcessInput(c *ishell.Context) {
	if len(c.RawArgs) == 0 {
		return
	}

	rawInput := strings.Join(c.RawArgs, " ")
	rawInput = strings.TrimSpace(rawInput)

	// Skip comment lines (same logic as script service)
	if strings.HasPrefix(rawInput, "%%") {
		logger.Debug("Skipping comment line", "input", rawInput)
		return
	}

	logger.Debug("Processing user input", "input", rawInput)

	// Parse input - interpolation will happen later at execution time
	cmd := parser.ParseInput(rawInput)
	logger.Debug("Parsed command", "name", cmd.Name, "message", cmd.Message, "options", cmd.Options)

	// Execute the command
	executeCommand(c, cmd)
}

// GetGlobalContext returns the global context instance for external access.
func GetGlobalContext() *context.NeuroContext {
	return context.GetGlobalContext().(*context.NeuroContext)
}

// InitializeServices sets up all required services for the NeuroShell environment.
func InitializeServices(testMode bool) error {
	// Get the global context singleton
	globalCtx := context.GetGlobalContext().(*context.NeuroContext)

	// Set test mode on global context
	globalCtx.SetTestMode(testMode)

	// Register all pure services
	if err := services.GetGlobalRegistry().RegisterService(services.NewScriptService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewVariableService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewExecutorService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewInterpolationService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewBashService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewHelpService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewEditorService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewChatSessionService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewModelService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewModelCatalogService()); err != nil {
		return err
	}

	// Register ClientFactory service
	if err := services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService()); err != nil {
		return err
	}

	// Use mock LLM service in test mode, new LLM service in production
	if testMode {
		if err := services.GetGlobalRegistry().RegisterService(services.NewMockLLMService()); err != nil {
			return err
		}
	} else {
		if err := services.GetGlobalRegistry().RegisterService(services.NewLLMService()); err != nil {
			return err
		}
	}

	// Register ThemeService if not already registered (needed for tests that clear the registry)
	if !services.GetGlobalRegistry().HasService("theme") {
		if err := services.GetGlobalRegistry().RegisterService(services.NewThemeService()); err != nil {
			return err
		}
	}

	// Register AutoCompleteService if not already registered
	if !services.GetGlobalRegistry().HasService("autocomplete") {
		if err := services.GetGlobalRegistry().RegisterService(services.NewAutoCompleteService()); err != nil {
			return err
		}
	}

	// Register MarkdownService if not already registered
	if !services.GetGlobalRegistry().HasService("markdown") {
		if err := services.GetGlobalRegistry().RegisterService(services.NewMarkdownService()); err != nil {
			return err
		}
	}

	// Initialize enhanced command service for priority-based command resolution
	// This is now managed within the commands package
	enhancedCommandService := commands.NewEnhancedCommandService()
	if err := enhancedCommandService.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize enhanced command service: %w", err)
	}

	// Set the global enhanced command service
	commands.SetGlobalEnhancedCommandService(enhancedCommandService)

	// Initialize all services
	if err := services.GetGlobalRegistry().InitializeAll(); err != nil {
		return err
	}

	logger.Info("All services initialized successfully")
	return nil
}

func executeCommand(c *ishell.Context, cmd *parser.Command) {
	logger.CommandExecution(cmd.Name, cmd.Options)

	// Get the global context singleton
	globalCtx := context.GetGlobalContext().(*context.NeuroContext)

	// Set global context for service access
	context.SetGlobalContext(globalCtx)

	// Get interpolation service
	interpolationService, err := services.GetGlobalRegistry().GetService("interpolation")
	if err != nil {
		logger.Error("Interpolation service not available", "error", err)
		c.Printf("Error: interpolation service not available: %s\n", err.Error())
		return
	}

	is := interpolationService.(*services.InterpolationService)

	// Interpolate command components using service
	interpolatedCmd, err := is.InterpolateCommand(cmd)
	if err != nil {
		logger.Error("Command interpolation failed", "command", cmd.Name, "error", err)
		c.Printf("Error: interpolation failed: %s\n", err.Error())
		return
	}

	logger.InterpolationStep(cmd.Message, interpolatedCmd.Message)

	// Prepare input for execution
	input := interpolatedCmd.Message

	// Execute command with enhanced resolution (with fallback to builtin registry)
	enhancedService := commands.GetGlobalEnhancedCommandService()
	if enhancedService == nil {
		// Fallback to builtin registry if enhanced service is not available
		logger.Debug("Enhanced command service not available, using builtin registry")
		err = commands.GetGlobalRegistry().Execute(interpolatedCmd.Name, interpolatedCmd.Options, input)
	} else {
		// Use enhanced command resolution
		err = enhancedService.Execute(interpolatedCmd.Name, interpolatedCmd.Options, input)
	}

	if err != nil {
		logger.Error("Command execution failed", "command", interpolatedCmd.Name, "error", err)
		c.Printf("Error: %s\n", err.Error())
		if cmd.Name != "help" {
			c.Println("Type \\help for available commands")
		}
	} else {
		logger.Debug("Command executed successfully", "command", interpolatedCmd.Name)
	}
}
