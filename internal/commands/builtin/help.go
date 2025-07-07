package builtin

import (
	"fmt"
	"regexp"
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
	return "\\help[command_name] or \\help command_name"
}

// HelpInfo returns structured help information for the help command.
func (c *HelpCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\help[command_name] or \\help command_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
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
				Description: "Show all available commands",
			},
			{
				Command:     "\\help[echo]",
				Description: "Show detailed help for the echo command",
			},
			{
				Command:     "\\help bash",
				Description: "Show detailed help for the bash command",
			},
		},
		Notes: []string{
			"Without arguments, shows all available commands",
			"With command name, shows detailed help for that specific command",
			"Set _style variable to 'dark1' for professional formatting with colors and borders",
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

	// Get _style variable for theme selection
	theme := ""
	variableService, err := services.GetGlobalVariableService()
	if err == nil {
		if styleValue, err := variableService.Get("_style"); err == nil {
			theme = styleValue
		}
	}

	// Check if a specific command was requested via bracket syntax: \help[command_name] or input: \help command_name
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
		return c.showCommandHelpNew(requestedCommand, helpService, theme)
	}

	// Otherwise, show all commands (original behavior)
	return c.showAllCommandsNew(helpService, theme)
}

// showCommandHelpNew displays detailed help information for a specific command using HelpInfo
func (c *HelpCommand) showCommandHelpNew(commandName string, helpService *services.HelpService, theme string) error {
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

	// Get render service to access themes
	renderService, err := services.GetGlobalRenderService()
	if err != nil {
		return fmt.Errorf("render service not available: %w", err)
	}

	// Get theme object (nil for plain text)
	themeObj, isValid := renderService.GetThemeByName(theme)
	if !isValid {
		// Invalid theme, fall back to plain text
		themeObj = nil
	}

	// Render help info using theme object
	output := c.renderHelpInfo(helpInfo, themeObj)
	fmt.Print(output)

	return nil
}

