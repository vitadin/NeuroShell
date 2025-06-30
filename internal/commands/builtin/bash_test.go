package builtin

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/types"
)

func TestBashCommand_Name(t *testing.T) {
	cmd := &BashCommand{}
	assert.Equal(t, "bash", cmd.Name())
}

func TestBashCommand_ParseMode(t *testing.T) {
	cmd := &BashCommand{}
	assert.Equal(t, types.ParseModeRaw, cmd.ParseMode())
}

func TestBashCommand_Description(t *testing.T) {
	cmd := &BashCommand{}
	assert.Equal(t, "Execute system command", cmd.Description())
}

func TestBashCommand_Usage(t *testing.T) {
	cmd := &BashCommand{}
	assert.Equal(t, "\\bash[command] or \\bash command", cmd.Usage())
}

func TestBashCommand_Execute(t *testing.T) {
	tests := []struct {
		name           string
		args           map[string]string
		input          string
		wantErr        bool
		errMsg         string
		expectedOutput string
	}{
		{
			name:           "execute with input",
			args:           map[string]string{},
			input:          "ls -la",
			wantErr:        false,
			expectedOutput: "Executing: ls -la (not implemented yet)\n",
		},
		{
			name:           "execute with complex command",
			args:           map[string]string{},
			input:          "find /tmp -name '*.txt' | grep test",
			wantErr:        false,
			expectedOutput: "Executing: find /tmp -name '*.txt' | grep test (not implemented yet)\n",
		},
		{
			name:           "execute with quotes and special characters",
			args:           map[string]string{},
			input:          `echo "Hello World!" && pwd`,
			wantErr:        false,
			expectedOutput: "Executing: echo \"Hello World!\" && pwd (not implemented yet)\n",
		},
		{
			name:    "no command provided",
			args:    map[string]string{},
			input:   "",
			wantErr: true,
			errMsg:  "Usage:",
		},
		{
			name:           "command with whitespace",
			args:           map[string]string{},
			input:          "   ls -la   ",
			wantErr:        false,
			expectedOutput: "Executing:    ls -la    (not implemented yet)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &BashCommand{}
			ctx := testutils.NewMockContext()

			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmd.Execute(tt.args, tt.input, ctx)

			// Restore stdout
			w.Close()
			os.Stdout = originalStdout

			// Read captured output
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.expectedOutput != "" {
					assert.Equal(t, tt.expectedOutput, outputStr)
				}
			}
		})
	}
}

func TestBashCommand_Execute_EmptyInputHandling(t *testing.T) {
	cmd := &BashCommand{}
	ctx := testutils.NewMockContext()

	tests := []struct {
		name  string
		args  map[string]string
		input string
	}{
		{
			name:  "completely empty",
			args:  map[string]string{},
			input: "",
		},
		{
			name:  "whitespace only",
			args:  map[string]string{},
			input: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input, ctx)

			if tt.input == "" {
				// Empty input should error
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "Usage:")
			} else {
				// Whitespace-only input should be treated as a command
				assert.NoError(t, err)
			}
		})
	}
}

func TestBashCommand_Execute_LongCommands(t *testing.T) {
	cmd := &BashCommand{}
	ctx := testutils.NewMockContext()

	// Test with a very long command
	longCommand := "echo " + strings.Repeat("very long command ", 100)

	// Capture stdout
	originalStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(map[string]string{}, longCommand, ctx)

	// Restore stdout
	w.Close()
	os.Stdout = originalStdout

	assert.NoError(t, err)
}

func TestBashCommand_Execute_SpecialCharacters(t *testing.T) {
	cmd := &BashCommand{}
	ctx := testutils.NewMockContext()

	specialCommands := []string{
		"echo $HOME",
		"ls -la | grep test",
		"find . -name '*.go' -exec echo {} \\;",
		"echo \"test with quotes\"",
		"echo 'single quotes'",
		"echo `backticks`",
		"command && another_command",
		"command || fallback",
		"command; another_command",
		"echo test > output.txt",
		"cat < input.txt",
		"command 2>&1",
	}

	for _, specialCmd := range specialCommands {
		t.Run(fmt.Sprintf("special_cmd_%s", specialCmd[:min(20, len(specialCmd))]), func(t *testing.T) {
			// Capture stdout
			originalStdout := os.Stdout
			_, w, _ := os.Pipe()
			os.Stdout = w

			err := cmd.Execute(map[string]string{}, specialCmd, ctx)

			// Restore stdout
			w.Close()
			os.Stdout = originalStdout

			// Should handle special characters without error (though not execute them)
			assert.NoError(t, err)
		})
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Benchmark tests
func BenchmarkBashCommand_Execute(b *testing.B) {
	cmd := &BashCommand{}
	ctx := testutils.NewMockContext()
	input := "echo test"

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull
	defer func() {
		devNull.Close()
		os.Stdout = originalStdout
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(map[string]string{}, input, ctx)
	}
}

func BenchmarkBashCommand_Execute_LongCommand(b *testing.B) {
	cmd := &BashCommand{}
	ctx := testutils.NewMockContext()
	longInput := "echo " + strings.Repeat("long ", 1000)

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull
	defer func() {
		devNull.Close()
		os.Stdout = originalStdout
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(map[string]string{}, longInput, ctx)
	}
}

// Test interface compliance
func TestBashCommand_Interface(t *testing.T) {
	var _ types.Command = &BashCommand{}

	cmd := &BashCommand{}

	// Test all interface methods return reasonable values
	assert.NotEmpty(t, cmd.Name())
	assert.NotEmpty(t, cmd.Description())
	assert.NotEmpty(t, cmd.Usage())

	// ParseMode should be Raw for bash commands
	assert.Equal(t, types.ParseModeRaw, cmd.ParseMode())
}

// Test metadata consistency
func TestBashCommand_ConsistentMetadata(t *testing.T) {
	cmd := &BashCommand{}

	// Test that multiple calls return the same values
	name1 := cmd.Name()
	name2 := cmd.Name()
	assert.Equal(t, name1, name2)

	desc1 := cmd.Description()
	desc2 := cmd.Description()
	assert.Equal(t, desc1, desc2)

	usage1 := cmd.Usage()
	usage2 := cmd.Usage()
	assert.Equal(t, usage1, usage2)

	mode1 := cmd.ParseMode()
	mode2 := cmd.ParseMode()
	assert.Equal(t, mode1, mode2)
}
