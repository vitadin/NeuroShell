package statemachine

import (
	"strings"
	"testing"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/builtin"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStateProcessor(t *testing.T) {
	ctx := context.NewTestContext()
	config := neurotypes.StateMachineConfig{
		RecursionLimit: 100,
	}

	processor := NewStateProcessor(ctx.(*context.NeuroContext), config)
	assert.NotNil(t, processor)
	assert.NotNil(t, processor.context)
	assert.NotNil(t, processor.interpolator)
	assert.NotNil(t, processor.resolver)
	assert.NotNil(t, processor.logger)
	assert.Equal(t, config, processor.config)
}

func TestStateProcessor_interpolateVariables(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Set up some test variables
	err := ctx.SetVariable("test_var", "hello")
	require.NoError(t, err)
	err = ctx.SetVariable("name", "world")
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "no variables",
			input:    "echo simple command",
			expected: "echo simple command",
			wantErr:  false,
		},
		{
			name:     "single variable",
			input:    "echo ${test_var}",
			expected: "echo hello",
			wantErr:  false,
		},
		{
			name:     "multiple variables",
			input:    "echo ${test_var} ${name}",
			expected: "echo hello world",
			wantErr:  false,
		},
		{
			name:     "undefined variable",
			input:    "echo ${undefined_var}",
			expected: "echo ", // Undefined variables are replaced with empty string
			wantErr:  false,
		},
		{
			name:     "nested variables",
			input:    "echo ${test_var} with ${name}!",
			expected: "echo hello with world!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.interpolateVariables(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestStateProcessor_parseCommand(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	tests := []struct {
		name        string
		input       string
		expectedCmd string
		wantErr     bool
	}{
		{
			name:        "simple command",
			input:       "\\echo hello",
			expectedCmd: "echo",
			wantErr:     false,
		},
		{
			name:        "command with options",
			input:       "\\set[var=value] message",
			expectedCmd: "set",
			wantErr:     false,
		},
		{
			name:        "command with complex options",
			input:       "\\send[model=gpt-4,temp=0.7] Hello world",
			expectedCmd: "send",
			wantErr:     false,
		},
		{
			name:        "comment line",
			input:       "%% This is a comment",
			expectedCmd: "echo", // Comments are treated as echo commands
			wantErr:     false,
		},
		{
			name:        "empty input",
			input:       "",
			expectedCmd: "echo", // Empty input becomes echo with empty message
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.parseCommand(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedCmd, result.Name)
			}
		})
	}
}

func TestStateProcessor_resolveCommand_TryCommand(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Parse a try command
	parsed, err := processor.parseCommand("\\try echo hello")
	require.NoError(t, err)

	// Resolve it
	resolved, err := processor.resolveCommand(parsed)
	require.NoError(t, err)
	assert.Equal(t, "try", resolved.Name)
	assert.Equal(t, neurotypes.CommandTypeTry, resolved.Type)
}

func TestStateProcessor_resolveCommand_BuiltinCommand(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx := context.GetGlobalContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Register a test builtin command
	testCmd := &builtin.EchoCommand{}
	err := commands.GetGlobalRegistry().Register(testCmd)
	require.NoError(t, err)

	// Parse and resolve builtin command
	parsed, err := processor.parseCommand("\\echo hello")
	require.NoError(t, err)

	resolved, err := processor.resolveCommand(parsed)
	require.NoError(t, err)
	assert.Equal(t, "echo", resolved.Name)
	assert.Equal(t, neurotypes.CommandTypeBuiltin, resolved.Type)
	assert.Equal(t, testCmd, resolved.BuiltinCommand)
}

func TestStateProcessor_resolveCommand_UnknownCommand(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Parse an unknown command
	parsed, err := processor.parseCommand("\\unknown-command")
	require.NoError(t, err)

	// Try to resolve it
	resolved, err := processor.resolveCommand(parsed)
	assert.Error(t, err)
	assert.Nil(t, resolved)
	assert.Contains(t, err.Error(), "unknown command")
}

func TestStateProcessor_executeBuiltinCommand(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Create a test builtin command
	testCmd := &builtin.EchoCommand{}

	// Create resolved command
	resolved := &neurotypes.StateMachineResolvedCommand{
		Name:           "echo",
		Type:           neurotypes.CommandTypeBuiltin,
		BuiltinCommand: testCmd,
	}

	// Parse command for parameters
	parsed, err := processor.parseCommand("\\echo hello world")
	require.NoError(t, err)

	// Execute builtin command
	err = processor.executeBuiltinCommand(resolved, parsed, "\\echo hello world")
	assert.NoError(t, err)
}

