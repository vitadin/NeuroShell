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
	
	// Parse input - interpolation will happen later at execution time
	cmd := parser.ParseInput(rawInput)
	
	// Execute the command
	executeCommand(c, cmd)
}

// Global context instance to persist across commands
var globalCtx = context.New()



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
	
	if err := services.GlobalRegistry.RegisterService(services.NewInterpolationService()); err != nil {
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
	// Get interpolation service
	interpolationService, err := services.GlobalRegistry.GetService("interpolation")
	if err != nil {
		c.Printf("Error: interpolation service not available: %s\n", err.Error())
		return
	}

	is := interpolationService.(*services.InterpolationService)

	// Interpolate command components using service
	interpolatedCmd, err := is.InterpolateCommand(cmd, globalCtx)
	if err != nil {
		c.Printf("Error: interpolation failed: %s\n", err.Error())
		return
	}

	// Prepare input for execution
	input := interpolatedCmd.Message
	if interpolatedCmd.Name == "bash" && interpolatedCmd.ParseMode == parser.ParseModeRaw && interpolatedCmd.BracketContent != "" {
		input = interpolatedCmd.BracketContent
	}

	// Execute command with interpolated values
	err = commands.GlobalRegistry.Execute(interpolatedCmd.Name, interpolatedCmd.Options, input, globalCtx)
	if err != nil {
		c.Printf("Error: %s\\n", err.Error())
		if cmd.Name != "help" {
			c.Println("Type \\help for available commands")
		}
	}
}

