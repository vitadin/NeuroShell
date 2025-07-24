package builtin

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
	"neuroshell/pkg/stringprocessing"
)

// IfNotCommand implements the \if-not command for inverted conditional execution.
// It evaluates a condition and executes a command when the condition is falsy.
type IfNotCommand struct{}

// Name returns the command name "if-not" for registration and lookup.
func (c *IfNotCommand) Name() string {
	return "if-not"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *IfNotCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the if-not command does.
func (c *IfNotCommand) Description() string {
	return "Conditionally execute commands when boolean conditions are false"
}

// Usage returns the syntax and usage examples for the if-not command.
func (c *IfNotCommand) Usage() string {
	return "\\if-not[condition=boolean_expression] command_to_execute"
}

// HelpInfo returns structured help information for the if-not command.
func (c *IfNotCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "condition",
				Description: "Expression to evaluate for falsiness (see condition evaluation rules below)",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\if-not[condition=false] \\set[var=value]",
				Description: "Execute when condition is explicitly false",
			},
			{
				Command:     "\\if-not[condition=true] \\echo This won't run",
				Description: "Command not executed when condition is explicitly true",
			},
			{
				Command:     "\\if-not[condition=\"\"] \\echo Empty means falsy",
				Description: "Empty strings are falsy - this will execute",
			},
			{
				Command:     "\\if-not[condition=hello] \\echo This won't run",
				Description: "Non-empty strings are truthy - this won't execute",
			},
			{
				Command:     "\\if-not[condition=${_session_id}] \\session-new",
				Description: "Create session if no active session (common pattern)",
			},
			{
				Command:     "\\if-not[condition=${#active_model_name}] \\model-new[catalog_id=O4M] default",
				Description: "Create default model if no active model",
			},
			{
				Command:     "\\if-not[condition=0] \\echo Zero is falsy",
				Description: "Number 0 is explicitly falsy - this will execute",
			},
			{
				Command:     "\\if-not[condition=1] \\echo This won't run",
				Description: "Number 1 is explicitly truthy - this won't execute",
			},
		},
		Notes: []string{
			"CONDITION EVALUATION RULES (case-insensitive):",
			"  Explicitly TRUTHY: 'true', '1', 'yes', 'on', 'enabled'",
			"  Explicitly FALSY: 'false', '0', 'no', 'off', 'disabled'",
			"  Empty strings: FALSY (\"\" or undefined variables)",
			"  Any other non-empty string: TRUTHY (including Unicode like '🌟✨')",
			"Variables are interpolated before evaluation - ${var} becomes var's value",
			"The command after \\if-not is only executed if the condition is falsy",
			"Result of condition evaluation is stored in #if_not_result system variable",
			"This is the logical inverse of \\if - useful for 'create if missing' patterns",
		},
	}
}

// Execute evaluates the condition and optionally executes the command.
// The condition is evaluated as a boolean expression and the command is executed if falsy.
// Options:
//   - condition: boolean expression to evaluate (required)
//
// The command after \if-not is only executed if the condition evaluates to false.
func (c *IfNotCommand) Execute(args map[string]string, input string) error {
	// Get condition parameter
	condition, exists := args["condition"]
	if !exists {
		return fmt.Errorf("condition parameter is required")
	}

	// Evaluate the condition using shared logic
	result := c.evaluateCondition(condition)

	// Store the result in system variable for debugging
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		_ = variableService.SetSystemVariable("#if_not_result", strconv.FormatBool(result))
	}

	// If condition is FALSE, push the command to the stack for execution (inverted logic)
	if !result && strings.TrimSpace(input) != "" {
		if stackService, err := services.GetGlobalStackService(); err == nil {
			stackService.PushCommand(input)
		}
	}

	return nil
}

// evaluateCondition evaluates a boolean expression string
// Note: Variable interpolation (${var}) is handled by the state machine before this command executes,
// so the condition parameter already contains the expanded variable values.
func (c *IfNotCommand) evaluateCondition(condition string) bool {
	// Trim whitespace and evaluate using shared logic
	// Variables have already been interpolated by the state machine
	return stringprocessing.IsTruthy(strings.TrimSpace(condition))
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&IfNotCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register if-not command: %v", err))
	}
}
