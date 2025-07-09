package services

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.findWordStart(tt.line, tt.pos)
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
	err = ctx.SetSystemVariable("_status", "0")
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
			name:             "complete user variables",
			prefix:           "${u",
			expectedContains: []string{"${username}"},
			expectedCount:    1,
		},
		{
			name:             "complete system variables",
			prefix:           "${_s",
			expectedContains: []string{"${_status}"},
			expectedCount:    -1, // don't check count, as there may be other _s variables
		},
		{
			name:             "no matches",
			prefix:           "${xyz",
			expectedContains: []string{},
			expectedCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getVariableCompletions(tt.prefix)
			if tt.expectedCount >= 0 {
				assert.Equal(t, tt.expectedCount, len(result))
			}
			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected)
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
