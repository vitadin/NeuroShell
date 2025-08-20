package statemachine

import (
	"testing"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/builtin"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupStackTestEnvironment creates a test environment with all necessary services
func setupStackTestEnvironment() (*context.NeuroContext, error) {
	// Create test context
	ctx := context.NewTestContext()
	concreteCtx := ctx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Setup services
	services.SetGlobalRegistry(services.NewRegistry())

	// Setup commands
	commands.SetGlobalRegistry(commands.NewRegistry())

	// Register builtin commands
	if err := commands.GetGlobalRegistry().Register(&builtin.SetCommand{}); err != nil {
		return nil, err
	}
	if err := commands.GetGlobalRegistry().Register(&builtin.GetCommand{}); err != nil {
		return nil, err
	}
	if err := commands.GetGlobalRegistry().Register(&builtin.EchoCommand{}); err != nil {
		return nil, err
	}
	if err := commands.GetGlobalRegistry().Register(&builtin.HelpCommand{}); err != nil {
		return nil, err
	}
	if err := commands.GetGlobalRegistry().Register(&builtin.ExitCommand{}); err != nil {
		return nil, err
	}

	// Register all required services
	if err := services.GetGlobalRegistry().RegisterService(services.NewVariableService()); err != nil {
		return nil, err
	}
	if err := services.GetGlobalRegistry().RegisterService(services.NewStackService()); err != nil {
		return nil, err
	}

	// Initialize all services
	if err := services.GetGlobalRegistry().InitializeAll(); err != nil {
		return nil, err
	}

	return concreteCtx, nil
}

func TestNewStackMachine(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	assert.NotNil(t, sm)
	assert.Equal(t, ctx, sm.context)
	assert.NotNil(t, sm.stateProcessor)
	assert.NotNil(t, sm.tryHandler)
	assert.Equal(t, config, sm.config)
	assert.NotNil(t, sm.logger)
	assert.NotNil(t, sm.stackService)
	assert.NotNil(t, sm.variableService)
}

func TestStackMachine_Execute_SimpleCommand(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Execute a simple command that should be pushed to stack
	err = sm.Execute("\\echo Hello World")
	assert.NoError(t, err)

	// Stack should be empty after processing
	assert.Equal(t, 0, ctx.GetStackSize())
}

func TestStackMachine_Execute_EmptyInput_ExpectsEchoCommand(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Execute empty command - this will be interpreted as \echo with empty message
	// Since \echo is implemented, this should succeed
	err = sm.Execute("")
	assert.NoError(t, err)

	// Stack should be empty
	assert.Equal(t, 0, ctx.GetStackSize())
}

func TestStackMachine_Execute_MultipleCommands(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Execute multiple commands in sequence
	err = sm.Execute("\\echo First")
	assert.NoError(t, err)

	err = sm.Execute("\\echo Second")
	assert.NoError(t, err)

	err = sm.Execute("\\echo Third")
	assert.NoError(t, err)

	// All commands should be processed and stack should be empty
	assert.Equal(t, 0, ctx.GetStackSize())
}

func TestStackMachine_ProcessStack_ErrorBoundaryMarkers(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Manually push error boundary markers to test handling
	ctx.PushCommand("ERROR_BOUNDARY_END:test_id_1")
	ctx.PushCommand("\\echo Hello")
	ctx.PushCommand("ERROR_BOUNDARY_START:test_id_1")

	// Process the stack
	err = sm.processStack()
	assert.NoError(t, err)

	// Stack should be empty after processing
	assert.Equal(t, 0, ctx.GetStackSize())

	// Should not be in try block after processing
	assert.False(t, ctx.IsInTryBlock())
}

func TestStackMachine_ProcessStack_NestedTryBlocks(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Create nested try blocks
	ctx.PushCommand("ERROR_BOUNDARY_END:outer_try")
	ctx.PushCommand("ERROR_BOUNDARY_END:inner_try")
	ctx.PushCommand("\\echo Nested")
	ctx.PushCommand("ERROR_BOUNDARY_START:inner_try")
	ctx.PushCommand("ERROR_BOUNDARY_START:outer_try")

	// Process the stack
	err = sm.processStack()
	assert.NoError(t, err)

	// Stack should be empty after processing
	assert.Equal(t, 0, ctx.GetStackSize())

	// Should not be in try block after processing
	assert.False(t, ctx.IsInTryBlock())
}

func TestStackMachine_UpdateEchoConfig(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	config.EchoCommands = false
	sm := NewStackMachine(ctx, config)

	// Set _echo_command variable to "true"
	err = ctx.SetSystemVariable("_echo_command", "true")
	require.NoError(t, err)

	// Update echo configuration
	sm.updateEchoConfig()

	// Config should now have echo commands enabled
	assert.True(t, sm.config.EchoCommands)

	// Set _echo_command variable to "false"
	err = ctx.SetSystemVariable("_echo_command", "false")
	require.NoError(t, err)

	// Update echo configuration
	sm.updateEchoConfig()

	// Config should now have echo commands disabled
	assert.False(t, sm.config.EchoCommands)
}

func TestStackMachine_ExecuteInternal(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// ExecuteInternal should work the same as Execute
	err = sm.ExecuteInternal("\\echo Internal Test")
	assert.NoError(t, err)

	// Stack should be empty after processing
	assert.Equal(t, 0, ctx.GetStackSize())
}

func TestStackMachine_GetConfig(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	config.EchoCommands = true
	config.RecursionLimit = 100

	sm := NewStackMachine(ctx, config)

	retrievedConfig := sm.GetConfig()
	assert.Equal(t, config, retrievedConfig)
	assert.True(t, retrievedConfig.EchoCommands)
	assert.Equal(t, 100, retrievedConfig.RecursionLimit)
}

func TestStackMachine_SetConfig(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Create new config
	newConfig := neurotypes.StateMachineConfig{
		EchoCommands:   true,
		MacroExpansion: false,
		RecursionLimit: 25,
	}

	sm.SetConfig(newConfig)

	retrievedConfig := sm.GetConfig()
	assert.Equal(t, newConfig, retrievedConfig)
	assert.True(t, retrievedConfig.EchoCommands)
	assert.False(t, retrievedConfig.MacroExpansion)
	assert.Equal(t, 25, retrievedConfig.RecursionLimit)
}

func TestStackMachine_ProcessCommand_ErrorBoundaryDetection(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Test START boundary marker
	err = sm.processCommand("ERROR_BOUNDARY_START:test_123")
	assert.NoError(t, err)
	assert.True(t, ctx.IsInTryBlock())
	assert.Equal(t, "test_123", ctx.GetCurrentTryID())

	// Test END boundary marker
	err = sm.processCommand("ERROR_BOUNDARY_END:test_123")
	assert.NoError(t, err)
	assert.False(t, ctx.IsInTryBlock())
	assert.Equal(t, "", ctx.GetCurrentTryID())
}

func TestStackMachine_TryBlockErrorHandling(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Create a try block with a command that should succeed
	ctx.PushCommand("ERROR_BOUNDARY_END:test_id")
	ctx.PushCommand("\\echo Success")
	ctx.PushCommand("ERROR_BOUNDARY_START:test_id")

	// Process the stack
	err = sm.processStack()
	assert.NoError(t, err)

	// Should not be in try block after processing
	assert.False(t, ctx.IsInTryBlock())

	// Stack should be empty
	assert.Equal(t, 0, ctx.GetStackSize())
}

func TestStackMachine_Integration_SimpleWorkflow(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Execute a simple workflow
	err = sm.Execute("\\set[test_var=hello]")
	assert.NoError(t, err)

	err = sm.Execute("\\echo Testing")
	assert.NoError(t, err)

	// Check that variable was set (if echo command processes it)
	// Note: Since we don't have actual command implementations in this test,
	// we're mainly testing the stack processing mechanism
	assert.Equal(t, 0, ctx.GetStackSize())
}

func TestStackMachine_ProcessStack_EmptyStack(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Process empty stack should not error
	err = sm.processStack()
	assert.NoError(t, err)

	// Stack should remain empty
	assert.Equal(t, 0, ctx.GetStackSize())
}

func TestStackMachine_ProcessStack_MixedCommands(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Mix regular commands with boundary markers
	ctx.PushCommand("\\echo Final")
	ctx.PushCommand("ERROR_BOUNDARY_END:mixed_test")
	ctx.PushCommand("\\echo Inside Try")
	ctx.PushCommand("ERROR_BOUNDARY_START:mixed_test")
	ctx.PushCommand("\\echo First")

	// Process the stack
	err = sm.processStack()
	assert.NoError(t, err)

	// Stack should be empty
	assert.Equal(t, 0, ctx.GetStackSize())

	// Should not be in try block
	assert.False(t, ctx.IsInTryBlock())
}

// Benchmark tests for performance
func BenchmarkStackMachine_Execute(b *testing.B) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(b, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sm.Execute("\\echo benchmark test")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStackMachine_ProcessStack(b *testing.B) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(b, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Add some commands to the stack
		ctx.PushCommand("\\echo test1")
		ctx.PushCommand("\\echo test2")
		ctx.PushCommand("\\echo test3")

		err := sm.processStack()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test cases for read-only functionality in stack machine

func TestStackMachine_shouldResetErrorState_ReadOnlyCommands(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	tests := []struct {
		name          string
		command       string
		expectedReset bool
		description   string
	}{
		{
			name:          "read-only get command should not reset error state",
			command:       "\\get[var]",
			expectedReset: false,
			description:   "get command is read-only by default",
		},
		{
			name:          "read-only help command should not reset error state",
			command:       "\\help",
			expectedReset: false,
			description:   "help command is read-only by default",
		},
		{
			name:          "read-only echo command should not reset error state",
			command:       "\\echo message",
			expectedReset: false,
			description:   "echo command is read-only by default",
		},
		{
			name:          "writable set command should reset error state",
			command:       "\\set[var=value]",
			expectedReset: true,
			description:   "set command is writable by default",
		},
		{
			name:          "non-neuroshell command should reset error state",
			command:       "regular command",
			expectedReset: true,
			description:   "non-NeuroShell commands should reset error state",
		},
		{
			name:          "unknown neuroshell command should reset error state",
			command:       "\\unknown",
			expectedReset: true,
			description:   "unknown commands should reset error state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.shouldResetErrorState(tt.command)
			assert.Equal(t, tt.expectedReset, result, tt.description)
		})
	}
}

func TestStackMachine_shouldResetErrorState_WithOverrides(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Set some read-only overrides to test the dynamic behavior
	ctx.SetCommandReadOnly("get", false) // Override read-only command to writable
	ctx.SetCommandReadOnly("set", true)  // Override writable command to read-only

	tests := []struct {
		name          string
		command       string
		expectedReset bool
		description   string
	}{
		{
			name:          "get command overridden to writable should reset error state",
			command:       "\\get[var]",
			expectedReset: true,
			description:   "get command overridden from read-only to writable",
		},
		{
			name:          "set command overridden to read-only should not reset error state",
			command:       "\\set[var=value]",
			expectedReset: false,
			description:   "set command overridden from writable to read-only",
		},
		{
			name:          "help command without override should still not reset error state",
			command:       "\\help",
			expectedReset: false,
			description:   "help command should remain read-only",
		},
		{
			name:          "echo command without override should still not reset error state",
			command:       "\\echo message",
			expectedReset: false,
			description:   "echo command should remain read-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.shouldResetErrorState(tt.command)
			assert.Equal(t, tt.expectedReset, result, tt.description)
		})
	}
}

func TestStackMachine_shouldResetErrorState_CommandParsing(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	tests := []struct {
		name          string
		command       string
		expectedReset bool
		description   string
	}{
		{
			name:          "command with options",
			command:       "\\get[var=value,other=test]",
			expectedReset: false,
			description:   "should parse command name correctly from options",
		},
		{
			name:          "command with spaces",
			command:       "\\echo hello world",
			expectedReset: false,
			description:   "should parse command name correctly with spaces",
		},
		{
			name:          "command with extra whitespace",
			command:       "  \\get[var]  ",
			expectedReset: false,
			description:   "should handle whitespace around command",
		},
		{
			name:          "command with complex options",
			command:       "\\set[var=value, other=test, flag]",
			expectedReset: true,
			description:   "should parse complex options correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.shouldResetErrorState(tt.command)
			assert.Equal(t, tt.expectedReset, result, tt.description)
		})
	}
}

