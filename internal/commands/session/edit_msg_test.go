package session

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestEditMessageCommand_Name(t *testing.T) {
	cmd := &EditMessageCommand{}
	assert.Equal(t, "session-edit-msg", cmd.Name())
}

func TestEditMessageCommand_ParseMode(t *testing.T) {
	cmd := &EditMessageCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestEditMessageCommand_Description(t *testing.T) {
	cmd := &EditMessageCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "edit")
	assert.Contains(t, strings.ToLower(desc), "message")
	assert.Contains(t, strings.ToLower(desc), "index")
}

func TestEditMessageCommand_Usage(t *testing.T) {
	cmd := &EditMessageCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\session-edit-msg")
	assert.Contains(t, usage, "idx=")
	assert.Contains(t, usage, "session=")
}

func TestEditMessageCommand_HelpInfo(t *testing.T) {
	cmd := &EditMessageCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "session-edit-msg", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.Contains(t, helpInfo.Usage, "session-edit-msg")
	assert.Len(t, helpInfo.Options, 2) // idx and session
	assert.True(t, len(helpInfo.Examples) >= 4)
	assert.True(t, len(helpInfo.StoredVariables) >= 9)
	assert.True(t, len(helpInfo.Notes) >= 5)
}

func TestEditMessageCommand_Execute_ParameterValidation(t *testing.T) {
	cmd := &EditMessageCommand{}
	ctx := context.New()
	setupEditMessageTestRegistry(t, ctx)

	// Create a test session with messages
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Test assistant"}, "test_session")
	require.NoError(t, err)

	addUserCmd := &AddUserMessageCommand{}
	err = addUserCmd.Execute(map[string]string{}, "Test message")
	require.NoError(t, err)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing idx parameter",
			args:        map[string]string{},
			input:       "New content",
			expectError: true,
			errorMsg:    "idx parameter is required",
		},
		{
			name:        "empty input content",
			args:        map[string]string{"idx": "1"},
			input:       "",
			expectError: true,
			errorMsg:    "new message content is required",
		},
		{
			name:        "whitespace only input",
			args:        map[string]string{"idx": "1"},
			input:       "   \t\n   ",
			expectError: true,
			errorMsg:    "new message content is required",
		},
		{
			name:        "invalid idx format - empty dot",
			args:        map[string]string{"idx": "."},
			input:       "New content",
			expectError: true,
			errorMsg:    "invalid normal order index format",
		},
		{
			name:        "invalid idx format - non-numeric",
			args:        map[string]string{"idx": "abc"},
			input:       "New content",
			expectError: true,
			errorMsg:    "invalid reverse order index number",
		},
		{
			name:        "invalid idx format - non-numeric with dot",
			args:        map[string]string{"idx": ".abc"},
			input:       "New content",
			expectError: true,
			errorMsg:    "invalid normal order index number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)

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

func TestEditMessageCommand_Execute_ReverseOrderIndexing(t *testing.T) {
	cmd := &EditMessageCommand{}
	ctx := context.New()
	setupEditMessageTestRegistry(t, ctx)

	// Create session with multiple messages
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Test assistant"}, "reverse_test")
	require.NoError(t, err)

	addUserCmd := &AddUserMessageCommand{}
	addAssistantCmd := &AddAssistantMessageCommand{}

	// Add messages: user1, assistant1, user2, assistant2
	err = addUserCmd.Execute(map[string]string{}, "First user message")
	require.NoError(t, err)
	err = addAssistantCmd.Execute(map[string]string{}, "First assistant message")
	require.NoError(t, err)
	err = addUserCmd.Execute(map[string]string{}, "Second user message")
	require.NoError(t, err)
	err = addAssistantCmd.Execute(map[string]string{}, "Second assistant message")
	require.NoError(t, err)

	tests := []struct {
		name               string
		idx                string
		newContent         string
		expectedPos        string
		expectedOldContent string
	}{
		{
			name:               "edit last message (idx=1)",
			idx:                "1",
			newContent:         "Edited last message",
			expectedPos:        "last message",
			expectedOldContent: "Second assistant message",
		},
		{
			name:               "edit second-to-last message (idx=2)",
			idx:                "2",
			newContent:         "Edited second-to-last",
			expectedPos:        "second-to-last message",
			expectedOldContent: "Second user message",
		},
		{
			name:               "edit third-to-last message (idx=3)",
			idx:                "3",
			newContent:         "Edited third-to-last",
			expectedPos:        "third-to-last message",
			expectedOldContent: "First assistant message",
		},
		{
			name:               "edit fourth-to-last message (idx=4)",
			idx:                "4",
			newContent:         "Edited fourth-to-last",
			expectedPos:        "4th from last message",
			expectedOldContent: "First user message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset session for each test
			ctx = context.New()
			setupEditMessageTestRegistry(t, ctx)

			err := newCmd.Execute(map[string]string{"system": "Test assistant"}, "reverse_test")
			require.NoError(t, err)

			err = addUserCmd.Execute(map[string]string{}, "First user message")
			require.NoError(t, err)
			err = addAssistantCmd.Execute(map[string]string{}, "First assistant message")
			require.NoError(t, err)
			err = addUserCmd.Execute(map[string]string{}, "Second user message")
			require.NoError(t, err)
			err = addAssistantCmd.Execute(map[string]string{}, "Second assistant message")
			require.NoError(t, err)

			// Execute the edit command
			outputStr := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(map[string]string{"idx": tt.idx}, tt.newContent)
				assert.NoError(t, err)
			})

			// Verify output message
			assert.Contains(t, outputStr, fmt.Sprintf("Edited message %s (%s)", tt.idx, tt.expectedPos))

			// Verify stored variables
			oldContent, err := ctx.GetVariable("#old_content")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOldContent, oldContent)

			newContent, err := ctx.GetVariable("#new_content")
			assert.NoError(t, err)
			assert.Equal(t, tt.newContent, newContent)

			position, err := ctx.GetVariable("#message_position")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPos, position)

			messageIndex, err := ctx.GetVariable("#message_index")
			assert.NoError(t, err)
			assert.Equal(t, tt.idx, messageIndex)
		})
	}
}