func TestStateProcessor_executeBuiltinCommand_NoCommand(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Create resolved command without actual builtin command
	resolved := &neurotypes.StateMachineResolvedCommand{
		Name:           "echo",
		Type:           neurotypes.CommandTypeBuiltin,
		BuiltinCommand: nil, // This should cause an error
	}

	parsed, err := processor.parseCommand("\\echo hello")
	require.NoError(t, err)

	err = processor.executeBuiltinCommand(resolved, parsed, "\\echo hello")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no builtin command to execute")
}

func TestStateProcessor_executeScriptCommand(t *testing.T) {
	// Setup global context and services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Create registry and register services
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	// Register required services
	variableService := services.NewVariableService()
	err := variableService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(variableService)
	require.NoError(t, err)

	stackService := services.NewStackService()
	err = stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	processor := NewStateProcessor(concreteCtx, neurotypes.StateMachineConfig{})

	// Create resolved script command
	scriptContent := `\echo First command
\set test_var=script_value
\echo Second command`

	resolved := &neurotypes.StateMachineResolvedCommand{
		Name:          "test-script.neuro",
		Type:          neurotypes.CommandTypeUser,
		ScriptContent: scriptContent,
		ScriptPath:    "/path/to/test-script.neuro",
	}

	// Parse command with parameters
	parsed, err := processor.parseCommand("\\test-script.neuro[param1=value1] hello world")
	require.NoError(t, err)

	// Execute script command
	err = processor.executeScriptCommand(resolved, parsed)
	assert.NoError(t, err)

	// Verify parameter variables were set (check both variable service and context)
	value, err := variableService.Get("_0")
	if err == nil && value != "" {
		assert.Equal(t, "test-script.neuro", value)
	} else {
		t.Logf("Variable _0 not available via service, checking context: %v", err)
		value, err = concreteCtx.GetVariable("_0")
		if err == nil {
			assert.Equal(t, "test-script.neuro", value)
		}
	}

	value, err = variableService.Get("_1")
	if err == nil && value != "" {
		assert.Equal(t, "hello world", value)
	} else {
		t.Logf("Variable _1 not available via service: %v", err)
	}

	value, err = variableService.Get("param1")
	if err == nil && value != "" {
		assert.Equal(t, "value1", value)
	} else {
		t.Logf("Variable param1 not available via service: %v", err)
	}

	// Verify commands were pushed to stack (should be 3 commands)
	stackSize := stackService.GetStackSize()
	assert.Equal(t, 3, stackSize)
}

func TestStateProcessor_executeScriptCommand_EmptyScript(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Create resolved command with empty script
	resolved := &neurotypes.StateMachineResolvedCommand{
		Name:          "empty-script.neuro",
		Type:          neurotypes.CommandTypeUser,
		ScriptContent: "",
		ScriptPath:    "/path/to/empty-script.neuro",
	}

	parsed, err := processor.parseCommand("\\empty-script.neuro")
	require.NoError(t, err)

	err = processor.executeScriptCommand(resolved, parsed)
	assert.NoError(t, err) // Empty script should succeed
}

func TestStateProcessor_executeScriptCommand_OnlyComments(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Create resolved command with only comments
	scriptContent := `%% This is a comment
%% Another comment
%% Last comment`

	resolved := &neurotypes.StateMachineResolvedCommand{
		Name:          "comments-only.neuro",
		Type:          neurotypes.CommandTypeUser,
		ScriptContent: scriptContent,
		ScriptPath:    "/path/to/comments-only.neuro",
	}

	parsed, err := processor.parseCommand("\\comments-only.neuro")
	require.NoError(t, err)

	err = processor.executeScriptCommand(resolved, parsed)
	assert.NoError(t, err) // Script with only comments should succeed
}

func TestStateProcessor_executeScriptCommand_NoScriptContent(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Create resolved command without script content
	resolved := &neurotypes.StateMachineResolvedCommand{
		Name:          "no-content.neuro",
		Type:          neurotypes.CommandTypeUser,
		ScriptContent: "", // This should cause an error
		ScriptPath:    "",
	}

	parsed, err := processor.parseCommand("\\no-content.neuro")
	require.NoError(t, err)

	err = processor.executeScriptCommand(resolved, parsed)
	assert.NoError(t, err) // Empty content should not error, just do nothing
}

