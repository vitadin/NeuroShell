package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// MockCommand for testing
type MockCommand struct {
	name        string
	description string
	usage       string
	options     []neurotypes.HelpOption
}

func (m *MockCommand) Name() string {
	return m.name
}

func (m *MockCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

func (m *MockCommand) Description() string {
	return m.description
}

func (m *MockCommand) Usage() string {
	return m.usage
}

func (m *MockCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     m.name,
		Description: m.description,
		Usage:       m.usage,
		Options:     m.options,
	}
}

func (m *MockCommand) Execute(_ map[string]string, _ string) error {
	return nil
}

func TestAutoCompleteService_Name(t *testing.T) {
	service := NewAutoCompleteService()
	assert.Equal(t, "autocomplete", service.Name())
}

func TestAutoCompleteService_Initialize(t *testing.T) {
	service := NewAutoCompleteService()
	assert.False(t, service.initialized)

	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
}

func TestAutoCompleteService_FindWordStart(t *testing.T) {
	service := NewAutoCompleteService()

	tests := []struct {
		name     string
		line     string
		pos      int
		expected int
	}{
		{
			name:     "start of line",
			line:     "\\send hello",
			pos:      5,
			expected: 0,
		},
		{
			name:     "after space",
			line:     "\\send hello",
			pos:      11,
			expected: 6,
		},
		{
			name:     "after bracket",
			line:     "\\set[name=test]",
			pos:      10,
			expected: 10,
		},
		{
			name:     "after equals",
			line:     "\\set[name=test]",
			pos:      15,
			expected: 15,
		},
		{
			name:     "variable reference - partial name",
			line:     "${abc",
			pos:      5,
			expected: 0, // Should start from $
		},
		{
			name:     "variable reference - in command",
			line:     "\\send ${user",
			pos:      12,
			expected: 6, // Should start from $
		},
		{
			name:     "variable reference - in brackets",
			line:     "\\set[msg=\"${greet\"]",
			pos:      17,
			expected: 10, // Should start from $
		},
		{
			name:     "variable reference - complete",
			line:     "${username}",
			pos:      11,
			expected: 0, // Should start from $
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.findWordStart(tt.line, tt.pos)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoCompleteService_IsInVariableReference(t *testing.T) {
	service := NewAutoCompleteService()

	tests := []struct {
		name     string
		line     string
		pos      int
		expected bool
	}{
		{
			name:     "inside variable reference - partial",
			line:     "${abc",
			pos:      5,
			expected: true,
		},
		{
			name:     "inside variable reference - complete",
			line:     "${username}",
			pos:      10,
			expected: true,
		},
		{
			name:     "outside variable reference - after closing",
			line:     "${username} hello",
			pos:      15,
			expected: false,
		},
		{
			name:     "inside variable reference - in command",
			line:     "\\send ${user",
			pos:      12,
			expected: true,
		},
		{
			name:     "inside variable reference - in brackets",
			line:     "\\set[msg=\"${greet\"]",
			pos:      17,
			expected: true,
		},
		{
			name:     "not in variable reference",
			line:     "\\send hello",
			pos:      8,
			expected: false,
		},
		{
			name:     "multiple variables - in second",
			line:     "${first} ${second",
			pos:      17,
			expected: true,
		},
		{
			name:     "multiple variables - after first",
			line:     "${first} ${second}",
			pos:      9,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isInVariableReference(tt.line, tt.pos)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoCompleteService_IsInsideBrackets(t *testing.T) {
	service := NewAutoCompleteService()

	tests := []struct {
		name     string
		line     string
		pos      int
		expected bool
	}{
		{
			name:     "inside brackets",
			line:     "\\set[name=test]",
			pos:      10,
			expected: true,
		},
		{
			name:     "outside brackets",
			line:     "\\set[name=test] hello",
			pos:      18,
			expected: false,
		},
		{
			name:     "before brackets",
			line:     "\\set[name=test]",
			pos:      4,
			expected: false,
		},
		{
			name:     "no brackets",
			line:     "\\send hello",
			pos:      8,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isInsideBrackets(tt.line, tt.pos)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoCompleteService_GetCommandCompletions(t *testing.T) {
	// Set up test environment
	setupAutoCompleteTestRegistry(t)

	service := NewAutoCompleteService()
	err := service.Initialize()
	require.NoError(t, err)

	// Register test commands via context
	ctx := context.GetGlobalContext()
	require.NotNil(t, ctx)

	// Cast to concrete type to access RegisterCommandWithInfo
	neuroCtx, ok := ctx.(*context.NeuroContext)
	require.True(t, ok)

	// Register test commands with context
	neuroCtx.RegisterCommandWithInfo(&MockCommand{name: "send", description: "Send message"})
	neuroCtx.RegisterCommandWithInfo(&MockCommand{name: "set", description: "Set variable"})
	neuroCtx.RegisterCommandWithInfo(&MockCommand{name: "session", description: "Session management"})

	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:     "complete all commands",
			prefix:   "\\",
			expected: []string{"\\send", "\\session", "\\set"},
		},
		{
			name:     "complete se commands",
			prefix:   "\\se",
			expected: []string{"\\send", "\\session", "\\set"},
		},
		{
			name:     "complete exact match",
			prefix:   "\\send",
			expected: []string{"\\send"},
		},
		{
			name:     "no matches",
			prefix:   "\\xyz",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getCommandCompletions(tt.prefix)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoCompleteService_GetOptionCompletions(t *testing.T) {
	// Set up test environment
	setupAutoCompleteTestRegistry(t)

	service := NewAutoCompleteService()
	err := service.Initialize()
	require.NoError(t, err)

	// Register test command with options via context
	ctx := context.GetGlobalContext()
	require.NotNil(t, ctx)

	// Cast to concrete type to access RegisterCommandWithInfo
	neuroCtx, ok := ctx.(*context.NeuroContext)
	require.True(t, ok)

	neuroCtx.RegisterCommandWithInfo(&MockCommand{
		name:        "set",
		description: "Set variable",
		options: []neurotypes.HelpOption{
			{Name: "name", Type: "string", Description: "Variable name"},
			{Name: "value", Type: "string", Description: "Variable value"},
			{Name: "global", Type: "bool", Description: "Global variable"},
		},
	})

	tests := []struct {
		name        string
		line        string
		pos         int
		currentWord string
		expected    []string
	}{
		{
			name:        "complete option names",
			line:        "\\set[n",
			pos:         7,
			currentWord: "n",
			expected:    []string{"name="},
		},
		{
			name:        "complete boolean option",
			line:        "\\set[g",
			pos:         7,
			currentWord: "g",
			expected:    []string{"global"},
		},
		{
			name:        "complete all options",
			line:        "\\set[",
			pos:         6,
			currentWord: "",
			expected:    []string{"global", "name=", "value="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getOptionCompletions(tt.line, tt.pos, tt.currentWord)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoCompleteService_GetVariableCompletions(t *testing.T) {
	// Set up test environment
	setupAutoCompleteTestRegistry(t)

	service := NewAutoCompleteService()
	err := service.Initialize()
	require.NoError(t, err)

	// Set up test context with variables
	ctx := context.New()
	context.SetGlobalContext(ctx)

	// Add some test variables
	err = ctx.SetVariable("username", "john")
	require.NoError(t, err)
	err = ctx.SetVariable("project", "neuroshell")
	require.NoError(t, err)
	err = ctx.SetVariable("abc", "test")
	require.NoError(t, err)
	err = ctx.SetVariable("abcd", "test2")
	require.NoError(t, err)
	err = ctx.SetSystemVariable("_status", "0")
	require.NoError(t, err)
	err = ctx.SetSystemVariable("@user", "testuser")
	require.NoError(t, err)

	// Register variable service
	err = GetGlobalRegistry().RegisterService(NewVariableService())
	require.NoError(t, err)
	err = GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	tests := []struct {
		name             string
		prefix           string
		expectedContains []string
		expectedCount    int
	}{
		{
			name:             "complete user variables starting with 'u'",
			prefix:           "${u",
			expectedContains: []string{"${username}"},
			expectedCount:    1,
		},
		{
			name:             "complete system variables starting with '_s'",
			prefix:           "${_s",
			expectedContains: []string{"${_status}"},
			expectedCount:    -1, // don't check count, as there may be other _s variables
		},
		{
			name:             "complete system variables starting with '@u'",
			prefix:           "${@u",
			expectedContains: []string{"${@user}"},
			expectedCount:    -1, // don't check count, as there may be other @u variables
		},
		{
			name:             "complete partial 'a' - multiple matches",
			prefix:           "${a",
			expectedContains: []string{"${abc}", "${abcd}"},
			expectedCount:    -1, // don't check exact count, focus on contains
		},
		{
			name:             "complete exact match 'abc'",
			prefix:           "${abc",
			expectedContains: []string{"${abc}", "${abcd}"},
			expectedCount:    -1,
		},
		{
			name:             "no matches for 'xyz'",
			prefix:           "${xyz",
			expectedContains: []string{},
			expectedCount:    0,
		},
		{
			name:             "complete in command context",
			prefix:           "\\send ${u",
			expectedContains: []string{"\\send ${username}"},
			expectedCount:    -1,
		},
		{
			name:             "complete in bracket context",
			prefix:           "\\set[msg=\"${u",
			expectedContains: []string{"\\set[msg=\"${username}"},
			expectedCount:    -1,
		},
		{
			name:             "complete just '${' - should match all variables",
			prefix:           "${",
			expectedContains: []string{"${username}", "${project}", "${abc}", "${abcd}"},
			expectedCount:    -1, // many system variables may exist too
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getVariableCompletions(tt.prefix)
			if tt.expectedCount >= 0 {
				assert.Equal(t, tt.expectedCount, len(result))
			}
			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "Expected completion '%s' not found in results: %v", expected, result)
			}
		})
	}
}

func TestAutoCompleteService_Do(t *testing.T) {
	// Set up test environment
	setupAutoCompleteTestRegistry(t)

	service := NewAutoCompleteService()
	err := service.Initialize()
	require.NoError(t, err)

	// Register test commands via context
	ctx := context.GetGlobalContext()
	require.NotNil(t, ctx)

	// Cast to concrete type to access RegisterCommandWithInfo
	neuroCtx, ok := ctx.(*context.NeuroContext)
	require.True(t, ok)

	neuroCtx.RegisterCommandWithInfo(&MockCommand{name: "send", description: "Send message"})
	neuroCtx.RegisterCommandWithInfo(&MockCommand{name: "set", description: "Set variable"})

	tests := []struct {
		name           string
		line           string
		pos            int
		expectedCount  int
		expectedSuffix string
	}{
		{
			name:           "complete command",
			line:           "\\se",
			pos:            3,
			expectedCount:  2,    // "\\send" and "\\set"
			expectedSuffix: "nd", // completion for "\\send"
		},
		{
			name:          "no completion needed",
			line:          "\\send hello",
			pos:           11,
			expectedCount: 0,
		},
		{
			name:          "uninitialized service",
			line:          "\\se",
			pos:           3,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "uninitialized service" {
				service.initialized = false
			}

			suggestions, offset := service.Do([]rune(tt.line), tt.pos)
			assert.Equal(t, tt.expectedCount, len(suggestions))

			if tt.expectedCount > 0 {
				assert.Greater(t, offset, 0)
				if tt.expectedSuffix != "" {
					found := false
					for _, suggestion := range suggestions {
						if string(suggestion) == tt.expectedSuffix {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected suffix not found in suggestions")
				}
			}
		})
	}
}

func TestAutoCompleteService_Do_VariableCompletion(t *testing.T) {
	// Set up test environment
	setupAutoCompleteTestRegistry(t)

	service := NewAutoCompleteService()
	err := service.Initialize()
	require.NoError(t, err)

	// Set up test context with variables
	ctx := context.New()
	context.SetGlobalContext(ctx)

	// Add some test variables
	err = ctx.SetVariable("abc", "test")
	require.NoError(t, err)
	err = ctx.SetVariable("abcd", "test2")
	require.NoError(t, err)
	err = ctx.SetVariable("username", "john")
	require.NoError(t, err)

	// Register variable service
	err = GetGlobalRegistry().RegisterService(NewVariableService())
	require.NoError(t, err)
	err = GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	tests := []struct {
		name                string
		line                string
		pos                 int
		expectedCount       int
		expectedContainment string // Check if result contains this string
	}{
		{
			name:                "complete variable - basic case",
			line:                "${a",
			pos:                 3,
			expectedCount:       2,    // abc and abcd
			expectedContainment: "bc", // Should suggest "bc}" to complete "${abc}"
		},
		{
			name:                "complete variable - in command",
			line:                "\\send ${u",
			pos:                 9,
			expectedCount:       1,
			expectedContainment: "sername}", // Should suggest "sername}" to complete "${username}"
		},
		{
			name:                "complete variable - exact match",
			line:                "${abc",
			pos:                 5,
			expectedCount:       2,   // abc and abcd both match
			expectedContainment: "}", // Should suggest "}" to complete "${abc}"
		},
		{
			name:          "no variable completion outside reference",
			line:          "hello world",
			pos:           5,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, offset := service.Do([]rune(tt.line), tt.pos)

			if tt.expectedCount > 0 {
				assert.Equal(t, tt.expectedCount, len(suggestions), "Unexpected number of suggestions: %v", suggestions)
				assert.Greater(t, offset, 0, "Expected positive offset")

				// Check if any suggestion contains the expected string
				if tt.expectedContainment != "" {
					found := false
					for _, suggestion := range suggestions {
						if strings.Contains(string(suggestion), tt.expectedContainment) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected containment '%s' not found in suggestions: %v", tt.expectedContainment, suggestions)
				}
			} else {
				assert.Equal(t, tt.expectedCount, len(suggestions), "Expected no suggestions but got: %v", suggestions)
			}
		})
	}
}

// TestAutoCompleteService_Integration_VariableCompletion provides an integration test
// that demonstrates the complete variable completion functionality working end-to-end.
func TestAutoCompleteService_Integration_VariableCompletion(t *testing.T) {
	// Set up test environment
	setupAutoCompleteTestRegistry(t)

	service := NewAutoCompleteService()
	err := service.Initialize()
	require.NoError(t, err)

	// Set up realistic test variables that match the use case in the issue
	ctx := context.New()
	context.SetGlobalContext(ctx)

	err = ctx.SetVariable("abc", "test-value")
	require.NoError(t, err)
	err = ctx.SetVariable("user", "john")
	require.NoError(t, err)
	err = ctx.SetVariable("greeting", "hello")
	require.NoError(t, err)

	// Register variable service
	err = GetGlobalRegistry().RegisterService(NewVariableService())
	require.NoError(t, err)
	err = GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Test case 1: User types "${a" and presses TAB - should suggest "${abc}"
	t.Run("user_types_${a_expects_abc_completion", func(t *testing.T) {
		line := "${a"
		pos := 3
		suggestions, offset := service.Do([]rune(line), pos)

		// Should get one suggestion
		require.Len(t, suggestions, 1, "Expected exactly one suggestion for '${a'")

		// The suggestion should complete to "${abc}"
		suggestion := string(suggestions[0])
		expectedCompletion := "bc}"
		assert.Equal(t, expectedCompletion, suggestion, "Expected completion to be 'bc}' to complete '${abc}'")
		assert.Equal(t, 3, offset, "Expected offset to be 3 (length of '${a')")
	})

	// Test case 2: User types "${" and presses TAB - should suggest all variables
	t.Run("user_types_${_expects_all_variables", func(t *testing.T) {
		line := "${"
		pos := 2
		suggestions, offset := service.Do([]rune(line), pos)

		// Should get multiple suggestions (at least our 3 user variables)
		assert.GreaterOrEqual(t, len(suggestions), 3, "Expected at least 3 suggestions for '${'")
		assert.Equal(t, 2, offset, "Expected offset to be 2 (length of '${')")

		// Should contain our expected variables
		suggestionStrings := make([]string, len(suggestions))
		for i, s := range suggestions {
			suggestionStrings[i] = string(s)
		}

		// Check for completions that would result in complete variable references
		expectedCompletions := []string{"abc}", "user}", "greeting}"}
		for _, expected := range expectedCompletions {
			found := false
			for _, actual := range suggestionStrings {
				if actual == expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Expected completion '%s' not found in suggestions: %v", expected, suggestionStrings)
		}
	})

	// Test case 3: Variable completion in command context
	t.Run("variable_completion_in_command_context", func(t *testing.T) {
		line := "\\send ${u"
		pos := 9
		suggestions, offset := service.Do([]rune(line), pos)

		// Should get one suggestion for 'user'
		require.Len(t, suggestions, 1, "Expected exactly one suggestion for '\\send ${u'")
		assert.Equal(t, 3, offset, "Expected offset to be 3 (length of '${u')")

		suggestion := string(suggestions[0])
		expectedCompletion := "ser}"
		assert.Equal(t, expectedCompletion, suggestion, "Expected completion to be 'ser}' to complete '${user}'")
	})
}

// setupAutoCompleteTestRegistry creates a clean test environment
func setupAutoCompleteTestRegistry(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := GetGlobalRegistry()
	SetGlobalRegistry(NewRegistry())

	// Reset the global context to clean state
	context.ResetGlobalContext()

	// Cleanup function to restore original registries
	t.Cleanup(func() {
		SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}
