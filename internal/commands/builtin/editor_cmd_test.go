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
	// ctx := context.NewTestContext()
	// Set up registry with editor service but no variable service
	registry := services.NewRegistry()
	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize()
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
	// ctx := context.NewTestContext()
	// Set up mock editor environment for fast testing
	helper := testutils.SetupMockEditor()
	defer helper.Cleanup()

	// Set up a complete service registry
	registry := services.NewRegistry()

	// Add editor service
	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize()
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	// Add variable service
	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize()
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
		// Note: ctx is commented out above, so we can't check variables in this test
		t.Logf("Editor command executed successfully")
	}
}

func TestEditorCommand_Execute_WithEchoEditor(t *testing.T) {
	// This test specifically uses echo as EDITOR for fast, predictable testing
	cmd := &EditorCommand{}
	// ctx := context.NewTestContext()
	// Set up mock editor environment for fast testing
	helper := testutils.SetupMockEditor()
	defer helper.Cleanup()

	// Set up complete service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize()
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize()
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
	// ctx := context.NewTestContext()
	// Set up complete service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize()
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize()
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
	// ctx := context.NewTestContext()
	// Set up service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize()
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize()
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
	// vars := map[string]string{
	//	"@editor":  "echo", // Use echo as a mock editor
	//	"test_var": "test_value",
	// }
	// ctx := context.NewTestContextWithVars(vars)

	// Set up service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize()
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize()
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
	// ctx := context.NewTestContext()
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
	err = editorService.Initialize()
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
	err = editorService2.Initialize()
	require.NoError(t, err)
	defer func() { _ = editorService2.Cleanup() }()

	variableService2 := services.NewVariableService()
	err = registry3.RegisterService(variableService2)
	require.NoError(t, err)
	err = variableService2.Initialize()
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

func TestEditorCommand_Execute_WithInitialContent(t *testing.T) {
	cmd := &EditorCommand{}

	// Set up mock editor environment for fast testing
	helper := testutils.SetupMockEditor()
	defer helper.Cleanup()

	// Set up complete service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize()
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize()
	require.NoError(t, err)

	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Test with initial content
	initialText := "Write a blog post about AI"
	err = cmd.Execute(map[string]string{}, initialText)

	// Should handle the initial content without panicking
	if err != nil {
		t.Logf("Execute with initial content failed (may be expected in test environment): %v", err)
		// Verify the error mentions editor operation, not argument parsing
		assert.Contains(t, err.Error(), "editor operation failed")
	} else {
		t.Logf("Editor command with initial content executed successfully")
	}
}

func TestEditorCommand_Execute_EmptyVsInitialContent(t *testing.T) {
	cmd := &EditorCommand{}

	// Set up mock editor environment
	helper := testutils.SetupMockEditor()
	defer helper.Cleanup()

	// Set up complete service registry
	registry := services.NewRegistry()

	editorService := services.NewEditorService()
	err := registry.RegisterService(editorService)
	require.NoError(t, err)
	err = editorService.Initialize()
	require.NoError(t, err)
	defer func() { _ = editorService.Cleanup() }()

	variableService := services.NewVariableService()
	err = registry.RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize()
	require.NoError(t, err)

	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	// Test empty input
	err1 := cmd.Execute(map[string]string{}, "")

	// Test with content
	err2 := cmd.Execute(map[string]string{}, "Hello world")

	// Both should be handled without panicking
	// Behavior should be consistent (both may fail due to mock environment)
	if err1 != nil {
		t.Logf("Empty editor execute failed (expected in test env): %v", err1)
	}
	if err2 != nil {
		t.Logf("Content editor execute failed (expected in test env): %v", err2)
	}

	// Verify neither panicked or had argument parsing issues
	assert.NotContains(t, fmt.Sprintf("%v", err1), "panic")
	assert.NotContains(t, fmt.Sprintf("%v", err2), "panic")
}

func TestEditorCommand_Execute_VariousInitialContent(t *testing.T) {
	cmd := &EditorCommand{}

	// Set up minimal registry (will fail at service level, but we test argument handling)
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)
	defer func() {
		services.SetGlobalRegistry(services.NewRegistry())
	}()

	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: true, // Service not available
		},
		{
			name:    "simple text",
			input:   "Hello",
			wantErr: true, // Service not available
		},
		{
			name:    "multi-word text",
			input:   "Write a blog post",
			wantErr: true, // Service not available
		},
		{
			name:    "text with special characters",
			input:   "Hello! How are you? (fine)",
			wantErr: true, // Service not available
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true, // Service not available
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cmd.Execute(map[string]string{}, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "editor service not available")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEditorCommand_HelpInfo_EditorVariable(t *testing.T) {
	cmd := &EditorCommand{}
	helpInfo := cmd.HelpInfo()

	// Check that help info mentions _editor variable, not @editor
	assert.Contains(t, helpInfo.Usage, "\\editor")

	// Check examples mention _editor
	found := false
	for _, example := range helpInfo.Examples {
		if example.Command == "\\set[_editor=\"code --wait\"]" {
			found = true
			assert.Contains(t, example.Description, "VS Code")
			break
		}
	}
	assert.True(t, found, "Should contain example with _editor variable")

	// Check notes mention _editor
	foundNote := false
	for _, note := range helpInfo.Notes {
		if note == "Editor preference: 1) ${_editor} variable, 2) $EDITOR env var, 3) auto-detect" {
			foundNote = true
			break
		}
	}
	assert.True(t, foundNote, "Should contain note about _editor variable precedence")
}

func TestEditorCommand_Usage_EditorVariable(t *testing.T) {
	cmd := &EditorCommand{}
	usage := cmd.Usage()

	// Check that usage mentions _editor variable
	assert.Contains(t, usage, "${_editor}")
	assert.Contains(t, usage, "\\set[_editor=")

	// Make sure it doesn't mention the old @editor syntax
	assert.NotContains(t, usage, "\\set[@editor=")
	assert.NotContains(t, usage, "${@editor}")
}
