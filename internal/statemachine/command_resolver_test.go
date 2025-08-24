package statemachine

import (
	"os"
	"path/filepath"
	"testing"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/builtin"
	"neuroshell/pkg/neurotypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommandResolver(t *testing.T) {
	resolver := NewCommandResolver()
	assert.NotNil(t, resolver)
	assert.NotNil(t, resolver.stdlibLoader)
	assert.NotNil(t, resolver.logger)
}

func TestCommandResolver_ResolveCommand_BuiltinCommands(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resolver := NewCommandResolver()

	// Register a test builtin command
	testCmd := &builtin.EchoCommand{}
	err := commands.GetGlobalRegistry().Register(testCmd)
	require.NoError(t, err)

	result, err := resolver.ResolveCommand("echo")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "echo", result.Name)
	assert.Equal(t, neurotypes.CommandTypeBuiltin, result.Type)
	assert.Equal(t, testCmd, result.BuiltinCommand)
	assert.Empty(t, result.ScriptContent)
	assert.Empty(t, result.ScriptPath)
}

func TestCommandResolver_ResolveCommand_StdlibCommands(t *testing.T) {
	resolver := NewCommandResolver()

	// Test with a known stdlib script (assuming some exist)
	// This test may need adjustment based on actual stdlib contents
	result, err := resolver.ResolveCommand("stdlib-test-script")
	if err != nil {
		// If no stdlib scripts exist, test the error path
		assert.Contains(t, err.Error(), "unknown command")
		return
	}

	// If stdlib script exists, verify the result
	assert.Equal(t, "stdlib-test-script", result.Name)
	assert.Equal(t, neurotypes.CommandTypeStdlib, result.Type)
	assert.NotEmpty(t, result.ScriptContent)
	assert.NotEmpty(t, result.ScriptPath)
}

func TestCommandResolver_ResolveCommand_UserScripts(t *testing.T) {
	resolver := NewCommandResolver()

	// Create a temporary script file with unique name to avoid stdlib conflict
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "unique-user-script.neuro")
	scriptContent := "\\echo Hello from user script\n\\set test_var=value"

	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(t, err)

	// Change to the temp directory to test relative path resolution
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Test with relative path
	result, err := resolver.ResolveCommand("unique-user-script.neuro")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "unique-user-script.neuro", result.Name)
	assert.Equal(t, neurotypes.CommandTypeUser, result.Type)
	assert.Equal(t, scriptContent, result.ScriptContent)
	// Check that path ends with our expected filename (handles symlink resolution differences)
	assert.Contains(t, result.ScriptPath, "unique-user-script.neuro")

	// Test with absolute path
	result2, err := resolver.ResolveCommand(scriptPath)
	require.NoError(t, err)
	require.NotNil(t, result2)

	assert.Contains(t, result2.Name, "unique-user-script.neuro")
	assert.Equal(t, neurotypes.CommandTypeUser, result2.Type)
	assert.Equal(t, scriptContent, result2.ScriptContent)
	assert.Contains(t, result2.ScriptPath, "unique-user-script.neuro")
}

func TestCommandResolver_ResolveCommand_UserScripts_NeuroRC(t *testing.T) {
	resolver := NewCommandResolver()

	// Create a temporary .neurorc file
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "init.neurorc")
	scriptContent := "\\set init_executed=true"

	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(t, err)

	// Test with absolute path
	result, err := resolver.ResolveCommand(scriptPath)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Name, "init.neurorc")
	assert.Equal(t, neurotypes.CommandTypeUser, result.Type)
	assert.Equal(t, scriptContent, result.ScriptContent)
	assert.Contains(t, result.ScriptPath, "init.neurorc")
}

func TestCommandResolver_ResolveCommand_UnknownCommand(t *testing.T) {
	resolver := NewCommandResolver()

	result, err := resolver.ResolveCommand("definitely-does-not-exist")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown command: definitely-does-not-exist")
}

func TestCommandResolver_ResolveCommand_Priority(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resolver := NewCommandResolver()

	// Register a builtin command with the same name as a potential script
	testCmd := &builtin.EchoCommand{}
	err := commands.GetGlobalRegistry().Register(testCmd) // EchoCommand returns "echo" as name
	require.NoError(t, err)

	// Create a script file with the same name
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test-priority.neuro")
	err = os.WriteFile(scriptPath, []byte("\\echo from script"), 0644)
	require.NoError(t, err)

	// Change to the temp directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Test that builtin takes priority over user script
	result, err := resolver.ResolveCommand("echo")
	require.NoError(t, err)
	assert.Equal(t, neurotypes.CommandTypeBuiltin, result.Type)
	assert.Equal(t, testCmd, result.BuiltinCommand)

	// Test that user script is resolved when using file extension
	result2, err := resolver.ResolveCommand("test-priority.neuro")
	require.NoError(t, err)
	assert.Equal(t, neurotypes.CommandTypeUser, result2.Type)
	assert.Equal(t, "\\echo from script", result2.ScriptContent)
}

