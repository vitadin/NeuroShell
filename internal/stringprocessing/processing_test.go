// Package stringprocessing provides text processing utilities for NeuroShell.
package stringprocessing

import (
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
