// Package shell provides the interactive shell interface and input processing for NeuroShell.
// It integrates the command system with the ishell interactive environment and handles user input routing.
package shell

import (
	"strings"

	"github.com/abiosoft/ishell/v2"
	_ "neuroshell/internal/commands/assert"  // Import assert commands (init functions)
	_ "neuroshell/internal/commands/builtin" // Import for side effects (init functions)
	_ "neuroshell/internal/commands/model"   // Import model commands (init functions)
	_ "neuroshell/internal/commands/render"  // Import render commands (init functions)
	_ "neuroshell/internal/commands/session" // Import session commands (init functions)
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
	"neuroshell/internal/statemachine"
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
		return
	}

	// Execute the command using state machine (handles parsing, interpolation, execution)
	executeCommand(c, rawInput)
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
	// NOTE: ScriptService removed - script execution is now handled by state machine

	// Register ConfigurationService first - other services may depend on it
	if err := services.GetGlobalRegistry().RegisterService(services.NewConfigurationService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewVariableService()); err != nil {
		return err
	}

	// Register QueueService for command queuing
	if err := services.GetGlobalRegistry().RegisterService(services.NewQueueService()); err != nil {
		return err
	}

	// Register StackService for stack-based execution
	if err := services.GetGlobalRegistry().RegisterService(services.NewStackService()); err != nil {
		return err
	}

	if err := services.GetGlobalRegistry().RegisterService(services.NewExecutorService()); err != nil {
		return err
	}

	// NOTE: InterpolationService removed - interpolation is now embedded in state machine

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

	// Enhanced command resolution will be implemented later

	// Initialize all services
	if err := services.GetGlobalRegistry().InitializeAll(); err != nil {
		return err
	}

	logger.Debug("Services initialized")
	return nil
}

func executeCommand(c *ishell.Context, rawInput string) {

	// Get the global context singleton
	globalCtx := GetGlobalContext()

	// Set global context for service access
	context.SetGlobalContext(globalCtx)

	// Create state machine with default configuration
	stateMachine := statemachine.NewStateMachineWithDefaults(globalCtx)

	// Execute through state machine (handles complete pipeline)
	err := stateMachine.Execute(rawInput)

	if err != nil {
		logger.Error("Command failed", "command", rawInput, "error", err)
		c.Printf("Error: %s\n", err.Error())
		// Check if this looks like a help command to avoid infinite loops
		if !strings.Contains(strings.ToLower(rawInput), "help") {
			c.Println("Type \\help for available commands")
		}
	}
}
