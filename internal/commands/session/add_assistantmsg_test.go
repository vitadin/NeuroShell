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

func TestAddAssistantMessageCommand_Name(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	assert.Equal(t, "session-add-assistantmsg", cmd.Name())
}

func TestAddAssistantMessageCommand_ParseMode(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestAddAssistantMessageCommand_Description(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	description := cmd.Description()
	assert.NotEmpty(t, description)
	assert.Contains(t, strings.ToLower(description), "assistant message")
	assert.Contains(t, strings.ToLower(description), "session")
}

func TestAddAssistantMessageCommand_Usage(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "session-add-assistantmsg")
	assert.Contains(t, usage, "session=session_id")
	assert.Contains(t, usage, "response_content")
}

func TestAddAssistantMessageCommand_HelpInfo(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	help := cmd.HelpInfo()

	assert.Equal(t, "session-add-assistantmsg", help.Command)
	assert.NotEmpty(t, help.Description)
	assert.NotEmpty(t, help.Usage)

	// Check options
	assert.Len(t, help.Options, 1)
	assert.Equal(t, "session", help.Options[0].Name)
	assert.False(t, help.Options[0].Required)

	// Check examples
	assert.Len(t, help.Examples, 3)
	assert.Contains(t, help.Examples[0].Command, "session-add-assistantmsg")
	assert.Contains(t, help.Examples[1].Command, "${session_id}")
	assert.Contains(t, help.Examples[2].Command, "${_output}")

	// Check notes
	assert.Len(t, help.Notes, 8)
}

func TestAddAssistantMessageCommand_Execute_Success(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create a test session
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("test-session", "Test prompt", "")
	require.NoError(t, err)

	// Execute command to add assistant message
	args := map[string]string{"session": session.ID}
	response := "I'm doing well, thank you for asking!"

	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, response)
		assert.NoError(t, err)
	})

	// Check output confirmation
	assert.Contains(t, outputStr, "Added assistant message to session")
	assert.Contains(t, outputStr, session.Name)

	// Verify message was added to session
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 1)
	assert.Equal(t, "assistant", updatedSession.Messages[0].Role)
	assert.Equal(t, response, updatedSession.Messages[0].Content)
	assert.NotEmpty(t, updatedSession.Messages[0].ID)
	assert.NotZero(t, updatedSession.Messages[0].Timestamp)

	// Verify message history variable was updated
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	historyValue, err := variableService.Get("1")
	assert.NoError(t, err)
	assert.Equal(t, response, historyValue)
}

func TestAddAssistantMessageCommand_Execute_BySessionName(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
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
	response := "Here's the solution to your problem."

	err = cmd.Execute(args, response)
	assert.NoError(t, err)

	// Verify message was added
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 1)
	assert.Equal(t, "assistant", updatedSession.Messages[0].Role)
	assert.Equal(t, response, updatedSession.Messages[0].Content)
}

func TestAddAssistantMessageCommand_Execute_MultipleMessages(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create a test session
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("multi-session", "Test prompt", "")
	require.NoError(t, err)

	// Add multiple assistant messages
	responses := []string{
		"First response",
		"Second response",
		"Third response",
	}

	for _, response := range responses {
		args := map[string]string{"session": session.ID}
		err = cmd.Execute(args, response)
		assert.NoError(t, err)
	}

	// Verify all messages were added in correct order
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 3)

	for i, response := range responses {
		assert.Equal(t, "assistant", updatedSession.Messages[i].Role)
		assert.Equal(t, response, updatedSession.Messages[i].Content)
	}

	// Verify ${1} contains the latest (last) response
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	historyValue, err := variableService.Get("1")
	assert.NoError(t, err)
	assert.Equal(t, responses[len(responses)-1], historyValue)
}

