package builtin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// TestRunCommand_Name tests that the command returns the correct name.
func TestRunCommand_Name(t *testing.T) {
	cmd := &RunCommand{}
	if cmd.Name() != "run" {
		t.Errorf("Expected name 'run', got '%s'", cmd.Name())
	}
}

// TestRunCommand_ParseMode tests that the command returns the correct parse mode.
func TestRunCommand_ParseMode(t *testing.T) {
	cmd := &RunCommand{}
	expected := neurotypes.ParseModeRaw
	if cmd.ParseMode() != expected {
		t.Errorf("Expected parse mode %v, got %v", expected, cmd.ParseMode())
	}
}

// TestRunCommand_Description tests that the command returns a non-empty description.
func TestRunCommand_Description(t *testing.T) {
	cmd := &RunCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}
	if !strings.Contains(desc, "script") {
		t.Error("Expected description to mention 'script'")
	}
}

// TestRunCommand_Usage tests that the command returns comprehensive usage information.
func TestRunCommand_Usage(t *testing.T) {
	cmd := &RunCommand{}
	usage := cmd.Usage()
	if usage == "" {
		t.Error("Expected non-empty usage")
	}
	if !strings.Contains(usage, "\\run") {
		t.Error("Expected usage to contain '\\run'")
	}
	if !strings.Contains(usage, "script_path") {
		t.Error("Expected usage to contain 'script_path'")
	}
}

// TestRunCommand_HelpInfo tests that the command returns structured help information.
func TestRunCommand_HelpInfo(t *testing.T) {
	cmd := &RunCommand{}
	helpInfo := cmd.HelpInfo()

	if helpInfo.Command != "run" {
		t.Errorf("Expected command name 'run', got '%s'", helpInfo.Command)
	}

	if len(helpInfo.Examples) == 0 {
		t.Error("Expected examples to be provided")
	}

	if len(helpInfo.Notes) == 0 {
		t.Error("Expected notes to be provided")
	}
}

// TestRunCommand_Execute_EmptyInput tests execution with empty script path.
func TestRunCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &RunCommand{}
	err := cmd.Execute(map[string]string{}, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script path is required")
}

// Helper function to setup test registry with stack service
func setupStackTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register stack service
	err := services.GetGlobalRegistry().RegisterService(services.NewStackService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

// TestRunCommand_Execute_WhitespaceInput tests execution with whitespace-only input.
func TestRunCommand_Execute_WhitespaceInput(t *testing.T) {
	cmd := &RunCommand{}
	err := cmd.Execute(map[string]string{}, "   \t\n  ")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script path is required")
}

// TestRunCommand_Execute_ValidScript tests execution with a valid script file.
func TestRunCommand_Execute_ValidScript(t *testing.T) {
	// Setup test environment
	ctx := context.NewTestContext()
	setupStackTestRegistry(t, ctx)

	// Create a temporary test script
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "test_script.neuro")
	scriptContent := `%% Test script for \run command
\echo Hello from test script
\set[test_run_var="success"]
\get[test_run_var]`

	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(t, err, "Failed to create test script")

	// Get stack service
	stackService, err := services.GetGlobalStackService()
	require.NoError(t, err, "Stack service not available")

	// Clear any existing commands in the stack
	stackService.ClearStack()

	cmd := &RunCommand{}
	err = cmd.Execute(map[string]string{}, scriptPath)
	assert.NoError(t, err)

	// Verify that the script path was pushed to the stack service
	assert.False(t, stackService.IsEmpty(), "Expected script path to be pushed to stack service")

	// Pop the command and verify it's the script path
	poppedCommand, hasMore := stackService.PopCommand()
	assert.True(t, hasMore, "Expected command to be available in stack")
	assert.Equal(t, "\\"+scriptPath, poppedCommand, "Expected script path with backslash prefix in stack")
}

// TestRunCommand_Execute_RelativePath tests execution with relative path.
func TestRunCommand_Execute_RelativePath(t *testing.T) {
	// Setup test environment
	ctx := context.NewTestContext()
	setupStackTestRegistry(t, ctx)

	// Create a temporary test script in current directory
	scriptContent := `%% Relative path test script
\echo Testing relative path execution`

	scriptName := "relative_test.neuro"
	err := os.WriteFile(scriptName, []byte(scriptContent), 0644)
	require.NoError(t, err, "Failed to create test script")
	defer func() {
		_ = os.Remove(scriptName) // Clean up, ignore error in test cleanup
	}()

	// Get stack service
	stackService, err := services.GetGlobalStackService()
	require.NoError(t, err, "Stack service not available")
	stackService.ClearStack()

	cmd := &RunCommand{}
	err = cmd.Execute(map[string]string{}, scriptName)
	assert.NoError(t, err)

	// Verify command was pushed to stack
	assert.False(t, stackService.IsEmpty(), "Expected script path to be pushed to stack service")
}

