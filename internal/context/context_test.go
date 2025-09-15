package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/pkg/neurotypes"
)

func TestNew(t *testing.T) {
	ctx := New()

	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.variables)
	assert.NotNil(t, ctx.history)
	assert.NotNil(t, ctx.scriptMetadata)
	assert.NotEmpty(t, ctx.sessionID)
	assert.Contains(t, ctx.sessionID, "session_")
	assert.False(t, ctx.testMode)
	assert.Equal(t, 3, len(ctx.variables)) // _style, _default_command, and _completion_mode are initialized by default

	// Verify that _style is initialized to empty string
	styleValue, err := ctx.GetVariable("_style")
	assert.NoError(t, err)
	assert.Equal(t, "", styleValue)

	// Verify that _default_command is initialized to "echo"
	defaultCmd, err := ctx.GetVariable("_default_command")
	assert.NoError(t, err)
	assert.Equal(t, "echo", defaultCmd)

	// Verify that _completion_mode is initialized to "tab"
	completionMode, err := ctx.GetVariable("_completion_mode")
	assert.NoError(t, err)
	assert.Equal(t, "tab", completionMode)
	assert.Equal(t, 0, len(ctx.history))
	assert.Equal(t, 0, len(ctx.scriptMetadata))
}

func TestGetVariable_UserVariables(t *testing.T) {
	ctx := New()

	tests := []struct {
		name     string
		varName  string
		varValue string
		wantErr  bool
	}{
		{
			name:     "simple variable",
			varName:  "test",
			varValue: "value",
			wantErr:  false,
		},
		{
			name:     "empty value",
			varName:  "empty",
			varValue: "",
			wantErr:  false,
		},
		{
			name:     "complex name with underscores",
			varName:  "complex_var_name",
			varValue: "complex value with spaces",
			wantErr:  false,
		},
		{
			name:     "unicode value",
			varName:  "unicode",
			varValue: "héllo wørld 🌍",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the variable
			err := ctx.SetVariable(tt.varName, tt.varValue)
			require.NoError(t, err)

			// Get the variable
			value, err := ctx.GetVariable(tt.varName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.varValue, value)
			}
		})
	}
}

func TestGetVariable_NonExistentVariable(t *testing.T) {
	ctx := New()

	value, err := ctx.GetVariable("nonexistent")
	assert.NoError(t, err)     // Should not error
	assert.Equal(t, "", value) // Should return empty string
}

func TestSetVariable_ValidVariables(t *testing.T) {
	ctx := New()

	tests := []struct {
		name     string
		varName  string
		varValue string
	}{
		{"simple", "test", "value"},
		{"empty_value", "empty", ""},
		{"numbers", "num", "123"},
		{"special_chars", "special", "!@#$%^&*()"},
		{"multiline", "multi", "line1\nline2\nline3"},
		{"unicode", "unicode", "héllo wørld 🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ctx.SetVariable(tt.varName, tt.varValue)
			assert.NoError(t, err)

			// Verify it was set
			value, err := ctx.GetVariable(tt.varName)
			assert.NoError(t, err)
			assert.Equal(t, tt.varValue, value)
		})
	}
}

func TestSetVariable_SystemVariables_Forbidden(t *testing.T) {
	ctx := New()

	tests := []struct {
		name    string
		varName string
	}{
		{"system_at", "@pwd"},
		{"system_hash", "#session_id"},
		{"system_underscore", "_output"},
		{"custom_at", "@custom"},
		{"custom_hash", "#custom"},
		{"custom_underscore", "_custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ctx.SetVariable(tt.varName, "value")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "cannot set system variable")
			assert.Contains(t, err.Error(), tt.varName)
		})
	}
}

