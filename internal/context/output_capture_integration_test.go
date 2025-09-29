package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputCaptureSystemVariables(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	// Initial state - both variables should be empty
	currentOutput, err := ctx.GetVariable("@current_output")
	assert.NoError(t, err)
	assert.Equal(t, "", currentOutput)

	lastOutput, err := ctx.GetVariable("@last_output")
	assert.NoError(t, err)
	assert.Equal(t, "", lastOutput)

	// Capture some output
	testOutput := "Test command output"
	ctx.CaptureOutput(testOutput)

	// Check @current_output
	currentOutput, err = ctx.GetVariable("@current_output")
	assert.NoError(t, err)
	assert.Equal(t, testOutput, currentOutput)

	// @last_output should still be empty
	lastOutput, err = ctx.GetVariable("@last_output")
	assert.NoError(t, err)
	assert.Equal(t, "", lastOutput)

	// Reset output (simulating new command)
	ctx.ResetOutput()

	// @current_output should be empty, @last_output should have previous value
	currentOutput, err = ctx.GetVariable("@current_output")
	assert.NoError(t, err)
	assert.Equal(t, "", currentOutput)

	lastOutput, err = ctx.GetVariable("@last_output")
	assert.NoError(t, err)
	assert.Equal(t, testOutput, lastOutput)

	// Capture new output
	newOutput := "New command output"
	ctx.CaptureOutput(newOutput)

	// Check both variables
	currentOutput, err = ctx.GetVariable("@current_output")
	assert.NoError(t, err)
	assert.Equal(t, newOutput, currentOutput)

	lastOutput, err = ctx.GetVariable("@last_output")
	assert.NoError(t, err)
	assert.Equal(t, testOutput, lastOutput)
}

func TestOutputCaptureInGetAllVariables(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	// Set some output
	ctx.CaptureOutput("Test output for GetAllVariables")

	// Get all variables
	allVars := ctx.GetAllVariables()

	// Check that output capture variables are included
	assert.Contains(t, allVars, "@current_output")
	assert.Contains(t, allVars, "@last_output")

	assert.Equal(t, "Test output for GetAllVariables", allVars["@current_output"])
	assert.Equal(t, "", allVars["@last_output"])

	// Reset and check again
	ctx.ResetOutput()
	allVars = ctx.GetAllVariables()

	assert.Equal(t, "", allVars["@current_output"])
	assert.Equal(t, "Test output for GetAllVariables", allVars["@last_output"])
}

func TestOutputCaptureVariablesInTestMode(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	// Test that variables work correctly in test mode
	ctx.CaptureOutput("Test mode output")

	// Should get the captured output
	value, ok := ctx.getSystemVariable("@current_output")
	assert.True(t, ok)
	assert.Equal(t, "Test mode output", value)

	value, ok = ctx.getSystemVariable("@last_output")
	assert.True(t, ok)
	assert.Equal(t, "", value)
}

func TestOutputCaptureVariablesInProductionMode(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(false) // Production mode

	// Test that variables work correctly in production mode
	ctx.CaptureOutput("Production mode output")

	// Should get the captured output
	value, ok := ctx.getSystemVariable("@current_output")
	assert.True(t, ok)
	assert.Equal(t, "Production mode output", value)

	value, ok = ctx.getSystemVariable("@last_output")
	assert.True(t, ok)
	assert.Equal(t, "", value)
}

func TestOutputCaptureWithComplexOutput(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	// Test with complex output including newlines, special characters
	complexOutput := "Line 1\nLine 2\n\tTabbed content\nSpecial chars: !@#$%^&*()\nUnicode: 你好 🌟"
	ctx.CaptureOutput(complexOutput)

	// Verify through system variables
	currentOutput, err := ctx.GetVariable("@current_output")
	assert.NoError(t, err)
	assert.Equal(t, complexOutput, currentOutput)

	// Verify GetAllVariables preserves the content
	allVars := ctx.GetAllVariables()
	assert.Equal(t, complexOutput, allVars["@current_output"])
}

func TestOutputCaptureMultipleResets(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	outputs := []string{
		"First command output",
		"Second command output",
		"Third command output",
	}

	for i, output := range outputs {
		// Capture output
		ctx.CaptureOutput(output)

		// Verify current output
		current, _ := ctx.GetVariable("@current_output")
		assert.Equal(t, output, current)

		// Verify last output (should be from previous iteration or empty)
		last, _ := ctx.GetVariable("@last_output")
		if i == 0 {
			assert.Equal(t, "", last)
		} else {
			assert.Equal(t, outputs[i-1], last)
		}

		// Reset for next iteration
		ctx.ResetOutput()
	}

	// After final reset, current should be empty and last should have the final output
	current, _ := ctx.GetVariable("@current_output")
	assert.Equal(t, "", current)

	last, _ := ctx.GetVariable("@last_output")
	assert.Equal(t, outputs[len(outputs)-1], last)
}

func TestOutputCaptureEmptyAndWhitespaceHandling(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	testCases := []struct {
		name   string
		output string
	}{
		{"empty string", ""},
		{"single space", " "},
		{"multiple spaces", "   "},
		{"tab character", "\t"},
		{"newline character", "\n"},
		{"mixed whitespace", " \t\n "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset to clean state
			ctx.ResetOutput()
			ctx.ResetOutput() // Double reset to clear last output

			// Capture the test output
			ctx.CaptureOutput(tc.output)

			// Verify it's captured exactly as provided
			current, _ := ctx.GetVariable("@current_output")
			assert.Equal(t, tc.output, current)

			// Reset and verify it moves to last output correctly
			ctx.ResetOutput()

			current, _ = ctx.GetVariable("@current_output")
			assert.Equal(t, "", current)

			last, _ := ctx.GetVariable("@last_output")
			assert.Equal(t, tc.output, last)
		})
	}
}
