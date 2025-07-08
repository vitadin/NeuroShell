package builtin

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss/list"
	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// VarsCommand implements the \vars command for listing variables with filtering capabilities.
// It supports filtering by regex pattern and variable type (user, system, or all).
type VarsCommand struct{}

// Name returns the command name "vars" for registration and lookup.
func (c *VarsCommand) Name() string {
	return "vars"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *VarsCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the vars command does.
func (c *VarsCommand) Description() string {
	return "List variables with optional filtering"
}

// Usage returns the syntax and usage examples for the vars command.
func (c *VarsCommand) Usage() string {
	return "\\vars[pattern=regex, type=user|system|all]"
}

// HelpInfo returns structured help information for the vars command.
func (c *VarsCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "pattern",
				Description: "Regular expression pattern to filter variable names",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "type",
				Description: "Filter by variable type: user, system, or all",
				Required:    false,
				Type:        "string",
				Default:     "all",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\vars",
				Description: "List all variables in the current session",
			},
			{
				Command:     "\\vars[type=user]",
				Description: "Show only user-defined variables",
			},
			{
				Command:     "\\vars[type=system]",
				Description: "Show only system variables (@, #, _ prefixed)",
			},
			{
				Command:     "\\vars[pattern=^name]",
				Description: "Show variables starting with 'name'",
			},
			{
				Command:     "\\vars[pattern=session, type=system]",
				Description: "Show system variables containing 'session'",
			},
		},
		Notes: []string{
			"User variables: custom variables set with \\set command",
			"System variables: @ (environment), # (metadata), _ (command output)",
			"Pattern uses regular expression syntax for flexible filtering",
			"Long values are truncated with length information displayed",
		},
	}
}

// Execute lists variables with optional filtering by pattern and type.
// It retrieves all variables from the variable service and applies filters as specified.
func (c *VarsCommand) Execute(args map[string]string, _ string) error {

	// Get variable service
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get all variables
	allVars, err := variableService.GetAllVariables()
	if err != nil {
		return fmt.Errorf("failed to get variables: %w", err)
	}

	// Parse filter options
	pattern := args["pattern"]
	varType := args["type"]
	if varType == "" {
		varType = "all" // Default to showing all variables
	}

	// Apply filters
	filteredVars, err := c.applyFilters(allVars, pattern, varType)
	if err != nil {
		return fmt.Errorf("failed to apply filters: %w", err)
	}

	// Get theme object for styling
	themeObj := c.getThemeObject()

	// Display results
	c.displayVariables(filteredVars, themeObj)

	return nil
}

// applyFilters applies pattern and type filters to the variable map
func (c *VarsCommand) applyFilters(allVars map[string]string, pattern, varType string) (map[string]string, error) {
	result := make(map[string]string)

	// Compile regex pattern if provided
	var regexPattern *regexp.Regexp
	if pattern != "" {
		var err error
		regexPattern, err = regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern, err)
		}
	}

	for name, value := range allVars {
		// Apply type filter
		if !c.matchesTypeFilter(name, varType) {
			continue
		}

		// Apply pattern filter
		if regexPattern != nil && !regexPattern.MatchString(name) {
			continue
		}

		result[name] = value
	}

	return result, nil
}

// matchesTypeFilter checks if a variable name matches the specified type filter
func (c *VarsCommand) matchesTypeFilter(name, varType string) bool {
	isSystemVar := strings.HasPrefix(name, "@") || strings.HasPrefix(name, "#") || strings.HasPrefix(name, "_")

	switch varType {
	case "user":
		return !isSystemVar
	case "system":
		return isSystemVar
	case "all":
		return true
	default:
		return true // Default to showing all if unknown type
	}
}

// displayVariables formats and displays the filtered variables
func (c *VarsCommand) displayVariables(vars map[string]string, themeObj *services.Theme) {
	if len(vars) == 0 {
		fmt.Println("No variables found matching the specified criteria.")
		return
	}

	// Check if styling is enabled (not plain theme)
	if themeObj.Name != "plain" && themeObj.Name != "" {
		c.displayVariablesStyled(vars, themeObj)
		return
	}

	// Fall back to plain output
	c.displayVariablesPlain(vars)
}

// displayVariablesPlain displays variables in plain text format (original implementation)
func (c *VarsCommand) displayVariablesPlain(vars map[string]string) {
	// Separate user and system variables
	userVars := make(map[string]string)
	systemVars := make(map[string]string)

	for name, value := range vars {
		if strings.HasPrefix(name, "@") || strings.HasPrefix(name, "#") || strings.HasPrefix(name, "_") {
			systemVars[name] = value
		} else {
			userVars[name] = value
		}
	}

	// Display user variables
	if len(userVars) > 0 {
		fmt.Println("User Variables:")
		c.displayVariableGroup(userVars)
	}

	// Display system variables
	if len(systemVars) > 0 {
		if len(userVars) > 0 {
			fmt.Println() // Add spacing between groups
		}
		fmt.Println("System Variables:")
		c.displayVariableGroup(systemVars)
	}

	// Display summary
	fmt.Printf("\nTotal: %d variables\n", len(vars))
}

