package builtin

import (
	"fmt"
	"regexp"
	"strings"

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

	// Get theme object once
	themeObj := c.getThemeObject()

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
		return c.showCommandHelpNew(requestedCommand, helpService, themeObj)
	}

	// Otherwise, show all commands
	return c.showAllCommandsNew(helpService, themeObj)
}

// showCommandHelpNew displays detailed help information for a specific command using HelpInfo
func (c *HelpCommand) showCommandHelpNew(commandName string, helpService *services.HelpService, themeObj *services.Theme) error {
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

	// Render help info using theme object
	output := c.renderHelpInfo(helpInfo, themeObj)
	fmt.Print(output)

	return nil
}

// showAllCommandsNew displays a list of all available commands using theme object
func (c *HelpCommand) showAllCommandsNew(helpService *services.HelpService, themeObj *services.Theme) error {
	// Get all commands from the help service
	allCommands, err := helpService.GetAllCommands()
	if err != nil {
		return fmt.Errorf("failed to get command list: %w", err)
	}

	// Always use styled rendering - theme object handles whether to apply styling or not
	return c.showAllCommandsStyled(allCommands, themeObj)
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
		"check": true, "editor": true, "render": true, "version": true, "change-log-show": true,
	}

	modelCommands := map[string]bool{
		"model-catalog": true, "model-new": true, "model-status": true,
	}

	sessionCommands := map[string]bool{
		"session-activate": true, "session-add-assistantmsg": true, "session-add-usermsg": true,
		"session-copy": true, "session-delete": true, "session-edit-msg": true, "session-edit-with-editor": true,
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
func (c *HelpCommand) displaySessionCommandsGrouped(sessionCommands []*neurotypes.HelpInfo, themeObj *services.Theme) {
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
		case "session-add-usermsg", "session-add-assistantmsg", "session-edit-msg":
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
			fmt.Printf("    %s:\n", themeObj.Info.Render(groupName))
			for _, cmdInfo := range commands {
				fmt.Printf("      %s - %s\n",
					themeObj.Command.Render(fmt.Sprintf("%-18s", "\\"+cmdInfo.Command)),
					cmdInfo.Description)
			}
			fmt.Println()
		}
	}
}

// showAllCommandsStyled displays all commands using only theme object semantic styles
func (c *HelpCommand) showAllCommandsStyled(allCommands []*neurotypes.HelpInfo, themeObj *services.Theme) error {

	// Title
	fmt.Println(themeObj.Success.Render("Neuro Shell - Available Commands"))
	fmt.Println()

	// Categorize commands
	categories := c.categorizeCommands(allCommands)

	// Display each category
	for i, category := range categories {
		if i > 0 {
			fmt.Println()
		}

		fmt.Println(themeObj.Warning.Render(category.Name + ":"))

		// Special handling for Session Management to show subcategories
		if category.Name == "Session Management" {
			c.displaySessionCommandsGrouped(category.Commands, themeObj)
		} else {
			for _, cmdInfo := range category.Commands {
				fmt.Printf("  %s - %s\n",
					themeObj.Command.Render(fmt.Sprintf("%-20s", "\\"+cmdInfo.Command)),
					themeObj.Info.Render(cmdInfo.Description))
			}
		}
	}

	fmt.Println()

	// Notes
	fmt.Println(themeObj.Info.Render("Note: Text without \\ prefix is sent to LLM automatically"))
	fmt.Println(themeObj.Info.Render("Use \\help[command] for detailed help on any command"))

	return nil
}

