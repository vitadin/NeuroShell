package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputCaptureSystemVariables(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	// Initial state - @last_output should be empty
	lastOutput, err := ctx.GetVariable("@last_output")
	assert.NoError(t, err)
	assert.Equal(t, "", lastOutput)

	// Capture some output
	testOutput := "Test command output"
	ctx.CaptureOutput(testOutput)

	// @last_output should now have the captured value
	lastOutput, err = ctx.GetVariable("@last_output")
	assert.NoError(t, err)
	assert.Equal(t, testOutput, lastOutput)

	// Capture new output (this overwrites the previous value)
	newOutput := "New command output"
	ctx.CaptureOutput(newOutput)

	// @last_output should now have the new value
	lastOutput, err = ctx.GetVariable("@last_output")
	assert.NoError(t, err)
	assert.Equal(t, newOutput, lastOutput)
}

func TestOutputCaptureInGetAllVariables(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	// Set some output
	ctx.CaptureOutput("Test output for GetAllVariables")

	// Get all variables
	allVars := ctx.GetAllVariables()

	// Check that @last_output is included
	assert.Contains(t, allVars, "@last_output")
	assert.Equal(t, "Test output for GetAllVariables", allVars["@last_output"])
}

func TestOutputCaptureVariablesInTestMode(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	// Test that variables work correctly in test mode
	ctx.CaptureOutput("Test mode output")

	// Should get the captured output
	value, ok := ctx.getSystemVariable("@last_output")
	assert.True(t, ok)
	assert.Equal(t, "Test mode output", value)
}

func TestOutputCaptureVariablesInProductionMode(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(false) // Production mode

	// Test that variables work correctly in production mode
	ctx.CaptureOutput("Production mode output")

	// Should get the captured output
	value, ok := ctx.getSystemVariable("@last_output")
	assert.True(t, ok)
	assert.Equal(t, "Production mode output", value)
}

func TestOutputCaptureWithComplexOutput(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	// Test with complex output including newlines, special characters
	complexOutput := "Line 1\nLine 2\n\tTabbed content\nSpecial chars: !@#$%^&*()\nUnicode: 你好 🌟"
	ctx.CaptureOutput(complexOutput)

	// Verify through system variables
	lastOutput, err := ctx.GetVariable("@last_output")
	assert.NoError(t, err)
	assert.Equal(t, complexOutput, lastOutput)

	// Verify GetAllVariables preserves the content
	allVars := ctx.GetAllVariables()
	assert.Equal(t, complexOutput, allVars["@last_output"])
}

func TestOutputCaptureMultipleCaptures(t *testing.T) {
	ctx := New()
	ctx.SetTestMode(true)

	outputs := []string{
		"First command output",
		"Second command output",
		"Third command output",
	}

	for _, output := range outputs {
		// Capture output (this directly sets @last_output)
		ctx.CaptureOutput(output)

		// Verify last output has the captured value
		last, _ := ctx.GetVariable("@last_output")
		assert.Equal(t, output, last)
	}

	// After all captures, @last_output should have the final output
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
			// Capture the test output
			ctx.CaptureOutput(tc.output)

			// Verify it's captured exactly as provided
			last, _ := ctx.GetVariable("@last_output")
			assert.Equal(t, tc.output, last)
		})
	}
}
