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

func TestAddUserMessageCommand_Name(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	assert.Equal(t, "session-add-usermsg", cmd.Name())
}

func TestAddUserMessageCommand_ParseMode(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestAddUserMessageCommand_Description(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	description := cmd.Description()
	assert.NotEmpty(t, description)
	assert.Contains(t, strings.ToLower(description), "user message")
	assert.Contains(t, strings.ToLower(description), "session")
}

func TestAddUserMessageCommand_Usage(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "session-add-usermsg")
	assert.Contains(t, usage, "session=session_id")
	assert.Contains(t, usage, "message_content")
}

func TestAddUserMessageCommand_HelpInfo(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	help := cmd.HelpInfo()

	assert.Equal(t, "session-add-usermsg", help.Command)
	assert.NotEmpty(t, help.Description)
	assert.NotEmpty(t, help.Usage)

	// Check options
	assert.Len(t, help.Options, 1)
	assert.Equal(t, "session", help.Options[0].Name)
	assert.False(t, help.Options[0].Required)

	// Check examples
	assert.Len(t, help.Examples, 3)
	assert.Contains(t, help.Examples[0].Command, "session-add-usermsg")
	assert.Contains(t, help.Examples[1].Command, "${session_id}")
	assert.Contains(t, help.Examples[2].Command, "work-session")

	// Check notes
	assert.Len(t, help.Notes, 7)
}

func TestAddUserMessageCommand_Execute_Success(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create a test session
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("test-session", "Test prompt", "")
	require.NoError(t, err)

	// Execute command to add user message
	args := map[string]string{"session": session.ID}
	message := "Hello, how are you today?"

	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, message)
		assert.NoError(t, err)
	})

	// Check output confirmation
	assert.Contains(t, outputStr, "Added user message to session")
	assert.Contains(t, outputStr, session.Name)

	// Verify message was added to session
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 1)
	assert.Equal(t, "user", updatedSession.Messages[0].Role)
	assert.Equal(t, message, updatedSession.Messages[0].Content)
	assert.NotEmpty(t, updatedSession.Messages[0].ID)
	assert.NotZero(t, updatedSession.Messages[0].Timestamp)
}

func TestAddUserMessageCommand_Execute_BySessionName(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create a test session
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("work-session", "Test prompt", "")
	require.NoError(t, err)

	// Execute command using session name instead of ID
	args := map[string]string{"session": "work-session"}
	message := "Can you help me with this task?"

	err = cmd.Execute(args, message)
	assert.NoError(t, err)

	// Verify message was added
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 1)
	assert.Equal(t, "user", updatedSession.Messages[0].Role)
	assert.Equal(t, message, updatedSession.Messages[0].Content)
}

func TestAddUserMessageCommand_Execute_MultipleMessages(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create a test session
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("multi-session", "Test prompt", "")
	require.NoError(t, err)

	// Add multiple user messages
	messages := []string{
		"First message",
		"Second message",
		"Third message",
	}

	for _, message := range messages {
		args := map[string]string{"session": session.ID}
		err = cmd.Execute(args, message)
		assert.NoError(t, err)
	}

	// Verify all messages were added in correct order
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 3)

	for i, message := range messages {
		assert.Equal(t, "user", updatedSession.Messages[i].Role)
		assert.Equal(t, message, updatedSession.Messages[i].Content)
	}
}

func TestAddUserMessageCommand_Execute_EmptyMessage(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	// Execute with empty message
	args := map[string]string{"session": "test-session"}
	message := ""

	err := cmd.Execute(args, message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message content is required")
	assert.Contains(t, err.Error(), "Usage:")
}

func TestAddUserMessageCommand_Execute_SessionNotFound(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	// Execute with non-existent session
	args := map[string]string{"session": "nonexistent-session"}
	message := "Hello"

	err := cmd.Execute(args, message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find session")
	assert.Contains(t, err.Error(), "nonexistent-session")
}

func TestAddUserMessageCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &AddUserMessageCommand{}

	// Don't set up test registry - this will cause service not available error
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry()) // Empty registry

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
	}()

	args := map[string]string{"session": "test-session"}
	message := "Hello"

	err := cmd.Execute(args, message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat session service not available")
}

func TestAddUserMessageCommand_Execute_DefaultToActiveSession(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create a test session and make it active
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("active-session", "Test prompt", "")
	require.NoError(t, err)

	// Execute command without session parameter (should use active session)
	args := map[string]string{} // No session parameter
	message := "Hello from active session!"

	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, message)
		assert.NoError(t, err)
	})

	// Check output confirmation mentions the active session
	assert.Contains(t, outputStr, "Added user message to session")
	assert.Contains(t, outputStr, session.Name)

	// Verify message was added to the active session
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 1)
	assert.Equal(t, "user", updatedSession.Messages[0].Role)
	assert.Equal(t, message, updatedSession.Messages[0].Content)
}

func TestAddUserMessageCommand_Execute_NoActiveSessionError(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Ensure no active session
	ctx.SetActiveSessionID("")

	// Execute command without session parameter (should fail)
	args := map[string]string{} // No session parameter
	message := "Hello"

	err := cmd.Execute(args, message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no session specified and no active session found")
	assert.Contains(t, err.Error(), "Usage:")
}

func TestAddUserMessageCommand_Execute_AutoActivation(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create two sessions
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session1, err := chatService.CreateSession("session1", "Test prompt", "")
	require.NoError(t, err)

	session2, err := chatService.CreateSession("session2", "Test prompt", "")
	require.NoError(t, err)

	// session2 should be active (last created)
	activeSession, err := chatService.GetActiveSession()
	require.NoError(t, err)
	assert.Equal(t, session2.ID, activeSession.ID)

	// Add message to session1 (should make it active)
	args := map[string]string{"session": session1.ID}
	message := "Hello to session1"

	err = cmd.Execute(args, message)
	assert.NoError(t, err)

	// Verify session1 is now active
	activeSession, err = chatService.GetActiveSession()
	require.NoError(t, err)
	assert.Equal(t, session1.ID, activeSession.ID)

	// Verify session1 has the message
	updatedSession1, err := chatService.GetSession(session1.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession1.Messages, 1)
	assert.Equal(t, message, updatedSession1.Messages[0].Content)
}

func TestAddUserMessageCommand_Execute_WhitespaceHandling(t *testing.T) {
	cmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create a test session
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("whitespace-test", "Test prompt", "")
	require.NoError(t, err)

	// Test message with leading/trailing whitespace (should be preserved)
	args := map[string]string{"session": session.ID}
	message := "  Hello with spaces  "

	err = cmd.Execute(args, message)
	assert.NoError(t, err)

	// Verify message preserves whitespace
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 1)
	assert.Equal(t, message, updatedSession.Messages[0].Content)
}

// setupAddMessageTestRegistry creates a clean test registry for add message command tests
func setupAddMessageTestRegistry(t *testing.T) {
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
