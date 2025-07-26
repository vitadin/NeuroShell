// Package assert provides assertion commands for testing and validation in NeuroShell.
// This package contains commands that compare values and set system variables based on results.
package assert

import (
	"fmt"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// EqualCommand implements the \assert-equal command for comparing two values.
// It supports variable interpolation and sets system variables for test results.
type EqualCommand struct{}

// Name returns the command name "assert-equal" for registration and lookup.
func (c *EqualCommand) Name() string {
	return "assert-equal"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *EqualCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the assert-equal command does.
func (c *EqualCommand) Description() string {
	return "Compare two values for equality"
}

// Usage returns the syntax and usage examples for the assert-equal command.
func (c *EqualCommand) Usage() string {
	return "\\assert-equal[expect=expected_value, actual=actual_value]"
}

// HelpInfo returns structured help information for the assert-equal command.
func (c *EqualCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "expect",
				Description: "Expected value for comparison",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "actual",
				Description: "Actual value to compare against expected",
				Required:    true,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\assert-equal[expect=\"hello\", actual=\"hello\"]",
				Description: "Compare two literal string values",
			},
			{
				Command:     "\\assert-equal[expect=\"${expected_result}\", actual=\"${_output}\"]",
				Description: "Compare variables with interpolation",
			},
			{
				Command:     "\\assert-equal[expect=\"5\", actual=\"${count}\"]",
				Description: "Validate command output against expected value",
			},
			{
				Command:     "\\assert-equal[expect=\"success\", actual=\"${_status}\"]",
				Description: "Check status of previous operation",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_status",
				Description: "Exit code: '0' for pass, '1' for fail",
				Type:        "command_output",
				Example:     "0",
			},
			{
				Name:        "_assert_result",
				Description: "Assertion result status",
				Type:        "command_output",
				Example:     "PASS",
			},
			{
				Name:        "_assert_expected",
				Description: "Expected value from comparison",
				Type:        "command_output",
				Example:     "hello",
			},
			{
				Name:        "_assert_actual",
				Description: "Actual value from comparison",
				Type:        "command_output",
				Example:     "hello",
			},
		},
		Notes: []string{
			"Values are compared as strings after variable interpolation by state machine",
			"Useful for testing and validation in .neuro scripts",
			"Supports whitespace and case-sensitive string comparison",
		},
	}
}

// Execute compares two values for equality.
// Values are pre-interpolated by the state machine before this method is called.
// It sets system variables _status, _assert_result, _assert_expected, and _assert_actual.
func (c *EqualCommand) Execute(args map[string]string, _ string) error {

	// Validate required arguments
	expected, hasExpected := args["expect"]
	actual, hasActual := args["actual"]

	if !hasExpected || !hasActual {
		return fmt.Errorf("Usage: %s", c.Usage())
	}

	// Compare the values (already interpolated by state machine)
	isEqual := expected == actual

	// Get variable service to set system variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Set system variables based on result
	if isEqual {
		// Assertion passed
		_ = variableService.SetSystemVariable("_status", "0")
		_ = variableService.SetSystemVariable("_assert_result", "PASS")
		_ = variableService.SetSystemVariable("_assert_expected", expected)
		_ = variableService.SetSystemVariable("_assert_actual", actual)

		// Output success message
		fmt.Printf("✓ Assertion passed: values are equal\n")
		fmt.Printf("  Value: %s\n", expected)
	} else {
		// Assertion failed
		_ = variableService.SetSystemVariable("_status", "1")
		_ = variableService.SetSystemVariable("_assert_result", "FAIL")
		_ = variableService.SetSystemVariable("_assert_expected", expected)
		_ = variableService.SetSystemVariable("_assert_actual", actual)

		// Output failure message with diff-style information
		fmt.Printf("✗ Assertion failed: values are not equal\n")
		fmt.Printf("  Expected: %s\n", expected)
		fmt.Printf("  Actual:   %s\n", actual)
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&EqualCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register assert-equal command: %v", err))
	}
}
