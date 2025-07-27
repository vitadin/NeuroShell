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

func TestIsAPIRelated(t *testing.T) {
	// Define the test provider list that matches the standard providers
	providers := []string{"openai", "anthropic", "openrouter", "moonshot", "gemini"}

	tests := []struct {
		name             string
		variableName     string
		expectedIsAPI    bool
		expectedProvider string
	}{
		// Provider names with API keywords - should match with provider
		{
			name:             "openai api key",
			variableName:     "OPENAI_API_KEY",
			expectedIsAPI:    true,
			expectedProvider: "openai",
		},
		{
			name:             "anthropic api key",
			variableName:     "ANTHROPIC_API_KEY",
			expectedIsAPI:    true,
			expectedProvider: "anthropic",
		},
		{
			name:             "openrouter api key",
			variableName:     "OPENROUTER_API_KEY",
			expectedIsAPI:    true,
			expectedProvider: "openrouter",
		},
		{
			name:             "moonshot api key",
			variableName:     "MOONSHOT_API_KEY",
			expectedIsAPI:    true,
			expectedProvider: "moonshot",
		},
		{
			name:             "case insensitive provider with key",
			variableName:     "OpenAI_Key",
			expectedIsAPI:    true,
			expectedProvider: "openai",
		},
		{
			name:             "mixed case with api",
			variableName:     "My_OpenAI_API_Token",
			expectedIsAPI:    true,
			expectedProvider: "openai",
		},
		{
			name:             "custom openai secret",
			variableName:     "MY_OPENAI_SECRET",
			expectedIsAPI:    true,
			expectedProvider: "openai",
		},
		{
			name:             "anthropic api token",
			variableName:     "WORK_ANTHROPIC_API_TOKEN",
			expectedIsAPI:    true,
			expectedProvider: "anthropic",
		},

		// API keywords without provider - should match as generic
		{
			name:             "custom api key",
			variableName:     "CUSTOM_API_KEY",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "secret token",
			variableName:     "MY_SECRET_TOKEN",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "just key",
			variableName:     "ACCESS_KEY",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "case insensitive api",
			variableName:     "Service_API_Token",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "case insensitive secret",
			variableName:     "App_SECRET",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},

		// Provider names without API keywords - should NOT match
		{
			name:             "openai debug flag",
			variableName:     "OPENAI_DEBUG",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "openai endpoint url",
			variableName:     "OPENAI_ENDPOINT",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "anthropic model name",
			variableName:     "ANTHROPIC_MODEL",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "just provider name",
			variableName:     "OPENAI",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "provider in middle without api keywords",
			variableName:     "SOME_OPENAI_CONFIG",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "moonshot timeout setting",
			variableName:     "MOONSHOT_TIMEOUT",
			expectedIsAPI:    false,
			expectedProvider: "",
		},

		// Non-API variables - should not match
		{
			name:             "database url",
			variableName:     "DATABASE_URL",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "random config",
			variableName:     "MY_CONFIG_VALUE",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "path variable",
			variableName:     "PATH",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "home directory",
			variableName:     "HOME",
			expectedIsAPI:    false,
			expectedProvider: "",
		},

		// Edge cases
		{
			name:             "empty string",
			variableName:     "",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "only spaces",
			variableName:     "   ",
			expectedIsAPI:    false,
			expectedProvider: "",
		},
		{
			name:             "api keyword at end",
			variableName:     "MYAPI",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "key keyword at start",
			variableName:     "KEYCHAIN_CONFIG",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "secret keyword in middle",
			variableName:     "MYSECRETCONFIG",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},

		// Provider detection with multiple providers in name
		{
			name:             "multiple providers with api keyword",
			variableName:     "OPENAI_ANTHROPIC_API_KEY",
			expectedIsAPI:    true,
			expectedProvider: "openai", // First provider found
		},

		// Case sensitivity tests
		{
			name:             "uppercase api",
			variableName:     "MY_API_TOKEN",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "lowercase key",
			variableName:     "my_access_key",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "mixed case secret",
			variableName:     "My_Secret_Value",
			expectedIsAPI:    true,
			expectedProvider: "generic",
		},
		{
			name:             "uppercase provider with lowercase api",
			variableName:     "OPENAI_api_key",
			expectedIsAPI:    true,
			expectedProvider: "openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAPI, provider := IsAPIRelated(tt.variableName, providers)
			assert.Equal(t, tt.expectedIsAPI, isAPI, "IsAPIRelated result should match expected")
			assert.Equal(t, tt.expectedProvider, provider, "detected provider should match expected")
		})
	}
}
