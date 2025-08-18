package session

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestDeleteMessageCommand_Name(t *testing.T) {
	cmd := &DeleteMessageCommand{}
	assert.Equal(t, "session-delete-msg", cmd.Name())
}

func TestDeleteMessageCommand_ParseMode(t *testing.T) {
	cmd := &DeleteMessageCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestDeleteMessageCommand_Description(t *testing.T) {
	cmd := &DeleteMessageCommand{}
	desc := cmd.Description()
	assert.Contains(t, desc, "Delete message by index")
	assert.Contains(t, desc, "dual indexing system")
}

func TestDeleteMessageCommand_Usage(t *testing.T) {
	cmd := &DeleteMessageCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "session-delete-msg")
	assert.Contains(t, usage, "idx=N")
	assert.Contains(t, usage, "idx=.N")
	assert.Contains(t, usage, "session=")
}

func TestDeleteMessageCommand_HelpInfo(t *testing.T) {
	cmd := &DeleteMessageCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "session-delete-msg", helpInfo.Command)
	assert.Contains(t, helpInfo.Description, "Delete message by index")
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Options)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.StoredVariables)
	assert.NotEmpty(t, helpInfo.Notes)

	// Check for required options
	var hasIdxOption bool
	for _, opt := range helpInfo.Options {
		if opt.Name == "idx" && opt.Required {
			hasIdxOption = true
			break
		}
	}
	assert.True(t, hasIdxOption, "Should have required idx option")
}

func TestDeleteMessageCommand_Execute_Success(t *testing.T) {
	tests := []struct {
		name              string
		idx               string
		confirm           string
		messageCount      int
		expectedIndex     int
		expectedRemaining int
	}{
		{
			name:              "reverse order - last message",
			idx:               "1",
			confirm:           "false",
			messageCount:      3,
			expectedIndex:     2, // Last message (0-based)
			expectedRemaining: 2,
		},
		{
			name:              "reverse order - second-to-last message",
			idx:               "2",
			confirm:           "false",
			messageCount:      3,
			expectedIndex:     1, // Second-to-last message (0-based)
			expectedRemaining: 2,
		},
		{
			name:              "normal order - first message",
			idx:               ".1",
			confirm:           "false",
			messageCount:      3,
			expectedIndex:     0, // First message (0-based)
			expectedRemaining: 2,
		},
		{
			name:              "normal order - second message",
			idx:               ".2",
			confirm:           "false",
			messageCount:      3,
			expectedIndex:     1, // Second message (0-based)
			expectedRemaining: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.New()
			setupDeleteMessageTestRegistry(t, ctx)

			chatService, err := services.GetGlobalChatSessionService()
			require.NoError(t, err)

			// Create test session with messages
			session, err := chatService.CreateSession("test-session", "Test system", "")
			require.NoError(t, err)

			// Add test messages
			for i := 0; i < tt.messageCount; i++ {
				role := "user"
				if i%2 == 1 {
					role = "assistant"
				}
				err = chatService.AddMessage(session.ID, role, fmt.Sprintf("Message %d content", i+1))
				require.NoError(t, err)
			}

			// Set as active session
			err = chatService.SetActiveSession(session.ID)
			require.NoError(t, err)

			// Get the original message that will be deleted
			updatedSession, err := chatService.GetSessionByNameOrID(session.ID)
			require.NoError(t, err)
			originalMessage := updatedSession.Messages[tt.expectedIndex]

			cmd := &DeleteMessageCommand{}
			args := map[string]string{
				"idx":     tt.idx,
				"confirm": tt.confirm,
			}

			// Execute command
			err = cmd.Execute(args, "")
			assert.NoError(t, err)

			// Verify message was deleted
			finalSession, err := chatService.GetSessionByNameOrID(session.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRemaining, len(finalSession.Messages))

			// Verify variables were set correctly
			deletedMessageID, err := ctx.GetVariable("#deleted_message_id")
			assert.NoError(t, err)
			assert.Equal(t, originalMessage.ID, deletedMessageID)

			deletedContent, err := ctx.GetVariable("#deleted_content")
			assert.NoError(t, err)
			assert.Equal(t, originalMessage.Content, deletedContent)

			remainingCount, err := ctx.GetVariable("#remaining_messages")
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("%d", tt.expectedRemaining), remainingCount)

			sessionName, err := ctx.GetVariable("#session_name")
			assert.NoError(t, err)
			assert.Equal(t, session.Name, sessionName)

			output, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "Deleted message")
			assert.Contains(t, output, tt.idx)
		})
	}
}