func TestAddAssistantMessageCommand_Execute_MixedConversation(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	userCmd := &AddUserMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Create a test session
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("conversation", "Test prompt", "")
	require.NoError(t, err)

	// Simulate a conversation: user -> assistant -> user -> assistant
	args := map[string]string{"session": session.ID}

	// User message 1
	err = userCmd.Execute(args, "Hello, how are you?")
	assert.NoError(t, err)

	// Assistant response 1
	err = cmd.Execute(args, "I'm doing well, thank you!")
	assert.NoError(t, err)

	// User message 2
	err = userCmd.Execute(args, "Can you help me with coding?")
	assert.NoError(t, err)

	// Assistant response 2
	err = cmd.Execute(args, "Of course! What programming language?")
	assert.NoError(t, err)

	// Verify conversation structure
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 4)

	// Check message roles and content in order
	assert.Equal(t, "user", updatedSession.Messages[0].Role)
	assert.Equal(t, "Hello, how are you?", updatedSession.Messages[0].Content)

	assert.Equal(t, "assistant", updatedSession.Messages[1].Role)
	assert.Equal(t, "I'm doing well, thank you!", updatedSession.Messages[1].Content)

	assert.Equal(t, "user", updatedSession.Messages[2].Role)
	assert.Equal(t, "Can you help me with coding?", updatedSession.Messages[2].Content)

	assert.Equal(t, "assistant", updatedSession.Messages[3].Role)
	assert.Equal(t, "Of course! What programming language?", updatedSession.Messages[3].Content)

	// Verify ${1} contains the latest assistant response
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	historyValue, err := variableService.Get("1")
	assert.NoError(t, err)
	assert.Equal(t, "Of course! What programming language?", historyValue)
}

func TestAddAssistantMessageCommand_Execute_DefaultToActiveSession(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
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
	response := "I'm doing well, thank you!"

	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, response)
		assert.NoError(t, err)
	})

	// Check output confirmation mentions the active session
	assert.Contains(t, outputStr, "Added assistant message to session")
	assert.Contains(t, outputStr, session.Name)

	// Verify message was added to the active session
	updatedSession, err := chatService.GetSession(session.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession.Messages, 1)
	assert.Equal(t, "assistant", updatedSession.Messages[0].Role)
	assert.Equal(t, response, updatedSession.Messages[0].Content)

	// Verify message history variable was updated
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	historyValue, err := variableService.Get("1")
	assert.NoError(t, err)
	assert.Equal(t, response, historyValue)
}

func TestAddAssistantMessageCommand_Execute_NoActiveSessionError(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	setupAddMessageTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Ensure no active session
	ctx.SetActiveSessionID("")

	// Execute command without session parameter (should fail)
	args := map[string]string{} // No session parameter
	response := "Hello"

	err := cmd.Execute(args, response)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no session specified and no active session found")
	assert.Contains(t, err.Error(), "Usage:")
}

func TestAddAssistantMessageCommand_Execute_AutoActivation(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
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
	response := "Hello to session1"

	err = cmd.Execute(args, response)
	assert.NoError(t, err)

	// Verify session1 is now active
	activeSession, err = chatService.GetActiveSession()
	require.NoError(t, err)
	assert.Equal(t, session1.ID, activeSession.ID)

	// Verify session1 has the message
	updatedSession1, err := chatService.GetSession(session1.ID)
	require.NoError(t, err)
	assert.Len(t, updatedSession1.Messages, 1)
	assert.Equal(t, response, updatedSession1.Messages[0].Content)

	// Verify message history variable was updated
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	historyValue, err := variableService.Get("1")
	assert.NoError(t, err)
	assert.Equal(t, response, historyValue)
}

func TestAddAssistantMessageCommand_Execute_EmptyResponse(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	setupAddMessageTestRegistry(t)

	// Execute with empty response
	args := map[string]string{"session": "test-session"}
	response := ""

	err := cmd.Execute(args, response)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "response content is required")
	assert.Contains(t, err.Error(), "Usage:")
}

func TestAddAssistantMessageCommand_Execute_SessionNotFound(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}
	setupAddMessageTestRegistry(t)

	// Execute with non-existent session
	args := map[string]string{"session": "nonexistent-session"}
	response := "Hello"

	err := cmd.Execute(args, response)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find session")
	assert.Contains(t, err.Error(), "nonexistent-session")
}

func TestAddAssistantMessageCommand_Execute_ChatServiceNotAvailable(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}

	// Don't set up test registry - this will cause service not available error
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry()) // Empty registry

	defer func() {
		services.SetGlobalRegistry(oldRegistry)
	}()

	args := map[string]string{"session": "test-session"}
	response := "Hello"

	err := cmd.Execute(args, response)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat session service not available")
}

func TestAddAssistantMessageCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	cmd := &AddAssistantMessageCommand{}

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

	// Create a session first (this should work)
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	session, err := chatService.CreateSession("test-session", "Test prompt", "")
	require.NoError(t, err)

	// This should fail due to missing variable service
	args := map[string]string{"session": session.ID}
	response := "Hello"

	err = cmd.Execute(args, response)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}
