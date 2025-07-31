package session

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

func TestCopyCommand_Name(t *testing.T) {
	cmd := &CopyCommand{}
	assert.Equal(t, "session-copy", cmd.Name())
}

func TestCopyCommand_ParseMode(t *testing.T) {
	cmd := &CopyCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestCopyCommand_Description(t *testing.T) {
	cmd := &CopyCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "copy")
	assert.Contains(t, strings.ToLower(desc), "session")
}

func TestCopyCommand_Usage(t *testing.T) {
	cmd := &CopyCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\session-copy")
	assert.Contains(t, usage, "source_session_id")
	assert.Contains(t, usage, "source_session_name")
	assert.Contains(t, usage, "target_session_name")
}

func TestCopyCommand_HelpInfo(t *testing.T) {
	cmd := &CopyCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "session-copy", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.Len(t, helpInfo.Options, 3) // source_session_id, source_session_name, target_session_name
	assert.True(t, len(helpInfo.Examples) >= 4)
	assert.True(t, len(helpInfo.StoredVariables) >= 8)
}

func TestCopyCommand_Execute_ParameterValidation(t *testing.T) {
	cmd := &CopyCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no source parameters",
			args:        map[string]string{"target_session_name": "copy_test"},
			expectError: true,
			errorMsg:    "must specify either source_session_id or source_session_name",
		},
		{
			name:        "both source parameters",
			args:        map[string]string{"source_session_id": "abc123", "source_session_name": "test"},
			expectError: true,
			errorMsg:    "cannot specify both source_session_id and source_session_name",
		},
		{
			name:        "only source_session_id",
			args:        map[string]string{"source_session_id": "nonexistent"},
			expectError: true, // Will fail because session doesn't exist
			errorMsg:    "source session lookup failed",
		},
		{
			name:        "only source_session_name",
			args:        map[string]string{"source_session_name": "nonexistent"},
			expectError: true, // Will fail because session doesn't exist
			errorMsg:    "source session lookup failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupCopyTestRegistry(t, ctx)
			err := cmd.Execute(tt.args, "")

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

func TestCopyCommand_Execute_CopyBySessionName(t *testing.T) {
	cmd := &CopyCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	// Create original session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "You are a test assistant"}, "original_session")
	require.NoError(t, err)

	// Add some messages to make the test more realistic
	addUserCmd := &AddUserMessageCommand{}
	err = addUserCmd.Execute(map[string]string{}, "Hello, test message!")
	require.NoError(t, err)

	addAssistantCmd := &AddAssistantMessageCommand{}
	err = addAssistantCmd.Execute(map[string]string{}, "Hello! This is a test response.")
	require.NoError(t, err)

	// Copy by session name with auto-generated target name
	outputStr := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(map[string]string{"source_session_name": "original_session"}, "")
		assert.NoError(t, err)
	})

	// Verify output message
	assert.Contains(t, outputStr, "Copied session 'original_session'")
	assert.Contains(t, outputStr, "with 2 messages")

	// Verify session variables were set
	sessionID, err := ctx.GetVariable("#session_id")
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionID)

	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionName)
	assert.NotEqual(t, "original_session", sessionName) // Should be different

	sourceSessionID, err := ctx.GetVariable("#source_session_id")
	assert.NoError(t, err)
	assert.NotEmpty(t, sourceSessionID)

	sourceSessionName, err := ctx.GetVariable("#source_session_name")
	assert.NoError(t, err)
	assert.Equal(t, "original_session", sourceSessionName)

	messageCount, err := ctx.GetVariable("#message_count")
	assert.NoError(t, err)
	assert.Equal(t, "2", messageCount)

	systemPrompt, err := ctx.GetVariable("#system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, "You are a test assistant", systemPrompt)
}

func TestCopyCommand_Execute_CopyBySessionID(t *testing.T) {
	cmd := &CopyCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	// Create original session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "You are a coding assistant"}, "code_session")
	require.NoError(t, err)

	// Get the session ID
	originalSessionID, err := ctx.GetVariable("#session_id")
	require.NoError(t, err)

	// Copy by session ID with custom target name
	outputStr := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(map[string]string{
			"source_session_id":   originalSessionID,
			"target_session_name": "code_session_copy",
		}, "")
		assert.NoError(t, err)
	})

	// Verify output message
	assert.Contains(t, outputStr, "Copied session 'code_session' to 'code_session_copy'")

	// Verify new session has different ID but same content
	newSessionID, err := ctx.GetVariable("#session_id")
	assert.NoError(t, err)
	assert.NotEqual(t, originalSessionID, newSessionID)

	newSessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "code_session_copy", newSessionName)

	systemPrompt, err := ctx.GetVariable("#system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, "You are a coding assistant", systemPrompt)
}

func TestCopyCommand_Execute_CustomTargetName(t *testing.T) {
	cmd := &CopyCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	// Create original session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "test_original")
	require.NoError(t, err)

	// Copy with custom target name
	err = cmd.Execute(map[string]string{
		"source_session_name": "test_original",
		"target_session_name": "custom_copy_name",
	}, "")
	assert.NoError(t, err)

	// Verify target name was used
	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "custom_copy_name", sessionName)
}

