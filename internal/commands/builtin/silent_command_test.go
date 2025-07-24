package builtin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestSilentCommand_Name(t *testing.T) {
	cmd := &SilentCommand{}
	assert.Equal(t, "silent", cmd.Name())
}

func TestSilentCommand_ParseMode(t *testing.T) {
	cmd := &SilentCommand{}
	assert.Equal(t, neurotypes.ParseModeRaw, cmd.ParseMode())
}

func TestSilentCommand_Description(t *testing.T) {
	cmd := &SilentCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "output")
	assert.Contains(t, strings.ToLower(desc), "suppress")
}

func TestSilentCommand_Usage(t *testing.T) {
	cmd := &SilentCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\silent")
	assert.Contains(t, usage, "command_to_execute")
}

func TestSilentCommand_HelpInfo(t *testing.T) {
	cmd := &SilentCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "silent", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeRaw, helpInfo.ParseMode)
	assert.Empty(t, helpInfo.Options) // Silent command has no options
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Check that examples contain expected commands
	exampleCommands := []string{
		"\\silent \\echo Hello World",
		"\\silent \\bash ls -la",
		"\\silent \\set[var=value]",
		"\\silent \\model-activate my-model",
		"\\silent",
	}

	for _, expectedCmd := range exampleCommands {
		found := false
		for _, example := range helpInfo.Examples {
			if example.Command == expectedCmd {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected example command not found: %s", expectedCmd)
	}
}

func TestSilentCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &SilentCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSilentTestRegistry(t, ctx)

	// Test empty silent command
	err := cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Empty silent command should not push anything to stack
	stackService, err := services.GetGlobalStackService()
	require.NoError(t, err)
	assert.Equal(t, 0, stackService.GetStackSize())
}

func TestSilentCommand_Execute_WithTargetCommand(t *testing.T) {
	cmd := &SilentCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSilentTestRegistry(t, ctx)

	tests := []struct {
		name          string
		input         string
		expectedStack []string
	}{
		{
			name:  "simple echo command",
			input: "\\echo hello",
			expectedStack: []string{
				"SILENT_BOUNDARY_START:silent_id_1",
				"\\echo hello",
				"SILENT_BOUNDARY_END:silent_id_1",
			},
		},
		{
			name:  "bash command",
			input: "\\bash ls -la",
			expectedStack: []string{
				"SILENT_BOUNDARY_START:silent_id_2",
				"\\bash ls -la",
				"SILENT_BOUNDARY_END:silent_id_2",
			},
		},
		{
			name:  "set command",
			input: "\\set[var=value]",
			expectedStack: []string{
				"SILENT_BOUNDARY_START:silent_id_3",
				"\\set[var=value]",
				"SILENT_BOUNDARY_END:silent_id_3",
			},
		},
		{
			name:  "complex command with spaces",
			input: "\\model-activate my-complex-model-name",
			expectedStack: []string{
				"SILENT_BOUNDARY_START:silent_id_4",
				"\\model-activate my-complex-model-name",
				"SILENT_BOUNDARY_END:silent_id_4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupSilentTestRegistry(t, ctx)

			err := cmd.Execute(map[string]string{}, tt.input)
			assert.NoError(t, err)

			// Verify stack contains expected commands in correct order
			stackService, err := services.GetGlobalStackService()
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedStack), stackService.GetStackSize())

			// Check stack contents (LIFO order - commands are popped in reverse order of push)
			actualStack := []string{}
			for !stackService.IsEmpty() {
				command, hasCommand := stackService.PopCommand()
				require.True(t, hasCommand)
				actualStack = append(actualStack, command)
			}

			assert.Equal(t, tt.expectedStack, actualStack)
		})
	}
}

func TestSilentCommand_Execute_UniqueIDs(t *testing.T) {
	cmd := &SilentCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSilentTestRegistry(t, ctx)

	// Execute multiple silent commands to ensure unique IDs
	commands := []string{"\\echo test1", "\\echo test2", "\\echo test3"}

	for _, command := range commands {
		err := cmd.Execute(map[string]string{}, command)
		require.NoError(t, err)

		// Check that each command gets a unique ID
		stackService, err := services.GetGlobalStackService()
		require.NoError(t, err)

		// Pop all commands for this execution
		var startMarker string
		for !stackService.IsEmpty() {
			command, hasCommand := stackService.PopCommand()
			require.True(t, hasCommand)
			if strings.HasPrefix(command, "SILENT_BOUNDARY_START:") {
				startMarker = command
				break
			}
		}

		expectedID := strings.TrimPrefix(startMarker, "SILENT_BOUNDARY_START:")
		assert.Contains(t, expectedID, "silent_id_")

		// Clear remaining commands for next test
		stackService.ClearStack()
	}
}

func TestSilentCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &SilentCommand{}

	// Set up empty registry to simulate missing services
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry()) // Empty registry

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
	}()

	// Don't setup services - should fail
	err := cmd.Execute(map[string]string{}, "\\echo test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stack service not available")
}

func TestSilentCommand_Execute_WhitespaceHandling(t *testing.T) {
	cmd := &SilentCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSilentTestRegistry(t, ctx)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "leading whitespace",
			input:    "   \\echo hello",
			expected: "\\echo hello",
		},
		{
			name:     "trailing whitespace",
			input:    "\\echo hello   ",
			expected: "\\echo hello",
		},
		{
			name:     "both leading and trailing",
			input:    "   \\echo hello world   ",
			expected: "\\echo hello world",
		},
		{
			name:     "tabs and spaces",
			input:    "\t  \\set[var=value]  \t",
			expected: "\\set[var=value]",
		},
		{
			name:     "only whitespace",
			input:    "   \t  ",
			expected: "", // Should be treated as empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupSilentTestRegistry(t, ctx)

			err := cmd.Execute(map[string]string{}, tt.input)
			assert.NoError(t, err)

			stackService, err := services.GetGlobalStackService()
			require.NoError(t, err)

			if tt.expected == "" {
				// Empty input should not push anything to stack
				assert.Equal(t, 0, stackService.GetStackSize())
			} else {
				// Should have 3 commands on stack
				assert.Equal(t, 3, stackService.GetStackSize())

				// Pop and verify the middle command (target command)
				stackService.PopCommand() // Skip END marker
				actualCommand, hasCommand := stackService.PopCommand()
				require.True(t, hasCommand)
				assert.Equal(t, tt.expected, actualCommand)
			}
		})
	}
}

// setupSilentTestRegistry sets up a test environment with required services for silent command tests
func setupSilentTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register required services
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

// Interface compliance check
var _ neurotypes.Command = (*SilentCommand)(nil)
