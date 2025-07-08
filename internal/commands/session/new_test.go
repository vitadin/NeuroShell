package session

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestNewCommand_Name(t *testing.T) {
	cmd := &NewCommand{}
	assert.Equal(t, "session-new", cmd.Name())
}

func TestNewCommand_ParseMode(t *testing.T) {
	cmd := &NewCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestNewCommand_Description(t *testing.T) {
	cmd := &NewCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "session")
}

func TestNewCommand_Usage(t *testing.T) {
	cmd := &NewCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\session-new")
	assert.Contains(t, usage, "session_name")
	assert.Contains(t, usage, "system=")
}

func TestNewCommand_Execute_BasicFunctionality(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSessionTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
	}{
		{
			name:        "create named session",
			args:        map[string]string{},
			input:       "test_session",
			expectError: false,
		},
		{
			name:        "create session with system prompt",
			args:        map[string]string{"system": "You are a helpful assistant"},
			input:       "assistant_session",
			expectError: false,
		},
		{
			name:        "create session with spaces in name",
			args:        map[string]string{},
			input:       "my project work",
			expectError: false,
		},
		{
			name:        "create session with custom system prompt",
			args:        map[string]string{"system": "You are a code reviewer"},
			input:       "code review session",
			expectError: false,
		},
		{
			name:        "missing session name",
			args:        map[string]string{},
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupSessionTestRegistry(t, ctx)

			err := cmd.Execute(tt.args, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify session was created by checking variables
			sessionID, err := ctx.GetVariable("#session_id")
			assert.NoError(t, err)
			assert.NotEmpty(t, sessionID)

			sessionName, err := ctx.GetVariable("#session_name")
			assert.NoError(t, err)
			assert.NotEmpty(t, sessionName)

			// Check session has correct name (from input)
			if tt.input != "" {
				assert.Equal(t, tt.input, sessionName)
			}

			// Check message count - no initial messages supported anymore
			messageCount, err := ctx.GetVariable("#message_count")
			assert.NoError(t, err)
			assert.Equal(t, "0", messageCount) // No initial message support

			// Check system prompt
			systemPrompt, err := ctx.GetVariable("#system_prompt")
			assert.NoError(t, err)
			if tt.args["system"] != "" {
				assert.Equal(t, tt.args["system"], systemPrompt)
			} else {
				assert.Equal(t, "You are a helpful assistant.", systemPrompt) // Default
			}

			// Check output variable
			output, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "Created session")
		})
	}
}

func TestNewCommand_Execute_InvalidSessionNames(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	tests := []struct {
		name        string
		sessionName string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty session name",
			sessionName: "",
			expectError: true,
			errorMsg:    "session name is required",
		},
		{
			name:        "session name too long",
			sessionName: strings.Repeat("a", 65),
			expectError: true,
			errorMsg:    "too long",
		},
		{
			name:        "reserved name auto-versioned",
			sessionName: "new",
			expectError: false,
			errorMsg:    "",
		},
		{
			name:        "control characters",
			sessionName: "test\x00session",
			expectError: true,
			errorMsg:    "invalid characters",
		},
		{
			name:        "valid name with spaces",
			sessionName: "my test session",
			expectError: false,
		},
		{
			name:        "valid name with underscores",
			sessionName: "test_session_01",
			expectError: false,
		},
		{
			name:        "valid name with hyphens",
			sessionName: "test-session-01",
			expectError: false,
		},
		{
			name:        "valid name with dots",
			sessionName: "test.session.01",
			expectError: false,
		},
		{
			name:        "valid short name",
			sessionName: "a",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupSessionTestRegistry(t, ctx)

			err := cmd.Execute(map[string]string{}, tt.sessionName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewCommand_Execute_DuplicateSessionNames(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create first session
	err := cmd.Execute(map[string]string{}, "duplicate_test")
	assert.NoError(t, err)

	// Try to create second session with same name - should succeed with auto-versioning
	err = cmd.Execute(map[string]string{}, "duplicate_test")
	assert.NoError(t, err)
}

func TestNewCommand_Execute_VariableInterpolation(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Set up test variables
	require.NoError(t, ctx.SetVariable("session_prefix", "test"))
	require.NoError(t, ctx.SetVariable("assistant_type", "code reviewer"))

	// Create session with variable interpolation in session name and system prompt
	args := map[string]string{
		"system": "You are a ${assistant_type} for ${session_prefix} purposes",
	}
	input := "${session_prefix}_session"

	err := cmd.Execute(args, input)
	assert.NoError(t, err)

	// Check that variables were interpolated
	systemPrompt, err := ctx.GetVariable("#system_prompt")
	assert.NoError(t, err)
	assert.Contains(t, systemPrompt, "code reviewer")
	assert.Contains(t, systemPrompt, "test purposes")

	// Check that session name was interpolated
	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "test_session", sessionName)
}

func TestNewCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &NewCommand{}

	// Don't setup services - should fail
	err := cmd.Execute(map[string]string{}, "test_session")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not available")
}

// setupSessionTestRegistry sets up a test environment with required services
func setupSessionTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewChatSessionService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewInterpolationService())
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
var _ neurotypes.Command = (*NewCommand)(nil)