// showAllCommandsNew displays a list of all available commands using HelpInfo
func (c *HelpCommand) showAllCommandsNew(helpService *services.HelpService, theme string) error {
	// Get all commands from the help service
	allCommands, err := helpService.GetAllCommands()
	if err != nil {
		return fmt.Errorf("failed to get command list: %w", err)
	}

	// Check if theme is specified (non-empty means styled output)
	if theme != "" {
		return c.showAllCommandsStyled(allCommands, helpService, theme)
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

// showAllCommandsStyled displays all commands with professional styling using the specified theme
func (c *HelpCommand) showAllCommandsStyled(allCommands []services.CommandInfo, _ *services.HelpService, theme string) error {
	renderService, err := services.GetGlobalRenderService()
	if err != nil {
		return fmt.Errorf("render service not available: %w", err)
	}

	// Get theme object using alias support
	themeObj, isValid := renderService.GetThemeByName(theme)
	if !isValid {
		// Invalid theme, fall back to default
		themeObj, _ = renderService.GetTheme("default")
	}
	if themeObj == nil {
		// Empty theme should not reach this styled method, fallback to default
		themeObj, _ = renderService.GetTheme("default")
	}

	// Title with border
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(themeObj.Success.GetForeground()).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(themeObj.Info.GetForeground()).
		Width(60).
		Align(lipgloss.Center)

	fmt.Println(titleStyle.Render("Neuro Shell Commands"))
	fmt.Println()

	// Commands in styled format
	for _, cmdInfo := range allCommands {
		cmdStyle := themeObj.Command.Bold(true)
		descStyle := themeObj.Info.Bold(false)

		fmt.Printf("  %s - %s\n",
			cmdStyle.Render(fmt.Sprintf("%-20s", cmdInfo.Usage)),
			descStyle.Render(cmdInfo.Description))
	}

	fmt.Println()

	// Styled examples section
	exampleHeaderStyle := themeObj.Bold.Foreground(themeObj.Warning.GetForeground())
	fmt.Println(exampleHeaderStyle.Render("Examples:"))

	exampleStyle := themeObj.Variable
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
	noteStyle := themeObj.Info.Italic(true)
	fmt.Println(noteStyle.Render("Note: Text without \\ prefix is sent to LLM automatically"))
	fmt.Println(noteStyle.Render("Use \\help[command] for detailed help on a specific command"))

	return nil
}

// renderHelpInfo renders help information using a theme object (nil for plain text)
func (c *HelpCommand) renderHelpInfo(helpInfo neurotypes.HelpInfo, theme *services.RenderTheme) string {
	if theme == nil {
		return c.renderHelpInfoPlain(helpInfo)
	}
	return c.renderHelpInfoStyled(helpInfo, theme)
}

// renderHelpInfoPlain renders help information as plain text
func (c *HelpCommand) renderHelpInfoPlain(helpInfo neurotypes.HelpInfo) string {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Command: %s\n", helpInfo.Command))
	result.WriteString(fmt.Sprintf("Description: %s\n", helpInfo.Description))
	result.WriteString(fmt.Sprintf("Usage: %s\n", helpInfo.Usage))
	result.WriteString(fmt.Sprintf("Parse Mode: %s\n", c.parseModeToString(helpInfo.ParseMode)))

	if len(helpInfo.Options) > 0 {
		result.WriteString("\nOptions:\n")
		for _, option := range helpInfo.Options {
			defaultStr := ""
			if option.Default != "" {
				defaultStr = fmt.Sprintf(" (default: %s)", option.Default)
			}
			requiredStr := ""
			if option.Required {
				requiredStr = " (required)"
			}
			result.WriteString(fmt.Sprintf("  %s - %s%s%s\n", option.Name, option.Description, defaultStr, requiredStr))
		}
	}

	if len(helpInfo.Examples) > 0 {
		result.WriteString("\nExamples:\n")
		for _, example := range helpInfo.Examples {
			result.WriteString(fmt.Sprintf("  %s\n", example.Command))
			if example.Description != "" {
				result.WriteString("    %% " + example.Description + "\n")
			}
		}
	}

	if len(helpInfo.Notes) > 0 {
		result.WriteString("\nNotes:\n")
		for _, note := range helpInfo.Notes {
			result.WriteString(fmt.Sprintf("  %s\n", note))
		}
	}

	return result.String()
}

// renderHelpInfoStyled renders help information with professional styling using the theme
func (c *HelpCommand) renderHelpInfoStyled(helpInfo neurotypes.HelpInfo, theme *services.RenderTheme) string {
	var result strings.Builder

	// Title with border
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Command.GetForeground()).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Info.GetForeground())

	title := fmt.Sprintf("Command: %s", helpInfo.Command)
	result.WriteString(titleStyle.Render(title))
	result.WriteString("\n\n")

	// Description
	descStyle := theme.Info.Bold(false)
	result.WriteString(descStyle.Render("Description: "))
	result.WriteString(helpInfo.Description)
	result.WriteString("\n\n")

	// Usage with syntax highlighting
	usageStyle := theme.Bold.Foreground(theme.Success.GetForeground())
	result.WriteString(usageStyle.Render("Usage: "))
	styledUsage := c.highlightNeuroShellSyntax(helpInfo.Usage, theme)
	result.WriteString(styledUsage)
	result.WriteString("\n\n")

	// Parse Mode
	parseModeStyle := theme.Info.Bold(false)
	result.WriteString(parseModeStyle.Render("Parse Mode: "))
	result.WriteString(c.parseModeToString(helpInfo.ParseMode))
	result.WriteString("\n")

	// Options section
	if len(helpInfo.Options) > 0 {
		result.WriteString("\n")
		optionHeaderStyle := theme.Bold.Foreground(theme.Warning.GetForeground())
		result.WriteString(optionHeaderStyle.Render("Options:"))
		result.WriteString("\n")

		for _, option := range helpInfo.Options {
			optionNameStyle := theme.Variable.Bold(true)
			result.WriteString("  ")
			result.WriteString(optionNameStyle.Render(option.Name))
			result.WriteString(" - ")
			result.WriteString(option.Description)

			if option.Default != "" {
				defaultStyle := theme.Info.Italic(true)
				result.WriteString(defaultStyle.Render(fmt.Sprintf(" (default: %s)", option.Default)))
			}
			if option.Required {
				requiredStyle := theme.Error.Bold(true)
				result.WriteString(requiredStyle.Render(" (required)"))
			}
			result.WriteString("\n")
		}
	}

	// Examples section
	if len(helpInfo.Examples) > 0 {
		result.WriteString("\n")
		exampleHeaderStyle := theme.Bold.Foreground(theme.Success.GetForeground())
		result.WriteString(exampleHeaderStyle.Render("Examples:"))
		result.WriteString("\n")

		for _, example := range helpInfo.Examples {
			result.WriteString("  ")
			styledExample := c.highlightNeuroShellSyntax(example.Command, theme)
			result.WriteString(styledExample)
			result.WriteString("\n")
			if example.Description != "" {
				commentStyle := theme.Info.Italic(true)
				result.WriteString("    ")
				result.WriteString(commentStyle.Render("%% " + example.Description))
				result.WriteString("\n")
			}
		}
	}

	// Notes section
	if len(helpInfo.Notes) > 0 {
		result.WriteString("\n")
		noteHeaderStyle := theme.Bold.Foreground(theme.Warning.GetForeground())
		result.WriteString(noteHeaderStyle.Render("Notes:"))
		result.WriteString("\n")

		for _, note := range helpInfo.Notes {
			noteStyle := theme.Info.Italic(true)
			result.WriteString("  ")
			result.WriteString(noteStyle.Render(note))
			result.WriteString("\n")
		}
	}

	return result.String()
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

// highlightNeuroShellSyntax applies syntax highlighting for NeuroShell-specific patterns
func (c *HelpCommand) highlightNeuroShellSyntax(text string, theme *services.RenderTheme) string {
	result := text

	// Highlight variables: ${variable_name}
	variablePattern := regexp.MustCompile(`\$\{[^}]+\}`)
	result = variablePattern.ReplaceAllStringFunc(result, func(match string) string {
		return theme.Variable.Render(match)
	})

	// Highlight commands: \command (but only if not already styled)
	commandPattern := regexp.MustCompile(`\\[a-zA-Z_][a-zA-Z0-9_-]*`)
	result = commandPattern.ReplaceAllStringFunc(result, func(match string) string {
		// Check if this text is already styled (contains ANSI sequences)
		if strings.Contains(match, "\x1b[") {
			return match // Already styled, don't re-style
		}
		return theme.Command.Render(match)
	})

	return result
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