func TestSetVariable_WhitelistedGlobalVariables(t *testing.T) {
	ctx := New()

	tests := []struct {
		name     string
		varName  string
		varValue string
		wantErr  bool
	}{
		{"whitelisted_style_variable", "_style", "dark", false},
		{"whitelisted_style_empty", "_style", "", false},
		{"whitelisted_style_overwrite", "_style", "light", false},
		{"non_whitelisted_underscore", "_config", "value", true},
		{"non_whitelisted_underscore_2", "_secret", "value", true},
		{"non_whitelisted_underscore_3", "_custom", "value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ctx.SetVariable(tt.varName, tt.varValue)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot set system variable")
				assert.Contains(t, err.Error(), tt.varName)
			} else {
				assert.NoError(t, err)

				// Verify the value was actually set
				actualValue, err := ctx.GetVariable(tt.varName)
				assert.NoError(t, err)
				assert.Equal(t, tt.varValue, actualValue)
			}
		})
	}
}

func TestGetMessageHistory(t *testing.T) {
	ctx := New()

	// Test empty history
	history := ctx.GetMessageHistory(5)
	assert.Equal(t, 0, len(history))

	// Add some messages
	messages := []neurotypes.Message{
		{ID: "1", Role: "user", Content: "Hello", Timestamp: time.Now()},
		{ID: "2", Role: "assistant", Content: "Hi there", Timestamp: time.Now()},
		{ID: "3", Role: "user", Content: "How are you?", Timestamp: time.Now()},
		{ID: "4", Role: "assistant", Content: "I'm fine", Timestamp: time.Now()},
		{ID: "5", Role: "user", Content: "Good!", Timestamp: time.Now()},
	}

	ctx.history = messages

	tests := []struct {
		name     string
		n        int
		expected int
	}{
		{"get_all_negative", -1, 5},
		{"get_all_zero", 0, 5},
		{"get_all_exact", 5, 5},
		{"get_all_more", 10, 5},
		{"get_last_three", 3, 3},
		{"get_last_one", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := ctx.GetMessageHistory(tt.n)
			assert.Equal(t, tt.expected, len(history))

			// If getting subset, verify we get the last N messages
			if tt.n > 0 && tt.n < len(messages) {
				expectedStart := len(messages) - tt.n
				for i, msg := range history {
					assert.Equal(t, messages[expectedStart+i].ID, msg.ID)
				}
			}
		})
	}
}

func TestGetSessionState(t *testing.T) {
	ctx := New()

	// Set some variables
	require.NoError(t, ctx.SetVariable("var1", "value1"))
	require.NoError(t, ctx.SetVariable("var2", "value2"))

	// Add some history
	ctx.history = []neurotypes.Message{
		{ID: "1", Role: "user", Content: "Hello", Timestamp: time.Now()},
	}

	state := ctx.GetSessionState()

	assert.Equal(t, ctx.sessionID, state.ID)
	assert.Equal(t, ctx.variables, state.Variables)
	assert.Equal(t, ctx.history, state.History)
	assert.NotZero(t, state.CreatedAt)
	assert.NotZero(t, state.UpdatedAt)
}

func TestGetSystemVariable(t *testing.T) {
	ctx := New()

	tests := []struct {
		name      string
		varName   string
		shouldGet bool
		validator func(value string) bool
	}{
		{
			name:      "pwd",
			varName:   "@pwd",
			shouldGet: true,
			validator: filepath.IsAbs,
		},
		{
			name:      "user",
			varName:   "@user",
			shouldGet: true,
			validator: func(value string) bool {
				return len(value) > 0
			},
		},
		{
			name:      "home",
			varName:   "@home",
			shouldGet: true,
			validator: filepath.IsAbs,
		},
		{
			name:      "date",
			varName:   "@date",
			shouldGet: true,
			validator: func(value string) bool {
				_, err := time.Parse("2006-01-02", value)
				return err == nil
			},
		},
		{
			name:      "time",
			varName:   "@time",
			shouldGet: true,
			validator: func(value string) bool {
				_, err := time.Parse("15:04:05", value)
				return err == nil
			},
		},
		{
			name:      "os",
			varName:   "@os",
			shouldGet: true,
			validator: func(value string) bool {
				return strings.Contains(value, "/")
			},
		},
		{
			name:      "session_id",
			varName:   "#session_id",
			shouldGet: true,
			validator: func(value string) bool {
				return strings.HasPrefix(value, "session_")
			},
		},
		{
			name:      "message_count",
			varName:   "#message_count",
			shouldGet: true,
			validator: func(value string) bool {
				return value == "0" // No messages in history
			},
		},
		{
			name:      "test_mode_false",
			varName:   "#test_mode",
			shouldGet: true,
			validator: func(value string) bool {
				return value == "false"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := ctx.getSystemVariable(tt.varName)
			assert.Equal(t, tt.shouldGet, ok)
			if tt.shouldGet && tt.validator != nil {
				assert.True(t, tt.validator(value), "Value validation failed for %s: %s", tt.varName, value)
			}
		})
	}
}

