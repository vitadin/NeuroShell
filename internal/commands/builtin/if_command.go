package builtin

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// IfCommand implements the \if command for conditional execution.
// It evaluates a condition and optionally executes a command based on the result.
type IfCommand struct{}

// Name returns the command name "if" for registration and lookup.
func (c *IfCommand) Name() string {
	return "if"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *IfCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the if command does.
func (c *IfCommand) Description() string {
	return "Conditionally execute commands based on boolean conditions"
}

// Usage returns the syntax and usage examples for the if command.
func (c *IfCommand) Usage() string {
	return "\\if[condition=boolean_expression] command_to_execute"
}

// HelpInfo returns structured help information for the if command.
func (c *IfCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "condition",
				Description: "Boolean expression to evaluate (true/false, variable existence, etc.)",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\if[condition=true] \\set[var=value]",
				Description: "Execute command when condition is true",
			},
			{
				Command:     "\\if[condition=${@user}] \\echo Hello ${@user}",
				Description: "Execute command if user variable exists and is truthy",
			},
			{
				Command:     "\\if[condition=${debug_mode}] \\echo Debug enabled",
				Description: "Execute command if debug_mode variable is true",
			},
			{
				Command:     "\\if[condition=${#test_mode}] \\echo Running in test mode",
				Description: "Execute command if test_mode system variable is true",
			},
		},
		Notes: []string{
			"Conditions are evaluated as boolean: true, 1, yes, on, enabled are truthy",
			"Empty strings, false, 0, no, off, disabled are falsy",
			"Variable existence is checked - non-empty variables are truthy",
			"System variables can be used in conditions",
			"The command after \\if is only executed if the condition is true",
			"Result of condition evaluation is stored in #if_result system variable",
		},
	}
}

// Execute evaluates the condition and optionally executes the command.
// The condition is evaluated as a boolean expression.
// Options:
//   - condition: boolean expression to evaluate (required)
//
// The command after \if is only executed if the condition evaluates to true.
func (c *IfCommand) Execute(args map[string]string, input string) error {
	// Get condition parameter
	condition, exists := args["condition"]
	if !exists {
		return fmt.Errorf("condition parameter is required")
	}

	// Evaluate the condition
	result := c.evaluateCondition(condition)

	// Store the result in system variable for debugging
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		_ = variableService.SetSystemVariable("#if_result", strconv.FormatBool(result))
	}

	// If condition is true, push the command to the stack for execution
	if result && strings.TrimSpace(input) != "" {
		if stackService, err := services.GetGlobalStackService(); err == nil {
			stackService.PushCommand(input)
		}
	}

	return nil
}

// evaluateCondition evaluates a boolean expression string
// Note: Variable interpolation (${var}) is handled by the state machine before this command executes,
// so the condition parameter already contains the expanded variable values.
func (c *IfCommand) evaluateCondition(condition string) bool {
	// Trim whitespace and evaluate directly
	// Variables have already been interpolated by the state machine
	return c.isTruthy(strings.TrimSpace(condition))
}

// isTruthy determines if a string represents a truthy value
func (c *IfCommand) isTruthy(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))

	// Empty string is falsy
	if value == "" {
		return false
	}

	// Common truthy values
	truthyValues := map[string]bool{
		"true":    true,
		"1":       true,
		"yes":     true,
		"on":      true,
		"enabled": true,
	}

	// Common falsy values
	falsyValues := map[string]bool{
		"false":    true,
		"0":        true,
		"no":       true,
		"off":      true,
		"disabled": true,
	}

	// Check explicit truthy/falsy values
	if truthyValues[value] {
		return true
	}
	if falsyValues[value] {
		return false
	}

	// Any non-empty string is considered truthy
	return true
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&IfCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register if command: %v", err))
	}
}
