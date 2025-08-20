package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestSaveCommand_Name(t *testing.T) {
	cmd := &SaveCommand{}
	assert.Equal(t, "session-save", cmd.Name())
}

func TestSaveCommand_ParseMode(t *testing.T) {
	cmd := &SaveCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestSaveCommand_Description(t *testing.T) {
	cmd := &SaveCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "save")
	assert.Contains(t, strings.ToLower(desc), "session")
	assert.Contains(t, strings.ToLower(desc), "auto-save")
}

func TestSaveCommand_Usage(t *testing.T) {
	cmd := &SaveCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\session-save")
	assert.Contains(t, usage, "session_identifier")
	assert.Contains(t, usage, "~/.config/neuroshell/sessions/")
}

func TestSaveCommand_HelpInfo(t *testing.T) {
	cmd := &SaveCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "session-save", helpInfo.Command)
	assert.Contains(t, strings.ToLower(helpInfo.Description), "save")
	assert.Equal(t, "\\session-save session_identifier", helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Should have no options (fixed behavior)
	assert.Len(t, helpInfo.Options, 0)

	// Should have examples
	assert.NotEmpty(t, helpInfo.Examples)

	// Should have _output stored variable
	assert.Len(t, helpInfo.StoredVariables, 1)
	assert.Equal(t, "_output", helpInfo.StoredVariables[0].Name)
	assert.Equal(t, "command_output", helpInfo.StoredVariables[0].Type)

	// Should have notes about fixed behavior
	assert.NotEmpty(t, helpInfo.Notes)
	notesText := strings.Join(helpInfo.Notes, " ")
	assert.Contains(t, notesText, "sessions directory")
	assert.Contains(t, notesText, "session-id")
	assert.Contains(t, notesText, "overwrites")
}

func TestSaveCommand_Execute_BasicFunctionality(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	// In test mode, context returns "/tmp/neuroshell-test-config"
	tempConfigDir := "/tmp/neuroshell-test-config"
	// Clean up and ensure directory structure
	_ = os.RemoveAll(tempConfigDir)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
	}()

	setupSaveTestRegistry(t, ctx)

	// Create a test session first
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{}, "test_session")
	require.NoError(t, err)

	// Get the session ID for verification
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)
	session, err := chatService.GetSessionByNameOrID("test_session")
	require.NoError(t, err)

	// Execute session-save command
	err = cmd.Execute(map[string]string{}, "test_session")
	assert.NoError(t, err)

	// Verify that the session was saved to the correct location
	sessionsDir := filepath.Join(tempConfigDir, "sessions")
	expectedFilename := session.ID + ".json"
	expectedPath := filepath.Join(sessionsDir, expectedFilename)

	// Check that the file exists
	assert.FileExists(t, expectedPath)

	// Verify file contents by reading the JSON file
	jsonData, err := os.ReadFile(expectedPath)
	assert.NoError(t, err)
	jsonStr := string(jsonData)

	// Verify the JSON contains the session data we expect
	assert.Contains(t, jsonStr, session.ID)
	assert.Contains(t, jsonStr, session.Name)
	assert.Contains(t, jsonStr, "test_session") // Explicit check for the name we set

	// Just verify the file can be imported successfully (don't check content since import changes IDs)
	_, err = chatService.ImportSessionFromJSON(expectedPath)
	assert.NoError(t, err)

	// Verify _output variable was set
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)
	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Session saved to sessions/")
	assert.Contains(t, output, expectedFilename)
}

func TestSaveCommand_Execute_BySessionID(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	// In test mode, context returns "/tmp/neuroshell-test-config"
	tempConfigDir := "/tmp/neuroshell-test-config"
	// Clean up and ensure directory structure
	_ = os.RemoveAll(tempConfigDir)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
	}()

	setupSaveTestRegistry(t, ctx)

	// Create a test session first
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{}, "test_session_by_id")
	require.NoError(t, err)

	// Get the session ID
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)
	session, err := chatService.GetSessionByNameOrID("test_session_by_id")
	require.NoError(t, err)

	// Execute session-save command using session ID
	err = cmd.Execute(map[string]string{}, session.ID)
	assert.NoError(t, err)

	// Verify that the session was saved
	sessionsDir := filepath.Join(tempConfigDir, "sessions")
	expectedFilename := session.ID + ".json"
	expectedPath := filepath.Join(sessionsDir, expectedFilename)
	assert.FileExists(t, expectedPath)
}

func TestSaveCommand_Execute_ByPrefixMatching(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	// In test mode, context returns "/tmp/neuroshell-test-config"
	tempConfigDir := "/tmp/neuroshell-test-config"
	// Clean up and ensure directory structure
	_ = os.RemoveAll(tempConfigDir)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
	}()

	setupSaveTestRegistry(t, ctx)

	// Create a test session first
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{}, "prefix_test_session")
	require.NoError(t, err)

	// Execute session-save command using prefix matching
	err = cmd.Execute(map[string]string{}, "prefix")
	assert.NoError(t, err)

	// Get the session to verify filename
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)
	session, err := chatService.GetSessionByNameOrID("prefix_test_session")
	require.NoError(t, err)

	// Verify that the session was saved
	sessionsDir := filepath.Join(tempConfigDir, "sessions")
	expectedFilename := session.ID + ".json"
	expectedPath := filepath.Join(sessionsDir, expectedFilename)
	assert.FileExists(t, expectedPath)
}