func TestEditMessageCommand_Execute_NormalOrderIndexing(t *testing.T) {
	cmd := &EditMessageCommand{}
	ctx := context.New()
	setupEditMessageTestRegistry(t, ctx)

	// Create session with multiple messages
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Test assistant"}, "normal_test")
	require.NoError(t, err)

	addUserCmd := &AddUserMessageCommand{}
	addAssistantCmd := &AddAssistantMessageCommand{}

	// Add messages: user1, assistant1, user2, assistant2
	err = addUserCmd.Execute(map[string]string{}, "First user message")
	require.NoError(t, err)
	err = addAssistantCmd.Execute(map[string]string{}, "First assistant message")
	require.NoError(t, err)
	err = addUserCmd.Execute(map[string]string{}, "Second user message")
	require.NoError(t, err)
	err = addAssistantCmd.Execute(map[string]string{}, "Second assistant message")
	require.NoError(t, err)

	tests := []struct {
		name               string
		idx                string
		newContent         string
		expectedPos        string
		expectedOldContent string
	}{
		{
			name:               "edit first message (idx=.1)",
			idx:                ".1",
			newContent:         "Edited first message",
			expectedPos:        "first message",
			expectedOldContent: "First user message",
		},
		{
			name:               "edit second message (idx=.2)",
			idx:                ".2",
			newContent:         "Edited second message",
			expectedPos:        "second message",
			expectedOldContent: "First assistant message",
		},
		{
			name:               "edit third message (idx=.3)",
			idx:                ".3",
			newContent:         "Edited third message",
			expectedPos:        "third message",
			expectedOldContent: "Second user message",
		},
		{
			name:               "edit fourth message (idx=.4)",
			idx:                ".4",
			newContent:         "Edited fourth message",
			expectedPos:        "4th message",
			expectedOldContent: "Second assistant message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset session for each test
			ctx = context.New()
			setupEditMessageTestRegistry(t, ctx)

			err := newCmd.Execute(map[string]string{"system": "Test assistant"}, "normal_test")
			require.NoError(t, err)

			err = addUserCmd.Execute(map[string]string{}, "First user message")
			require.NoError(t, err)
			err = addAssistantCmd.Execute(map[string]string{}, "First assistant message")
			require.NoError(t, err)
			err = addUserCmd.Execute(map[string]string{}, "Second user message")
			require.NoError(t, err)
			err = addAssistantCmd.Execute(map[string]string{}, "Second assistant message")
			require.NoError(t, err)

			// Execute the edit command
			outputStr := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(map[string]string{"idx": tt.idx}, tt.newContent)
				assert.NoError(t, err)
			})

			// Verify output message
			assert.Contains(t, outputStr, fmt.Sprintf("Edited message %s (%s)", tt.idx, tt.expectedPos))

			// Verify stored variables
			oldContent, err := ctx.GetVariable("#old_content")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOldContent, oldContent)

			newContent, err := ctx.GetVariable("#new_content")
			assert.NoError(t, err)
			assert.Equal(t, tt.newContent, newContent)

			position, err := ctx.GetVariable("#message_position")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPos, position)

			messageIndex, err := ctx.GetVariable("#message_index")
			assert.NoError(t, err)
			assert.Equal(t, tt.idx, messageIndex)
		})
	}
}

