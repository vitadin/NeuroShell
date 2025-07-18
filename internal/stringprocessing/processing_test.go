// Package stringprocessing provides text processing utilities for NeuroShell.
package stringprocessing

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterpretEscapeSequences(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "newline escape",
			input:    "line1\\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "tab escape",
			input:    "col1\\tcol2",
			expected: "col1\tcol2",
		},
		{
			name:     "carriage return escape",
			input:    "line1\\rline2",
			expected: "line1\rline2",
		},
		{
			name:     "backslash escape",
			input:    "path\\\\file",
			expected: "path\\file",
		},
		{
			name:     "quote escapes",
			input:    "He said \\\"Hello\\\" and I said \\'Hi\\'",
			expected: "He said \"Hello\" and I said 'Hi'",
		},
		{
			name:     "multiple escapes",
			input:    "line1\\nline2\\tcolumn\\\\path",
			expected: "line1\nline2\tcolumn\\path",
		},
		{
			name:     "no escapes",
			input:    "plain text",
			expected: "plain text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := InterpretEscapeSequences(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCleanContinuationMarkers(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple continuation",
			input:    "line1\n... line2",
			expected: "line1\nline2",
		},
		{
			name:     "multiple continuations",
			input:    "line1\n... line2\n... line3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "standalone continuation markers",
			input:    "line1\n...\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "mixed continuation patterns",
			input:    "line1\n... line2\n...\n... line3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "continuation marker with spaces",
			input:    "line1\n...    line2   ",
			expected: "line1\nline2",
		},
		{
			name:     "no continuation markers",
			input:    "line1\nline2\nline3",
			expected: "line1\nline2\nline3",
		},
		{
			name:     "dots that are not continuation markers",
			input:    "This is... not a continuation\nNormal line",
			expected: "This is... not a continuation\nNormal line",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CleanContinuationMarkers(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestProcessTextForMarkdown(t *testing.T) {
	testCases := []struct {
		name             string
		input            string
		interpretEscapes bool
		expected         string
	}{
		{
			name:             "raw mode with escape sequences",
			input:            "line1\\nline2\\n... line3",
			interpretEscapes: false,
			expected:         "line1\\nline2\\nline3",
		},
		{
			name:             "interpret escapes mode",
			input:            "line1\\nline2\\n... line3",
			interpretEscapes: true,
			expected:         "line1\nline2\nline3",
		},
		{
			name:             "actual newlines with continuations",
			input:            "line1\n... line2\n... line3",
			interpretEscapes: false,
			expected:         "line1\nline2\nline3",
		},
		{
			name:             "mixed newlines and escapes",
			input:            "line1\\n... line2\n... line3",
			interpretEscapes: true,
			expected:         "line1\nline2\nline3",
		},
		{
			name:             "standalone continuation markers",
			input:            "line1\\n...\\nline2",
			interpretEscapes: false,
			expected:         "line1\\nline2",
		},
		{
			name:             "no continuation markers, raw mode",
			input:            "line1\\nline2\\nline3",
			interpretEscapes: false,
			expected:         "line1\\nline2\\nline3",
		},
		{
			name:             "no continuation markers, interpret escapes",
			input:            "line1\\nline2\\nline3",
			interpretEscapes: true,
			expected:         "line1\nline2\nline3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ProcessTextForMarkdown(tc.input, tc.interpretEscapes)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
		hasError bool
	}{
		// Explicitly truthy values
		{"true", "true", true, false},
		{"TRUE uppercase", "TRUE", true, false},
		{"True mixed case", "True", true, false},
		{"1", "1", true, false},
		{"yes", "yes", true, false},
		{"YES uppercase", "YES", true, false},
		{"on", "on", true, false},
		{"enabled", "enabled", true, false},

		// Explicitly falsy values
		{"false", "false", false, false},
		{"FALSE uppercase", "FALSE", false, false},
		{"0", "0", false, false},
		{"no", "no", false, false},
		{"NO uppercase", "NO", false, false},
		{"off", "off", false, false},
		{"disabled", "disabled", false, false},

		// Empty and whitespace
		{"empty string", "", false, false},
		{"whitespace only", "   ", false, false},
		{"tabs and spaces", "\t  \n  ", false, false},

		// Non-empty strings (truthy)
		{"random string", "hello", true, false},
		{"random uppercase", "HELLO", true, false},
		{"numbers as string", "123", true, false},
		{"special characters", "!@#$%", true, false},
		{"unicode", "🌟✨", true, false},

		// Whitespace around values
		{"spaced true", "  true  ", true, false},
		{"spaced false", "  false  ", false, false},
		{"spaced random", "  hello  ", true, false},

		// Edge cases for malformed conditions
		{"normal length string", strings.Repeat("a", 100), true, false},
		{"boundary length string", strings.Repeat("a", 200), true, false},
		{"too long string", strings.Repeat("a", 201), false, true},
		{"very long string", strings.Repeat("malformed;", 50), false, true},

		// Simulate malformed condition from semicolon bug
		{"malformed condition", `false"; \echo "done"`, true, false},
		{"long malformed condition", strings.Repeat(`false"; \echo "done"; `, 10), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := IsTruthy(tt.value)

			if tt.hasError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "condition too long")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsTruthy_ErrorMessages(t *testing.T) {
	// Test that error messages are helpful
	longCondition := strings.Repeat("a", 250)
	_, err := IsTruthy(longCondition)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "condition too long")
	assert.Contains(t, err.Error(), "250 chars")
	assert.Contains(t, err.Error(), "malformed condition")
}

func TestIsTruthy_SafetyProtection(t *testing.T) {
	// Test that the safety protection would prevent the infinite loop scenario

	// Simulate what happens when semicolon causes variable corruption
	malformedConditions := []string{
		`false"; \echo "done"`,
		`false"; \echo "done"; \echo "done"`,
		`false"; \echo "done"; \echo "done"; \echo "done"`,
		strings.Repeat(`false"; \echo "done"; `, 20), // Very long
	}

	for i, condition := range malformedConditions {
		t.Run(fmt.Sprintf("malformed_condition_%d", i), func(t *testing.T) {
			result, err := IsTruthy(condition)

			if len(condition) > 200 {
				// Should be caught by length protection
				assert.Error(t, err)
			} else {
				// Shorter malformed conditions are still truthy but at least don't error
				assert.NoError(t, err)
				assert.True(t, result) // They're truthy (non-empty) but won't infinite loop due to length protection eventually
			}
		})
	}
}