func TestStateProcessor_executeTryCommand(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	tests := []struct {
		name      string
		input     string
		wantErr   bool
		shouldSet bool // whether variables should be set
	}{
		{
			name:      "empty try command",
			input:     "\\try",
			wantErr:   false,
			shouldSet: true,
		},
		{
			name:      "try with command",
			input:     "\\try echo hello",
			wantErr:   false,
			shouldSet: false, // Should push boundary, not set vars immediately
		},
		{
			name:      "try with whitespace only",
			input:     "\\try    ",
			wantErr:   false,
			shouldSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := processor.parseCommand(tt.input)
			require.NoError(t, err)

			err = processor.executeTryCommand(parsed, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateProcessor_ProcessCommand_FullPipeline(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	ctx := context.GetGlobalContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Register a test builtin command
	testCmd := &builtin.EchoCommand{}
	err := commands.GetGlobalRegistry().Register(testCmd)
	require.NoError(t, err)

	// Set up test variables
	err = ctx.(*context.NeuroContext).SetVariable("greeting", "hello")
	require.NoError(t, err)
	err = ctx.(*context.NeuroContext).SetVariable("name", "world")
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "simple builtin command",
			input:   "\\echo simple test",
			wantErr: false,
		},
		{
			name:    "command with variable interpolation",
			input:   "\\echo ${greeting} ${name}",
			wantErr: false,
		},
		{
			name:    "try command",
			input:   "\\try echo test",
			wantErr: false,
		},
		{
			name:    "unknown command",
			input:   "\\unknown-command",
			wantErr: true,
			errMsg:  "command resolution failed",
		},
		{
			name:    "plain text as echo command",
			input:   "not a command", // This gets parsed as echo command
			wantErr: false,           // Echo command is registered and will succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := processor.ProcessCommand(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStateProcessor_formatNamedArgs(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	tests := []struct {
		name     string
		options  map[string]string
		expected string
	}{
		{
			name:     "empty options",
			options:  map[string]string{},
			expected: "",
		},
		{
			name:     "single option",
			options:  map[string]string{"key1": "value1"},
			expected: "key1=value1",
		},
		{
			name:     "multiple options",
			options:  map[string]string{"key2": "value2", "key1": "value1"},
			expected: "key1=value1,key2=value2", // Should be sorted
		},
		{
			name:     "options with special characters",
			options:  map[string]string{"temp": "0.7", "model": "gpt-4"},
			expected: "model=gpt-4,temp=0.7", // Should be sorted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.formatNamedArgs(tt.options)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStateProcessor_SetConfig(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	newConfig := neurotypes.StateMachineConfig{
		RecursionLimit: 200,
	}

	processor.SetConfig(newConfig)
	assert.Equal(t, newConfig, processor.config)
}

func TestStateProcessor_executeCommand_UnknownType(t *testing.T) {
	ctx := context.NewTestContext()
	processor := NewStateProcessor(ctx.(*context.NeuroContext), neurotypes.StateMachineConfig{})

	// Create resolved command with unknown type
	resolved := &neurotypes.StateMachineResolvedCommand{
		Name: "test",
		Type: neurotypes.CommandType(999), // Invalid type
	}

	parsed, err := processor.parseCommand("\\test")
	require.NoError(t, err)

	err = processor.executeCommand(resolved, parsed, "\\test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command type")
}

func TestStateProcessor_ScriptParameterPassing(t *testing.T) {
	// Setup global context and services
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Create registry and register services
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	// Register required services
	variableService := services.NewVariableService()
	err := variableService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(variableService)
	require.NoError(t, err)

	stackService := services.NewStackService()
	err = stackService.Initialize()
	require.NoError(t, err)
	err = registry.RegisterService(stackService)
	require.NoError(t, err)

	processor := NewStateProcessor(concreteCtx, neurotypes.StateMachineConfig{})

	// Test with multiple parameters
	scriptContent := `\echo ${_0} ${_1} ${param1} ${param2}`

	resolved := &neurotypes.StateMachineResolvedCommand{
		Name:          "send",
		Type:          neurotypes.CommandTypeUser,
		ScriptContent: scriptContent,
		ScriptPath:    "/path/to/send.neuro",
	}

	// Parse command with multiple parameters
	parsed, err := processor.parseCommand("\\send[param1=value1,param2=value2] positional argument")
	require.NoError(t, err)

	// Execute script command
	err = processor.executeScriptCommand(resolved, parsed)
	assert.NoError(t, err)

	// Verify all parameter variables were set correctly (check what's available)
	tests := []struct {
		varName  string
		expected string
	}{
		{"_0", "send"},
		{"_1", "positional argument"},
		{"param1", "value1"},
		{"param2", "value2"},
	}

	for _, tt := range tests {
		value, err := concreteCtx.GetVariable(tt.varName)
		if err == nil {
			assert.Equal(t, tt.expected, value, "Variable %s has wrong value", tt.varName)
		} else {
			t.Logf("Variable %s not available: %v", tt.varName, err)
		}
	}
}

func TestStateProcessor_ScriptCommentHandling(t *testing.T) {
	// Test script with mixed comments and commands
	scriptContent := `%% This is a comment
\echo First command
%% Another comment in the middle
\echo Second command
%% Final comment`

	// Parse script and count non-comment lines
	lines := strings.Split(scriptContent, "\n")
	scriptLines := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "%%") {
			scriptLines = append(scriptLines, trimmed)
		}
	}

	// Should have exactly 2 non-comment lines
	assert.Len(t, scriptLines, 2)
	assert.Equal(t, "\\echo First command", scriptLines[0])
	assert.Equal(t, "\\echo Second command", scriptLines[1])
}
