package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
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
		return c.showCommandHelpNew(requestedCommand, helpService)
	}

	// Otherwise, show all commands
	return c.showAllCommandsNew(helpService)
}

// showCommandHelpNew displays detailed help information for a specific command using HelpInfo
func (c *HelpCommand) showCommandHelpNew(commandName string, helpService *services.HelpService) error {
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

	// Create output printer with optional style injection
	var styleProvider output.StyleProvider
	if themeService, err := services.GetGlobalThemeService(); err == nil {
		styleProvider = themeService
	}
	printer := output.NewPrinter(output.WithStyles(styleProvider))

	// Render help info using printer with semantic types
	c.renderHelpInfoWithPrinter(helpInfo, printer)

	return nil
}

// showAllCommandsNew displays a list of all available commands using semantic output
func (c *HelpCommand) showAllCommandsNew(helpService *services.HelpService) error {
	// Get all commands from the help service
	allCommands, err := helpService.GetAllCommands()
	if err != nil {
		return fmt.Errorf("failed to get command list: %w", err)
	}

	// Always use semantic rendering through output system
	return c.showAllCommandsStyled(allCommands)
}

// CommandCategory represents a category of commands
type CommandCategory struct {
	Name     string
	Commands []*neurotypes.HelpInfo
}

// categorizeCommands groups commands into logical categories
func (c *HelpCommand) categorizeCommands(allCommands []*neurotypes.HelpInfo) []CommandCategory {
	categories := []CommandCategory{
		{Name: "Core Commands", Commands: []*neurotypes.HelpInfo{}},
		{Name: "Session Management", Commands: []*neurotypes.HelpInfo{}},
		{Name: "Model Management", Commands: []*neurotypes.HelpInfo{}},
		{Name: "System & Tools", Commands: []*neurotypes.HelpInfo{}},
		{Name: "Testing & Debugging", Commands: []*neurotypes.HelpInfo{}},
	}

	// Define command categories
	coreCommands := map[string]bool{
		"bash": true, "echo": true, "exit": true, "get": true, "get-env": true,
		"help": true, "run": true, "send": true, "set": true, "set-env": true,
		"silent": true, "try": true, "vars": true,
	}

	systemCommands := map[string]bool{
		"check": true, "clip": true,

		"editor": true, "render": true, "version": true, "license": true, "change-log-show": true,
	}

	modelCommands := map[string]bool{
		"model-catalog": true, "model-new": true, "model-status": true,
	}

	sessionCommands := map[string]bool{
		"session-activate": true, "session-add-assistantmsg": true, "session-add-usermsg": true,
		"session-copy": true, "session-delete": true,
		"session-edit-msg": true, "session-delete-msg": true,
		"session-edit-with-editor": true,
		"session-rename":           true, "session-edit-system": true,
		"session-export": true, "session-import": true,
		"session-json-export": true, "session-json-import": true, "session-list": true,
		"session-new": true, "session-show": true,
	}

	testingCommands := map[string]bool{
		"assert-equal": true,
	}

	// Categorize commands
	for _, cmdInfo := range allCommands {
		switch {
		case coreCommands[cmdInfo.Command]:
			categories[0].Commands = append(categories[0].Commands, cmdInfo) // Core Commands
		case sessionCommands[cmdInfo.Command]:
			categories[1].Commands = append(categories[1].Commands, cmdInfo) // Session Management
		case modelCommands[cmdInfo.Command]:
			categories[2].Commands = append(categories[2].Commands, cmdInfo) // Model Management
		case systemCommands[cmdInfo.Command]:
			categories[3].Commands = append(categories[3].Commands, cmdInfo) // System & Tools
		case testingCommands[cmdInfo.Command]:
			categories[4].Commands = append(categories[4].Commands, cmdInfo) // Testing & Debugging
		default:
			// Unknown commands go to Core Commands category
			categories[0].Commands = append(categories[0].Commands, cmdInfo)
		}
	}

	// Filter out empty categories
	result := []CommandCategory{}
	for _, category := range categories {
		if len(category.Commands) > 0 {
			result = append(result, category)
		}
	}

	return result
}

// displaySessionCommandsGrouped displays session commands organized by functionality
func (c *HelpCommand) displaySessionCommandsGrouped(sessionCommands []*neurotypes.HelpInfo, printer *output.Printer) {
	// Group session commands by functionality
	sessionGroups := map[string][]*neurotypes.HelpInfo{
		"Basic Management": {},
		"Conversation":     {},
		"Import/Export":    {},
	}

	// Categorize session commands
	for _, cmdInfo := range sessionCommands {
		switch cmdInfo.Command {
		case "session-new", "session-list", "session-activate", "session-copy", "session-delete", "session-show":
			sessionGroups["Basic Management"] = append(sessionGroups["Basic Management"], cmdInfo)
		case "session-add-usermsg", "session-add-assistantmsg", "session-edit-msg", "session-delete-msg":
			sessionGroups["Conversation"] = append(sessionGroups["Conversation"], cmdInfo)
		case "session-export", "session-import", "session-json-export", "session-json-import":
			sessionGroups["Import/Export"] = append(sessionGroups["Import/Export"], cmdInfo)
		default:
			sessionGroups["Basic Management"] = append(sessionGroups["Basic Management"], cmdInfo)
		}
	}

	// Display each group
	groupOrder := []string{"Basic Management", "Conversation", "Import/Export"}
	for _, groupName := range groupOrder {
		commands := sessionGroups[groupName]
		if len(commands) > 0 {
			printer.Print("    ")
			printer.Info(groupName + ":")
			for _, cmdInfo := range commands {
				printer.Print("      ")
				printer.Code(fmt.Sprintf("\\%-18s", cmdInfo.Command))
				printer.Print(" - ")
				printer.Println(cmdInfo.Description)
			}
			printer.Println("")
		}
	}
}

