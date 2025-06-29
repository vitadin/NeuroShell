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
	
	// Parse input - this never fails, always returns a valid command
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