func TestCopyCommand_Execute_AutoGeneratedTargetName(t *testing.T) {
	cmd := &CopyCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	// Create original session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "test_original")
	require.NoError(t, err)

	// Copy with auto-generated target name
	err = cmd.Execute(map[string]string{"source_session_name": "test_original"}, "")
	assert.NoError(t, err)

	// Verify auto-generated name was used
	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.NotEmpty(t, sessionName)
	assert.NotEqual(t, "test_original", sessionName)
	// Should follow auto-generation pattern
	assert.True(t, strings.HasPrefix(sessionName, "Session ") ||
		strings.HasPrefix(sessionName, "Chat ") ||
		strings.HasPrefix(sessionName, "Work ") ||
		strings.HasPrefix(sessionName, "Project "))
}

func TestCopyCommand_Execute_DuplicateTargetName(t *testing.T) {
	cmd := &CopyCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	// Create original session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "original")
	require.NoError(t, err)

	// Create another session with name we'll try to use as target
	err = newCmd.Execute(map[string]string{}, "existing_target")
	require.NoError(t, err)

	// Try to copy with existing target name - should fail
	err = cmd.Execute(map[string]string{
		"source_session_name": "original",
		"target_session_name": "existing_target",
	}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in use")
}

func TestCopyCommand_Execute_DeepCopyVerification(t *testing.T) {
	cmd := &CopyCommand{}
	listCmd := &ListCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	// Create original session with multiple messages
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{"system": "Deep copy test assistant"}, "deep_copy_test")
	require.NoError(t, err)

	// Add multiple messages
	addUserCmd := &AddUserMessageCommand{}
	err = addUserCmd.Execute(map[string]string{}, "First user message")
	require.NoError(t, err)

	addAssistantCmd := &AddAssistantMessageCommand{}
	err = addAssistantCmd.Execute(map[string]string{}, "First assistant response")
	require.NoError(t, err)

	err = addUserCmd.Execute(map[string]string{}, "Second user message")
	require.NoError(t, err)

	err = addAssistantCmd.Execute(map[string]string{}, "Second assistant response")
	require.NoError(t, err)

	// Copy the session
	err = cmd.Execute(map[string]string{
		"source_session_name": "deep_copy_test",
		"target_session_name": "deep_copy_result",
	}, "")
	assert.NoError(t, err)

	// Verify both sessions exist
	outputStr := stringprocessing.CaptureOutput(func() {
		err := listCmd.Execute(map[string]string{}, "")
		assert.NoError(t, err)
	})

	assert.Contains(t, outputStr, "deep_copy_test")
	assert.Contains(t, outputStr, "deep_copy_result")

	// Verify copied session has correct message count
	messageCount, err := ctx.GetVariable("#message_count")
	assert.NoError(t, err)
	assert.Equal(t, "4", messageCount)

	// Verify system prompt was copied
	systemPrompt, err := ctx.GetVariable("#system_prompt")
	assert.NoError(t, err)
	assert.Equal(t, "Deep copy test assistant", systemPrompt)
}

func TestCopyCommand_Execute_MultipleCopiesFromSameSource(t *testing.T) {
	cmd := &CopyCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	// Reset test counters to ensure predictable auto-generated names
	testutils.ResetTestCounters()

	// Create original session
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "source_for_multiple_copies")
	require.NoError(t, err)

	// Create first copy
	err = cmd.Execute(map[string]string{
		"source_session_name": "source_for_multiple_copies",
		"target_session_name": "copy_1",
	}, "")
	assert.NoError(t, err)

	// Create second copy
	err = cmd.Execute(map[string]string{
		"source_session_name": "source_for_multiple_copies",
		"target_session_name": "copy_2",
	}, "")
	assert.NoError(t, err)

	// Verify all sessions exist and are different
	listCmd := &ListCommand{}
	outputStr := stringprocessing.CaptureOutput(func() {
		err := listCmd.Execute(map[string]string{}, "")
		assert.NoError(t, err)
	})

	assert.Contains(t, outputStr, "source_for_multiple_copies")
	assert.Contains(t, outputStr, "copy_1")
	assert.Contains(t, outputStr, "copy_2")

	// Verify the last copy is active
	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "copy_2", sessionName)
}

func TestCopyCommand_Execute_PrefixMatching(t *testing.T) {
	cmd := &CopyCommand{}
	ctx := context.New()
	setupCopyTestRegistry(t, ctx)

	// Create session with specific name for prefix testing
	newCmd := &NewCommand{}
	err := newCmd.Execute(map[string]string{}, "project_alpha_main")
	require.NoError(t, err)

	// Copy using prefix matching
	err = cmd.Execute(map[string]string{
		"source_session_name": "project_alpha", // Should match "project_alpha_main"
		"target_session_name": "prefix_copy_test",
	}, "")
	assert.NoError(t, err)

	// Verify source session name in variables
	sourceSessionName, err := ctx.GetVariable("#source_session_name")
	assert.NoError(t, err)
	assert.Equal(t, "project_alpha_main", sourceSessionName)

	// Verify target session name
	sessionName, err := ctx.GetVariable("#session_name")
	assert.NoError(t, err)
	assert.Equal(t, "prefix_copy_test", sessionName)
}

// setupSessionTestRegistry sets up a test environment with required services for session copy tests
func setupCopyTestRegistry(t *testing.T, ctx neurotypes.Context) {
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