func TestEditMessageCommand_Execute_OutOfBoundsErrors(t *testing.T) {
	cmd := &EditMessageCommand{}
	ctx := context.New()
	setupEditMessageTestRegistry(t, ctx)

	// Create session with 2 messages
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Test assistant"}, "bounds_test")
	require.NoError(t, err)

	addUserCmd := &AddUserMessageCommand{}
	err = addUserCmd.Execute(map[string]string{}, "First message")
	require.NoError(t, err)
	err = addUserCmd.Execute(map[string]string{}, "Second message")
	require.NoError(t, err)

	tests := []struct {
		name     string
		idx      string
		errorMsg string
	}{
		{
			name:     "reverse order out of bounds - too high",
			idx:      "3",
			errorMsg: "reverse order index 3 is out of bounds (session has 2 messages)",
		},
		{
			name:     "reverse order out of bounds - zero",
			idx:      "0",
			errorMsg: "reverse order index 0 is out of bounds (session has 2 messages)",
		},
		{
			name:     "normal order out of bounds - too high",
			idx:      ".3",
			errorMsg: "normal order index 3 is out of bounds (session has 2 messages)",
		},
		{
			name:     "normal order out of bounds - zero",
			idx:      ".0",
			errorMsg: "normal order index 0 is out of bounds (session has 2 messages)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(map[string]string{"idx": tt.idx}, "New content")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestEditMessageCommand_Execute_EmptySession(t *testing.T) {
	cmd := &EditMessageCommand{}
	ctx := context.New()
	setupEditMessageTestRegistry(t, ctx)

	// Create empty session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Test assistant"}, "empty_test")
	require.NoError(t, err)

	// Try to edit message in empty session
	err = cmd.Execute(map[string]string{"idx": "1"}, "New content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session 'empty_test' has no messages to edit")
}

func TestEditMessageCommand_Execute_SpecificSession(t *testing.T) {
	cmd := &EditMessageCommand{}
	ctx := context.New()
	setupEditMessageTestRegistry(t, ctx)

	// Create two sessions
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "First assistant"}, "session1")
	require.NoError(t, err)

	addUserCmd := &AddUserMessageCommand{}
	err = addUserCmd.Execute(map[string]string{}, "Message in session1")
	require.NoError(t, err)

	err = newCmd.Execute(map[string]string{"system": "Second assistant"}, "session2")
	require.NoError(t, err)

	err = addUserCmd.Execute(map[string]string{}, "Message in session2")
	require.NoError(t, err)

	// Edit message in specific session (session1)
	outputStr := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(map[string]string{"session": "session1", "idx": "1"}, "Edited session1 message")
		assert.NoError(t, err)
	})

	// Verify output references correct session
	assert.Contains(t, outputStr, "in session 'session1'")

	// Verify session variables
	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "session1", sessionName)

	oldContent, err := ctx.GetVariable("#old_content")
	assert.NoError(t, err)
	assert.Equal(t, "Message in session1", oldContent)
}

func TestEditMessageCommand_Execute_MessageMetadataPreservation(t *testing.T) {
	cmd := &EditMessageCommand{}
	ctx := context.New()
	ctx.SetTestMode(true)
	setupEditMessageTestRegistry(t, ctx)

	// Create session with message
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Test assistant"}, "metadata_test")
	require.NoError(t, err)

	addUserCmd := &AddUserMessageCommand{}
	err = addUserCmd.Execute(map[string]string{}, "Original message")
	require.NoError(t, err)

	// Get original message details
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	originalSession, err := chatService.GetActiveSession()
	require.NoError(t, err)
	originalMessage := originalSession.Messages[0]
	originalUpdatedAt := originalSession.UpdatedAt

	// Edit the message
	err = cmd.Execute(map[string]string{"idx": "1"}, "Edited message content")
	assert.NoError(t, err)

	// Get updated session and message
	updatedSession, err := chatService.GetActiveSession()
	require.NoError(t, err)
	updatedMessage := updatedSession.Messages[0]

	// Verify metadata preservation
	assert.Equal(t, originalMessage.ID, updatedMessage.ID, "Message ID should be preserved")
	assert.Equal(t, originalMessage.Role, updatedMessage.Role, "Message role should be preserved")
	assert.Equal(t, originalMessage.Timestamp, updatedMessage.Timestamp, "Message timestamp should be preserved")

	// Verify content was changed
	assert.NotEqual(t, originalMessage.Content, updatedMessage.Content, "Message content should be changed")
	assert.Equal(t, "Edited message content", updatedMessage.Content, "Message content should match new content")

	// Verify session UpdatedAt was updated
	assert.True(t, updatedSession.UpdatedAt.After(originalUpdatedAt), "Session UpdatedAt should be updated")
}

