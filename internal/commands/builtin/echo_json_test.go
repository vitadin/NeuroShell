package builtin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestEchoJSONCommand_Name(t *testing.T) {
	cmd := &EchoJSONCommand{}
	assert.Equal(t, "echo-json", cmd.Name())
}

func TestEchoJSONCommand_ParseMode(t *testing.T) {
	cmd := &EchoJSONCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestEchoJSONCommand_Description(t *testing.T) {
	cmd := &EchoJSONCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "json")
}

func TestEchoJSONCommand_Usage(t *testing.T) {
	cmd := &EchoJSONCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\echo-json")
}

func TestEchoJSONCommand_HelpInfo(t *testing.T) {
	cmd := &EchoJSONCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, cmd.Name(), helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, cmd.ParseMode(), helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.NotEmpty(t, helpInfo.Options)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Verify specific options
	optionNames := make(map[string]bool)
	for _, option := range helpInfo.Options {
		optionNames[option.Name] = true
	}
	assert.True(t, optionNames["to"])
}

func TestEchoJSONCommand_Execute_ValidJSON(t *testing.T) {
	cmd := &EchoJSONCommand{}
	setupEchoJSONTestRegistry(t)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:  "simple object",
			input: `{"key": "value"}`,
			expected: `{
  "key": "value"
}`,
		},
		{
			name:  "nested object",
			input: `{"outer": {"inner": "value", "number": 42}}`,
			expected: `{
  "outer": {
    "inner": "value",
    "number": 42
  }
}`,
		},
		{
			name:  "array",
			input: `["item1", "item2", 123]`,
			expected: `[
  "item1",
  "item2",
  123
]`,
		},
		{
			name:  "complex structure",
			input: `{"users": [{"name": "John", "age": 30}, {"name": "Jane", "age": 25}], "total": 2}`,
			// Note: JSON key order is not guaranteed, so we check structure rather than exact format
			expected: "", // Will be verified differently
		},
		{
			name:  "whitespace handling",
			input: `  {"key":"value"}  `,
			expected: `{
  "key": "value"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute command
			err := cmd.Execute(map[string]string{}, tt.input)
			require.NoError(t, err)

			// Verify result was stored in _output
			variableService, _ := services.GetGlobalVariableService()
			result, err := variableService.Get("_output")
			require.NoError(t, err)

			if tt.expected == "" && tt.name == "complex structure" {
				// For complex structure, just verify it's properly formatted JSON
				assert.Contains(t, result, "{\n  ")
				assert.Contains(t, result, "\"users\": [")
				assert.Contains(t, result, "\"total\": 2")
				assert.Contains(t, result, "\"name\": \"John\"")
				assert.Contains(t, result, "\"age\": 30")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestEchoJSONCommand_Execute_InvalidJSON(t *testing.T) {
	cmd := &EchoJSONCommand{}
	setupEchoJSONTestRegistry(t)

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "not json at all",
			input: "hello world",
		},
		{
			name:  "malformed json - missing quote",
			input: `{"key: "value"}`,
		},
		{
			name:  "malformed json - trailing comma",
			input: `{"key": "value",}`,
		},
		{
			name:  "malformed json - unclosed brace",
			input: `{"key": "value"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute command
			err := cmd.Execute(map[string]string{}, tt.input)
			require.NoError(t, err) // Command should never fail

			// Verify error message was stored in _output
			variableService, _ := services.GetGlobalVariableService()
			result, err := variableService.Get("_output")
			require.NoError(t, err)
			assert.Contains(t, result, "Error: Invalid JSON")
			assert.Contains(t, result, tt.input) // Should include original input
		})
	}
}

func TestEchoJSONCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &EchoJSONCommand{}
	setupEchoJSONTestRegistry(t)

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "completely empty",
			input: "",
		},
		{
			name:  "only whitespace",
			input: "   \t\n  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute command
			err := cmd.Execute(map[string]string{}, tt.input)
			require.NoError(t, err)

			// Verify error message was stored
			variableService, _ := services.GetGlobalVariableService()
			result, err := variableService.Get("_output")
			require.NoError(t, err)
			assert.Equal(t, "Error: No JSON data provided", result)
		})
	}
}

func TestEchoJSONCommand_Execute_CustomTargetVariable(t *testing.T) {
	cmd := &EchoJSONCommand{}
	setupEchoJSONTestRegistry(t)

	tests := []struct {
		name      string
		targetVar string
		input     string
	}{
		{
			name:      "user variable",
			targetVar: "my_json",
			input:     `{"test": "value"}`,
		},
		{
			name:      "system variable",
			targetVar: "_error",
			input:     `{"error": "none"}`,
		},
		{
			name:      "custom user variable",
			targetVar: "formatted_data",
			input:     `{"data": [1, 2, 3]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute command with custom target
			args := map[string]string{"to": tt.targetVar}
			err := cmd.Execute(args, tt.input)
			require.NoError(t, err)

			// Verify result was stored in target variable
			variableService, _ := services.GetGlobalVariableService()
			result, err := variableService.Get(tt.targetVar)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			// Should be properly formatted JSON
			assert.Contains(t, result, "{\n  ")
		})
	}
}

