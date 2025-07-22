package session

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestGetCommand_Name(t *testing.T) {
	cmd := &GetCommand{}
	assert.Equal(t, "session-get", cmd.Name())
}

func TestGetCommand_ParseMode(t *testing.T) {
	cmd := &GetCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestGetCommand_Description(t *testing.T) {
	cmd := &GetCommand{}
	description := cmd.Description()
	assert.NotEmpty(t, description)
	assert.Contains(t, strings.ToLower(description), "session")
}

func TestGetCommand_Usage(t *testing.T) {
	cmd := &GetCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "session-get")
	assert.Contains(t, usage, "name_or_prefix")
}

func TestGetCommand_HelpInfo(t *testing.T) {
	cmd := &GetCommand{}
	help := cmd.HelpInfo()

	assert.Equal(t, "session-get", help.Command)
	assert.NotEmpty(t, help.Description)
	assert.NotEmpty(t, help.Usage)

	// Check options
	assert.Len(t, help.Options, 1)
	assert.Equal(t, "name_or_prefix", help.Options[0].Name)
	assert.False(t, help.Options[0].Required)

	// Check examples
	assert.Len(t, help.Examples, 3)
	assert.Contains(t, help.Examples[0].Command, "\\session-get")
	assert.Contains(t, help.Examples[1].Command, "\\session-get[project1]")
	assert.Contains(t, help.Examples[2].Command, "\\session-get proj")

	// Check notes
	assert.Len(t, help.Notes, 4)
}

func TestGetCommand_Execute_GetActiveSession(t *testing.T) {
	cmd := &GetCommand{}
	setupSessionGetTestRegistry(t)

	// Create a test session and set it as active
	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test session
	session, err := chatSessionService.CreateSession("test-session", "Test session for testing", "")
	require.NoError(t, err)

	// Execute command with no parameters (should get active session)
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(map[string]string{}, "")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "Active session: test-session")
	assert.Contains(t, outputStr, session.ID)

	// Check that _session_id variable was set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	sessionID, err := variableService.Get("_session_id")
	assert.NoError(t, err)
	assert.Equal(t, session.ID, sessionID)
}

func TestGetCommand_Execute_FindSessionByExactName(t *testing.T) {
	cmd := &GetCommand{}
	setupSessionGetTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create multiple test sessions
	_, err = chatSessionService.CreateSession("project-alpha", "Alpha project session", "")
	require.NoError(t, err)

	session2, err := chatSessionService.CreateSession("project-beta", "Beta project session", "")
	require.NoError(t, err)

	// Execute command with exact name match
	args := map[string]string{"project-beta": ""}
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(args, "")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "Session activated: project-beta")
	assert.Contains(t, outputStr, session2.ID)

	// Check that _session_id variable was set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	sessionID, err := variableService.Get("_session_id")
	assert.NoError(t, err)
	assert.Equal(t, session2.ID, sessionID)

	// Verify that project-beta is now the active session
	activeSession, err := chatSessionService.GetActiveSession()
	assert.NoError(t, err)
	assert.Equal(t, "project-beta", activeSession.Name)
	assert.Equal(t, session2.ID, activeSession.ID)
}

func TestGetCommand_Execute_FindSessionByPrefix(t *testing.T) {
	cmd := &GetCommand{}
	setupSessionGetTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test sessions with unique prefixes
	_, err = chatSessionService.CreateSession("alpha-project", "Alpha project", "")
	require.NoError(t, err)

	session2, err := chatSessionService.CreateSession("beta-testing", "Beta testing session", "")
	require.NoError(t, err)

	// Execute command with unique prefix
	args := map[string]string{"beta": ""}
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(args, "")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "Session activated: beta-testing")
	assert.Contains(t, outputStr, session2.ID)

	// Check that _session_id variable was set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	sessionID, err := variableService.Get("_session_id")
	assert.NoError(t, err)
	assert.Equal(t, session2.ID, sessionID)
}

func TestGetCommand_Execute_SpaceSyntax(t *testing.T) {
	cmd := &GetCommand{}
	setupSessionGetTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test session
	session, err := chatSessionService.CreateSession("space-test", "Test session for space syntax", "")
	require.NoError(t, err)

	// Execute command using space syntax
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(map[string]string{}, "space-test")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "Session activated: space-test")
	assert.Contains(t, outputStr, session.ID)

	// Check that _session_id variable was set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	sessionID, err := variableService.Get("_session_id")
	assert.NoError(t, err)
	assert.Equal(t, session.ID, sessionID)
}

func TestGetCommand_Execute_NoActiveSession(t *testing.T) {
	cmd := &GetCommand{}
	setupSessionGetTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Ensure no active session
	ctx.SetActiveSessionID("")

	// Execute command with no parameters (should fail)
	err := cmd.Execute(map[string]string{}, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active session found")
	assert.Contains(t, err.Error(), "Usage:")
}

func TestGetCommand_Execute_SessionNotFound(t *testing.T) {
	cmd := &GetCommand{}
	setupSessionGetTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Execute command with non-existent session
	args := map[string]string{"nonexistent": ""}
	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session lookup failed")
	assert.Contains(t, err.Error(), "Usage:")
}

func TestGetCommand_Execute_AmbiguousPrefix(t *testing.T) {
	cmd := &GetCommand{}
	setupSessionGetTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create sessions with same prefix
	_, err = chatSessionService.CreateSession("project-alpha", "Alpha project", "")
	require.NoError(t, err)

	_, err = chatSessionService.CreateSession("project-beta", "Beta project", "")
	require.NoError(t, err)

	// Execute command with ambiguous prefix
	args := map[string]string{"project": ""}
	err = cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple sessions match prefix")
	assert.Contains(t, err.Error(), "project-alpha")
	assert.Contains(t, err.Error(), "project-beta")
}

func TestGetCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &GetCommand{}

	// Don't set up test registry - this will cause service not available error
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry()) // Empty registry

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
	}()

	args := map[string]string{"test": ""}
	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat session service not available")
}

func TestGetCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	cmd := &GetCommand{}

	// Set up registry with chat session service but no variable service
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Register only chat session service
	err := services.GetGlobalRegistry().RegisterService(services.NewChatSessionService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()

	// This should fail due to missing variable service
	err = cmd.Execute(map[string]string{}, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

// setupSessionGetTestRegistry creates a clean test registry for session-get command tests
func setupSessionGetTestRegistry(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Create a test context
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewChatSessionService())
	require.NoError(t, err)

	// Initialize all services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original state
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}
