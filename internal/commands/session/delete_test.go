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

func TestDeleteCommand_Name(t *testing.T) {
	cmd := &DeleteCommand{}
	assert.Equal(t, "session-delete", cmd.Name())
}

func TestDeleteCommand_ParseMode(t *testing.T) {
	cmd := &DeleteCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestDeleteCommand_Description(t *testing.T) {
	cmd := &DeleteCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "delete")
	assert.Contains(t, strings.ToLower(desc), "session")
}

func TestDeleteCommand_Usage(t *testing.T) {
	cmd := &DeleteCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\session-delete")
	assert.Contains(t, usage, "name=")
	assert.Contains(t, usage, "Examples:")
}

func TestDeleteCommand_Execute_BasicFunctionality(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupSessionTestRegistry(t, ctx)

	// Create a session first to delete
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{"name": "test_session"}, "", ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
	}{
		{
			name:        "delete session by name option",
			args:        map[string]string{"name": "test_session"},
			input:       "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create session before each test
			ctx = context.New()
			setupSessionTestRegistry(t, ctx)
			err := createCmd.Execute(map[string]string{"name": "test_session"}, "", ctx)
			require.NoError(t, err)

			err = cmd.Execute(tt.args, tt.input, ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify session was deleted by checking output
			output, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "Deleted session")
			assert.Contains(t, output, "test_session")
		})
	}
}

func TestDeleteCommand_Execute_DeleteByInput(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create a session first
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{"name": "input_test"}, "", ctx)
	require.NoError(t, err)

	// Delete by input parameter
	err = cmd.Execute(map[string]string{}, "input_test", ctx)
	assert.NoError(t, err)

	// Verify deletion
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Deleted session")
	assert.Contains(t, output, "input_test")
}

func TestDeleteCommand_Execute_DeleteBySessionID(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create a session first
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{"name": "id_test"}, "", ctx)
	require.NoError(t, err)

	// Get the chat service to access session directly
	chatService, err := cmd.getChatSessionService()
	require.NoError(t, err)

	// Get the session by name to get the actual ID
	session, err := chatService.GetSessionByName("id_test")
	require.NoError(t, err)
	actualSessionID := session.ID

	// Delete by actual session ID
	err = cmd.Execute(map[string]string{}, actualSessionID, ctx)
	assert.NoError(t, err)

	// Verify deletion
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Deleted session")
	assert.Contains(t, output, "id_test")
}

func TestDeleteCommand_Execute_MissingSessionIdentifier(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Try to delete without providing session name or ID
	err := cmd.Execute(map[string]string{}, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session name or ID is required")
	assert.Contains(t, err.Error(), "Usage:")
}

func TestDeleteCommand_Execute_SessionNotFound(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Try to delete non-existent session
	err := cmd.Execute(map[string]string{"name": "nonexistent"}, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteCommand_Execute_VariableInterpolation(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create a session
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{"name": "var_test"}, "", ctx)
	require.NoError(t, err)

	// Set up test variable
	require.NoError(t, ctx.SetVariable("session_to_delete", "var_test"))

	// Delete using variable interpolation
	err = cmd.Execute(map[string]string{"name": "${session_to_delete}"}, "", ctx)
	assert.NoError(t, err)

	// Verify deletion
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Deleted session")
	assert.Contains(t, output, "var_test")
}

func TestDeleteCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()

	// Don't setup services - should fail
	err := cmd.Execute(map[string]string{"name": "test"}, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not available")
}

func TestDeleteCommand_Execute_ChatServiceNotAvailable(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()

	// Setup only variable service, not chat service
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().InitializeAll(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})

	// Should fail due to missing chat service
	err = cmd.Execute(map[string]string{"name": "test"}, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat session service not available")
}

func TestDeleteCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()

	// Setup only chat service, not variable service
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	err := services.GetGlobalRegistry().RegisterService(services.NewChatSessionService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().InitializeAll(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})

	// Should fail due to missing variable service
	err = cmd.Execute(map[string]string{"name": "test"}, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestDeleteCommand_Execute_SessionVariableManagement(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create two sessions
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{"name": "session1"}, "", ctx)
	require.NoError(t, err)

	err = createCmd.Execute(map[string]string{"name": "session2"}, "", ctx)
	require.NoError(t, err)

	// Verify session variables are set for session2 (current)
	sessionName, err := ctx.GetVariable("#session_name")
	require.NoError(t, err)
	assert.Equal(t, "session2", sessionName)

	// Delete session1 (not current)
	err = cmd.Execute(map[string]string{"name": "session1"}, "", ctx)
	assert.NoError(t, err)

	// Session variables should still be set for session2
	sessionName, err = ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "session2", sessionName)

	// Delete session2 (current)
	err = cmd.Execute(map[string]string{"name": "session2"}, "", ctx)
	assert.NoError(t, err)

	// Session variables should be cleared or updated
	sessionName, err = ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	// Session name should be empty if no active session
	assert.Equal(t, "", sessionName)
}

func TestDeleteCommand_Execute_PriorityOfArguments(t *testing.T) {
	cmd := &DeleteCommand{}
	ctx := context.New()
	setupSessionTestRegistry(t, ctx)

	// Create two sessions
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{"name": "priority_test1"}, "", ctx)
	require.NoError(t, err)

	err = createCmd.Execute(map[string]string{"name": "priority_test2"}, "", ctx)
	require.NoError(t, err)

	// Test that name argument takes priority over input
	err = cmd.Execute(map[string]string{"name": "priority_test1"}, "priority_test2", ctx)
	assert.NoError(t, err)

	// Verify that priority_test1 was deleted (name argument took priority)
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "priority_test1")
	assert.NotContains(t, output, "priority_test2")
}

// Interface compliance check
var _ neurotypes.Command = (*DeleteCommand)(nil)
