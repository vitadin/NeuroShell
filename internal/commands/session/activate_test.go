package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestActivateCommand_Name(t *testing.T) {
	cmd := &ActivateCommand{}
	assert.Equal(t, "session-activate", cmd.Name())
}

func TestActivateCommand_ParseMode(t *testing.T) {
	cmd := &ActivateCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestActivateCommand_Description(t *testing.T) {
	cmd := &ActivateCommand{}
	assert.Contains(t, cmd.Description(), "Activate session")
}

func TestActivateCommand_Usage(t *testing.T) {
	cmd := &ActivateCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "\\session-activate")
	assert.Contains(t, usage, "id=false")
	assert.Contains(t, usage, "id=true")
}

func TestActivateCommand_HelpInfo(t *testing.T) {
	cmd := &ActivateCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, cmd.Name(), helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, cmd.ParseMode(), helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.NotEmpty(t, helpInfo.Options)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Verify id option
	optionNames := make(map[string]bool)
	for _, option := range helpInfo.Options {
		optionNames[option.Name] = true
	}
	assert.True(t, optionNames["id"])
}

// Test no-parameter behavior: show active session
func TestActivateCommand_Execute_ShowActiveSession(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	// Create a test session and set it as active
	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test session
	session, err := chatSessionService.CreateSession("test-session", "Test session for testing", "")
	require.NoError(t, err)

	// Execute command with no parameters (should show active session)
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(map[string]string{}, "")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "Active session: test-session")
	assert.Contains(t, outputStr, session.ID[:8])

	// Check that _session_id variable was set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	sessionID, err := variableService.Get("_session_id")
	assert.NoError(t, err)
	assert.Equal(t, session.ID, sessionID)
}

// Test no-parameter behavior: no sessions exist
func TestActivateCommand_Execute_NoSessionsExist(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Ensure no active session and no sessions exist
	ctx.SetActiveSessionID("")
	ctx.SetChatSessions(make(map[string]*neurotypes.ChatSession))

	// Execute command with no parameters
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(map[string]string{}, "")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "No sessions found. Use \\session-new to create a session.")
}

// Test no-parameter behavior: auto-activate latest session
func TestActivateCommand_Execute_AutoActivateLatest(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create multiple sessions with different timestamps
	_, err = chatSessionService.CreateSession("old-session", "Old session", "")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // Ensure different timestamps

	session2, err := chatSessionService.CreateSession("latest-session", "Latest session", "")
	require.NoError(t, err)

	// Clear active session to trigger auto-activation
	ctx.SetActiveSessionID("")

	// Execute command with no parameters (should auto-activate latest)
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(map[string]string{}, "")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "No active session found. Activated most recent session: latest-session")
	assert.Contains(t, outputStr, session2.ID[:8])

	// Verify latest session is now active
	activeSession, err := chatSessionService.GetActiveSession()
	assert.NoError(t, err)
	assert.Equal(t, session2.ID, activeSession.ID)

	// Check activation variables were set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	activeName, err := variableService.Get("#active_session_name")
	assert.NoError(t, err)
	assert.Equal(t, "latest-session", activeName)

	activeID, err := variableService.Get("#active_session_id")
	assert.NoError(t, err)
	assert.Equal(t, session2.ID, activeID)
}

// Test activation by name (default behavior)
func TestActivateCommand_Execute_ActivateByName(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test sessions
	_, err = chatSessionService.CreateSession("project-alpha", "Alpha project session", "")
	require.NoError(t, err)

	session2, err := chatSessionService.CreateSession("project-beta", "Beta project session", "")
	require.NoError(t, err)

	// Execute command with name search (default)
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(map[string]string{}, "beta")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "Activated session 'project-beta'")
	assert.Contains(t, outputStr, session2.ID[:8])

	// Verify activation variables were set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	activeName, err := variableService.Get("#active_session_name")
	assert.NoError(t, err)
	assert.Equal(t, "project-beta", activeName)

	activeID, err := variableService.Get("#active_session_id")
	assert.NoError(t, err)
	assert.Equal(t, session2.ID, activeID)

	messageCount, err := variableService.Get("#active_session_message_count")
	assert.NoError(t, err)
	assert.Equal(t, "0", messageCount)
}

// Test activation by ID prefix
func TestActivateCommand_Execute_ActivateByID(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test sessions
	session1, err := chatSessionService.CreateSession("session1", "First session", "")
	require.NoError(t, err)

	_, err = chatSessionService.CreateSession("session2", "Second session", "")
	require.NoError(t, err)

	// Execute command with ID search
	idPrefix := session1.ID[:8]
	args := map[string]string{"id": "true"}
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(args, idPrefix)
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "Activated session 'session1'")
	assert.Contains(t, outputStr, session1.ID[:8])

	// Verify session1 is now active
	activeSession, err := chatSessionService.GetActiveSession()
	assert.NoError(t, err)
	assert.Equal(t, session1.ID, activeSession.ID)
}

// Test no matches for name search
func TestActivateCommand_Execute_NoMatchesName(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test session
	_, err = chatSessionService.CreateSession("test-session", "Test session", "")
	require.NoError(t, err)

	// Execute command with non-matching name
	err = cmd.Execute(map[string]string{}, "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No sessions found matching name 'nonexistent'")
	assert.Contains(t, err.Error(), "Available sessions:")
	assert.Contains(t, err.Error(), "test-session")
}