func TestDeleteMessageCommand_Execute_WithSpecificSession(t *testing.T) {
	// Setup
	ctx := context.New()
	setupDeleteMessageTestRegistry(t, ctx)

	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test session
	session, err := chatService.CreateSession("specific-session", "Test system", "")
	require.NoError(t, err)

	// Add messages
	err = chatService.AddMessage(session.ID, "user", "First message")
	require.NoError(t, err)
	err = chatService.AddMessage(session.ID, "assistant", "Second message")
	require.NoError(t, err)

	// Create another session as active to test session parameter
	activeSession, err := chatService.CreateSession("active-session", "Active system", "")
	require.NoError(t, err)
	err = chatService.SetActiveSession(activeSession.ID)
	require.NoError(t, err)

	cmd := &DeleteMessageCommand{}
	args := map[string]string{
		"idx":     "1",
		"session": session.Name,
		"confirm": "false",
	}

	// Execute command
	err = cmd.Execute(args, "")
	assert.NoError(t, err)

	// Verify correct session was modified
	updatedSession, err := chatService.GetSessionByNameOrID(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(updatedSession.Messages)) // One message deleted

	// Verify active session was not modified
	activeSessionUpdated, err := chatService.GetSessionByNameOrID(activeSession.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, len(activeSessionUpdated.Messages)) // Should still be empty
}

