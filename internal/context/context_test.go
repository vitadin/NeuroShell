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
	assert.NotNil(t, ctx.executionQueue)
	assert.NotNil(t, ctx.scriptMetadata)
	assert.NotEmpty(t, ctx.sessionID)
	assert.Contains(t, ctx.sessionID, "session_")
	assert.False(t, ctx.testMode)
	assert.Equal(t, 0, len(ctx.variables))
	assert.Equal(t, 0, len(ctx.history))
	assert.Equal(t, 0, len(ctx.executionQueue))
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
	assert.Error(t, err)
	assert.Equal(t, "", value)
	assert.Contains(t, err.Error(), "variable nonexistent not found")
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
	ctx.SetVariable("var1", "value1")
	ctx.SetVariable("var2", "value2")

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
			validator: func(value string) bool {
				return filepath.IsAbs(value)
			},
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
			validator: func(value string) bool {
				return filepath.IsAbs(value)
			},
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
		name     string
		varName  string
		expected string
	}{
		{"message_1", "1", "message_1_placeholder"},
		{"message_2", "2", "message_2_placeholder"},
		{"message_123", "123", "message_123_placeholder"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := ctx.getSystemVariable(tt.varName)
			assert.True(t, ok)
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
	ctx.SetVariable("name", "Alice")
	ctx.SetVariable("age", "30")
	ctx.SetVariable("empty", "")

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
	ctx.SetVariable("var1", "${var2}")
	ctx.SetVariable("var2", "final_value")
	ctx.SetVariable("prefix", "var")
	ctx.SetVariable("suffix", "2")

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
	ctx.SetVariable("var1", "${var2}")
	ctx.SetVariable("var2", "${var1}")

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
			ctx.SetVariable(fmt.Sprintf("var%d", i), "final")
		} else {
			ctx.SetVariable(fmt.Sprintf("var%d", i), fmt.Sprintf("${var%d}", i+1))
		}
	}

	result := ctx.InterpolateVariables("${var1}")
	// Should stop after maxIterations, so won't reach "final"
	assert.NotEqual(t, "final", result)
	assert.Contains(t, result, "${var") // Should still contain unresolved variables
}

func TestInterpolateOnce(t *testing.T) {
	ctx := New()
	ctx.SetVariable("name", "Alice")
	ctx.SetVariable("nested", "${name}")

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

func TestQueueCommand(t *testing.T) {
	ctx := New()

	assert.Equal(t, 0, ctx.GetQueueSize())

	ctx.QueueCommand("command1")
	assert.Equal(t, 1, ctx.GetQueueSize())

	ctx.QueueCommand("command2")
	ctx.QueueCommand("command3")
	assert.Equal(t, 3, ctx.GetQueueSize())
}

func TestDequeueCommand(t *testing.T) {
	ctx := New()

	// Test empty queue
	cmd, ok := ctx.DequeueCommand()
	assert.False(t, ok)
	assert.Equal(t, "", cmd)

	// Add commands and test FIFO behavior
	ctx.QueueCommand("first")
	ctx.QueueCommand("second")
	ctx.QueueCommand("third")

	cmd, ok = ctx.DequeueCommand()
	assert.True(t, ok)
	assert.Equal(t, "first", cmd)
	assert.Equal(t, 2, ctx.GetQueueSize())

	cmd, ok = ctx.DequeueCommand()
	assert.True(t, ok)
	assert.Equal(t, "second", cmd)
	assert.Equal(t, 1, ctx.GetQueueSize())

	cmd, ok = ctx.DequeueCommand()
	assert.True(t, ok)
	assert.Equal(t, "third", cmd)
	assert.Equal(t, 0, ctx.GetQueueSize())

	// Queue should be empty now
	cmd, ok = ctx.DequeueCommand()
	assert.False(t, ok)
	assert.Equal(t, "", cmd)
}

func TestClearQueue(t *testing.T) {
	ctx := New()

	ctx.QueueCommand("command1")
	ctx.QueueCommand("command2")
	ctx.QueueCommand("command3")
	assert.Equal(t, 3, ctx.GetQueueSize())

	ctx.ClearQueue()
	assert.Equal(t, 0, ctx.GetQueueSize())

	// Should be able to add commands after clearing
	ctx.QueueCommand("new_command")
	assert.Equal(t, 1, ctx.GetQueueSize())
}

func TestPeekQueue(t *testing.T) {
	ctx := New()

	// Test empty queue
	queue := ctx.PeekQueue()
	assert.Equal(t, 0, len(queue))

	// Add commands
	ctx.QueueCommand("command1")
	ctx.QueueCommand("command2")
	ctx.QueueCommand("command3")

	queue = ctx.PeekQueue()
	expected := []string{"command1", "command2", "command3"}
	assert.Equal(t, expected, queue)

	// Verify queue size unchanged
	assert.Equal(t, 3, ctx.GetQueueSize())

	// Verify returned slice is a copy (modifying it shouldn't affect original)
	queue[0] = "modified"
	newQueue := ctx.PeekQueue()
	assert.Equal(t, "command1", newQueue[0]) // Should still be original value
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
	ctx.SetVariable("normal", "value")

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
			expected: "}", // Inner ${normal} resolves to 'value', leaving outer ${value} to resolve to empty (missing var) = '}'
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
	os.Unsetenv("GOOS")
	os.Unsetenv("GOARCH")

	value, ok := ctx.getSystemVariable("@os")
	assert.True(t, ok)
	assert.Equal(t, "/", value) // Should be "/" when both are empty

	// Test with set environment
	os.Setenv("GOOS", "linux")
	os.Setenv("GOARCH", "amd64")

	value, ok = ctx.getSystemVariable("@os")
	assert.True(t, ok)
	assert.Equal(t, "linux/amd64", value)

	// Restore original values
	if origGOOS != "" {
		os.Setenv("GOOS", origGOOS)
	}
	if origGOARCH != "" {
		os.Setenv("GOARCH", origGOARCH)
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
	ctx.SetVariable("name", "Alice")
	text := "Hello ${name}, how are you today?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.InterpolateVariables(text)
	}
}

func BenchmarkInterpolateVariables_MultipleVariables(b *testing.B) {
	ctx := New()
	ctx.SetVariable("name", "Alice")
	ctx.SetVariable("age", "30")
	ctx.SetVariable("city", "New York")
	text := "${name} is ${age} years old and lives in ${city}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.InterpolateVariables(text)
	}
}

func BenchmarkGetVariable_UserVariable(b *testing.B) {
	ctx := New()
	ctx.SetVariable("test", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.GetVariable("test")
	}
}

func BenchmarkGetVariable_SystemVariable(b *testing.B) {
	ctx := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.GetVariable("#session_id")
	}
}

func BenchmarkQueueOperations(b *testing.B) {
	ctx := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.QueueCommand("test_command")
		ctx.DequeueCommand()
	}
}

func BenchmarkLargeVariableMap(b *testing.B) {
	ctx := New()

	// Setup large variable map
	for i := 0; i < 1000; i++ {
		ctx.SetVariable(fmt.Sprintf("var%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx.GetVariable("var500") // Access middle variable
	}
}
