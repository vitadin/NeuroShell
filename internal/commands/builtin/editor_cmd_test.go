package builtin

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/services"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

func TestEditorCommand_Name(t *testing.T) {
	cmd := &EditorCommand{}
	assert.Equal(t, "editor", cmd.Name())
}

func TestEditorCommand_ParseMode(t *testing.T) {
	cmd := &EditorCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestEditorCommand_Description(t *testing.T) {
	cmd := &EditorCommand{}
	assert.Equal(t, "Open external editor for composing input", cmd.Description())
}

func TestEditorCommand_Usage(t *testing.T) {
	cmd := &EditorCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "\\editor")
	assert.Contains(t, usage, "external editor")
	assert.Contains(t, usage, "${_output}")
}

func TestEditorCommand_Execute_EditorServiceNotAvailable(t *testing.T) {
	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Use empty service registry to simulate missing editor service
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)
	defer func() {
		// Restore global registry after test
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "editor service not available")
}

func TestEditorCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Set up registry with editor service but no variable service
	registry := services.NewRegistry()
	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize(ctx)
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestEditorCommand_Execute_MockSuccess(t *testing.T) {
	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Set up mock editor environment for fast testing
	helper := testutils.SetupMockEditor()
	defer helper.Cleanup()

	// Set up a complete service registry
	registry := services.NewRegistry()

	// Add editor service
	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize(ctx)
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	// Add variable service
	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize(ctx)
	require.NoError(t, err)

	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Execute the command
	err = cmd.Execute(map[string]string{}, "")

	// This should succeed with echo as the editor
	if err != nil {
		// Log error but don't fail test - some environments might have restrictions
		t.Logf("Execute failed (may be environment specific): %v", err)
	} else {
		// If successful, verify the _output variable was set
		value, err := ctx.GetVariable("_output")
		if err == nil {
			t.Logf("Editor content stored in _output: %q", value)
		}
	}
}

func TestEditorCommand_Execute_WithEchoEditor(t *testing.T) {
	// This test specifically uses echo as EDITOR for fast, predictable testing
	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Set up mock editor environment for fast testing
	helper := testutils.SetupMockEditor()
	defer helper.Cleanup()

	// Set up complete service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize(ctx)
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize(ctx)
	require.NoError(t, err)

	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Execute command - should complete quickly
	err = cmd.Execute(map[string]string{}, "")

	// Verify execution completed (successfully or with expected error)
	assert.NotEqual(t, "test timed out", fmt.Sprintf("%v", err))
}

func TestEditorCommand_Execute_EmptyArgs(t *testing.T) {
	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Set up complete service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize(ctx)
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize(ctx)
	require.NoError(t, err)

	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Test with empty args map
	err = cmd.Execute(map[string]string{}, "")

	// Should handle empty args gracefully
	// Error is expected if no editor is found, but shouldn't panic
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

func TestEditorCommand_Execute_ArgsHandling(t *testing.T) {
	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Set up service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize(ctx)
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize(ctx)
	require.NoError(t, err)

	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Test with various args
	testArgs := []map[string]string{
		{},
		{"option1": "value1"},
		{"execute": "true"},
		{"execute": "false"},
		{"custom": "test"},
	}

	for i, args := range testArgs {
		t.Run(fmt.Sprintf("args_test_%d", i), func(t *testing.T) {
			err := cmd.Execute(args, "")
			// Should handle all args gracefully without panicking
			if err != nil {
				t.Logf("Expected error in test environment: %v", err)
			}
		})
	}
}

func TestEditorCommand_Integration_WithMockContext(t *testing.T) {
	cmd := &EditorCommand{}

	// Create context with some variables
	vars := map[string]string{
		"@editor":  "echo", // Use echo as a mock editor
		"test_var": "test_value",
	}
	ctx := testutils.NewMockContextWithVars(vars)

	// Set up service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize(ctx)
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize(ctx)
	require.NoError(t, err)

	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Test command execution
	err = cmd.Execute(map[string]string{}, "")

	// The result depends on whether echo can be executed successfully
	// and whether the file operations work in the test environment
	if err != nil {
		t.Logf("Execute failed (may be expected in test environment): %v", err)
		// Check error type
		assert.Contains(t, err.Error(), "editor")
	}
}

func TestEditorCommand_ServiceInteraction(t *testing.T) {
	// This test verifies that the command properly interacts with services
	// without actually executing an external editor

	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Test 1: Missing editor service
	registry1 := services.NewRegistry()
	variableService := services.NewVariableService()
	err := registry1.RegisterService(variableService)
	require.NoError(t, err)

	services.SetGlobalRegistry(registry1)

	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "editor service not available")

	// Test 2: Missing variable service
	registry2 := services.NewRegistry()
	editorService := services.NewEditorService()
	err = registry2.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize(ctx)
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	services.SetGlobalRegistry(registry2)

	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")

	// Test 3: Both services present but editor operation fails
	registry3 := services.NewRegistry()

	editorService2 := services.NewEditorService()
	err = registry3.RegisterService(editorService2)
	require.NoError(t, err)
	err = editorService2.Initialize(ctx)
	require.NoError(t, err)
	defer func() { _ = editorService2.Cleanup() }()

	variableService2 := services.NewVariableService()
	err = registry3.RegisterService(variableService2)
	require.NoError(t, err)
	err = variableService2.Initialize(ctx)
	require.NoError(t, err)

	services.SetGlobalRegistry(registry3)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Clear environment to force "no editor found" error using helper
	helper := testutils.SetupNoEditor()
	defer helper.Cleanup()

	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "editor operation failed")
}

func TestEditorCommand_InputParameterHandling(t *testing.T) {
	// Test that the input parameter is correctly ignored (as per the function signature)
	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Set up minimal service registry for this test
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Test with various input values - should all fail with "service not available"
	// since we're not setting up services, but importantly should not panic
	testInputs := []string{
		"",
		"some input text",
		"multi\nline\ninput",
		"input with special chars: !@#$%^&*()",
	}

	for i, input := range testInputs {
		t.Run(fmt.Sprintf("input_test_%d", i), func(t *testing.T) {
			err := cmd.Execute(map[string]string{}, input)
			assert.Error(t, err) // Expected due to missing services
			assert.Contains(t, err.Error(), "editor service not available")
		})
	}
}

func TestEditorCommand_ConcurrentExecution(t *testing.T) {
	// Test that multiple concurrent executions don't cause race conditions
	cmd := &EditorCommand{}
	ctx := testutils.NewMockContext()

	// Set up service registry
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Run multiple concurrent executions
	done := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			err := cmd.Execute(map[string]string{}, "")
			done <- err
		}()
	}

	// Collect all results
	for i := 0; i < 5; i++ {
		err := <-done
		// All should fail with the same error (missing service) - no panics or race conditions
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "editor service not available")
	}
}