func TestEchoJSONCommand_formatJSON(t *testing.T) {
	cmd := &EchoJSONCommand{}

	tests := []struct {
		name        string
		input       string
		indent      int
		expected    string
		expectError bool
	}{
		{
			name:     "simple formatting - 2 spaces",
			input:    `{"key":"value"}`,
			indent:   2,
			expected: "{\n  \"key\": \"value\"\n}",
		},
		{
			name:     "simple formatting - 4 spaces",
			input:    `{"key":"value"}`,
			indent:   4,
			expected: "{\n    \"key\": \"value\"\n}",
		},
		{
			name:     "compact formatting",
			input:    `{"key":"value"}`,
			indent:   0,
			expected: "{\"key\":\"value\"}",
		},
		{
			name:        "invalid json",
			input:       `{"key":}`,
			indent:      2,
			expectError: true,
		},
		{
			name:     "array formatting",
			input:    `[1,2,3]`,
			indent:   2,
			expected: "[\n  1,\n  2,\n  3\n]",
		},
		{
			name:     "whitespace trimming",
			input:    "  {\"trimmed\":true}  ",
			indent:   2,
			expected: "{\n  \"trimmed\": true\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cmd.formatJSON(tt.input, tt.indent)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestEchoJSONCommand_storeResult(t *testing.T) {
	cmd := &EchoJSONCommand{}
	setupEchoJSONTestRegistry(t)

	tests := []struct {
		name      string
		targetVar string
		result    string
	}{
		{
			name:      "system variable",
			targetVar: "_output",
			result:    "test result",
		},
		{
			name:      "user variable",
			targetVar: "custom_var",
			result:    "custom result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store result
			cmd.storeResult(tt.targetVar, tt.result)

			// Verify it was stored
			variableService, _ := services.GetGlobalVariableService()
			retrieved, err := variableService.Get(tt.targetVar)
			require.NoError(t, err)
			assert.Equal(t, tt.result, retrieved)
		})
	}
}

func TestEchoJSONCommand_storeResult_NoVariableService(t *testing.T) {
	cmd := &EchoJSONCommand{}
	// Don't set up registry - this should test graceful degradation

	// This should not panic or error
	require.NotPanics(t, func() {
		cmd.storeResult("test_var", "test_result")
	})
}

func TestEchoJSONCommand_InterfaceCompliance(_ *testing.T) {
	var _ neurotypes.Command = (*EchoJSONCommand)(nil)
}

func TestEchoJSONCommand_Execute_IndentOption(t *testing.T) {
	cmd := &EchoJSONCommand{}
	setupEchoJSONTestRegistry(t)

	tests := []struct {
		name         string
		input        string
		indentOption string
		expectString string
	}{
		{
			name:         "default 2-space indent",
			input:        `{"key": "value"}`,
			indentOption: "",
			expectString: "{\n  \"key\": \"value\"\n}",
		},
		{
			name:         "4-space indent",
			input:        `{"key": "value"}`,
			indentOption: "4",
			expectString: "{\n    \"key\": \"value\"\n}",
		},
		{
			name:         "compact format",
			input:        `{"key": "value"}`,
			indentOption: "0",
			expectString: "{\"key\":\"value\"}",
		},
		{
			name:         "invalid indent falls back to default",
			input:        `{"key": "value"}`,
			indentOption: "invalid",
			expectString: "{\n  \"key\": \"value\"\n}",
		},
		{
			name:         "negative indent falls back to default",
			input:        `{"key": "value"}`,
			indentOption: "-1",
			expectString: "{\n  \"key\": \"value\"\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{}
			if tt.indentOption != "" {
				args["indent"] = tt.indentOption
			}

			// Execute command
			err := cmd.Execute(args, tt.input)
			require.NoError(t, err)

			// Verify result was stored correctly
			variableService, _ := services.GetGlobalVariableService()
			result, err := variableService.Get("_output")
			require.NoError(t, err)
			assert.Equal(t, tt.expectString, result)
		})
	}
}

func TestEchoJSONCommand_Execute_DebugNetworkExample(t *testing.T) {
	cmd := &EchoJSONCommand{}
	setupEchoJSONTestRegistry(t)

	// Simulate typical debug network data structure
	debugNetworkJSON := `{
		"http_request": {
			"method": "POST",
			"url": "https://api.openai.com/v1/chat/completions",
			"headers": {"Content-Type": "application/json"},
			"body": "{\"model\":\"gpt-4\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}]}"
		},
		"http_response": {
			"status_code": 200,
			"headers": {"Content-Type": "application/json"},
			"body": "{\"choices\":[{\"message\":{\"content\":\"Hello there!\"}}]}"
		},
		"timing": {
			"request_time": "2024-01-01T12:00:00Z",
			"response_time": "2024-01-01T12:00:01Z",
			"duration_ms": 1000
		}
	}`

	// Execute command
	err := cmd.Execute(map[string]string{}, debugNetworkJSON)
	require.NoError(t, err)

	// Verify result is properly formatted
	variableService, _ := services.GetGlobalVariableService()
	result, err := variableService.Get("_output")
	require.NoError(t, err)

	// Should be properly indented
	assert.Contains(t, result, "{\n  \"http_request\":")
	assert.Contains(t, result, "  \"timing\": {")
	assert.Contains(t, result, "    \"duration_ms\": 1000")
}

// setupEchoJSONTestRegistry creates a test registry with required services for echo-json testing
func setupEchoJSONTestRegistry(t *testing.T) {
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	services.SetGlobalRegistry(registry)
	context.SetGlobalContext(ctx)

	// Store cleanup for test teardown
	t.Cleanup(func() {
		// Reset to previous state if needed
		// For now, we just let the test state persist since each test is isolated
	})
}
