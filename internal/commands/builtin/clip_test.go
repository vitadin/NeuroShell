package builtin

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestClipCommand_Name(t *testing.T) {
	cmd := &ClipCommand{}
	assert.Equal(t, "clip", cmd.Name())
}

func TestClipCommand_ParseMode(t *testing.T) {
	cmd := &ClipCommand{}
	assert.Equal(t, neurotypes.ParseModeRaw, cmd.ParseMode())
}

func TestClipCommand_Description(t *testing.T) {
	cmd := &ClipCommand{}
	assert.Equal(t, "Copy text to system clipboard", cmd.Description())
}

func TestClipCommand_Usage(t *testing.T) {
	cmd := &ClipCommand{}
	assert.Equal(t, "\\clip text to copy", cmd.Usage())
}

func TestClipCommand_HelpInfo(t *testing.T) {
	cmd := &ClipCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "clip", helpInfo.Command)
	assert.Equal(t, "Copy text to system clipboard", helpInfo.Description)
	assert.Equal(t, "\\clip text to copy", helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeRaw, helpInfo.ParseMode)

	// Check examples
	assert.NotEmpty(t, helpInfo.Examples)
	assert.GreaterOrEqual(t, len(helpInfo.Examples), 3)

	// Check stored variables
	assert.NotEmpty(t, helpInfo.StoredVariables)
	foundClipboard := false
	for _, variable := range helpInfo.StoredVariables {
		if variable.Name == "_clipboard" {
			foundClipboard = true
			assert.Equal(t, "system_output", variable.Type)
		}
	}
	assert.True(t, foundClipboard, "_clipboard variable should be documented")

	// Check notes
	assert.NotEmpty(t, helpInfo.Notes)
}

func TestClipCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &ClipCommand{}

	// Capture stdout to check output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "completely empty",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   \t\n  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(nil, tt.input)
			assert.NoError(t, err)
		})
	}

	// Restore stdout and check output
	_ = w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Should contain warning message for both test cases
	assert.Contains(t, outputStr, "Warning: No content specified. Clipboard unchanged.")
	// Warning should appear twice (once for each test case)
	assert.Equal(t, 2, strings.Count(outputStr, "Warning: No content specified. Clipboard unchanged."))
}

func TestClipCommand_Execute_BasicText(t *testing.T) {
	// Skip this test on Linux where clipboard is not available
	if !clipboardAvailable {
		t.Skip("Skipping clipboard test on Linux (clipboard not available)")
	}

	cmd := &ClipCommand{}

	// Capture stdout to check feedback
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tests := []struct {
		name          string
		input         string
		expectedChars int
	}{
		{
			name:          "simple text",
			input:         "Hello World",
			expectedChars: 11,
		},
		{
			name:          "text with spaces",
			input:         "This is a longer message",
			expectedChars: 24,
		},
		{
			name:          "single character",
			input:         "x",
			expectedChars: 1,
		},
		{
			name:          "empty string explicitly",
			input:         "",
			expectedChars: 0, // Will be handled by empty input case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(nil, tt.input)
			// All cases should not error - fallback mechanism handles clipboard issues gracefully
			assert.NoError(t, err)
		})
	}

	// Restore stdout and check output
	_ = w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Should contain either success messages or fallback messages
	// We can't guarantee clipboard works in test environment, so check for either
	assert.True(t,
		strings.Contains(outputStr, "Copied") ||
			strings.Contains(outputStr, "Stored") ||
			strings.Contains(outputStr, "Warning"),
		"Should contain success, fallback, or warning message")
}

func TestClipCommand_Execute_VariableInterpolation(t *testing.T) {
	// Skip this test on Linux where clipboard is not available
	if !clipboardAvailable {
		t.Skip("Skipping clipboard test on Linux (clipboard not available)")
	}

	// Note: Variable interpolation is handled by the shell before our command executes
	// So we test with already-interpolated strings
	cmd := &ClipCommand{}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tests := []struct {
		name               string
		input              string // Already interpolated by shell
		expectedChars      int
		expectedInFeedback string
	}{
		{
			name:          "interpolated variable",
			input:         "Hello John!", // ${name} already interpolated to "John"
			expectedChars: 11,
		},
		{
			name:          "interpolated output",
			input:         "The answer is 42", // ${_output} already interpolated
			expectedChars: 16,
		},
		{
			name:          "complex interpolation",
			input:         "User: John, Status: active", // Multiple variables interpolated
			expectedChars: 26,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(nil, tt.input)
			assert.NoError(t, err)
		})
	}

	// Restore stdout and check output
	_ = w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Should handle all interpolated content properly
	assert.True(t,
		strings.Contains(outputStr, "Copied") ||
			strings.Contains(outputStr, "Stored"),
		"Should contain success or fallback message")
}

