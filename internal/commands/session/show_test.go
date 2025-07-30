package session

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestShowCommand_Name(t *testing.T) {
	cmd := &ShowCommand{}
	assert.Equal(t, "session-show", cmd.Name())
}

func TestShowCommand_Description(t *testing.T) {
	cmd := &ShowCommand{}
	description := cmd.Description()
	assert.Contains(t, description, "Display detailed session information")
	assert.Contains(t, description, "smart content rendering")
}

func TestShowCommand_ParseMode(t *testing.T) {
	cmd := &ShowCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestShowCommand_Usage(t *testing.T) {
	cmd := &ShowCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "\\session-show[id=false] session_text")
	assert.Contains(t, usage, "Examples:")
	assert.Contains(t, usage, "Options:")
}

func TestShowCommand_HelpInfo(t *testing.T) {
	cmd := &ShowCommand{}
	helpInfo := cmd.HelpInfo()
	assert.Equal(t, "session-show", helpInfo.Command)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)
	assert.Len(t, helpInfo.Options, 1)
	assert.Equal(t, "id", helpInfo.Options[0].Name)
}

// setupShowCommandTestRegistry creates a clean test registry for session-show command tests
func setupShowCommandTestRegistry(t *testing.T) {
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
	err = services.GetGlobalRegistry().RegisterService(services.NewThemeService())
	require.NoError(t, err)

	// Initialize all services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Clean up function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
	})
}

func TestShowCommand_Execute_NoSessions(t *testing.T) {
	cmd := &ShowCommand{}
	setupShowCommandTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	// Ensure no sessions exist
	ctx.SetActiveSessionID("")
	ctx.SetChatSessions(make(map[string]*neurotypes.ChatSession))

	// Test with no sessions
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(map[string]string{}, "")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "No sessions found. Use \\session-new to create a session.")
}

func TestShowCommand_Execute_ShowActiveSession(t *testing.T) {
	cmd := &ShowCommand{}
	setupShowCommandTestRegistry(t)

	ctx := context.GetGlobalContext()
	ctx.SetTestMode(true)

	chatSessionService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create and activate a session
	session, err := chatSessionService.CreateSession("active-session", "You are a helpful assistant", "")
	require.NoError(t, err)

	err = chatSessionService.SetActiveSession(session.ID)
	require.NoError(t, err)

	// Test showing active session
	var executeErr error
	outputStr := stringprocessing.CaptureOutput(func() {
		executeErr = cmd.Execute(map[string]string{}, "")
	})

	assert.NoError(t, executeErr)
	assert.Contains(t, outputStr, "Session: active-session")
	assert.Contains(t, outputStr, "System: You are a helpful assistant")
	assert.Contains(t, outputStr, "Messages: 0 total")
}

func TestShowCommand_TruncateContent_ShortContent(t *testing.T) {
	cmd := &ShowCommand{}
	content := "Short content"
	result := cmd.truncateContent(content, 100)
	assert.Equal(t, content, result)
}

func TestShowCommand_TruncateContent_LongContent(t *testing.T) {
	cmd := &ShowCommand{}
	longContent := strings.Repeat("a", 200) // 200 chars
	result := cmd.truncateContent(longContent, 100)

	// Should be truncated with ellipsis and char count
	assert.Contains(t, result, "...")
	assert.Contains(t, result, "(200 chars)")
	assert.True(t, len(result) > 100) // Result includes char count info
}

func TestShowCommand_TruncateContent_EdgeCase(t *testing.T) {
	cmd := &ShowCommand{}
	content := "ab"
	result := cmd.truncateContent(content, 1)
	assert.Equal(t, "...", result) // Very short max length
}

func TestShowCommand_RenderMessages_ManyMessages(t *testing.T) {
	cmd := &ShowCommand{}

	// Create many messages (more than MaxMessagesShown)
	messages := make([]neurotypes.Message, 15)
	now := time.Now()
	for i := 0; i < 15; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		messages[i] = neurotypes.Message{
			ID:        fmt.Sprintf("msg%d", i+1),
			Role:      role,
			Content:   fmt.Sprintf("Message content %d", i+1),
			Timestamp: now.Add(time.Duration(i) * time.Second),
		}
	}

	// Get theme service for test
	themeService, err := services.GetGlobalThemeService()
	require.NoError(t, err)
	theme := themeService.GetThemeByName("")

	outputStr := stringprocessing.CaptureOutput(func() {
		cmd.renderMessages(messages, theme)
	})

	// Should show first 5, separator, and last 5
	assert.Contains(t, outputStr, "[1] user")
	assert.Contains(t, outputStr, "[5] user")
	assert.Contains(t, outputStr, "... (5 more messages) ...")
	assert.Contains(t, outputStr, "[11] user")
	assert.Contains(t, outputStr, "[15] user")
}

func TestShowCommand_RenderSingleMessage_LongContent(t *testing.T) {
	cmd := &ShowCommand{}

	longContent := strings.Repeat("a", 300)
	msg := neurotypes.Message{
		ID:        "msg1",
		Role:      "user",
		Content:   longContent,
		Timestamp: time.Now(),
	}

	// Get theme service for test
	themeService, err := services.GetGlobalThemeService()
	require.NoError(t, err)
	theme := themeService.GetThemeByName("")

	outputStr := stringprocessing.CaptureOutput(func() {
		cmd.renderSingleMessage(1, msg, theme)
	})

	assert.Contains(t, outputStr, "[1] user")
	assert.Contains(t, outputStr, "...")
	assert.Contains(t, outputStr, "(300 chars)")
}

func TestShowCommand_FindSessionsByName(t *testing.T) {
	cmd := &ShowCommand{}

	sessions := []*neurotypes.ChatSession{
		{Name: "work-project", ID: "1"},
		{Name: "debug-session", ID: "2"},
		{Name: "work-analysis", ID: "3"},
	}

	// Test exact match
	matches := cmd.findSessionsByName(sessions, "debug-session")
	assert.Len(t, matches, 1)
	assert.Equal(t, "debug-session", matches[0].Name)

	// Test partial match
	matches = cmd.findSessionsByName(sessions, "work")
	assert.Len(t, matches, 2)

	// Test case insensitive
	matches = cmd.findSessionsByName(sessions, "WORK")
	assert.Len(t, matches, 2)

	// Test no match
	matches = cmd.findSessionsByName(sessions, "nonexistent")
	assert.Len(t, matches, 0)
}

func TestShowCommand_FindSessionsByIDPrefix(t *testing.T) {
	cmd := &ShowCommand{}

	sessions := []*neurotypes.ChatSession{
		{Name: "session1", ID: "abc12345"},
		{Name: "session2", ID: "abc67890"},
		{Name: "session3", ID: "def12345"},
	}

	// Test prefix match
	matches := cmd.findSessionsByIDPrefix(sessions, "abc")
	assert.Len(t, matches, 2)

	// Test specific match
	matches = cmd.findSessionsByIDPrefix(sessions, "abc123")
	assert.Len(t, matches, 1)
	assert.Equal(t, "session1", matches[0].Name)

	// Test case insensitive
	matches = cmd.findSessionsByIDPrefix(sessions, "ABC")
	assert.Len(t, matches, 2)

	// Test no match
	matches = cmd.findSessionsByIDPrefix(sessions, "xyz")
	assert.Len(t, matches, 0)
}
