package builtin

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

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
	return "\\vars [pattern=regex] [type=user|system|all]"
}

// Execute lists variables with optional filtering by pattern and type.
// It retrieves all variables from the variable service and applies filters as specified.
func (c *VarsCommand) Execute(args map[string]string, _ string, ctx neurotypes.Context) error {
	// Get variable service
	variableService, err := c.getVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get all variables
	allVars, err := variableService.GetAllVariables(ctx)
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

	// Display results
	c.displayVariables(filteredVars)

	return nil
}

// getVariableService retrieves the variable service from the global registry
func (c *VarsCommand) getVariableService() (*services.VariableService, error) {
	service, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return nil, err
	}

	variableService, ok := service.(*services.VariableService)
	if !ok {
		return nil, fmt.Errorf("variable service has incorrect type")
	}

	return variableService, nil
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
func (c *VarsCommand) displayVariables(vars map[string]string) {
	if len(vars) == 0 {
		fmt.Println("No variables found matching the specified criteria.")
		return
	}

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

func init() {
	if err := commands.GlobalRegistry.Register(&VarsCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register vars command: %v", err))
	}
}