func TestGetSystemVariable_TestMode(t *testing.T) {
	ctx := New()

	// Test false mode
	value, ok := ctx.getSystemVariable("#test_mode")
	assert.True(t, ok)
	assert.Equal(t, "false", value)

	// Test true mode
	ctx.SetTestMode(true)
	value, ok = ctx.getSystemVariable("#test_mode")
	assert.True(t, ok)
	assert.Equal(t, "true", value)
}

func TestGetSystemVariable_MessageHistory(t *testing.T) {
	ctx := New()

	tests := []struct {
		name        string
		varName     string
		expected    string
		shouldExist bool
	}{
		{"message_1", "1", "", false},
		{"message_2", "2", "", false},
		{"message_123", "123", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := ctx.getSystemVariable(tt.varName)
			assert.Equal(t, tt.shouldExist, ok)
			assert.Equal(t, tt.expected, value)
		})
	}
}

func TestGetSystemVariable_NonExistent(t *testing.T) {
	ctx := New()

	tests := []string{
		"@nonexistent",
		"#nonexistent",
		"_nonexistent",
		"regular_var",
		"",
		"@",
		"#",
		"_",
	}

	for _, varName := range tests {
		t.Run(varName, func(t *testing.T) {
			value, ok := ctx.getSystemVariable(varName)
			assert.False(t, ok)
			assert.Equal(t, "", value)
		})
	}
}

func TestInterpolateVariables_NoVariables(t *testing.T) {
	ctx := New()

	tests := []string{
		"plain text",
		"no variables here",
		"$not_a_variable",
		"${incomplete",
		"missing}",
		"",
	}

	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			result := ctx.InterpolateVariables(text)
			assert.Equal(t, text, result)
		})
	}
}

func TestInterpolateVariables_SimpleVariables(t *testing.T) {
	ctx := New()
	require.NoError(t, ctx.SetVariable("name", "Alice"))
	require.NoError(t, ctx.SetVariable("age", "30"))
	require.NoError(t, ctx.SetVariable("empty", ""))

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single variable",
			input:    "Hello ${name}",
			expected: "Hello Alice",
		},
		{
			name:     "multiple variables",
			input:    "${name} is ${age} years old",
			expected: "Alice is 30 years old",
		},
		{
			name:     "empty variable",
			input:    "Value: ${empty}",
			expected: "Value: ",
		},
		{
			name:     "missing variable",
			input:    "Missing: ${missing}",
			expected: "Missing: ",
		},
		{
			name:     "variable at start",
			input:    "${name} says hello",
			expected: "Alice says hello",
		},
		{
			name:     "variable at end",
			input:    "Age is ${age}",
			expected: "Age is 30",
		},
		{
			name:     "only variable",
			input:    "${name}",
			expected: "Alice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.InterpolateVariables(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInterpolateVariables_SystemVariables(t *testing.T) {
	ctx := New()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "session_id",
			input:    "Session: ${#session_id}",
			contains: "session_",
		},
		{
			name:     "message_count",
			input:    "Count: ${#message_count}",
			contains: "Count: 0",
		},
		{
			name:     "test_mode",
			input:    "Test: ${#test_mode}",
			contains: "Test: false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.InterpolateVariables(tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestInterpolateVariables_NestedVariables(t *testing.T) {
	ctx := New()
	require.NoError(t, ctx.SetVariable("var1", "${var2}"))
	require.NoError(t, ctx.SetVariable("var2", "final_value"))
	require.NoError(t, ctx.SetVariable("prefix", "var"))
	require.NoError(t, ctx.SetVariable("suffix", "2"))

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple nested",
			input:    "${var1}",
			expected: "final_value",
		},
		{
			name:     "nested with text",
			input:    "Result: ${var1}",
			expected: "Result: final_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.InterpolateVariables(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInterpolateVariables_CircularReference(t *testing.T) {
	ctx := New()
	require.NoError(t, ctx.SetVariable("var1", "${var2}"))
	require.NoError(t, ctx.SetVariable("var2", "${var1}"))

	// Should not cause infinite loop due to maxIterations limit
	result := ctx.InterpolateVariables("${var1}")
	// Exact result depends on implementation but should not hang
	assert.NotEmpty(t, result)
}