// renderHelpInfo renders help information using only theme object semantic styles
func (c *HelpCommand) renderHelpInfo(helpInfo neurotypes.HelpInfo, theme *services.Theme) string {
	var result strings.Builder

	// Title
	title := fmt.Sprintf("Command: %s", helpInfo.Command)
	result.WriteString(theme.Command.Render(title))
	result.WriteString("\n\n")

	// Description
	result.WriteString(theme.Info.Render("Description: "))
	result.WriteString(helpInfo.Description)
	result.WriteString("\n\n")

	// Usage with syntax highlighting
	result.WriteString(theme.Success.Render("Usage: "))
	styledUsage := c.highlightNeuroShellSyntax(helpInfo.Usage, theme)
	result.WriteString(styledUsage)
	result.WriteString("\n\n")

	// Parse Mode
	result.WriteString(theme.Info.Render("Parse Mode: "))
	result.WriteString(c.parseModeToString(helpInfo.ParseMode))
	result.WriteString("\n")

	// Options section
	if len(helpInfo.Options) > 0 {
		result.WriteString("\n")
		result.WriteString(theme.Warning.Render("Options:"))
		result.WriteString("\n")

		for _, option := range helpInfo.Options {
			result.WriteString("  ")
			result.WriteString(theme.Variable.Render(option.Name))
			result.WriteString(" - ")
			result.WriteString(option.Description)

			if option.Default != "" {
				result.WriteString(theme.Info.Render(fmt.Sprintf(" (default: %s)", option.Default)))
			}
			if option.Required {
				result.WriteString(theme.Error.Render(" (required)"))
			}
			result.WriteString("\n")
		}
	}

	// Examples section
	if len(helpInfo.Examples) > 0 {
		result.WriteString("\n")
		result.WriteString(theme.Success.Render("Examples:"))
		result.WriteString("\n")

		for _, example := range helpInfo.Examples {
			result.WriteString("  ")
			styledExample := c.highlightNeuroShellSyntax(example.Command, theme)
			result.WriteString(styledExample)
			result.WriteString("\n")
			if example.Description != "" {
				result.WriteString("    ")
				result.WriteString(theme.Info.Render("%% " + example.Description))
				result.WriteString("\n")
			}
		}
	}

	// Stored Variables section
	if len(helpInfo.StoredVariables) > 0 {
		result.WriteString("\n")
		result.WriteString(theme.Warning.Render("Stored Variables:"))
		result.WriteString("\n")

		for _, storedVar := range helpInfo.StoredVariables {
			result.WriteString("  ")
			result.WriteString(theme.Variable.Render("${" + storedVar.Name + "}"))
			result.WriteString(" - ")
			result.WriteString(storedVar.Description)

			if storedVar.Type != "" {
				result.WriteString(" ")
				result.WriteString(theme.Info.Render("(" + storedVar.Type + ")"))
			}

			if storedVar.Example != "" {
				result.WriteString("\n    ")
				result.WriteString(theme.Info.Render("Example: " + storedVar.Example))
			}
			result.WriteString("\n")
		}
	}

	// Notes section
	if len(helpInfo.Notes) > 0 {
		result.WriteString("\n")
		result.WriteString(theme.Warning.Render("Notes:"))
		result.WriteString("\n")

		for _, note := range helpInfo.Notes {
			result.WriteString("  ")
			result.WriteString(theme.Info.Render(note))
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
// This method handles command-specific syntax highlighting using semantic theme styles.
func (c *HelpCommand) highlightNeuroShellSyntax(text string, theme *services.Theme) string {
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

// getThemeObject retrieves the theme object based on the _style variable
func (c *HelpCommand) getThemeObject() *services.Theme {
	// Get _style variable for theme selection
	styleValue := ""
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		if value, err := variableService.Get("_style"); err == nil {
			styleValue = value
		}
	}

	// Get theme service and theme object (always returns valid theme)
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		// This should rarely happen, but we need to return something
		panic(fmt.Sprintf("theme service not available: %v", err))
	}

	return themeService.GetThemeByName(styleValue)
}

// getHelpService retrieves the help service from the global registry
func (c *HelpCommand) getHelpService() (*services.HelpService, error) {
	return services.GetGlobalHelpService()
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&HelpCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register help command: %v", err))
	}
}