func TestDeleteMessageCommand_Execute_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (string, error) // Returns session ID
		args        map[string]string
		expectedErr string
	}{
		{
			name: "missing idx parameter",
			setupFunc: func() (string, error) {
				return createTestSessionWithMessages(3)
			},
			args: map[string]string{
				"confirm": "false",
			},
			expectedErr: "idx parameter is required",
		},
		{
			name: "invalid index format",
			setupFunc: func() (string, error) {
				return createTestSessionWithMessages(3)
			},
			args: map[string]string{
				"idx":     "invalid",
				"confirm": "false",
			},
			expectedErr: "invalid index 'invalid'",
		},
		{
			name: "index out of bounds - reverse order",
			setupFunc: func() (string, error) {
				return createTestSessionWithMessages(2)
			},
			args: map[string]string{
				"idx":     "5",
				"confirm": "false",
			},
			expectedErr: "reverse order index 5 is out of bounds",
		},
		{
			name: "index out of bounds - normal order",
			setupFunc: func() (string, error) {
				return createTestSessionWithMessages(2)
			},
			args: map[string]string{
				"idx":     ".5",
				"confirm": "false",
			},
			expectedErr: "normal order index 5 is out of bounds",
		},
		{
			name: "session with no messages",
			setupFunc: func() (string, error) {
				return createTestSessionWithMessages(0)
			},
			args: map[string]string{
				"idx":     "1",
				"confirm": "false",
			},
			expectedErr: "has no messages to delete",
		},
		{
			name: "non-existent session",
			setupFunc: func() (string, error) {
				return createTestSessionWithMessages(2)
			},
			args: map[string]string{
				"idx":     "1",
				"session": "non-existent",
				"confirm": "false",
			},
			expectedErr: "failed to find session 'non-existent'",
		},
		{
			name: "confirmation required (default)",
			setupFunc: func() (string, error) {
				return createTestSessionWithMessages(2)
			},
			args: map[string]string{
				"idx": "1",
				// No confirm parameter - should default to true
			},
			expectedErr: "deletion cancelled for safety",
		},
		{
			name: "explicit confirmation required",
			setupFunc: func() (string, error) {
				return createTestSessionWithMessages(2)
			},
			args: map[string]string{
				"idx":     "1",
				"confirm": "true",
			},
			expectedErr: "deletion cancelled for safety",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.New()
			setupDeleteMessageTestRegistry(t, ctx)

			sessionID, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			if sessionID != "" {
				chatService, err := services.GetGlobalChatSessionService()
				require.NoError(t, err)
				err = chatService.SetActiveSession(sessionID)
				require.NoError(t, err)
			}

			cmd := &DeleteMessageCommand{}

			// Execute command
			err = cmd.Execute(tt.args, "")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestDeleteMessageCommand_Execute_NoActiveSession(t *testing.T) {
	// Setup fresh context and services
	ctx := context.New()
	setupDeleteMessageTestRegistry(t, ctx)

	// Ensure no active session is set
	ctx.SetActiveSessionID("")

	cmd := &DeleteMessageCommand{}
	args := map[string]string{
		"idx":     "1",
		"confirm": "false",
	}

	// Execute command
	err := cmd.Execute(args, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no session specified and no active session found")
}

func TestDeleteMessageCommand_Execute_ServiceNotAvailable(t *testing.T) {
	// Setup without initializing services (clear any existing registry)
	services.SetGlobalRegistry(services.NewRegistry())

	cmd := &DeleteMessageCommand{}
	args := map[string]string{
		"idx":     "1",
		"confirm": "false",
	}

	// Execute command
	err := cmd.Execute(args, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not available")
}

func TestDeleteMessageCommand_truncateContent(t *testing.T) {
	tests := []struct {
		content   string
		maxLength int
		expected  string
	}{
		{"Short text", 20, "Short text"},
		{"This is a very long text that should be truncated", 20, "This is a very long ..."},
		{"", 10, ""},
		{"Exact", 5, "Exact"},
		{"TooLong", 5, "TooLo..."},
	}

	for _, tt := range tests {
		result := truncateContent(tt.content, tt.maxLength)
		assert.Equal(t, tt.expected, result)
	}
}

// Helper function to create a test session with specified number of messages
func createTestSessionWithMessages(messageCount int) (string, error) {
	chatService, err := services.GetGlobalChatSessionService()
	if err != nil {
		return "", err
	}

	session, err := chatService.CreateSession("test-session", "Test system", "")
	if err != nil {
		return "", err
	}

	for i := 0; i < messageCount; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		err = chatService.AddMessage(session.ID, role, fmt.Sprintf("Test message %d", i+1))
		if err != nil {
			return "", err
		}
	}

	return session.ID, nil
}

func TestDeleteMessageCommand_updateDeletedMessageVariables(t *testing.T) {
	// Setup
	ctx := context.New()
	setupDeleteMessageTestRegistry(t, ctx)

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	// Create test data
	timestamp := time.Date(2024, 1, 15, 14, 30, 25, 0, time.UTC)
	testMessage := &neurotypes.Message{
		ID:        "msg-123",
		Role:      "user",
		Content:   "Test message content",
		Timestamp: timestamp,
	}

	testSession := &neurotypes.ChatSession{
		ID:       "session-456",
		Name:     "test-session",
		Messages: []neurotypes.Message{}, // Empty after deletion
	}

	cmd := &DeleteMessageCommand{}

	// Execute
	err = cmd.updateDeletedMessageVariables(testSession, testMessage, "1", "last message", variableService)
	assert.NoError(t, err)

	// Verify variables using context
	expectedVars := map[string]string{
		"#deleted_message_id":        "msg-123",
		"#deleted_message_role":      "user",
		"#deleted_message_index":     "1",
		"#deleted_message_position":  "last message",
		"#deleted_content":           "Test message content",
		"#deleted_message_timestamp": "2024-01-15 14:30:25",
		"#session_id":                "session-456",
		"#session_name":              "test-session",
		"#remaining_messages":        "0",
	}

	for varName, expectedValue := range expectedVars {
		actualValue, err := ctx.GetVariable(varName)
		assert.NoError(t, err, "Failed to get variable %s", varName)
		assert.Equal(t, expectedValue, actualValue, "Variable %s has wrong value", varName)
	}
}

// setupDeleteMessageTestRegistry sets up a test environment with required services for delete message tests
func setupDeleteMessageTestRegistry(t *testing.T, ctx neurotypes.Context) {
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

	// Initialize all services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)
}
