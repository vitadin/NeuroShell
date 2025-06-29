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
	
	if strings.HasPrefix(rawInput, "\\") {
		handleNeuroCommand(c, rawInput)
	} else {
		handleSystemCommand(c, rawInput)
	}
}

func handleNeuroCommand(c *ishell.Context, input string) {
	cmd, err := parser.ParseCommand(input)
	if err != nil {
		c.Printf("Error parsing command: %v\n", err)
		return
	}
	
	switch cmd.Name {
	case "send":
		handleSendCommand(c, cmd)
	case "set":
		handleSetCommand(c, cmd)
	case "get":
		handleGetCommand(c, cmd)
	case "help":
		handleHelpCommand(c, cmd)
	default:
		c.Printf("Unknown command: \\%s\n", cmd.Name)
		c.Println("Type \\help for available commands")
	}
}

func handleSystemCommand(c *ishell.Context, input string) {
	c.Printf("System command not implemented: %s\n", input)
	c.Println("Use \\bash for system commands")
}

func handleSendCommand(c *ishell.Context, cmd *parser.Command) {
	if cmd.Message == "" {
		c.Println("Usage: \\send message")
		return
	}
	c.Printf("Sending: %s\n", cmd.Message)
}

func handleSetCommand(c *ishell.Context, cmd *parser.Command) {
	if len(cmd.Options) == 0 {
		c.Println("Usage: \\set[var=value] or \\set var value")
		return
	}
	
	for key, value := range cmd.Options {
		if value == "" && cmd.Message != "" {
			value = cmd.Message
		}
		c.Printf("Setting %s = %s\n", key, value)
	}
}

func handleGetCommand(c *ishell.Context, cmd *parser.Command) {
	if len(cmd.Options) == 0 && cmd.Message == "" {
		c.Println("Usage: \\get[var] or \\get var")
		return
	}
	
	var variable string
	if len(cmd.Options) > 0 {
		for key := range cmd.Options {
			variable = key
			break
		}
	} else {
		variable = cmd.Message
	}
	
	c.Printf("Getting %s (not implemented yet)\n", variable)
}

func handleHelpCommand(c *ishell.Context, cmd *parser.Command) {
	c.Println("Neuro Shell Commands:")
	c.Println("  \\send message          - Send message to LLM agent")
	c.Println("  \\set[var=value]        - Set a variable")
	c.Println("  \\get[var]              - Get a variable")
	c.Println("  \\help                  - Show this help")
	c.Println("")
	c.Println("Examples:")
	c.Println("  \\send Hello world")
	c.Println("  \\set[name=\"John\"]")
	c.Println("  \\get[name]")
}