// showAllCommandsStyled displays all commands using semantic output types
func (c *HelpCommand) showAllCommandsStyled(allCommands []*neurotypes.HelpInfo) error {
	// Create output printer with optional style injection
	var styleProvider output.StyleProvider
	if themeService, err := services.GetGlobalThemeService(); err == nil {
		styleProvider = themeService
	}
	printer := output.NewPrinter(output.WithStyles(styleProvider))

	// Title
	printer.Success("Neuro Shell - Available Commands")
	printer.Println("")

	// Categorize commands
	categories := c.categorizeCommands(allCommands)

	// Display each category
	for i, category := range categories {
		if i > 0 {
			printer.Println("")
		}

		printer.Warning(category.Name + ":")

		// Special handling for Session Management to show subcategories
		if category.Name == "Session Management" {
			c.displaySessionCommandsGrouped(category.Commands, printer)
		} else {
			for _, cmdInfo := range category.Commands {
				printer.Print("  ")
				printer.Code(fmt.Sprintf("\\%-20s", cmdInfo.Command))
				printer.Print(" - ")
				printer.Println(cmdInfo.Description)
			}
		}
	}

	printer.Println("")

	// Notes
	printer.Info("Note: Text without \\ prefix is sent to LLM automatically")
	printer.Info("Use \\help[command] for detailed help on any command")

	return nil
}

// renderHelpInfoWithPrinter renders help information using semantic output types
func (c *HelpCommand) renderHelpInfoWithPrinter(helpInfo neurotypes.HelpInfo, printer *output.Printer) {
	// Title - use command styling for the command name
	printer.Print("Command: ")
	printer.Code(helpInfo.Command)
	printer.Println("")
	printer.Println("")

	// Description
	printer.Print("Description: ")
	printer.Println(helpInfo.Description)
	printer.Println("")

	// Usage - highlight the usage syntax as code
	printer.Print("Usage: ")
	printer.CodeBlock(helpInfo.Usage)
	printer.Println("")

	// Parse Mode
	printer.Print("Parse Mode: ")
	printer.Info(c.parseModeToString(helpInfo.ParseMode))

	// Options section
	if len(helpInfo.Options) > 0 {
		printer.Println("")
		printer.Warning("Options:")

		for _, option := range helpInfo.Options {
			printer.Print("  ")
			printer.Code(option.Name)
			printer.Print(" - ")
			printer.Print(option.Description)
			if option.Default != "" {
				printer.Print(" ")
				printer.Info(fmt.Sprintf("(default: %s)", option.Default))
			}
			if option.Required {
				printer.Print(" ")
				printer.Error("(required)")
			}
			printer.Println("")
		}
	}

	// Examples section
	if len(helpInfo.Examples) > 0 {
		printer.Println("")
		printer.Success("Examples:")

		for _, example := range helpInfo.Examples {
			printer.Print("  ")
			printer.CodeBlock(example.Command)
			if example.Description != "" {
				printer.Comment("%% " + example.Description)
			}
		}
	}

	// Stored Variables section
	if len(helpInfo.StoredVariables) > 0 {
		printer.Println("")
		printer.Warning("Stored Variables:")

		for _, storedVar := range helpInfo.StoredVariables {
			printer.Print("  ")
			printer.Code("${" + storedVar.Name + "}")
			printer.Print(" - ")
			printer.Print(storedVar.Description)

			if storedVar.Type != "" {
				printer.Print(" ")
				printer.Info(fmt.Sprintf("(%s)", storedVar.Type))
			}
			printer.Println("")

			if storedVar.Example != "" {
				printer.Print("    Example: ")
				printer.Code(storedVar.Example)
				printer.Println("")
			}
		}
	}

	// Notes section
	if len(helpInfo.Notes) > 0 {
		printer.Println("")
		printer.Warning("Notes:")

		for _, note := range helpInfo.Notes {
			printer.Print("  ")
			printer.Info(note)
		}
	}
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

// getHelpService retrieves the help service from the global registry
func (c *HelpCommand) getHelpService() (*services.HelpService, error) {
	return services.GetGlobalHelpService()
}

// IsReadOnly returns true as the help command doesn't modify system state.
func (c *HelpCommand) IsReadOnly() bool {
	return true
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&HelpCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register help command: %v", err))
	}
}