func TestEditMessageCommand_ParseMessageIndex(t *testing.T) {

	tests := []struct {
		name         string
		idxStr       string
		messageCount int
		expectedIdx  int
		expectedPos  string
		expectError  bool
		errorMsg     string
	}{
		// Reverse order tests
		{
			name:         "reverse order - last message",
			idxStr:       "1",
			messageCount: 5,
			expectedIdx:  4,
			expectedPos:  "last message",
			expectError:  false,
		},
		{
			name:         "reverse order - second-to-last",
			idxStr:       "2",
			messageCount: 5,
			expectedIdx:  3,
			expectedPos:  "second-to-last message",
			expectError:  false,
		},
		{
			name:         "reverse order - third-to-last",
			idxStr:       "3",
			messageCount: 5,
			expectedIdx:  2,
			expectedPos:  "third-to-last message",
			expectError:  false,
		},
		{
			name:         "reverse order - 5th from last",
			idxStr:       "5",
			messageCount: 10,
			expectedIdx:  5,
			expectedPos:  "5th from last message",
			expectError:  false,
		},
		// Normal order tests
		{
			name:         "normal order - first message",
			idxStr:       ".1",
			messageCount: 5,
			expectedIdx:  0,
			expectedPos:  "first message",
			expectError:  false,
		},
		{
			name:         "normal order - second message",
			idxStr:       ".2",
			messageCount: 5,
			expectedIdx:  1,
			expectedPos:  "second message",
			expectError:  false,
		},
		{
			name:         "normal order - third message",
			idxStr:       ".3",
			messageCount: 5,
			expectedIdx:  2,
			expectedPos:  "third message",
			expectError:  false,
		},
		{
			name:         "normal order - 5th message",
			idxStr:       ".5",
			messageCount: 10,
			expectedIdx:  4,
			expectedPos:  "5th message",
			expectError:  false,
		},
		// Error cases
		{
			name:         "reverse order - out of bounds high",
			idxStr:       "6",
			messageCount: 5,
			expectError:  true,
			errorMsg:     "reverse order index 6 is out of bounds",
		},
		{
			name:         "normal order - out of bounds high",
			idxStr:       ".6",
			messageCount: 5,
			expectError:  true,
			errorMsg:     "normal order index 6 is out of bounds",
		},
		{
			name:         "reverse order - zero",
			idxStr:       "0",
			messageCount: 5,
			expectError:  true,
			errorMsg:     "reverse order index 0 is out of bounds",
		},
		{
			name:         "normal order - zero",
			idxStr:       ".0",
			messageCount: 5,
			expectError:  true,
			errorMsg:     "normal order index 0 is out of bounds",
		},
		{
			name:         "invalid format - empty dot",
			idxStr:       ".",
			messageCount: 5,
			expectError:  true,
			errorMsg:     "invalid normal order index format",
		},
		{
			name:         "invalid format - non-numeric",
			idxStr:       "abc",
			messageCount: 5,
			expectError:  true,
			errorMsg:     "invalid reverse order index number",
		},
		{
			name:         "invalid format - non-numeric with dot",
			idxStr:       ".abc",
			messageCount: 5,
			expectError:  true,
			errorMsg:     "invalid normal order index number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := stringprocessing.ParseMessageIndex(tt.idxStr, tt.messageCount)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedIdx, result.ZeroBasedIndex)
				assert.Equal(t, tt.expectedPos, result.PositionDescription)
			}
		})
	}
}

func TestEditMessageCommand_GetOrdinalSuffix(t *testing.T) {
	tests := []struct {
		num      int
		expected string
	}{
		{1, "st"},
		{2, "nd"},
		{3, "rd"},
		{4, "th"},
		{5, "th"},
		{10, "th"},
		{11, "th"}, // Special case
		{12, "th"}, // Special case
		{13, "th"}, // Special case
		{21, "st"},
		{22, "nd"},
		{23, "rd"},
		{24, "th"},
		{101, "st"},
		{102, "nd"},
		{103, "rd"},
		{111, "th"}, // Special case
		{112, "th"}, // Special case
		{113, "th"}, // Special case
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ordinal_%d", tt.num), func(t *testing.T) {
			result := stringprocessing.GetOrdinalSuffix(tt.num)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// setupEditMessageTestRegistry sets up a test environment with required services for edit message tests
func setupEditMessageTestRegistry(t *testing.T, ctx neurotypes.Context) {
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

	err = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	require.NoError(t, err)

	// Initialize services
	for _, service := range services.GetGlobalRegistry().GetAllServices() {
		err := service.Initialize()
		require.NoError(t, err)
	}

	// Cleanup function
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})
}
