package shell

import (
	"log"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"neuroshell/internal/commands"
	_ "neuroshell/internal/commands/builtin" // Import for side effects (init functions)
	"neuroshell/internal/context"
	"neuroshell/internal/parser"
	"neuroshell/internal/services"
)

func ProcessInput(c *ishell.Context) {
	if len(c.RawArgs) == 0 {
		return
	}
	
	rawInput := strings.Join(c.RawArgs, " ")
	rawInput = strings.TrimSpace(rawInput)
	
	// Debug: print what we received
	// c.Printf("DEBUG: RawArgs=%v, rawInput='%s'\n", c.RawArgs, rawInput)
	
	// Interpolate variables in the raw input before parsing
	interpolatedInput, err := interpolateVariables(rawInput)
	if err != nil {
		c.Printf("Variable interpolation error: %s\n", err.Error())
		return
	}
	
	// Parse the interpolated input
	cmd := parser.ParseInput(interpolatedInput)
	
	// Execute the command
	executeCommand(c, cmd)
}

// Global context instance to persist across commands
var globalCtx = context.New()

// interpolateVariables performs variable interpolation on raw input using VariableService
func interpolateVariables(input string) (string, error) {
	// Get variable service from global registry
	variableService, err := services.GlobalRegistry.GetService("variable")
	if err != nil {
		// If variable service is not available, return input unchanged
		// This allows the system to work even if services fail to initialize
		return input, nil
	}

	// Cast to VariableService and interpolate
	vs := variableService.(*services.VariableService)
	interpolated, err := vs.InterpolateString(input, globalCtx)
	if err != nil {
		return input, err
	}

	return interpolated, nil
}

func InitializeServices() error {
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

	// Initialize all services with the global context
	if err := services.GlobalRegistry.InitializeAll(globalCtx); err != nil {
		return err
	}

	log.Println("All services initialized successfully")
	return nil
}

func executeCommand(c *ishell.Context, cmd *parser.Command) {
	// Prepare args and input for the new interface
	args := cmd.Options
	input := cmd.Message
	
	// Special handling for bash command to pass raw bracket content
	if cmd.Name == "bash" && cmd.ParseMode == parser.ParseModeRaw && cmd.BracketContent != "" {
		input = cmd.BracketContent
	}
	
	// Execute command through registry
	err := commands.GlobalRegistry.Execute(cmd.Name, args, input, globalCtx)
	if err != nil {
		c.Printf("Error: %s\\n", err.Error())
		if cmd.Name != "help" {
			c.Println("Type \\help for available commands")
		}
	}
}

