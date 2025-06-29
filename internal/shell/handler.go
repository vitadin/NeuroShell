package shell

import (
	"strings"

	"github.com/abiosoft/ishell/v2"
	"neuroshell/internal/parser"
)

func ProcessInput(c *ishell.Context) {
	if len(c.RawArgs) == 0 {
		return
	}
	
	rawInput := strings.Join(c.RawArgs, " ")
	rawInput = strings.TrimSpace(rawInput)
	
	// Parse input - this never fails, always returns a valid command
	cmd := parser.ParseInput(rawInput)
	
	// Execute the command
	executeCommand(c, cmd)
}

func executeCommand(c *ishell.Context, cmd *parser.Command) {
	switch cmd.Name {
	case "send":
		handleSendCommand(c, cmd)
	case "set":
		handleSetCommand(c, cmd)
	case "get":
		handleGetCommand(c, cmd)
	case "bash":
		handleBashCommand(c, cmd)
	case "help":
		handleHelpCommand(c, cmd)
	default:
		c.Printf("Unknown command: \\%s\n", cmd.Name)
		c.Println("Type \\help for available commands")
	}
}

func handleSendCommand(c *ishell.Context, cmd *parser.Command) {
	if cmd.Message == "" {
		c.Println("Usage: \\send message")
		return
	}
	c.Printf("Sending: %s\n", cmd.Message)
}

func handleSetCommand(c *ishell.Context, cmd *parser.Command) {
	if len(cmd.Options) == 0 && cmd.Message == "" {
		c.Println("Usage: \\set[var=value] or \\set var value")
		return
	}
	
	// Handle bracket syntax: \set[var=value]
	if len(cmd.Options) > 0 {
		for key, value := range cmd.Options {
			c.Printf("Setting %s = %s\n", key, value)
		}
		return
	}
	
	// Handle space syntax: \set var value
	if cmd.Message != "" {
		parts := strings.SplitN(cmd.Message, " ", 2)
		if len(parts) == 2 {
			c.Printf("Setting %s = %s\n", parts[0], parts[1])
		} else {
			c.Printf("Setting %s = \n", parts[0])
		}
	}
}

func handleGetCommand(c *ishell.Context, cmd *parser.Command) {
	var variable string
	
	// Handle bracket syntax: \get[var]
	if len(cmd.Options) > 0 {
		for key := range cmd.Options {
			variable = key
			break
		}
	} else if cmd.Message != "" {
		// Handle space syntax: \get var
		variable = strings.Fields(cmd.Message)[0]
	}
	
	if variable == "" {
		c.Println("Usage: \\get[var] or \\get var")
		return
	}
	
	c.Printf("Getting %s (not implemented yet)\n", variable)
}

func handleBashCommand(c *ishell.Context, cmd *parser.Command) {
	var command string
	
	if cmd.ParseMode == parser.ParseModeRaw && cmd.BracketContent != "" {
		// Raw bracket content: \bash[ls -la]
		command = cmd.BracketContent
	} else if cmd.Message != "" {
		// Message: \bash ls -la
		command = cmd.Message
	}
	
	if command == "" {
		c.Println("Usage: \\bash[command] or \\bash command")
		return
	}
	
	c.Printf("Executing: %s (not implemented yet)\n", command)
}

func handleHelpCommand(c *ishell.Context, cmd *parser.Command) {
	c.Println("Neuro Shell Commands:")
	c.Println("  \\send message          - Send message to LLM agent")
	c.Println("  \\set[var=value]        - Set a variable")
	c.Println("  \\get[var]              - Get a variable")
	c.Println("  \\bash[command]         - Execute system command")
	c.Println("  \\help                  - Show this help")
	c.Println("")
	c.Println("Examples:")
	c.Println("  \\send Hello world")
	c.Println("  \\set[name=\"John\"]")
	c.Println("  \\get[name]")
	c.Println("  \\bash[ls -la]")
	c.Println("")
	c.Println("Note: Text without \\ prefix is sent to LLM automatically")
}