func TestCommandResolver_resolvePathSafely_DirectoryTraversal(t *testing.T) {
	resolver := NewCommandResolver()

	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "simple directory traversal",
			filePath: "../../../etc/passwd",
			wantErr:  true,
			errMsg:   "parent directory access not allowed",
		},
		{
			name:     "complex directory traversal",
			filePath: "safe/../../../dangerous/file",
			wantErr:  true,
			errMsg:   "parent directory access not allowed",
		},
		{
			name:     "hidden directory traversal",
			filePath: "..\\..\\windows\\system32\\config\\sam",
			wantErr:  true,
			errMsg:   "parent directory access not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolver.resolvePathSafely(tt.filePath)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestCommandResolver_resolvePathSafely_AbsolutePath(t *testing.T) {
	resolver := NewCommandResolver()

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.neuro")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test with absolute path to existing file
	result, err := resolver.resolvePathSafely(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, tmpFile, result)

	// Test with absolute path to non-existent file
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.neuro")
	result, err = resolver.resolvePathSafely(nonExistentPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
	assert.Empty(t, result)
}

func TestCommandResolver_resolvePathSafely_RelativePath(t *testing.T) {
	resolver := NewCommandResolver()

	// Create a temporary file in current working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	tmpFile := "test-relative.neuro"
	err = os.WriteFile(tmpFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Test with relative path to existing file
	result, err := resolver.resolvePathSafely(tmpFile)
	assert.NoError(t, err)
	// Check that result contains our filename (handles symlink resolution differences)
	assert.Contains(t, result, tmpFile)

	// Test with relative path to non-existent file
	result, err = resolver.resolvePathSafely("does-not-exist.neuro")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
	assert.Empty(t, result)
}

func TestCommandResolver_resolveUserFilePath_FileReadError(t *testing.T) {
	resolver := NewCommandResolver()

	// Create a directory instead of a file to trigger read error
	tmpDir := t.TempDir()
	dirPath := filepath.Join(tmpDir, "not-a-file.neuro")
	err := os.Mkdir(dirPath, 0755)
	require.NoError(t, err)

	result, err := resolver.resolveUserFilePath(dirPath)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to read script file")
}

func TestCommandResolver_resolveUserFilePath_InvalidPath(t *testing.T) {
	resolver := NewCommandResolver()

	result, err := resolver.resolveUserFilePath("../../../invalid/path.neuro")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid file path")
}

func TestCommandResolver_ResolveCommand_NonScriptFile(t *testing.T) {
	resolver := NewCommandResolver()

	// Test command names that don't end with .neuro or .neurorc
	result, err := resolver.ResolveCommand("regular-command-name")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown command")

	result, err = resolver.ResolveCommand("file.txt")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unknown command")
}

// Test edge cases for path resolution
func TestCommandResolver_EdgeCases(t *testing.T) {
	resolver := NewCommandResolver()

	t.Run("empty command name", func(t *testing.T) {
		result, err := resolver.ResolveCommand("")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unknown command")
	})

	t.Run("whitespace command name", func(t *testing.T) {
		result, err := resolver.ResolveCommand("   ")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unknown command")
	})

	t.Run("command with spaces", func(t *testing.T) {
		result, err := resolver.ResolveCommand("command with spaces")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unknown command")
	})
}

// Test interaction between different command types
func TestCommandResolver_CommandTypeInteraction(t *testing.T) {
	cleanup := setupTestEnvironment(t)
	defer cleanup()

	resolver := NewCommandResolver()

	// Create a temporary script
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalWd) }()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	scriptPath := "interaction-test.neuro"
	err = os.WriteFile(scriptPath, []byte("\\echo user script"), 0644)
	require.NoError(t, err)

	// Test user script resolution
	result, err := resolver.ResolveCommand(scriptPath)
	require.NoError(t, err)
	assert.Equal(t, neurotypes.CommandTypeUser, result.Type)

	// Register a builtin with similar name (should not conflict due to extension)
	testCmd := &builtin.EchoCommand{}
	err = commands.GetGlobalRegistry().Register(testCmd) // EchoCommand returns "echo" as name
	require.NoError(t, err)

	// Test that builtin is resolved for command without extension
	result2, err := resolver.ResolveCommand("echo")
	require.NoError(t, err)
	assert.Equal(t, neurotypes.CommandTypeBuiltin, result2.Type)

	// Test that user script is still resolved with extension
	result3, err := resolver.ResolveCommand(scriptPath)
	require.NoError(t, err)
	assert.Equal(t, neurotypes.CommandTypeUser, result3.Type)
}
