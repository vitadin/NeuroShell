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

func TestRenameCommand_Name(t *testing.T) {
	cmd := &RenameCommand{}
	assert.Equal(t, "session-rename", cmd.Name())
}

func TestRenameCommand_ParseMode(t *testing.T) {
	cmd := &RenameCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestRenameCommand_Description(t *testing.T) {
	cmd := &RenameCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "change")
	assert.Contains(t, strings.ToLower(desc), "session")
	assert.Contains(t, strings.ToLower(desc), "name")
}

func TestRenameCommand_Usage(t *testing.T) {
	cmd := &RenameCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\session-rename")
	assert.Contains(t, usage, "session=")
	assert.Contains(t, usage, "new_session_name")
}

func TestRenameCommand_HelpInfo(t *testing.T) {
	cmd := &RenameCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "session-rename", helpInfo.Command)
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
	assert.True(t, variableNames["#old_session_name"])
	assert.True(t, variableNames["#new_session_name"])
	assert.True(t, variableNames["_output"])

	// Check notes
	assert.NotEmpty(t, helpInfo.Notes)
}

// setupRenameTestRegistry sets up a test environment with required services for rename tests
func setupRenameTestRegistry(t *testing.T, ctx neurotypes.Context) {
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

func TestRenameCommand_Execute_ActiveSession_Success(t *testing.T) {
	cmd := &RenameCommand{}
	ctx := context.New()
	setupRenameTestRegistry(t, ctx)

	// Create a test session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "System prompt"}, "old_session_name")
	require.NoError(t, err)

	// Execute command to rename session
	args := map[string]string{} // No session specified, should use active
	newName := "My New Session Name"

	err = cmd.Execute(args, newName)
	require.NoError(t, err)

	// Verify variables were set
	oldName, err := ctx.GetVariable("#old_session_name")
	assert.NoError(t, err)
	assert.Equal(t, "old_session_name", oldName)

	newSessionName, err := ctx.GetVariable("#new_session_name")
	assert.NoError(t, err)
	assert.Equal(t, newName, newSessionName)

	output, err := ctx.GetVariable("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Renamed session from")
	assert.Contains(t, output, "old_session_name")
	assert.Contains(t, output, newName)
}

func TestRenameCommand_Execute_EmptyName_Error(t *testing.T) {
	cmd := &RenameCommand{}
	ctx := context.New()
	setupRenameTestRegistry(t, ctx)

	// Create a test session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "System prompt"}, "test_session")
	require.NoError(t, err)

	// Execute command with empty new name
	args := map[string]string{}
	emptyName := ""

	err = cmd.Execute(args, emptyName)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "new session name is required")
}
