package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"neuroshell/internal/context"
)

func TestNewQueueService(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)
	assert.NotNil(t, service)
	assert.Equal(t, "queue", service.Name())
}

func TestQueueService_Initialize(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	assert.NoError(t, err)
}

func TestQueueService_QueueCommand(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test queuing a single command
	service.QueueCommand("set var1=value1")

	// Verify command was queued
	commands := concreteCtx.PeekQueue()
	assert.Len(t, commands, 1)
	assert.Equal(t, "set var1=value1", commands[0])
}

func TestQueueService_QueueCommands(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test queuing multiple commands
	commands := []string{
		"set var1=value1",
		"set var2=value2",
		"get var1",
	}

	service.QueueCommands(commands)

	// Verify all commands were queued
	queuedCommands := concreteCtx.PeekQueue()
	assert.Len(t, queuedCommands, 3)
	assert.Equal(t, commands, queuedCommands)
}

func TestQueueService_GetQueueSize(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test empty queue
	size := service.GetQueueSize()
	assert.Equal(t, 0, size)

	// Add commands
	service.QueueCommand("set var1=value1")
	service.QueueCommand("set var2=value2")

	// Test non-empty queue
	size = service.GetQueueSize()
	assert.Equal(t, 2, size)
}

func TestQueueService_ClearQueue(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Add commands
	service.QueueCommand("set var1=value1")
	service.QueueCommand("set var2=value2")

	// Verify queue has commands
	assert.Equal(t, 2, service.GetQueueSize())

	// Clear queue
	service.ClearQueue()

	// Verify queue is empty
	assert.Equal(t, 0, service.GetQueueSize())
}

func TestQueueService_DequeueCommand(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Add commands
	service.QueueCommands([]string{"cmd1", "cmd2", "cmd3"})

	// Dequeue commands
	cmd1, ok := service.DequeueCommand()
	assert.True(t, ok)
	assert.Equal(t, "cmd1", cmd1)

	cmd2, ok := service.DequeueCommand()
	assert.True(t, ok)
	assert.Equal(t, "cmd2", cmd2)

	// Verify remaining queue size
	assert.Equal(t, 1, service.GetQueueSize())
}

func TestQueueService_PeekQueue(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test empty queue
	commands := service.PeekQueue()
	assert.Empty(t, commands)

	// Add commands
	service.QueueCommands([]string{"cmd1", "cmd2", "cmd3"})

	// Test peek
	commands = service.PeekQueue()
	assert.Len(t, commands, 3)
	assert.Equal(t, []string{"cmd1", "cmd2", "cmd3"}, commands)

	// Verify queue wasn't modified (peek should not dequeue)
	assert.Equal(t, 3, service.GetQueueSize())
}

func TestQueueService_EmptyQueueOperations(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test dequeue from empty queue
	cmd, ok := service.DequeueCommand()
	assert.False(t, ok)
	assert.Empty(t, cmd)

	// Test peek empty queue
	commands := service.PeekQueue()
	assert.Empty(t, commands)
}

func TestQueueService_LargeQueue(t *testing.T) {
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	service := NewQueueService(concreteCtx)

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test with large number of commands
	largeCommands := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeCommands[i] = "set var" + string(rune('a'+i%26)) + "=value" + string(rune('0'+i%10))
	}

	service.QueueCommands(largeCommands)
	assert.Equal(t, 1000, service.GetQueueSize())

	// Test dequeue all
	for i := 0; i < 1000; i++ {
		cmd, ok := service.DequeueCommand()
		assert.True(t, ok)
		assert.Equal(t, largeCommands[i], cmd)
	}

	assert.Equal(t, 0, service.GetQueueSize())
}
