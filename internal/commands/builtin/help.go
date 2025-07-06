package builtin

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
	return "\\help[styled=true, command_name] or \\help[styled=true] command_name"
}

// HelpInfo returns structured help information for the help command.
func (c *HelpCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\help[styled=true, command_name] or \\help[styled=true] command_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "styled",
				Description: "Enable styled output with colors and formatting",
				Required:    false,
				Type:        "bool",
				Default:     "false",
			},
			{
				Name:        "command",
				Description: "Specific command to show help for",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\help",
				Description: "Show all available commands in plain text",
			},
			{
				Command:     "\\help[styled=true]",
				Description: "Show all available commands with styling",
			},
			{
				Command:     "\\help[echo]",
				Description: "Show detailed help for the echo command",
			},
			{
				Command:     "\\help[styled=true, echo]",
				Description: "Show styled detailed help for the echo command",
			},
		},
		Notes: []string{
			"Without arguments, shows all available commands",
			"With command name, shows detailed help for that specific command",
			"Use styled=true for professional formatting with colors and borders",
			"Styled output leverages the built-in rendering service themes",
		},
	}
}

// Execute displays help information for commands. If a specific command is requested via args,
// it shows detailed help for that command. Otherwise, it shows all available commands.
func (c *HelpCommand) Execute(args map[string]string, input string) error {
	// Get help service
	helpService, err := c.getHelpService()
	if err != nil {
		return fmt.Errorf("help service not available: %w", err)
	}

	// Check for styled option
	styled := false
	if styledValue, exists := args["styled"]; exists {
		styled = styledValue == "true"
		delete(args, "styled") // Remove styled from args so it doesn't interfere with command detection
	}

	// Check if a specific command was requested via bracket syntax: \help[command_name] or input: \help[styled=true] command_name
	var requestedCommand string

	// First check for command in remaining args (bracket syntax)
	if len(args) > 0 {
		// Get the first key from args (command name)
		for key := range args {
			requestedCommand = key
			break
		}
	}

	// If no command found in args, check input string (space syntax)
	if requestedCommand == "" && strings.TrimSpace(input) != "" {
		// Parse the first word from input as command name
		fields := strings.Fields(strings.TrimSpace(input))
		if len(fields) > 0 {
			requestedCommand = fields[0]
		}
	}

	// If specific command requested, show detailed help for that command
	if requestedCommand != "" {
		return c.showCommandHelpNew(requestedCommand, helpService, styled)
	}

	// Otherwise, show all commands (original behavior)
	return c.showAllCommandsNew(helpService, styled)
}

// showCommandHelpNew displays detailed help information for a specific command using HelpInfo
func (c *HelpCommand) showCommandHelpNew(commandName string, helpService *services.HelpService, styled bool) error {
	// Get the command directly from the help service
	_, err := helpService.GetCommand(commandName)
	if err != nil {
		return fmt.Errorf("command '%s' not found. Use \\help to see all available commands", commandName)
	}

	// Get the actual command to access its HelpInfo method
	command, exists := commands.GlobalRegistry.Get(commandName)
	if !exists {
		// Fallback to simple output if we can't get the command instance
		return fmt.Errorf("command '%s' not found in registry", commandName)
	}

	// Get structured help info
	helpInfo := command.HelpInfo()

	// Get render service for styling
	if styled {
		renderService, err := services.GetGlobalRenderService()
		if err != nil {
			return fmt.Errorf("render service not available: %w", err)
		}

		styledOutput, err := renderService.RenderHelp(helpInfo, true)
		if err != nil {
			return fmt.Errorf("failed to render styled help: %w", err)
		}

		fmt.Print(styledOutput)
	} else {
		renderService, err := services.GetGlobalRenderService()
		if err != nil {
			return fmt.Errorf("render service not available: %w", err)
		}

		plainOutput, err := renderService.RenderHelp(helpInfo, false)
		if err != nil {
			return fmt.Errorf("failed to render plain help: %w", err)
		}

		fmt.Print(plainOutput)
	}

	return nil
}

// showAllCommandsNew displays a list of all available commands using HelpInfo
func (c *HelpCommand) showAllCommandsNew(helpService *services.HelpService, styled bool) error {
	// Get all commands from the help service
	allCommands, err := helpService.GetAllCommands()
	if err != nil {
		return fmt.Errorf("failed to get command list: %w", err)
	}

	if styled {
		return c.showAllCommandsStyled(allCommands, helpService)
	}

	// Plain text output (existing behavior)
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

// showAllCommandsStyled displays all commands with professional styling
func (c *HelpCommand) showAllCommandsStyled(allCommands []services.CommandInfo, _ *services.HelpService) error {
	renderService, err := services.GetGlobalRenderService()
	if err != nil {
		return fmt.Errorf("render service not available: %w", err)
	}

	theme, exists := renderService.GetTheme("default")
	if !exists {
		return fmt.Errorf("default theme not found")
	}

	// Title with border
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Success.GetForeground()).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Info.GetForeground()).
		Width(60).
		Align(lipgloss.Center)

	fmt.Println(titleStyle.Render("Neuro Shell Commands"))
	fmt.Println()

	// Commands in styled format
	for _, cmdInfo := range allCommands {
		cmdStyle := theme.Command.Bold(true)
		descStyle := theme.Info.Bold(false)

		fmt.Printf("  %s - %s\n",
			cmdStyle.Render(fmt.Sprintf("%-20s", cmdInfo.Usage)),
			descStyle.Render(cmdInfo.Description))
	}

	fmt.Println()

	// Styled examples section
	exampleHeaderStyle := theme.Bold.Foreground(theme.Warning.GetForeground())
	fmt.Println(exampleHeaderStyle.Render("Examples:"))

	exampleStyle := theme.Variable
	examples := []string{
		"\\send Hello world",
		"\\set[name=\"John\"]",
		"\\get[name]",
		"\\bash[ls -la]",
	}

	for _, example := range examples {
		styledExample, _ := renderService.HighlightKeywords(example, []string{})
		fmt.Printf("  %s\n", exampleStyle.Render(styledExample))
	}

	fmt.Println()

	// Styled notes
	noteStyle := theme.Info.Italic(true)
	fmt.Println(noteStyle.Render("Note: Text without \\ prefix is sent to LLM automatically"))
	fmt.Println(noteStyle.Render("Use \\help[command] for detailed help on a specific command"))

	return nil
}

// getHelpService retrieves the help service from the global registry
func (c *HelpCommand) getHelpService() (*services.HelpService, error) {
	return services.GetGlobalHelpService()
}

func init() {
	if err := commands.GlobalRegistry.Register(&HelpCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register help command: %v", err))
	}
}
