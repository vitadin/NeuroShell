package builtin

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// HelpCommand implements the \help command for displaying available commands and usage information.
// It lists all registered commands with their descriptions and provides examples.
type HelpCommand struct{}

// Name returns the command name "help" for registration and lookup.
func (c *HelpCommand) Name() string {
	return "help"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *HelpCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the help command does.
func (c *HelpCommand) Description() string {
	return "Show command help"
}

// Usage returns the syntax and usage examples for the help command.
func (c *HelpCommand) Usage() string {
	return "\\help [command]"
}

// Execute displays help information for commands. If a specific command is requested via args,
// it shows detailed help for that command. Otherwise, it shows all available commands.
func (c *HelpCommand) Execute(args map[string]string, _ string, _ neurotypes.Context) error {
	// Get help service
	helpService, err := c.getHelpService()
	if err != nil {
		return fmt.Errorf("help service not available: %w", err)
	}

	// Check if a specific command was requested via bracket syntax: \help[command_name]
	var requestedCommand string
	if len(args) > 0 {
		// Get the first key from args (command name)
		for key := range args {
			requestedCommand = key
			break
		}
	}

	// If specific command requested, show detailed help for that command
	if requestedCommand != "" {
		return c.showCommandHelp(requestedCommand, helpService)
	}

	// Otherwise, show all commands (original behavior)
	return c.showAllCommands(helpService)
}

// showCommandHelp displays detailed help information for a specific command
func (c *HelpCommand) showCommandHelp(commandName string, helpService *services.HelpService) error {
	cmdInfo, err := helpService.GetCommand(commandName)
	if err != nil {
		return fmt.Errorf("command '%s' not found. Use \\help to see all available commands", commandName)
	}

	fmt.Printf("Command: %s\n", cmdInfo.Name)
	fmt.Printf("Description: %s\n", cmdInfo.Description)
	fmt.Printf("Usage: %s\n", cmdInfo.Usage)
	fmt.Printf("Parse Mode: %s\n", c.parseModeToString(cmdInfo.ParseMode))

	// Show examples specific to this command
	c.showCommandExamples(cmdInfo.Name)

	return nil
}

// showAllCommands displays a list of all available commands (original behavior)
func (c *HelpCommand) showAllCommands(helpService *services.HelpService) error {
	// Get all commands from the help service
	allCommands, err := helpService.GetAllCommands()
	if err != nil {
		return fmt.Errorf("failed to get command list: %w", err)
	}

	fmt.Println("Neuro Shell Commands:")
	for _, cmdInfo := range allCommands {
		fmt.Printf("  %-20s - %s\n", cmdInfo.Usage, cmdInfo.Description)
	}

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  \\send Hello world")
	fmt.Println("  \\set[name=\"John\"]")
	fmt.Println("  \\get[name]")
	fmt.Println("  \\bash[ls -la]")
	fmt.Println()
	fmt.Println("Note: Text without \\ prefix is sent to LLM automatically")
	fmt.Println("Use \\help[command] for detailed help on a specific command")

	return nil
}

// getHelpService retrieves the help service from the global registry
func (c *HelpCommand) getHelpService() (*services.HelpService, error) {
	service, err := services.GetGlobalRegistry().GetService("help")
	if err != nil {
		return nil, err
	}

	helpService, ok := service.(*services.HelpService)
	if !ok {
		return nil, fmt.Errorf("help service has incorrect type")
	}

	return helpService, nil
}

// parseModeToString converts parse mode enum to readable string
func (c *HelpCommand) parseModeToString(mode neurotypes.ParseMode) string {
	switch mode {
	case neurotypes.ParseModeKeyValue:
		return "Key-Value (supports [key=value] syntax)"
	case neurotypes.ParseModeRaw:
		return "Raw (passes input directly without parsing)"
	case neurotypes.ParseModeWithOptions:
		return "With Options (supports additional options)"
	default:
		return "Unknown"
	}
}

// showCommandExamples shows command-specific examples
func (c *HelpCommand) showCommandExamples(commandName string) {
	fmt.Println("\nExamples:")

	switch commandName {
	case "send":
		fmt.Println("  \\send Hello, how are you?")
		fmt.Println("  \\send Can you help me with this code?")
	case "set":
		fmt.Println("  \\set[name=\"John\"]")
		fmt.Println("  \\set[count=42]")
		fmt.Println("  \\set name John Doe")
	case "get":
		fmt.Println("  \\get[name]")
		fmt.Println("  \\get name")
		fmt.Println("  \\get @user")
	case "bash":
		fmt.Println("  \\bash ls -la")
		fmt.Println("  \\bash echo \"Hello World\"")
		fmt.Println("  \\bash pwd")
	case "help":
		fmt.Println("  \\help")
		fmt.Println("  \\help[bash]")
		fmt.Println("  \\help[send]")
	case "run":
		fmt.Println("  \\run[file=\"script.neuro\"]")
		fmt.Println("  \\run script.neuro")
	case "exit":
		fmt.Println("  \\exit")
		fmt.Println("  \\exit[code=0]")
	default:
		fmt.Printf("  \\%s\n", commandName)
		fmt.Printf("  \\%s[option=value]\n", commandName)
	}
}

func init() {
	if err := commands.GlobalRegistry.Register(&HelpCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register help command: %v", err))
	}
}