// Test no matches for ID search
func TestActivateCommand_Execute_NoMatchesID(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test session
	session, err := chatSessionService.CreateSession("test-session", "Test session", "")
	require.NoError(t, err)

	// Execute command with non-matching ID prefix
	args := map[string]string{"id": "true"}
	err = cmd.Execute(args, "xyz")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No sessions found matching ID prefix 'xyz'")
	assert.Contains(t, err.Error(), "Available sessions:")
	assert.Contains(t, err.Error(), session.ID[:8])
}

// Test multiple matches for name search
func TestActivateCommand_Execute_MultipleMatchesName(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create sessions with similar names
	_, err = chatSessionService.CreateSession("project-alpha", "Alpha project", "")
	require.NoError(t, err)

	_, err = chatSessionService.CreateSession("project-beta", "Beta project", "")
	require.NoError(t, err)

	// Execute command with ambiguous name
	err = cmd.Execute(map[string]string{}, "project")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Multiple sessions match name 'project'")
	assert.Contains(t, err.Error(), "Please be more specific:")
	assert.Contains(t, err.Error(), "project-alpha")
	assert.Contains(t, err.Error(), "project-beta")
	assert.Contains(t, err.Error(), "Tip: Use the full name")
}

// Test multiple matches for ID search
func TestActivateCommand_Execute_MultipleMatchesID(t *testing.T) {
	cmd := &ActivateCommand{}
	setupSessionActivateTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	_, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create sessions and manually set IDs with same prefix for testing
	sessions := ctx.GetChatSessions()

	session1 := &neurotypes.ChatSession{
		ID:        "abc12345-1111-1111-1111-111111111111",
		Name:      "session1",
		Messages:  []neurotypes.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	session2 := &neurotypes.ChatSession{
		ID:        "abc67890-2222-2222-2222-222222222222",
		Name:      "session2",
		Messages:  []neurotypes.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	sessions[session1.ID] = session1
	sessions[session2.ID] = session2
	ctx.SetChatSessions(sessions)

	// Execute command with ambiguous ID prefix
	args := map[string]string{"id": "true"}
	err = cmd.Execute(args, "abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Multiple sessions match ID prefix 'abc'")
	assert.Contains(t, err.Error(), "Please be more specific:")
	assert.Contains(t, err.Error(), "session1")
	assert.Contains(t, err.Error(), "session2")
}

// Test service unavailable errors
func TestActivateCommand_Execute_ChatSessionServiceNotAvailable(t *testing.T) {
	cmd := &ActivateCommand{}

	// Don't set up test registry - this will cause service not available error
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry()) // Empty registry

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
	}()

	err := cmd.Execute(map[string]string{}, "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat session service not available")
}

func TestActivateCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	cmd := &ActivateCommand{}

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

// Test helper methods
func TestActivateCommand_findSessionsByName(t *testing.T) {
	cmd := &ActivateCommand{}

	// Create test sessions
	sessions := []*neurotypes.ChatSession{
		{
			ID:   "id1",
			Name: "my-project-session",
		},
		{
			ID:   "id2",
			Name: "test-session",
		},
		{
			ID:   "id3",
			Name: "another-project-config",
		},
	}

	// Test exact match
	matches := cmd.findSessionsByName(sessions, "my-project-session")
	assert.Len(t, matches, 1)
	assert.Equal(t, "my-project-session", matches[0].Name)

	// Test partial match
	matches = cmd.findSessionsByName(sessions, "project")
	assert.Len(t, matches, 2)

	// Test case insensitive match
	matches = cmd.findSessionsByName(sessions, "TEST")
	assert.Len(t, matches, 1)
	assert.Equal(t, "test-session", matches[0].Name)

	// Test no match
	matches = cmd.findSessionsByName(sessions, "nonexistent")
	assert.Len(t, matches, 0)
}

func TestActivateCommand_findSessionsByIDPrefix(t *testing.T) {
	cmd := &ActivateCommand{}

	// Create test sessions
	sessions := []*neurotypes.ChatSession{
		{
			ID:   "abc12345",
			Name: "session1",
		},
		{
			ID:   "abc67890",
			Name: "session2",
		},
		{
			ID:   "def12345",
			Name: "session3",
		},
	}

	// Test unique prefix match
	matches := cmd.findSessionsByIDPrefix(sessions, "abc1")
	assert.Len(t, matches, 1)
	assert.Equal(t, "abc12345", matches[0].ID)

	// Test multiple prefix matches
	matches = cmd.findSessionsByIDPrefix(sessions, "abc")
	assert.Len(t, matches, 2)

	// Test case insensitive match
	matches = cmd.findSessionsByIDPrefix(sessions, "ABC1")
	assert.Len(t, matches, 1)
	assert.Equal(t, "abc12345", matches[0].ID)

	// Test no match
	matches = cmd.findSessionsByIDPrefix(sessions, "xyz")
	assert.Len(t, matches, 0)
}

func TestActivateCommand_findLatestSessionByTimestamp(t *testing.T) {
	cmd := &ActivateCommand{}

	now := time.Now()
	sessions := []*neurotypes.ChatSession{
		{
			ID:        "old",
			Name:      "old-session",
			UpdatedAt: now.Add(-time.Hour),
		},
		{
			ID:        "recent",
			Name:      "recent-session",
			UpdatedAt: now.Add(-time.Minute),
		},
		{
			ID:        "latest",
			Name:      "latest-session",
			UpdatedAt: now,
		},
	}

	latest := cmd.findLatestSessionByTimestamp(sessions)
	assert.NotNil(t, latest)
	assert.Equal(t, "latest", latest.ID)
	assert.Equal(t, "latest-session", latest.Name)
}

// setupSessionActivateTestRegistry creates a clean test registry for session-activate command tests
func setupSessionActivateTestRegistry(t *testing.T) {
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