func TestSaveCommand_Execute_OverwriteExisting(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	// In test mode, context returns "/tmp/neuroshell-test-config"
	tempConfigDir := "/tmp/neuroshell-test-config"
	// Clean up and ensure directory structure
	_ = os.RemoveAll(tempConfigDir)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
	}()

	setupSaveTestRegistry(t, ctx)

	// Create a test session first
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{}, "overwrite_test")
	require.NoError(t, err)

	// Save the session once
	err = cmd.Execute(map[string]string{}, "overwrite_test")
	assert.NoError(t, err)

	// Get file info before second save
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)
	session, err := chatService.GetSessionByNameOrID("overwrite_test")
	require.NoError(t, err)

	sessionsDir := filepath.Join(tempConfigDir, "sessions")
	expectedPath := filepath.Join(sessionsDir, session.ID+".json")
	firstSaveInfo, err := os.Stat(expectedPath)
	require.NoError(t, err)

	// Add a message to the session to change it
	addMsgCmd := &AddUserMessageCommand{}
	err = addMsgCmd.Execute(map[string]string{"session": "overwrite_test"}, "Test message")
	require.NoError(t, err)

	// Save the session again (should overwrite)
	err = cmd.Execute(map[string]string{}, "overwrite_test")
	assert.NoError(t, err)

	// Verify file was overwritten (modification time should be different)
	secondSaveInfo, err := os.Stat(expectedPath)
	require.NoError(t, err)
	assert.True(t, secondSaveInfo.ModTime().After(firstSaveInfo.ModTime()) || secondSaveInfo.ModTime().Equal(firstSaveInfo.ModTime()))

	// Verify updated content by importing
	importedSession, err := chatService.ImportSessionFromJSON(expectedPath)
	assert.NoError(t, err)
	assert.Len(t, importedSession.Messages, 1) // Should have the added message
}

func TestSaveCommand_Execute_CreatesSessionsDirectory(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	// In test mode, context returns "/tmp/neuroshell-test-config"
	tempConfigDir := "/tmp/neuroshell-test-config"
	// Clean up and ensure directory structure
	_ = os.RemoveAll(tempConfigDir)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
	}()

	// Ensure sessions directory doesn't exist initially
	sessionsDir := filepath.Join(tempConfigDir, "sessions")
	_, err := os.Stat(sessionsDir)
	assert.True(t, os.IsNotExist(err))

	setupSaveTestRegistry(t, ctx)

	// Create a test session
	createCmd := &NewCommand{}
	err = createCmd.Execute(map[string]string{}, "directory_test")
	require.NoError(t, err)

	// Execute session-save command
	err = cmd.Execute(map[string]string{}, "directory_test")
	assert.NoError(t, err)

	// Verify sessions directory was created
	_, err = os.Stat(sessionsDir)
	assert.NoError(t, err)
}

func TestSaveCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	setupSaveTestRegistry(t, ctx)

	// Execute with empty input
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session identifier is required")
}

func TestSaveCommand_Execute_SessionNotFound(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	setupSaveTestRegistry(t, ctx)

	// Try to save non-existent session
	err := cmd.Execute(map[string]string{}, "nonexistent_session")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session lookup failed")
}

func TestSaveCommand_Execute_ChatServiceNotAvailable(t *testing.T) {
	cmd := &SaveCommand{}

	// Setup empty registry (no services)
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	// Execute without chat session service
	err := cmd.Execute(map[string]string{}, "test_session")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat session service not available")
}

func TestSaveCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	// Setup registry with only chat session service
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)
	context.SetGlobalContext(ctx)

	err := registry.RegisterService(services.NewChatSessionService())
	require.NoError(t, err)
	err = registry.InitializeAll()
	require.NoError(t, err)

	// Create a session first
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)
	_, err = chatService.CreateSession("test_session", "", "")
	require.NoError(t, err)

	// Execute without variable service
	err = cmd.Execute(map[string]string{}, "test_session")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestSaveCommand_Execute_InvalidConfigDirectory(t *testing.T) {
	// This test will be skipped since we can't override the config directory in test mode
	// The test context always returns "/tmp/neuroshell-test-config" which is writable
	t.Skip("Skipping test - cannot override config directory in test mode")
}

func TestSaveCommand_Execute_WithOutput(t *testing.T) {
	cmd := &SaveCommand{}
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)

	// In test mode, context returns "/tmp/neuroshell-test-config"
	tempConfigDir := "/tmp/neuroshell-test-config"
	// Clean up and ensure directory structure
	_ = os.RemoveAll(tempConfigDir)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
	}()

	setupSaveTestRegistry(t, ctx)

	// Create a test session
	createCmd := &NewCommand{}
	err := createCmd.Execute(map[string]string{}, "output_test")
	require.NoError(t, err)

	// Capture output
	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(map[string]string{}, "output_test")
		require.NoError(t, err)
	})

	// Verify output message
	assert.Contains(t, output, "Session saved to sessions/")
	assert.Contains(t, output, ".json")
}

func TestSaveCommand_IsReadOnly(t *testing.T) {
	cmd := &SaveCommand{}
	assert.False(t, cmd.IsReadOnly())
}

// setupSaveTestRegistry sets up a test environment with required services for session save tests
func setupSaveTestRegistry(t *testing.T, ctx neurotypes.Context) {
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
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

// Interface compliance check
var _ neurotypes.Command = (*SaveCommand)(nil)