func TestInterpolateVariables_MaxIterations(t *testing.T) {
	ctx := New()

	// Create a chain longer than maxIterations (10)
	for i := 1; i <= 15; i++ {
		if i == 15 {
			require.NoError(t, ctx.SetVariable(fmt.Sprintf("var%d", i), "final"))
		} else {
			require.NoError(t, ctx.SetVariable(fmt.Sprintf("var%d", i), fmt.Sprintf("${var%d}", i+1)))
		}
	}

	result := ctx.InterpolateVariables("${var1}")
	// Should stop after maxIterations, so won't reach "final"
	assert.NotEqual(t, "final", result)
	assert.Contains(t, result, "${var") // Should still contain unresolved variables
}

func TestInterpolateOnce(t *testing.T) {
	ctx := New()
	require.NoError(t, ctx.SetVariable("name", "Alice"))
	require.NoError(t, ctx.SetVariable("nested", "${name}"))

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single pass simple",
			input:    "Hello ${name}",
			expected: "Hello Alice",
		},
		{
			name:     "single pass nested - first level only",
			input:    "${nested}",
			expected: "${name}", // Only one pass, so nested variable isn't resolved
		},
		{
			name:     "multiple variables single pass",
			input:    "${name} and ${nested}",
			expected: "Alice and ${name}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.interpolateOnce(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestScriptMetadata(t *testing.T) {
	ctx := New()

	// Test getting non-existent metadata
	value, exists := ctx.GetScriptMetadata("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, value)

	// Test setting and getting various neurotypes
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"string", "str_key", "string_value"},
		{"int", "int_key", 42},
		{"bool", "bool_key", true},
		{"slice", "slice_key", []string{"a", "b", "c"}},
		{"map", "map_key", map[string]int{"one": 1, "two": 2}},
		{"nil", "nil_key", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx.SetScriptMetadata(tt.key, tt.value)

			value, exists := ctx.GetScriptMetadata(tt.key)
			assert.True(t, exists)
			assert.Equal(t, tt.value, value)
		})
	}
}

func TestClearScriptMetadata(t *testing.T) {
	ctx := New()

	// Set some metadata
	ctx.SetScriptMetadata("key1", "value1")
	ctx.SetScriptMetadata("key2", "value2")

	// Verify they exist
	_, exists := ctx.GetScriptMetadata("key1")
	assert.True(t, exists)
	_, exists = ctx.GetScriptMetadata("key2")
	assert.True(t, exists)

	// Clear metadata
	ctx.ClearScriptMetadata()

	// Verify they're gone
	_, exists = ctx.GetScriptMetadata("key1")
	assert.False(t, exists)
	_, exists = ctx.GetScriptMetadata("key2")
	assert.False(t, exists)

	// Should be able to set new metadata after clearing
	ctx.SetScriptMetadata("new_key", "new_value")
	value, exists := ctx.GetScriptMetadata("new_key")
	assert.True(t, exists)
	assert.Equal(t, "new_value", value)
}

