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
	assert.Contains(t, usage, "name=")
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
			name:        "create anonymous session",
			args:        map[string]string{},
			input:       "",
			expectError: false,
		},
		{
			name:        "create named session",
			args:        map[string]string{"name": "test_session"},
			input:       "",
			expectError: false,
		},
		{
			name:        "create session with system prompt",
			args:        map[string]string{"system": "You are a helpful assistant"},
			input:       "",
			expectError: false,
		},
		{
			name:        "create session with initial message",
			args:        map[string]string{},
			input:       "Hello world",
			expectError: false,
		},
		{
			name:        "create session with all options",
			args:        map[string]string{"name": "full_test", "system": "You are a code reviewer"},
			input:       "Review this code",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupSessionTestRegistry(t, ctx)

			err := cmd.Execute(tt.args, tt.input, ctx)

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

			// Check if named session has correct name
			if tt.args["name"] != "" {
				assert.Equal(t, tt.args["name"], sessionName)
			}

			// Check message count
			messageCount, err := ctx.GetVariable("#message_count")
			assert.NoError(t, err)
			if tt.input != "" {
				assert.Equal(t, "1", messageCount) // Initial message added
			} else {
				assert.Equal(t, "0", messageCount) // No initial message
			}

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
			name:        "session name too short",
			sessionName: "ab",
			expectError: true,
			errorMsg:    "too short",
		},
		{
			name:        "session name too long",
			sessionName: strings.Repeat("a", 65),
			expectError: true,
			errorMsg:    "too long",
		},
		{
			name:        "invalid characters",
			sessionName: "test@session",
			expectError: true,
			errorMsg:    "invalid session name",
		},
		{
			name:        "reserved name",
			sessionName: "new",
			expectError: true,
			errorMsg:    "reserved",
		},
		{
			name:        "starts with special char",
			sessionName: "_test",
			expectError: true,
			errorMsg:    "invalid session name",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupSessionTestRegistry(t, ctx)

			args := map[string]string{"name": tt.sessionName}
			err := cmd.Execute(args, "", ctx)

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
	args := map[string]string{"name": "duplicate_test"}
	err := cmd.Execute(args, "", ctx)
	assert.NoError(t, err)

	// Try to create second session with same name
	err = cmd.Execute(args, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in use")
}

func TestNewCommand_Execute_VariableInterpolation(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Set up test variables
	require.NoError(t, ctx.SetVariable("session_prefix", "test"))
	require.NoError(t, ctx.SetVariable("assistant_type", "code reviewer"))

	// Create session with variable interpolation in system prompt
	args := map[string]string{
		"name":   "var_test",
		"system": "You are a ${assistant_type} for ${session_prefix} purposes",
	}
	input := "Hello ${session_prefix} session"

	err := cmd.Execute(args, input, ctx)
	assert.NoError(t, err)

	// Check that variables were interpolated
	systemPrompt, err := ctx.GetVariable("#system_prompt")
	assert.NoError(t, err)
	assert.Contains(t, systemPrompt, "code reviewer")
	assert.Contains(t, systemPrompt, "test purposes")
}

func TestNewCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()

	// Don't setup services - should fail
	err := cmd.Execute(map[string]string{}, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not available")
}

// setupSessionTestRegistry sets up a test environment with required services
func setupSessionTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewChatSessionService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll(ctx)
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})
}

// Interface compliance check
var _ neurotypes.Command = (*NewCommand)(nil)