// formatValueForDisplay formats a variable value for concise display
// Shows full value if short, otherwise shows beginning + "..." + end + length info
func (c *VarsCommand) formatValueForDisplay(value string) string {
	const (
		maxDisplayWidth = 80
		firstPartLength = 30
		lastPartLength  = 20
	)

	// Handle newlines by replacing with \n for display
	displayValue := strings.ReplaceAll(value, "\n", "\\n")
	displayValue = strings.ReplaceAll(displayValue, "\r", "\\r")
	displayValue = strings.ReplaceAll(displayValue, "\t", "\\t")

	// If value is short enough, show it entirely
	if len(displayValue) <= maxDisplayWidth {
		return displayValue
	}

	// For long values, show: first part + "..." + last part + length info
	firstPart := displayValue[:firstPartLength]
	lastPart := displayValue[len(displayValue)-lastPartLength:]
	lengthInfo := fmt.Sprintf("(length: %d chars)", len(value))

	// Calculate space needed for "..." + lastPart + " " + lengthInfo
	ellipsisAndEnd := "..." + lastPart + " " + lengthInfo

	// If the truncated version would be longer than original, just show original
	if len(firstPart)+len(ellipsisAndEnd) >= len(displayValue) {
		return displayValue
	}

	return firstPart + "..." + lastPart + " " + lengthInfo
}

// displayVariableGroup displays a group of variables sorted by name
func (c *VarsCommand) displayVariableGroup(vars map[string]string) {
	// Sort variable names for consistent output
	names := make([]string, 0, len(vars))
	for name := range vars {
		names = append(names, name)
	}
	sort.Strings(names)

	// Display each variable
	for _, name := range names {
		value := vars[name]
		formattedValue := c.formatValueForDisplay(value)
		fmt.Printf("  %-20s = %s\n", name, formattedValue)
	}
}

// displayVariablesStyled displays variables using lipgloss list styling
func (c *VarsCommand) displayVariablesStyled(vars map[string]string, themeObj *services.Theme) {
	// Title
	fmt.Println(themeObj.Success.Render("Variables"))
	fmt.Println()

	// Separate user and system variables
	userVars := make(map[string]string)
	systemVars := make(map[string]string)

	for name, value := range vars {
		if strings.HasPrefix(name, "@") || strings.HasPrefix(name, "#") || strings.HasPrefix(name, "_") {
			systemVars[name] = value
		} else {
			userVars[name] = value
		}
	}

	// Create groups for display
	if len(userVars) > 0 {
		fmt.Println(themeObj.Warning.Render("User Variables:"))
		userList := c.createVariableListStyled(userVars, themeObj)
		fmt.Print(userList.String())
		fmt.Println()
	}

	if len(systemVars) > 0 {
		if len(userVars) > 0 {
			fmt.Println() // Add spacing between groups
		}
		fmt.Println(themeObj.Warning.Render("System Variables:"))

		// Group system variables by prefix
		envVars := make(map[string]string)    // @
		metaVars := make(map[string]string)   // #
		outputVars := make(map[string]string) // _

		for name, value := range systemVars {
			switch {
			case strings.HasPrefix(name, "@"):
				envVars[name] = value
			case strings.HasPrefix(name, "#"):
				metaVars[name] = value
			case strings.HasPrefix(name, "_"):
				outputVars[name] = value
			}
		}

		// Display environment variables
		if len(envVars) > 0 {
			fmt.Printf("  %s\n", themeObj.Info.Render("Environment (@):"))
			envList := c.createVariableListStyled(envVars, themeObj)
			fmt.Print(c.indentList(envList.String()))
		}

		// Display metadata variables
		if len(metaVars) > 0 {
			fmt.Printf("  %s\n", themeObj.Info.Render("Metadata (#):"))
			metaList := c.createVariableListStyled(metaVars, themeObj)
			fmt.Print(c.indentList(metaList.String()))
		}

		// Display output variables
		if len(outputVars) > 0 {
			fmt.Printf("  %s\n", themeObj.Info.Render("Command Outputs (_):"))
			outputList := c.createVariableListStyled(outputVars, themeObj)
			fmt.Print(c.indentList(outputList.String()))
		}
	}

	// Display summary
	fmt.Printf("\n%s\n", themeObj.Success.Render(fmt.Sprintf("Total: %d variables", len(vars))))
}

// createVariableListStyled creates a styled list for variables
func (c *VarsCommand) createVariableListStyled(vars map[string]string, themeObj *services.Theme) *list.List {
	// Sort variable names for consistent output
	names := make([]string, 0, len(vars))
	for name := range vars {
		names = append(names, name)
	}
	sort.Strings(names)

	// Create list
	l := themeObj.CreateList()
	for _, name := range names {
		value := vars[name]
		formattedValue := c.formatValueForDisplay(value)
		formattedVar := fmt.Sprintf("%s = %s",
			themeObj.Variable.Render(name),
			themeObj.Info.Render(formattedValue))
		l.Item(formattedVar)
	}
	return l
}

// indentList indents each line of the list output for nested display
func (c *VarsCommand) indentList(listOutput string) string {
	lines := strings.Split(listOutput, "\n")
	var indentedLines []string
	for _, line := range lines {
		if line != "" {
			indentedLines = append(indentedLines, "    "+line)
		} else {
			indentedLines = append(indentedLines, line)
		}
	}
	return strings.Join(indentedLines, "\n")
}

// getThemeObject retrieves the theme object based on the _style variable
func (c *VarsCommand) getThemeObject() *services.Theme {
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

func init() {
	if err := commands.GlobalRegistry.Register(&VarsCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register vars command: %v", err))
	}
}
