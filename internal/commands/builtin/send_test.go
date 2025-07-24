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

func TestSendCommand_Name(t *testing.T) {
	cmd := &SendCommand{}
	assert.Equal(t, "send", cmd.Name())
}

func TestSendCommand_ParseMode(t *testing.T) {
	cmd := &SendCommand{}
	assert.Equal(t, neurotypes.ParseModeRaw, cmd.ParseMode())
}

func TestSendCommand_Description(t *testing.T) {
	cmd := &SendCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "send")
	assert.Contains(t, strings.ToLower(desc), "message")
}

func TestSendCommand_Usage(t *testing.T) {
	cmd := &SendCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\send")
	assert.Contains(t, usage, "message")
}

func TestSendCommand_HelpInfo(t *testing.T) {
	cmd := &SendCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "send", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeRaw, helpInfo.ParseMode)
	assert.Empty(t, helpInfo.Options) // Send command has no options
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Check that examples contain expected commands
	expectedCommands := []string{
		"\\send Hello, how are you?",
		"\\send Analyze this data: ${data_variable}",
		"\\send ${_output}",
		"\\send Please review this code:\\n${code_content}",
		"\\set[_reply_way=stream] && \\send Tell me a story",
		"\\set[_reply_way=sync] && \\send What is 2+2?",
	}

	for _, expectedCmd := range expectedCommands {
		found := false
		for _, example := range helpInfo.Examples {
			if example.Command == expectedCmd {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected example command not found: %s", expectedCmd)
	}

	// Check that notes contain key information
	notesText := strings.Join(helpInfo.Notes, " ")
	assert.Contains(t, notesText, "session")
	assert.Contains(t, notesText, "model")
	assert.Contains(t, notesText, "_reply_way")
	assert.Contains(t, notesText, "API key")
}

func TestSendCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &SendCommand{}
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Usage")
}

func TestSendCommand_Execute_ValidInput(t *testing.T) {
	cmd := &SendCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSendTestRegistry(t, ctx)

	err := cmd.Execute(map[string]string{}, "Hello, world!")
	assert.NoError(t, err)

	// Verify command was pushed to stack
	stackService, err := services.GetGlobalStackService()
	require.NoError(t, err)
	assert.Equal(t, 1, stackService.GetStackSize())

	// Verify the pushed command
	command, hasCommand := stackService.PopCommand()
	require.True(t, hasCommand)
	assert.Equal(t, "\\_send Hello, world!", command)
}

func TestSendCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &SendCommand{}

	// Set up empty registry to simulate missing services
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry()) // Empty registry

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
	}()

	err := cmd.Execute(map[string]string{}, "test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stack service not available")
}

func TestSendCommand_Execute_MultipleMessages(t *testing.T) {
	cmd := &SendCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSendTestRegistry(t, ctx)

	messages := []string{
		"First message",
		"Second message with ${variable}",
		"Multi-line\nmessage content",
	}

	for i, msg := range messages {
		// Reset context for each test
		ctx = context.New()
		setupSendTestRegistry(t, ctx)

		err := cmd.Execute(map[string]string{}, msg)
		assert.NoError(t, err)

		// Verify command was pushed to stack
		stackService, err := services.GetGlobalStackService()
		require.NoError(t, err)
		assert.Equal(t, 1, stackService.GetStackSize())

		// Verify the pushed command
		command, hasCommand := stackService.PopCommand()
		require.True(t, hasCommand)
		assert.Equal(t, "\\_send "+msg, command, "Message %d failed", i+1)
	}
}

func TestSendCommand_Execute_WhitespaceHandling(t *testing.T) {
	cmd := &SendCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSendTestRegistry(t, ctx)

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "leading whitespace",
			input: "   Hello world",
		},
		{
			name:  "trailing whitespace",
			input: "Hello world   ",
		},
		{
			name:  "both leading and trailing",
			input: "   Hello world   ",
		},
		{
			name:  "tabs and spaces",
			input: "\t  Hello world  \t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupSendTestRegistry(t, ctx)

			err := cmd.Execute(map[string]string{}, tt.input)
			assert.NoError(t, err)

			// Verify command was pushed to stack exactly as provided (no trimming in delegation layer)
			stackService, err := services.GetGlobalStackService()
			require.NoError(t, err)
			assert.Equal(t, 1, stackService.GetStackSize())

			command, hasCommand := stackService.PopCommand()
			require.True(t, hasCommand)
			assert.Equal(t, "\\_send "+tt.input, command)
		})
	}
}

// setupSendTestRegistry sets up a test environment with required services for send command tests
func setupSendTestRegistry(t *testing.T, ctx neurotypes.Context) {
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
var _ neurotypes.Command = (*SendCommand)(nil)