func TestSetTestMode(t *testing.T) {
	ctx := New()

	// Default should be false
	assert.False(t, ctx.IsTestMode())

	// Set to true
	ctx.SetTestMode(true)
	assert.True(t, ctx.IsTestMode())

	// Set back to false
	ctx.SetTestMode(false)
	assert.False(t, ctx.IsTestMode())
}

func TestInterpolateVariables_EdgeCases(t *testing.T) {
	ctx := New()
	require.NoError(t, ctx.SetVariable("normal", "value"))

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "malformed variable start",
			input:    "$normal}",
			expected: "$normal}",
		},
		{
			name:     "malformed variable end",
			input:    "${normal",
			expected: "${normal",
		},
		{
			name:     "empty variable name",
			input:    "${}",
			expected: "${}",
		},
		{
			name:     "nested braces",
			input:    "${${normal}}",
			expected: "", // Inner ${normal} resolves to 'value', then ${value} resolves to empty (missing var)
		},
		{
			name:     "multiple dollar signs",
			input:    "$${normal}",
			expected: "$value",
		},
		{
			name:     "escaped looking variable",
			input:    "\\${normal}",
			expected: "\\value",
		},
		{
			name:     "unicode in variable name",
			input:    "${🙂}",
			expected: "",
		},
		{
			name:     "whitespace in variable name",
			input:    "${nor mal}",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.InterpolateVariables(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSystemVariables_OSEnvironment(t *testing.T) {
	ctx := New()

	// Test that we can get system variables even when OS calls might fail
	// by temporarily modifying environment

	// Save original values
	origGOOS := os.Getenv("GOOS")
	origGOARCH := os.Getenv("GOARCH")

	// Test with empty environment
	require.NoError(t, os.Unsetenv("GOOS"))
	require.NoError(t, os.Unsetenv("GOARCH"))

	value, ok := ctx.getSystemVariable("@os")
	assert.True(t, ok)
	assert.Equal(t, "/", value) // Should be "/" when both are empty

	// Test with set environment
	require.NoError(t, os.Setenv("GOOS", "linux"))
	require.NoError(t, os.Setenv("GOARCH", "amd64"))

	value, ok = ctx.getSystemVariable("@os")
	assert.True(t, ok)
	assert.Equal(t, "linux/amd64", value)

	// Restore original values
	if origGOOS != "" {
		require.NoError(t, os.Setenv("GOOS", origGOOS))
	}
	if origGOARCH != "" {
		require.NoError(t, os.Setenv("GOARCH", origGOARCH))
	}
}

func TestMockContext_SetVariableWithValidation(t *testing.T) {
	// Test that the context validation method works correctly
	ResetGlobalContext()
	ctx := GetGlobalContext()
	ctx.SetTestMode(true)

	tests := []struct {
		name     string
		varName  string
		varValue string
		wantErr  bool
	}{
		{"whitelisted_style_variable", "_style", "dark", false},
		{"whitelisted_style_empty", "_style", "", false},
		{"non_whitelisted_underscore", "_config", "value", true},
		{"system_at_variable", "@pwd", "value", true},
		{"system_hash_variable", "#session", "value", true},
		{"regular_variable", "normal", "value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ctx.SetVariableWithValidation(tt.varName, tt.varValue)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot set system variable")
				assert.Contains(t, err.Error(), tt.varName)
			} else {
				assert.NoError(t, err)

				// Verify the value was actually set
				actualValue, err := ctx.GetVariable(tt.varName)
				assert.NoError(t, err)
				assert.Equal(t, tt.varValue, actualValue)
			}
		})
	}
}

// Benchmark tests
func BenchmarkInterpolateVariables_NoVariables(b *testing.B) {
	ctx := New()
	text := "This is a long text with no variables to interpolate at all"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.InterpolateVariables(text)
	}
}

func BenchmarkInterpolateVariables_SingleVariable(b *testing.B) {
	ctx := New()
	if err := ctx.SetVariable("name", "Alice"); err != nil {
		b.Fatal(err)
	}
	text := "Hello ${name}, how are you today?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.InterpolateVariables(text)
	}
}