// TestRunCommand_Execute_NonexistentScript tests execution with nonexistent script.
func TestRunCommand_Execute_NonexistentScript(t *testing.T) {
	// Setup test environment
	ctx := context.NewTestContext()
	setupStackTestRegistry(t, ctx)

	cmd := &RunCommand{}
	err := cmd.Execute(map[string]string{}, "nonexistent_script.neuro")

	// Note: Since we're just pushing to the stack service, the RunCommand itself
	// won't validate file existence - that's handled by the state machine resolver.
	// The RunCommand should succeed in pushing the path to the stack.
	assert.NoError(t, err, "RunCommand should succeed in pushing path to stack, even if file doesn't exist")

	// Verify command was still pushed to stack (validation happens later)
	stackService, err := services.GetGlobalStackService()
	require.NoError(t, err, "Stack service not available")

	assert.False(t, stackService.IsEmpty(), "Expected script path to be pushed to stack service")
}

// TestRunCommand_Execute_MultipleScripts tests pushing multiple scripts to stack.
func TestRunCommand_Execute_MultipleScripts(t *testing.T) {
	// Setup test environment
	ctx := context.NewTestContext()
	setupStackTestRegistry(t, ctx)

	// Create multiple test scripts
	tempDir := t.TempDir()
	script1 := filepath.Join(tempDir, "script1.neuro")
	script2 := filepath.Join(tempDir, "script2.neuro")

	err := os.WriteFile(script1, []byte(`\echo First script`), 0644)
	require.NoError(t, err, "Failed to create first test script")

	err = os.WriteFile(script2, []byte(`\echo Second script`), 0644)
	require.NoError(t, err, "Failed to create second test script")

	stackService, err := services.GetGlobalStackService()
	require.NoError(t, err, "Stack service not available")
	stackService.ClearStack()

	cmd := &RunCommand{}

	// Execute first script
	err = cmd.Execute(map[string]string{}, script1)
	assert.NoError(t, err)

	// Execute second script
	err = cmd.Execute(map[string]string{}, script2)
	assert.NoError(t, err)

	// Verify both scripts are in the stack (LIFO order)
	command1, hasMore := stackService.PopCommand()
	assert.True(t, hasMore, "Expected first command in stack")
	assert.Equal(t, "\\"+script2, command1, "Expected second script with backslash prefix at top of stack (LIFO)")

	command2, hasMore := stackService.PopCommand()
	assert.True(t, hasMore, "Expected second command in stack")
	assert.Equal(t, "\\"+script1, command2, "Expected first script with backslash prefix in stack")
}

// TestRunCommand_Execute_StackServiceIntegration tests integration with stack service.
func TestRunCommand_Execute_StackServiceIntegration(t *testing.T) {
	// Setup test environment
	ctx := context.NewTestContext()
	setupStackTestRegistry(t, ctx)

	// Create test script
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "integration_test.neuro")
	scriptContent := `%% Integration test script
\echo Testing stack service integration
\set[integration_var="working"]`

	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(t, err, "Failed to create integration test script")

	// Get stack service and verify it's working
	stackService, err := services.GetGlobalStackService()
	require.NoError(t, err, "Stack service not available")

	// Clear stack and add a command to verify LIFO behavior
	stackService.ClearStack()
	stackService.PushCommand("\\echo before")

	cmd := &RunCommand{}
	err = cmd.Execute(map[string]string{}, scriptPath)
	assert.NoError(t, err)

	// Verify LIFO order: script path should be popped first
	firstCommand, hasMore := stackService.PopCommand()
	assert.True(t, hasMore, "Expected commands in stack")
	assert.Equal(t, "\\"+scriptPath, firstCommand, "Expected script path with backslash prefix at top of stack")

	secondCommand, hasMore := stackService.PopCommand()
	assert.True(t, hasMore, "Expected second command in stack")
	assert.Equal(t, "\\echo before", secondCommand, "Expected previous command in stack")
}

// TestRunCommand_Execute_PathVariations tests various path formats.
func TestRunCommand_Execute_PathVariations(t *testing.T) {
	// Setup test environment
	ctx := context.NewTestContext()
	setupStackTestRegistry(t, ctx)

	// Create test scripts with different path formats
	tempDir := t.TempDir()

	testCases := []struct {
		name       string
		scriptName string
		pathInput  string
	}{
		{
			name:       "simple filename",
			scriptName: "simple.neuro",
			pathInput:  "simple.neuro",
		},
		{
			name:       "filename with underscores",
			scriptName: "test_script_name.neuro",
			pathInput:  "test_script_name.neuro",
		},
		{
			name:       "filename with dashes",
			scriptName: "test-script-name.neuro",
			pathInput:  "test-script-name.neuro",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create script file
			scriptPath := filepath.Join(tempDir, tc.scriptName)
			err := os.WriteFile(scriptPath, []byte(`\echo `+tc.name), 0644)
			require.NoError(t, err, "Failed to create test script")

			// Change to temp directory for relative path testing
			originalDir, err := os.Getwd()
			require.NoError(t, err, "Failed to get current directory")

			err = os.Chdir(tempDir)
			require.NoError(t, err, "Failed to change to temp directory")
			defer func() {
				_ = os.Chdir(originalDir) // Restore directory, ignore error in test cleanup
			}()

			// Test the command
			stackService, err := services.GetGlobalStackService()
			require.NoError(t, err, "Stack service not available")
			stackService.ClearStack()

			cmd := &RunCommand{}
			err = cmd.Execute(map[string]string{}, tc.pathInput)
			assert.NoError(t, err)

			// Verify command was pushed
			assert.False(t, stackService.IsEmpty(), "Expected script path to be pushed to stack for '%s'", tc.name)
		})
	}
}
