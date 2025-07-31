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

func TestEditSystemCommand_Name(t *testing.T) {
	cmd := &EditSystemCommand{}
	assert.Equal(t, "session-edit-system", cmd.Name())
}

func TestEditSystemCommand_ParseMode(t *testing.T) {
	cmd := &EditSystemCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestEditSystemCommand_Description(t *testing.T) {
	cmd := &EditSystemCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "edit")
	assert.Contains(t, strings.ToLower(desc), "system")
	assert.Contains(t, strings.ToLower(desc), "prompt")
}

func TestEditSystemCommand_Usage(t *testing.T) {
	cmd := &EditSystemCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\session-edit-system")
	assert.Contains(t, usage, "session=")
	assert.Contains(t, usage, "system_prompt")
}

func TestEditSystemCommand_HelpInfo(t *testing.T) {
	cmd := &EditSystemCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "session-edit-system", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check options
	assert.Len(t, helpInfo.Options, 1)
	sessionOption := helpInfo.Options[0]
	assert.Equal(t, "session", sessionOption.Name)
	assert.False(t, sessionOption.Required)
	assert.Equal(t, "string", sessionOption.Type)

	// Check examples
	assert.NotEmpty(t, helpInfo.Examples)
	assert.True(t, len(helpInfo.Examples) >= 3)

	// Check stored variables
	assert.NotEmpty(t, helpInfo.StoredVariables)
	variableNames := make(map[string]bool)
	for _, variable := range helpInfo.StoredVariables {
		variableNames[variable.Name] = true
	}
	assert.True(t, variableNames["#session_id"])
	assert.True(t, variableNames["#session_name"])
	assert.True(t, variableNames["#old_system_prompt"])
	assert.True(t, variableNames["#new_system_prompt"])
	assert.True(t, variableNames["_output"])

	// Check notes
	assert.NotEmpty(t, helpInfo.Notes)
}

// setupEditSystemTestRegistry sets up a test environment with required services for edit system tests
func setupEditSystemTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	services.SetGlobalRegistry(services.NewRegistry())
	// Set the test context as global context
	context.SetGlobalContext(ctx)
	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().RegisterService(services.NewChatSessionService())
	require.NoError(t, err)
	err = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	require.NoError(t, err)
	// Initialize services
	for _, service := range services.GetGlobalRegistry().GetAllServices() {
		err := service.Initialize()
		require.NoError(t, err)
	}
}

func TestEditSystemCommand_Execute_ActiveSession_Success(t *testing.T) {
	cmd := &EditSystemCommand{}
	ctx := context.New()
	setupEditSystemTestRegistry(t, ctx)

	// Create a test session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Original system prompt"}, "test_session")
	require.NoError(t, err)

	// Execute command to update system prompt
	args := map[string]string{} // No session specified, should use active
	newSystemPrompt := "You are a helpful coding assistant"

	err = cmd.Execute(args, newSystemPrompt)
	require.NoError(t, err)

	// Verify variables were set
	sessionID, err := ctx.GetVariable("#session_id")
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "test_session", sessionName)

	oldPrompt, err := ctx.GetVariable("#old_system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, "Original system prompt", oldPrompt)

	newPrompt, err := ctx.GetVariable("#new_system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, newSystemPrompt, newPrompt)

	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Updated system prompt")
	assert.Contains(t, output, "test_session")
}

func TestEditSystemCommand_Execute_SpecificSession_Success(t *testing.T) {
	cmd := &EditSystemCommand{}
	ctx := context.New()
	setupEditSystemTestRegistry(t, ctx)

	// Create multiple test sessions
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "System prompt 1"}, "session1")
	require.NoError(t, err)

	err = newCmd.Execute(map[string]string{"system": "System prompt 2"}, "session2")
	require.NoError(t, err)

	// Execute command to update session2's system prompt
	args := map[string]string{"session": "session2"}
	newSystemPrompt := "You are a specialized assistant"

	err = cmd.Execute(args, newSystemPrompt)
	require.NoError(t, err)

	// Verify variables were set for session2
	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "session2", sessionName)

	oldPrompt, err := ctx.GetVariable("#old_system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, "System prompt 2", oldPrompt)

	newPrompt, err := ctx.GetVariable("#new_system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, newSystemPrompt, newPrompt)
}

func TestEditSystemCommand_Execute_ClearSystemPrompt_Success(t *testing.T) {
	cmd := &EditSystemCommand{}
	ctx := context.New()
	setupEditSystemTestRegistry(t, ctx)

	// Create a test session with system prompt
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Original system prompt"}, "test_session")
	require.NoError(t, err)

	// Execute command to clear system prompt (empty string)
	args := map[string]string{}
	emptyPrompt := ""

	err = cmd.Execute(args, emptyPrompt)
	require.NoError(t, err)

	// Verify appropriate output message
	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Cleared system prompt")
	assert.Contains(t, output, "test_session")

	// Verify old prompt was stored
	oldPrompt, err := ctx.GetVariable("#old_system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, "Original system prompt", oldPrompt)

	newPrompt, err := ctx.GetVariable("#new_system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, "", newPrompt)
}

func TestEditSystemCommand_Execute_NoActiveSession_Error(t *testing.T) {
	cmd := &EditSystemCommand{}
	ctx := context.New()
	setupEditSystemTestRegistry(t, ctx)

	// Don't create any sessions
	args := map[string]string{} // No session specified
	newSystemPrompt := "Test prompt"

	err := cmd.Execute(args, newSystemPrompt)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active session found")
}