func BenchmarkInterpolateVariables_MultipleVariables(b *testing.B) {
	ctx := New()
	if err := ctx.SetVariable("name", "Alice"); err != nil {
		b.Fatal(err)
	}
	if err := ctx.SetVariable("age", "30"); err != nil {
		b.Fatal(err)
	}
	if err := ctx.SetVariable("city", "New York"); err != nil {
		b.Fatal(err)
	}
	text := "${name} is ${age} years old and lives in ${city}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.InterpolateVariables(text)
	}
}

func BenchmarkGetVariable_UserVariable(b *testing.B) {
	ctx := New()
	if err := ctx.SetVariable("test", "value"); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.GetVariable("test")
	}
}

func BenchmarkGetVariable_SystemVariable(b *testing.B) {
	ctx := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.GetVariable("#session_id")
	}
}

func BenchmarkLargeVariableMap(b *testing.B) {
	ctx := New()

	// Setup large variable map
	for i := 0; i < 1000; i++ {
		if err := ctx.SetVariable(fmt.Sprintf("var%d", i), fmt.Sprintf("value%d", i)); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ctx.GetVariable("var500") // Access middle variable
	}
}

// Test cases for read-only command functionality

// MockCommand for testing read-only functionality
type MockTestCommand struct {
	name     string
	readOnly bool
}

func (m *MockTestCommand) Name() string                                { return m.name }
func (m *MockTestCommand) ParseMode() neurotypes.ParseMode             { return neurotypes.ParseModeKeyValue }
func (m *MockTestCommand) Description() string                         { return "Test command" }
func (m *MockTestCommand) Usage() string                               { return "\\test" }
func (m *MockTestCommand) Execute(_ map[string]string, _ string) error { return nil }
func (m *MockTestCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     m.Name(),
		Description: m.Description(),
		Usage:       m.Usage(),
		ParseMode:   m.ParseMode(),
		Examples:    []neurotypes.HelpExample{},
	}
}
func (m *MockTestCommand) IsReadOnly() bool { return m.readOnly }

func TestNeuroContext_ReadOnlyOverrides(t *testing.T) {
	ctx := New()

	// Initially should have no overrides
	overrides := ctx.GetReadOnlyOverrides()
	assert.Empty(t, overrides)

	// Set a read-only override
	ctx.SetCommandReadOnly("test-command", true)

	// Should now appear in overrides
	overrides = ctx.GetReadOnlyOverrides()
	assert.Contains(t, overrides, "test-command")
	assert.True(t, overrides["test-command"])

	// Set another override with different value
	ctx.SetCommandReadOnly("another-command", false)

	// Should now have both overrides
	overrides = ctx.GetReadOnlyOverrides()
	assert.Len(t, overrides, 2)
	assert.True(t, overrides["test-command"])
	assert.False(t, overrides["another-command"])

	// Remove first override
	ctx.RemoveCommandReadOnlyOverride("test-command")

	// Should now only have the second override
	overrides = ctx.GetReadOnlyOverrides()
	assert.Len(t, overrides, 1)
	assert.Contains(t, overrides, "another-command")
	assert.NotContains(t, overrides, "test-command")
}

func TestNeuroContext_IsCommandReadOnly_SelfDeclared(t *testing.T) {
	ctx := New()

	tests := []struct {
		name             string
		command          neurotypes.Command
		expectedReadOnly bool
	}{
		{
			name:             "read-only command",
			command:          &MockTestCommand{name: "get", readOnly: true},
			expectedReadOnly: true,
		},
		{
			name:             "writable command",
			command:          &MockTestCommand{name: "set", readOnly: false},
			expectedReadOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ctx.IsCommandReadOnly(tt.command)
			assert.Equal(t, tt.expectedReadOnly, result)
		})
	}
}

