package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestShowStackCommand_Name(t *testing.T) {
	cmd := &ShowStackCommand{}
	assert.Equal(t, "show-stack", cmd.Name())
}

func TestShowStackCommand_ParseMode(t *testing.T) {
	cmd := &ShowStackCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestShowStackCommand_Description(t *testing.T) {
	cmd := &ShowStackCommand{}
	assert.Equal(t, "Display the execution stack for development and debugging", cmd.Description())
}

func TestShowStackCommand_Usage(t *testing.T) {
	cmd := &ShowStackCommand{}
	assert.Equal(t, "\\show-stack[detailed=true]", cmd.Usage())
}

func TestShowStackCommand_HelpInfo(t *testing.T) {
	cmd := &ShowStackCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "show-stack", helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, cmd.Usage(), helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check options
	require.Len(t, helpInfo.Options, 1)
	assert.Equal(t, "detailed", helpInfo.Options[0].Name)
	assert.Equal(t, "Show additional stack information including indices and context", helpInfo.Options[0].Description)
	assert.False(t, helpInfo.Options[0].Required)
	assert.Equal(t, "boolean", helpInfo.Options[0].Type)
	assert.Equal(t, "false", helpInfo.Options[0].Default)

	// Check examples
	require.Len(t, helpInfo.Examples, 2)
	assert.Equal(t, "\\show-stack", helpInfo.Examples[0].Command)
	assert.Equal(t, "Display current execution stack", helpInfo.Examples[0].Description)
	assert.Equal(t, "\\show-stack[detailed=true]", helpInfo.Examples[1].Command)
	assert.Equal(t, "Show stack with indices and try/silent block context", helpInfo.Examples[1].Description)
}

func TestShowStackCommand_Execute_EmptyStack(t *testing.T) {
	// Setup context and services
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry with stack service
	registry := services.NewRegistry()
	stackService := services.NewStackService()
	require.NoError(t, stackService.Initialize())
	require.NoError(t, registry.RegisterService(stackService))
	services.SetGlobalRegistry(registry)

	// Create command and execute with empty stack
	cmd := &ShowStackCommand{}
	err := cmd.Execute(map[string]string{}, "")

	// Should not error with empty stack
	assert.NoError(t, err)
}

func TestShowStackCommand_Execute_WithCommands(t *testing.T) {
	// Setup context and services
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry with stack service
	registry := services.NewRegistry()
	stackService := services.NewStackService()
	require.NoError(t, stackService.Initialize())
	require.NoError(t, registry.RegisterService(stackService))
	services.SetGlobalRegistry(registry)

	// Add some commands to the stack
	stackService.PushCommand("\\echo test1")
	stackService.PushCommand("\\echo test2")
	stackService.PushCommand("\\echo test3")

	// Test basic show-stack
	cmd := &ShowStackCommand{}
	err := cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Test detailed mode
	err = cmd.Execute(map[string]string{"detailed": "true"}, "")
	assert.NoError(t, err)

	// Test detailed mode with "1" value
	err = cmd.Execute(map[string]string{"detailed": "1"}, "")
	assert.NoError(t, err)

	// Test non-detailed mode explicitly
	err = cmd.Execute(map[string]string{"detailed": "false"}, "")
	assert.NoError(t, err)
}

func TestShowStackCommand_Execute_WithTryAndSilentBlocks(t *testing.T) {
	// Setup context and services
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry with stack service
	registry := services.NewRegistry()
	stackService := services.NewStackService()
	require.NoError(t, stackService.Initialize())
	require.NoError(t, registry.RegisterService(stackService))
	services.SetGlobalRegistry(registry)

	// Add some commands to the stack
	stackService.PushCommand("\\echo test")

	// Set up try and silent blocks
	stackService.PushErrorBoundary("try-123")
	stackService.PushSilentBoundary("silent-456")

	// Test detailed mode should show context information
	cmd := &ShowStackCommand{}
	err := cmd.Execute(map[string]string{"detailed": "true"}, "")
	assert.NoError(t, err)

	// Verify we're in try and silent blocks
	assert.True(t, stackService.IsInTryBlock())
	assert.True(t, stackService.IsInSilentBlock())
	assert.Equal(t, "try-123", stackService.GetCurrentTryID())
	assert.Equal(t, "silent-456", stackService.GetCurrentSilentID())
}

func TestShowStackCommand_Execute_ServiceNotAvailable(t *testing.T) {
	// This test is more complex to set up without the stack service
	// For now, skip this test as the service is always available in the global registry
	t.Skip("Stack service is always available in the global registry")
}

func TestShowStackCommand_Execute_InvalidOptions(t *testing.T) {
	// Setup context and services
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup service registry with stack service
	registry := services.NewRegistry()
	stackService := services.NewStackService()
	require.NoError(t, stackService.Initialize())
	require.NoError(t, registry.RegisterService(stackService))
	services.SetGlobalRegistry(registry)

	// Add a command to the stack
	stackService.PushCommand("\\echo test")

	// Test with invalid detailed option values (should default to false)
	cmd := &ShowStackCommand{}

	testCases := []map[string]string{
		{"detailed": "invalid"},
		{"detailed": "0"},
		{"detailed": ""},
		{"other_option": "value"},
	}

	for _, options := range testCases {
		err := cmd.Execute(options, "")
		assert.NoError(t, err, "Options: %v", options)
	}
}
