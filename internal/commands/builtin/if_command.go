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
				Description: "Expression to evaluate for truthiness (see condition evaluation rules below)",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\if[condition=true] \\set[var=value]",
				Description: "Execute when condition is explicitly true",
			},
			{
				Command:     "\\if[condition=false] \\echo This won't run",
				Description: "Command not executed when condition is explicitly false",
			},
			{
				Command:     "\\if[condition=hello] \\echo Any text is truthy",
				Description: "Non-empty strings (including 'hello') are truthy",
			},
			{
				Command:     "\\if[condition=\"\"] \\echo Empty string",
				Description: "Empty strings are falsy - this won't execute",
			},
			{
				Command:     "\\if[condition=${@user}] \\echo Hello ${@user}",
				Description: "Execute if user variable exists and is non-empty",
			},
			{
				Command:     "\\if[condition=${debug_mode}] \\echo Debug enabled",
				Description: "Execute if debug_mode variable contains truthy value",
			},
			{
				Command:     "\\if[condition=1] \\echo Numeric true",
				Description: "Number 1 is explicitly truthy",
			},
			{
				Command:     "\\if[condition=0] \\echo This won't run",
				Description: "Number 0 is explicitly falsy",
			},
		},
		Notes: []string{
			"CONDITION EVALUATION RULES (case-insensitive):",
			"  Explicitly TRUTHY: 'true', '1', 'yes', 'on', 'enabled'",
			"  Explicitly FALSY: 'false', '0', 'no', 'off', 'disabled'",
			"  Empty strings: FALSY (\"\" or undefined variables)",
			"  Any other non-empty string: TRUTHY (including Unicode like '🌟✨')",
			"Variables are interpolated before evaluation - ${var} becomes var's value",
			"The command after \\if is only executed if the condition is truthy",
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
	// Trim whitespace and evaluate using shared logic
	// Variables have already been interpolated by the state machine
	return stringprocessing.IsTruthy(strings.TrimSpace(condition))
}

// IsReadOnly returns false as the if command modifies system state.
func (c *IfCommand) IsReadOnly() bool {
	return false
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&IfCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register if command: %v", err))
	}
}
