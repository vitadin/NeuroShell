package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"neuroshell/internal/context"
)

func TestNewStackService(t *testing.T) {
	service := NewStackService()
	assert.NotNil(t, service)
	assert.Equal(t, "stack", service.Name())
}

func TestStackService_Initialize(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	assert.NoError(t, err)
}

func TestStackService_PushCommand(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test pushing a single command
	service.PushCommand("set var1=value1")

	// Verify command was pushed
	commands := concreteCtx.PeekStack()
	assert.Len(t, commands, 1)
	assert.Equal(t, "set var1=value1", commands[0])
}

func TestStackService_PushCommands(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test pushing multiple commands
	commands := []string{
		"set var1=value1",
		"set var2=value2",
		"get var1",
	}

	service.PushCommands(commands)

	// Verify all commands were pushed (LIFO order, so last pushed is first)
	stackedCommands := concreteCtx.PeekStack()
	assert.Len(t, stackedCommands, 3)
	// Due to LIFO behavior, the commands will be in reverse order
	expectedOrder := []string{"get var1", "set var2=value2", "set var1=value1"}
	assert.Equal(t, expectedOrder, stackedCommands)
}

func TestStackService_GetStackSize(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test empty stack
	size := service.GetStackSize()
	assert.Equal(t, 0, size)

	// Add commands
	service.PushCommand("set var1=value1")
	service.PushCommand("set var2=value2")

	// Test non-empty stack
	size = service.GetStackSize()
	assert.Equal(t, 2, size)
}

func TestStackService_ClearStack(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Add commands
	service.PushCommand("set var1=value1")
	service.PushCommand("set var2=value2")

	// Verify stack has commands
	assert.Equal(t, 2, service.GetStackSize())

	// Clear stack
	service.ClearStack()

	// Verify stack is empty
	assert.Equal(t, 0, service.GetStackSize())
}

func TestStackService_PopCommand(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Add commands (LIFO: last in, first out)
	service.PushCommands([]string{"cmd1", "cmd2", "cmd3"})

	// Pop commands (should come out in reverse order due to LIFO)
	cmd1, ok := service.PopCommand()
	assert.True(t, ok)
	assert.Equal(t, "cmd3", cmd1) // Last pushed, first popped

	cmd2, ok := service.PopCommand()
	assert.True(t, ok)
	assert.Equal(t, "cmd2", cmd2)

	// Verify remaining stack size
	assert.Equal(t, 1, service.GetStackSize())
}

func TestStackService_PeekStack(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test empty stack
	commands := service.PeekStack()
	assert.Empty(t, commands)

	// Add commands
	service.PushCommands([]string{"cmd1", "cmd2", "cmd3"})

	// Test peek (LIFO order)
	commands = service.PeekStack()
	assert.Len(t, commands, 3)
	expectedOrder := []string{"cmd3", "cmd2", "cmd1"} // Reverse order due to LIFO
	assert.Equal(t, expectedOrder, commands)

	// Verify stack wasn't modified (peek should not pop)
	assert.Equal(t, 3, service.GetStackSize())
}

func TestStackService_EmptyStackOperations(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test pop from empty stack
	cmd, ok := service.PopCommand()
	assert.False(t, ok)
	assert.Empty(t, cmd)

	// Test peek empty stack
	commands := service.PeekStack()
	assert.Empty(t, commands)
}

func TestStackService_LargeStack(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test with large number of commands
	largeCommands := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeCommands[i] = "set var" + string(rune('a'+i%26)) + "=value" + string(rune('0'+i%10))
	}

	service.PushCommands(largeCommands)
	assert.Equal(t, 1000, service.GetStackSize())

	// Test pop all (should come out in reverse order due to LIFO)
	for i := 999; i >= 0; i-- {
		cmd, ok := service.PopCommand()
		assert.True(t, ok)
		assert.Equal(t, largeCommands[i], cmd)
	}

	assert.Equal(t, 0, service.GetStackSize())
}

func TestStackService_MaxStackDepthProtection(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Set a low stack limit for testing
	err := concreteCtx.SetSystemVariable("_max_stack_depth", "5")
	require.NoError(t, err)

	service := NewStackService()

	// Initialize service
	err = service.Initialize()
	require.NoError(t, err)

	// Test single command push with stack limit
	// Fill stack to just under limit
	for i := 0; i < 4; i++ {
		service.PushCommand("cmd" + string(rune('0'+i)))
	}
	assert.Equal(t, 4, service.GetStackSize())

	// This should work (at limit)
	service.PushCommand("cmd4")
	assert.Equal(t, 5, service.GetStackSize())

	// This should be prevented (exceeds limit)
	service.PushCommand("cmd5")
	assert.Equal(t, 5, service.GetStackSize()) // Should not increase
}

func TestStackService_MaxStackDepthProtection_PushCommands(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Set a low stack limit for testing
	err := concreteCtx.SetSystemVariable("_max_stack_depth", "3")
	require.NoError(t, err)

	service := NewStackService()

	// Initialize service
	err = service.Initialize()
	require.NoError(t, err)

	// Test multiple commands push with stack limit
	// Add one command first
	service.PushCommand("cmd0")
	assert.Equal(t, 1, service.GetStackSize())

	// Try to add 3 more commands (would exceed limit of 3)
	commands := []string{"cmd1", "cmd2", "cmd3"}
	service.PushCommands(commands)

	// Should still be 1 (the batch was rejected)
	assert.Equal(t, 1, service.GetStackSize())

	// Try to add 2 commands (should work - total would be 3)
	commands = []string{"cmd1", "cmd2"}
	service.PushCommands(commands)

	// Should now be 3
	assert.Equal(t, 3, service.GetStackSize())
}

func TestStackService_DefaultMaxStackDepth(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Don't set _max_stack_depth, should use default (1000)
	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test that default limit is 1000
	maxDepth := service.getMaxStackDepth(concreteCtx)
	assert.Equal(t, 1000, maxDepth)

	// Test we can add many commands (up to default limit)
	for i := 0; i < 999; i++ {
		service.PushCommand("cmd" + string(rune('0'+i%10)))
	}
	assert.Equal(t, 999, service.GetStackSize())

	// This should work (at limit)
	service.PushCommand("cmd999")
	assert.Equal(t, 1000, service.GetStackSize())

	// This should be prevented (exceeds limit)
	service.PushCommand("cmd1000")
	assert.Equal(t, 1000, service.GetStackSize()) // Should not increase
}

func TestStackService_InvalidMaxStackDepth(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test invalid values fall back to default
	testCases := []string{"", "invalid", "-1", "0", "abc123"}

	for _, testCase := range testCases {
		err := concreteCtx.SetSystemVariable("_max_stack_depth", testCase)
		require.NoError(t, err)

		maxDepth := service.getMaxStackDepth(concreteCtx)
		assert.Equal(t, 1000, maxDepth, "Invalid value '%s' should fall back to default 1000", testCase)
	}
}

func TestStackService_UserConfigurableMaxDepth(t *testing.T) {
	// Setup global context
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	service := NewStackService()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test various valid user configurations
	testCases := []struct {
		value    string
		expected int
	}{
		{"1", 1},
		{"10", 10},
		{"100", 100},
		{"2000", 2000},
		{"50", 50},
	}

	for _, testCase := range testCases {
		err := concreteCtx.SetSystemVariable("_max_stack_depth", testCase.value)
		require.NoError(t, err)

		maxDepth := service.getMaxStackDepth(concreteCtx)
		assert.Equal(t, testCase.expected, maxDepth, "Value '%s' should result in max depth %d", testCase.value, testCase.expected)
	}
}
