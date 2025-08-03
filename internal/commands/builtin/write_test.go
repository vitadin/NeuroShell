package builtin

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/pkg/neurotypes"
)

func TestWriteCommand_Name(t *testing.T) {
	cmd := &WriteCommand{}
	assert.Equal(t, "write", cmd.Name())
}

func TestWriteCommand_ParseMode(t *testing.T) {
	cmd := &WriteCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestWriteCommand_Description(t *testing.T) {
	cmd := &WriteCommand{}
	assert.Equal(t, "Write content to a file with overwrite or append modes", cmd.Description())
}

func TestWriteCommand_Usage(t *testing.T) {
	cmd := &WriteCommand{}
	assert.Equal(t, "\\write[file=path, mode=append] content", cmd.Usage())
}

func TestWriteCommand_HelpInfo(t *testing.T) {
	cmd := &WriteCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "write", helpInfo.Command)
	assert.Equal(t, "Write content to a file with overwrite or append modes", helpInfo.Description)
	assert.Equal(t, "\\write[file=path, mode=append] content", helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check options
	assert.Len(t, helpInfo.Options, 3)

	// Check file option
	fileOption := helpInfo.Options[0]
	assert.Equal(t, "file", fileOption.Name)
	assert.True(t, fileOption.Required)
	assert.Equal(t, "string", fileOption.Type)

	// Check mode option
	modeOption := helpInfo.Options[1]
	assert.Equal(t, "mode", modeOption.Name)
	assert.False(t, modeOption.Required)
	assert.Equal(t, "overwrite", modeOption.Default)

	// Check silent option
	silentOption := helpInfo.Options[2]
	assert.Equal(t, "silent", silentOption.Name)
	assert.False(t, silentOption.Required)
	assert.Equal(t, "false", silentOption.Default)

	// Check examples
	assert.NotEmpty(t, helpInfo.Examples)
	assert.GreaterOrEqual(t, len(helpInfo.Examples), 4)

	// Check notes
	assert.NotEmpty(t, helpInfo.Notes)
}

func TestWriteCommand_Execute_MissingFileParameter(t *testing.T) {
	cmd := &WriteCommand{}

	tests := []struct {
		name string
		args map[string]string
	}{
		{
			name: "no file parameter",
			args: map[string]string{},
		},
		{
			name: "empty file parameter",
			args: map[string]string{"file": ""},
		},
		{
			name: "whitespace only file parameter",
			args: map[string]string{"file": "   \t\n  "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, "test content")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "file parameter is required")
		})
	}
}

func TestWriteCommand_Execute_InvalidMode(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	tests := []struct {
		name string
		mode string
	}{
		{
			name: "invalid mode",
			mode: "invalid",
		},
		{
			name: "unsupported mode",
			mode: "prepend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{
				"file": testFile,
				"mode": tt.mode,
			}
			err := cmd.Execute(args, "test content")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid mode")
		})
	}
}

func TestWriteCommand_Execute_OverwriteMode(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Capture stdout to check feedback
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	tests := []struct {
		name     string
		args     map[string]string
		content  string
		expected string
	}{
		{
			name: "explicit overwrite mode",
			args: map[string]string{
				"file": testFile,
				"mode": "overwrite",
			},
			content:  "Hello World!",
			expected: "Hello World!",
		},
		{
			name: "default mode (overwrite)",
			args: map[string]string{
				"file": testFile,
			},
			content:  "Default mode test",
			expected: "Default mode test",
		},
		{
			name: "overwrite existing content",
			args: map[string]string{
				"file": testFile,
				"mode": "overwrite",
			},
			content:  "New content",
			expected: "New content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.content)
			assert.NoError(t, err)

			// Verify file content
			content, err := os.ReadFile(testFile)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(content))
		})
	}

	// Restore stdout and check output
	_ = w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Should contain feedback messages
	assert.Contains(t, outputStr, "Wrote")
	assert.Contains(t, outputStr, "bytes to")
}

func TestWriteCommand_Execute_AppendMode(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Write initial content
	err := os.WriteFile(testFile, []byte("Initial content\n"), 0644)
	require.NoError(t, err)

	// Capture stdout to check feedback
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Append content
	args := map[string]string{
		"file": testFile,
		"mode": "append",
	}
	err = cmd.Execute(args, "Appended content")
	assert.NoError(t, err)

	// Restore stdout and check output
	_ = w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Check feedback message
	assert.Contains(t, outputStr, "Appended")
	assert.Contains(t, outputStr, "bytes to")

	// Verify file content
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Initial content\nAppended content", string(content))
}

