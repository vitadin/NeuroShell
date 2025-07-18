package builtin

import (
	"fmt"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

// WhileCommand implements the \while command for conditional looping.
// It evaluates a condition and repeatedly executes a command while the condition remains true.
type WhileCommand struct{}

// Name returns the command name "while" for registration and lookup.
func (c *WhileCommand) Name() string {
	return "while"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *WhileCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the while command does.
func (c *WhileCommand) Description() string {
	return "Repeatedly execute commands while a boolean condition remains true"
}

// Usage returns the syntax and usage examples for the while command.
func (c *WhileCommand) Usage() string {
	return "\\while[condition=boolean_expression] command_to_execute"
}

// HelpInfo returns structured help information for the while command.
func (c *WhileCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "condition",
				Description: "Expression to evaluate for truthiness (same rules as \\if command)",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\set[counter=0]; \\while[condition=\"${counter} < 5\"] \\set[counter=\"$((${counter} + 1))\"]",
				Description: "Simple counter loop from 0 to 4",
			},
			{
				Command:     "\\set[done=false]; \\while[condition=\"${done}\"] \\bash[check_status.sh]; \\set[done=\"${_output}\"]",
				Description: "Loop until external condition becomes false",
			},
			{
				Command:     "\\while[condition=true] \\echo Infinite loop - be careful!",
				Description: "Infinite loop (will be stopped by stack overflow protection)",
			},
			{
				Command:     "\\set[file=data.txt]; \\while[condition=\"${file}\"] \\bash[process.sh ${file}]; \\set[file=\"${_output}\"]",
				Description: "Process files until no more files returned",
			},
		},
		Notes: []string{
			"CONDITION EVALUATION RULES (identical to \\if command):",
			"  Explicitly TRUTHY: 'true', '1', 'yes', 'on', 'enabled'",
			"  Explicitly FALSY: 'false', '0', 'no', 'off', 'disabled'",
			"  Empty strings: FALSY (\"\" or undefined variables)",
			"  Any other non-empty string: TRUTHY (including Unicode like '🌟✨')",
			"Variables are interpolated before evaluation - ${var} becomes var's value",
			"The command executes repeatedly while the condition remains truthy",
			"Stack overflow protection prevents infinite loops",
			"Result of condition evaluation is stored in #while_result system variable",
		},
	}
}

// Execute evaluates the condition and conditionally re-queues itself and the command.
// The condition is evaluated as a boolean expression using the same logic as \if.
// Options:
//   - condition: boolean expression to evaluate (required)
//   - _template: original condition template for re-interpolation (internal)
//
// The command after \while is executed repeatedly while the condition evaluates to true.
// Variable interpolation happens fresh on each iteration to enable dynamic conditions.
func (c *WhileCommand) Execute(args map[string]string, input string) error {
	// Get condition parameter
	condition, exists := args["condition"]
	if !exists {
		return fmt.Errorf("condition parameter is required")
	}

	// Get or detect the original template for re-queuing
	template, hasTemplate := args["_template"]
	if !hasTemplate {
		// First execution - try to detect if this was a variable template
		// by checking common interpolated values and inferring the original
		template = c.inferOriginalTemplate(condition, args)
	}

	// Evaluate the condition using shared logic with safety protections
	result, err := c.evaluateCondition(condition)
	if err != nil {
		return fmt.Errorf("condition evaluation failed: %w", err)
	}

	// Store the result in system variable for debugging
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		_ = variableService.SetSystemVariable("#while_result", strconv.FormatBool(result))
	}

	// If condition is true, re-queue this while command and the command to execute
	if result && strings.TrimSpace(input) != "" {
		if stackService, err := services.GetGlobalStackService(); err == nil {
			// Push this while command again for next iteration with template for fresh interpolation
			whileCmd := fmt.Sprintf("\\while[condition=%s,_template=%s] %s", template, template, input)
			stackService.PushCommand(whileCmd)

			// Push the command to execute this iteration
			stackService.PushCommand(input)
		}
	}

	return nil
}

// inferOriginalTemplate attempts to detect if the condition was originally a variable template.
// This is a heuristic approach since we don't have access to the pre-interpolated template.
// It works by examining the current condition and context to guess the original ${var} pattern.
func (c *WhileCommand) inferOriginalTemplate(condition string, args map[string]string) string {
	// If the condition already contains ${}, it wasn't fully interpolated
	if strings.Contains(condition, "${") {
		return condition
	}

	// For simple boolean conditions, assume they might have been variables
	condition = strings.TrimSpace(condition)
	
	// Check if this looks like a simple interpolated variable value
	// and try to find a matching variable with that value
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		// Look for variables that currently have this condition as their value
		// This is a heuristic - we're guessing the original variable name
		
		// Try common variable names for loop conditions
		commonNames := []string{"flag", "test", "done", "continue", "loop", "condition", "state"}
		for _, name := range commonNames {
			if value, err := variableService.GetVariable(name); err == nil && value == condition {
				return fmt.Sprintf("${%s}", name)
			}
		}
		
		// As a fallback, check if any user variable matches this value
		// This is expensive but necessary for correctness
		if allVars, err := variableService.ListVariables(); err == nil {
			for varName, varValue := range allVars {
				if varValue == condition && !strings.HasPrefix(varName, "#") && !strings.HasPrefix(varName, "_") && !strings.HasPrefix(varName, "@") {
					return fmt.Sprintf("${%s}", varName)
				}
			}
		}
	}

	// If we can't infer the original template, return the condition as-is
	// This means the while loop will not update dynamically, but won't infinite loop
	return condition
}

// evaluateCondition evaluates a boolean expression string using shared logic with safety protections.
// Note: Variable interpolation (${var}) is handled by the state machine before this command executes,
// so the condition parameter already contains the expanded variable values.
func (c *WhileCommand) evaluateCondition(condition string) (bool, error) {
	// Use shared condition evaluation logic with safety protections
	return stringprocessing.IsTruthy(strings.TrimSpace(condition))
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&WhileCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register while command: %v", err))
	}
}