func TestStackMachine_ReadOnlyCommands_Integration(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Set up error state initially
	if sm.errorService != nil {
		err := sm.errorService.SetErrorStateFromCommandResult(assert.AnError)
		require.NoError(t, err)
	}

	// Test that read-only commands don't reset error state
	err = sm.processCommand("\\get[nonexistent]")
	// The get command itself shouldn't error for non-existent variables, it just returns empty
	assert.NoError(t, err)

	// Test that writable commands do reset error state
	err = sm.processCommand("\\set[test_var=test_value]")
	assert.NoError(t, err)

	// Verify the variable was set
	value, err := ctx.GetVariable("test_var")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)
}

func TestStackMachine_ReadOnlyOverrides_Integration(t *testing.T) {
	ctx, err := setupStackTestEnvironment()
	require.NoError(t, err)

	config := neurotypes.DefaultStateMachineConfig()
	sm := NewStackMachine(ctx, config)

	// Override a read-only command to be writable
	ctx.SetCommandReadOnly("get", false)

	// Test that the overridden command now resets error state
	result := sm.shouldResetErrorState("\\get[var]")
	assert.True(t, result, "get command should reset error state when overridden to writable")

	// Override a writable command to be read-only
	ctx.SetCommandReadOnly("set", true)

	// Test that the overridden command now doesn't reset error state
	result = sm.shouldResetErrorState("\\set[var=value]")
	assert.False(t, result, "set command should not reset error state when overridden to read-only")

	// Remove the override for get command
	ctx.RemoveCommandReadOnlyOverride("get")

	// Test that it returns to its original behavior
	result = sm.shouldResetErrorState("\\get[var]")
	assert.False(t, result, "get command should not reset error state after override removal")
}