func TestWriteCommand_Execute_SilentMode(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	// Capture stdout to verify no output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	args := map[string]string{
		"file":   testFile,
		"silent": "true",
	}
	err := cmd.Execute(args, "Silent test content")
	assert.NoError(t, err)

	// Restore stdout and check output
	_ = w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	// Should have no output in silent mode
	assert.Empty(t, outputStr)

	// Verify file was still created with correct content
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "Silent test content", string(content))
}

func TestWriteCommand_Execute_DirectoryCreation(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	nestedPath := filepath.Join(tempDir, "level1", "level2", "level3", "test.txt")

	args := map[string]string{
		"file": nestedPath,
	}
	err := cmd.Execute(args, "Content in nested directories")
	assert.NoError(t, err)

	// Verify directories were created
	assert.DirExists(t, filepath.Dir(nestedPath))

	// Verify file content
	content, err := os.ReadFile(nestedPath)
	require.NoError(t, err)
	assert.Equal(t, "Content in nested directories", string(content))
}

func TestWriteCommand_Execute_EmptyContent(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty.txt")

	args := map[string]string{
		"file": testFile,
	}
	err := cmd.Execute(args, "")
	assert.NoError(t, err)

	// Verify empty file was created
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "", string(content))

	// Check file exists
	assert.FileExists(t, testFile)
}

func TestWriteCommand_Execute_SpecialCharacters(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "special.txt")

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "unicode content",
			content: "Hello 🌍 Unicode! 文字",
		},
		{
			name:    "newlines and tabs",
			content: "Line 1\nLine 2\tTabbed",
		},
		{
			name:    "special symbols",
			content: "Special: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{
				"file": testFile,
			}
			err := cmd.Execute(args, tt.content)
			assert.NoError(t, err)

			// Verify file content
			content, err := os.ReadFile(testFile)
			require.NoError(t, err)
			assert.Equal(t, tt.content, string(content))
		})
	}
}

func TestWriteCommand_Execute_BooleanParsing(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	tests := []struct {
		name      string
		silent    string
		expectOut bool
	}{
		{
			name:      "silent true",
			silent:    "true",
			expectOut: false,
		},
		{
			name:      "silent false",
			silent:    "false",
			expectOut: true,
		},
		{
			name:      "silent invalid - defaults to false",
			silent:    "invalid",
			expectOut: true,
		},
		{
			name:      "silent empty - defaults to false",
			silent:    "",
			expectOut: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			args := map[string]string{
				"file":   testFile,
				"silent": tt.silent,
			}
			err := cmd.Execute(args, "test content")
			assert.NoError(t, err)

			// Restore stdout and check output
			_ = w.Close()
			os.Stdout = oldStdout
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			if tt.expectOut {
				assert.Contains(t, outputStr, "Wrote")
			} else {
				assert.Empty(t, outputStr)
			}
		})
	}
}

func TestWriteCommand_Execute_ModeVariations(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	tests := []struct {
		name        string
		mode        string
		expectError bool
	}{
		{
			name:        "lowercase overwrite",
			mode:        "overwrite",
			expectError: false,
		},
		{
			name:        "uppercase overwrite",
			mode:        "OVERWRITE",
			expectError: false,
		},
		{
			name:        "mixed case append",
			mode:        "Append",
			expectError: false,
		},
		{
			name:        "with spaces",
			mode:        " append ",
			expectError: false,
		},
		{
			name:        "invalid mode",
			mode:        "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{
				"file": testFile,
				"mode": tt.mode,
			}
			err := cmd.Execute(args, "test content")
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteCommand_Execute_PermissionError(t *testing.T) {
	cmd := &WriteCommand{}

	// Try to write to a directory that should cause permission error
	// Note: This test may be platform-specific and might be skipped in some environments
	if os.Getuid() == 0 {
		t.Skip("Test skipped when running as root")
	}

	args := map[string]string{
		"file": "/root/cannot_write_here.txt",
	}
	err := cmd.Execute(args, "test content")

	// Should get some kind of error (permission denied or directory creation failure)
	assert.Error(t, err)
	assert.True(t,
		strings.Contains(err.Error(), "permission denied") ||
			strings.Contains(err.Error(), "failed to create directory") ||
			strings.Contains(err.Error(), "failed to write file"),
		"Expected permission-related error, got: %v", err)
}

func TestWriteCommand_Execute_LargeContent(t *testing.T) {
	cmd := &WriteCommand{}

	// Create temporary directory for test
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "large.txt")

	// Create large content (1MB)
	largeContent := strings.Repeat("A", 1024*1024)

	args := map[string]string{
		"file": testFile,
	}
	err := cmd.Execute(args, largeContent)
	assert.NoError(t, err)

	// Verify file content
	content, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, largeContent, string(content))
	assert.Equal(t, 1024*1024, len(content))
}

// Interface compliance check
var _ neurotypes.Command = (*WriteCommand)(nil)