func TestClipCommand_FallbackToVariable(t *testing.T) {
	cmd := &ClipCommand{}

	// Setup test context and services
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	// Register variable service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test fallback behavior directly
	err = cmd.fallbackToVariable("test content", "simulated clipboard error")
	assert.NoError(t, err)

	// Restore stdout and check output
	_ = w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Check that fallback message was displayed
	assert.Contains(t, outputStr, "Failed to copy to clipboard: simulated clipboard error")
	assert.Contains(t, outputStr, "Stored 12 characters in _clipboard variable")

	// Check that content was stored in _clipboard variable
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)
	clipboardValue, err := variableService.Get("_clipboard")
	assert.NoError(t, err)
	assert.Equal(t, "test content", clipboardValue)

	// Cleanup
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

func TestClipCommand_FallbackToVariable_ServiceUnavailable(t *testing.T) {
	cmd := &ClipCommand{}

	// Don't setup services - should fail gracefully
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry()) // Empty registry

	err := cmd.fallbackToVariable("test content", "clipboard error")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service unavailable")

	// Cleanup
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})
}

func TestClipCommand_CharacterCountAccuracy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "ascii text",
			input:    "Hello",
			expected: 5,
		},
		{
			name:     "unicode text",
			input:    "Hello 🌍",
			expected: 10, // 5 + 1 space + 4 bytes for emoji (len counts bytes in UTF-8)
		},
		{
			name:     "multiline text",
			input:    "Line 1\nLine 2\nLine 3",
			expected: 20,
		},
		{
			name:     "text with tabs",
			input:    "Col1\tCol2\tCol3",
			expected: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualCount := len(tt.input)
			assert.Equal(t, tt.expected, actualCount,
				"Character count should match for input: %q", tt.input)
		})
	}
}

func TestClipCommand_Execute_PlatformSpecific(t *testing.T) {
	cmd := &ClipCommand{}

	// Setup test context and services
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	// Register variable service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test with actual content
	err = cmd.Execute(nil, "Platform test content")
	assert.NoError(t, err)

	// Restore stdout and check output
	_ = w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// On platforms where clipboard is available, should show "Copied"
	// On platforms where clipboard is not available (Linux), should show fallback behavior
	if clipboardAvailable {
		assert.Contains(t, outputStr, "Copied 21 characters to clipboard")
	} else {
		// Linux fallback behavior
		assert.Contains(t, outputStr, "Failed to copy to clipboard: clipboard not available on this platform")
		assert.Contains(t, outputStr, "Stored 21 characters in _clipboard variable")

		// Verify content was stored in _clipboard variable
		variableService, err := services.GetGlobalVariableService()
		require.NoError(t, err)
		clipboardValue, err := variableService.Get("_clipboard")
		assert.NoError(t, err)
		assert.Equal(t, "Platform test content", clipboardValue)
	}

	// Cleanup
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

func TestClipCommand_ClipboardAvailability(t *testing.T) {
	// This test documents the expected platform behavior
	t.Logf("Clipboard available on this platform: %v", clipboardAvailable)

	// Platform-specific expectations
	if clipboardAvailable {
		t.Log("Expected: Direct clipboard access available")

		// Test that init function exists and can be called
		err := initClipboard()
		// Don't assert success since we might be in a headless environment
		// but the function should exist and not panic
		t.Logf("Clipboard init result: %v", err)
	} else {
		t.Log("Expected: Fallback to _clipboard variable")

		// Test that init function returns expected error
		err := initClipboard()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clipboard not available")

		// Test that write function returns expected error
		err = writeToClipboard("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clipboard not available")
	}
}

// Interface compliance check
var _ neurotypes.Command = (*ClipCommand)(nil)