func TestNeuroContext_IsCommandReadOnly_WithOverrides(t *testing.T) {
	ctx := New()

	// Create test commands with different self-declared read-only status
	readOnlyCmd := &MockTestCommand{name: "get", readOnly: true}
	writableCmd := &MockTestCommand{name: "set", readOnly: false}

	tests := []struct {
		name             string
		command          neurotypes.Command
		override         *bool // nil means no override
		expectedReadOnly bool
	}{
		{
			name:             "read-only command without override",
			command:          readOnlyCmd,
			override:         nil,
			expectedReadOnly: true,
		},
		{
			name:             "read-only command overridden to writable",
			command:          readOnlyCmd,
			override:         &[]bool{false}[0],
			expectedReadOnly: false,
		},
		{
			name:             "writable command without override",
			command:          writableCmd,
			override:         nil,
			expectedReadOnly: false,
		},
		{
			name:             "writable command overridden to read-only",
			command:          writableCmd,
			override:         &[]bool{true}[0],
			expectedReadOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous overrides
			ctx.ClearAllReadOnlyOverrides()

			// Set override if specified
			if tt.override != nil {
				ctx.SetCommandReadOnly(tt.command.Name(), *tt.override)
			}

			result := ctx.IsCommandReadOnly(tt.command)
			assert.Equal(t, tt.expectedReadOnly, result)
		})
	}
}

func TestNeuroContext_ReadOnlyOverrides_ThreadSafety(t *testing.T) {
	ctx := New()

	// Test concurrent access to read-only overrides
	done := make(chan bool)
	numGoroutines := 10

	// Start multiple goroutines that read and write overrides
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			commandName := fmt.Sprintf("command-%d", id)

			// Set override
			ctx.SetCommandReadOnly(commandName, id%2 == 0)

			// Read override
			overrides := ctx.GetReadOnlyOverrides()
			assert.Contains(t, overrides, commandName)

			// Remove override
			ctx.RemoveCommandReadOnlyOverride(commandName)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// All overrides should be removed
	overrides := ctx.GetReadOnlyOverrides()
	assert.Empty(t, overrides)
}

func TestNeuroContext_ReadOnlyOverrides_Integration(t *testing.T) {
	ctx := New()

	// Create commands with different self-declared read-only status
	commands := []neurotypes.Command{
		&MockTestCommand{name: "get", readOnly: true},
		&MockTestCommand{name: "set", readOnly: false},
		&MockTestCommand{name: "vars", readOnly: true},
		&MockTestCommand{name: "bash", readOnly: false},
	}

	// Test initial state (should use self-declared status)
	assert.True(t, ctx.IsCommandReadOnly(commands[0]))  // get: read-only
	assert.False(t, ctx.IsCommandReadOnly(commands[1])) // set: writable
	assert.True(t, ctx.IsCommandReadOnly(commands[2]))  // vars: read-only
	assert.False(t, ctx.IsCommandReadOnly(commands[3])) // bash: writable

	// Apply some overrides
	ctx.SetCommandReadOnly("get", false) // Override read-only to writable
	ctx.SetCommandReadOnly("set", true)  // Override writable to read-only

	// Test overridden state
	assert.False(t, ctx.IsCommandReadOnly(commands[0])) // get: overridden to writable
	assert.True(t, ctx.IsCommandReadOnly(commands[1]))  // set: overridden to read-only
	assert.True(t, ctx.IsCommandReadOnly(commands[2]))  // vars: unchanged (read-only)
	assert.False(t, ctx.IsCommandReadOnly(commands[3])) // bash: unchanged (writable)

	// Remove one override
	ctx.RemoveCommandReadOnlyOverride("get")

	// Test state after removing override
	assert.True(t, ctx.IsCommandReadOnly(commands[0]))  // get: back to self-declared (read-only)
	assert.True(t, ctx.IsCommandReadOnly(commands[1]))  // set: still overridden (read-only)
	assert.True(t, ctx.IsCommandReadOnly(commands[2]))  // vars: unchanged (read-only)
	assert.False(t, ctx.IsCommandReadOnly(commands[3])) // bash: unchanged (writable)

	// Verify overrides map contains only remaining override
	overrides := ctx.GetReadOnlyOverrides()
	assert.Len(t, overrides, 1)
	assert.Contains(t, overrides, "set")
	assert.True(t, overrides["set"])
}
