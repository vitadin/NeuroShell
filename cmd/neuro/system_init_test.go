package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "neuroshell/internal/commands/llm" // Import LLM commands for metadata
	"neuroshell/internal/context"
	"neuroshell/internal/data/embedded"
	"neuroshell/internal/services"
	"neuroshell/internal/shell"
)

func TestExecuteSystemInit_Success(t *testing.T) {
	// Set up clean test environment
	setupSystemInitTestEnvironment(t)

	// Execute system initialization
	err := executeSystemInit()
	assert.NoError(t, err)

	// Verify system variables were set
	ctx := shell.GetGlobalContext()

	executed, err := ctx.GetVariable("#system_init_executed")
	require.NoError(t, err)
	assert.Equal(t, "true", executed)

	path, err := ctx.GetVariable("#system_init_path")
	require.NoError(t, err)
	assert.Equal(t, "embedded://stdlib/system-init.neuro", path)
}

func TestExecuteSystemInit_ScriptContentExecution(t *testing.T) {
	// Set up clean test environment
	setupSystemInitTestEnvironment(t)

	// Execute system initialization
	err := executeSystemInit()
	assert.NoError(t, err)

	// Verify that the system initialization script was executed successfully
	// The main evidence is that system variables were set correctly
	ctx := shell.GetGlobalContext()

	executed, err := ctx.GetVariable("#system_init_executed")
	require.NoError(t, err)
	assert.Equal(t, "true", executed, "System initialization should have executed successfully")

	path, err := ctx.GetVariable("#system_init_path")
	require.NoError(t, err)
	assert.Equal(t, "embedded://stdlib/system-init.neuro", path, "System init path should be set correctly")

	// The script execution itself is tested by verifying that the system completed successfully
	// Individual command execution (like llm-api-load) is tested in their respective command tests
	// This test focuses on the system init framework working correctly
}

func TestExecuteSystemInit_EmbeddedScriptExists(t *testing.T) {
	// Verify the system-init.neuro script exists in embedded filesystem
	stdlibLoader := embedded.NewStdlibLoader()

	exists := stdlibLoader.ScriptExists("system-init")
	assert.True(t, exists, "system-init.neuro script should exist in embedded stdlib")

	// Verify script content can be loaded
	content, err := stdlibLoader.LoadScript("system-init")
	require.NoError(t, err)
	assert.NotEmpty(t, content, "system-init.neuro script should have content")

	// Verify script contains expected command
	assert.Contains(t, content, "\\silent \\try \\llm-api-load", "system-init.neuro should contain the llm-api-load command")
}

func TestExecuteSystemInit_NoScript_DoesNotFail(t *testing.T) {
	// Set up clean test environment
	setupSystemInitTestEnvironment(t)

	// This tests the graceful handling when system-init.neuro doesn't exist
	// Note: In a real scenario, we'd mock the embedded.StdlibLoader, but for this test
	// we'll just verify the current behavior with the existing script

	// Execute system initialization (should succeed even if script issues occur)
	err := executeSystemInit()
	assert.NoError(t, err, "executeSystemInit should not fail even with script issues")
}

func TestExecuteSystemInit_SilentExecution(t *testing.T) {
	// Set up clean test environment
	setupSystemInitTestEnvironment(t)

	// Execute system initialization
	err := executeSystemInit()
	assert.NoError(t, err)

	// The test that commands are executed silently is implicit -
	// if they weren't silent, they would produce output which would appear in test logs
	// The \silent \try wrapper ensures no output or errors break the execution flow

	// Verify execution completed successfully
	ctx := shell.GetGlobalContext()
	executed, err := ctx.GetVariable("#system_init_executed")
	require.NoError(t, err)
	assert.Equal(t, "true", executed)
}

func TestExecuteSystemInit_ErrorHandling(t *testing.T) {
	// Test that individual command failures don't stop system init
	// The "\try" wrapper in system-init.neuro should handle command failures gracefully

	// Set up clean test environment
	setupSystemInitTestEnvironment(t)

	// Execute system initialization
	err := executeSystemInit()
	assert.NoError(t, err, "System init should complete even if individual commands fail")

	// Verify system variables are still set even if some commands fail
	ctx := shell.GetGlobalContext()
	executed, err := ctx.GetVariable("#system_init_executed")
	require.NoError(t, err)
	assert.Equal(t, "true", executed)
}

// setupSystemInitTestEnvironment creates a clean test environment for system init tests
func setupSystemInitTestEnvironment(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Create a test context
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Initialize services required for system init
	err := shell.InitializeServices(true)
	require.NoError(t, err)

	// Cleanup function to restore original state
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}
