package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArrayValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple array",
			input:    "[\\get, \\set, \\list]",
			expected: []string{"\\get", "\\set", "\\list"},
		},
		{
			name:     "array with spaces",
			input:    "[ \\get , \\set , \\list ]",
			expected: []string{"\\get", "\\set", "\\list"},
		},
		{
			name:     "array with quoted items",
			input:    "[\"\\get\", '\\set', \\list]",
			expected: []string{"\\get", "\\set", "\\list"},
		},
		{
			name:     "empty array",
			input:    "[]",
			expected: []string{},
		},
		{
			name:     "single item not in array",
			input:    "\\get",
			expected: []string{"\\get"},
		},
		{
			name:     "quoted single item",
			input:    "\"\\get\"",
			expected: []string{"\\get"},
		},
		{
			name:     "array with one item",
			input:    "[\\get]",
			expected: []string{"\\get"},
		},
		{
			name:     "mixed content array",
			input:    "[command1, \"quoted command\", command3]",
			expected: []string{"command1", "quoted command", "command3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseArrayValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitByCommaWithArrays(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple key-value pairs",
			input:    "key1=value1, key2=value2",
			expected: []string{"key1=value1", " key2=value2"},
		},
		{
			name:     "array value with other options",
			input:    "keywords=[\\get, \\set], style=bold, theme=dark",
			expected: []string{"keywords=[\\get, \\set]", " style=bold", " theme=dark"},
		},
		{
			name:     "nested brackets",
			input:    "keywords=[cmd1, [nested, array]], other=value",
			expected: []string{"keywords=[cmd1, [nested, array]]", " other=value"},
		},
		{
			name:     "quoted strings with commas",
			input:    "text=\"hello, world\", keywords=[\\get, \\set]",
			expected: []string{"text=\"hello, world\"", " keywords=[\\get, \\set]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByComma(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseInputWithArrays(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedMsg  string
		expectedOpts map[string]string
	}{
		{
			name:         "render command with keyword array",
			input:        "\\render[keywords=[\\get, \\set], style=bold] This is a test message with \\get and \\set commands",
			expectedName: "render",
			expectedMsg:  "This is a test message with \\get and \\set commands",
			expectedOpts: map[string]string{
				"keywords": "[\\get, \\set]",
				"style":    "bold",
			},
		},
		{
			name:         "multiple array parameters",
			input:        "\\highlight[keywords=[\\get, \\set], colors=[red, blue], theme=dark] test",
			expectedName: "highlight",
			expectedMsg:  "test",
			expectedOpts: map[string]string{
				"keywords": "[\\get, \\set]",
				"colors":   "[red, blue]",
				"theme":    "dark",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseInput(tt.input)
			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.Equal(t, tt.expectedMsg, cmd.Message)
			assert.Equal(t, tt.expectedOpts, cmd.Options)
		})
	}
}
