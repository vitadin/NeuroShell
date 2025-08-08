package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"

	"github.com/stretchr/testify/require"
)

func TestJSONImportCommand_Name(t *testing.T) {
	cmd := &JSONImportCommand{}
	if cmd.Name() != "session-json-import" {
		t.Errorf("Expected name 'session-json-import', got '%s'", cmd.Name())
	}
}

func setupJSONImportTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Save original registry and restore it after the test
	originalRegistry := services.GetGlobalRegistry()
	t.Cleanup(func() {
		services.SetGlobalRegistry(originalRegistry)
	})

	// Create a new registry for testing
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	neuroshellcontext.SetGlobalContext(ctx)

	// Register required services
	variableService := services.NewVariableService()
	err := variableService.Initialize()
	require.NoError(t, err)
	err = services.GetGlobalRegistry().RegisterService(variableService)
	require.NoError(t, err)

	chatService := services.NewChatSessionService()
	err = chatService.Initialize()
	require.NoError(t, err)
	err = services.GetGlobalRegistry().RegisterService(chatService)
	require.NoError(t, err)

	stackService := services.NewStackService()
	err = stackService.Initialize()
	require.NoError(t, err)
	err = services.GetGlobalRegistry().RegisterService(stackService)
	require.NoError(t, err)
}

func TestJSONImportCommand_Execute_Success(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONImportTestRegistry(t, ctx)

	// Get services
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test JSON data
	testSession := &neurotypes.ChatSession{
		ID:           "original-id",
		Name:         "original-session",
		SystemPrompt: "You are a helpful assistant",
		Messages: []neurotypes.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Hello there!",
				Timestamp: testutils.GetCurrentTime(ctx),
			},
			{
				ID:        "msg-2",
				Role:      "assistant",
				Content:   "Hello! How can I help you?",
				Timestamp: testutils.GetCurrentTime(ctx),
			},
		},
		CreatedAt: testutils.GetCurrentTime(ctx),
		UpdatedAt: testutils.GetCurrentTime(ctx),
		IsActive:  false,
	}

	// Write test session to JSON file
	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "test-import.json")

	jsonData, err := json.MarshalIndent(testSession, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test session: %v", err)
	}

	err = os.WriteFile(importFile, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	// Create command
	cmd := &JSONImportCommand{}

	// Execute command
	args := map[string]string{
		"file": importFile,
	}
	err = cmd.Execute(args, "")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify session was imported
	sessions := chatService.ListSessions()
	if len(sessions) != 1 {
		t.Fatalf("Expected 1 imported session, got %d", len(sessions))
	}

	importedSession := sessions[0]

	// Verify imported session has new identity
	if importedSession.ID == testSession.ID {
		t.Error("Imported session should have different ID")
	}
	if importedSession.Name == testSession.Name {
		t.Error("Imported session should have different name")
	}

	// Verify content was preserved
	if importedSession.SystemPrompt != testSession.SystemPrompt {
		t.Errorf("Expected system prompt '%s', got '%s'", testSession.SystemPrompt, importedSession.SystemPrompt)
	}
	if len(importedSession.Messages) != len(testSession.Messages) {
		t.Errorf("Expected %d messages, got %d", len(testSession.Messages), len(importedSession.Messages))
	}

	// Verify session is active
	if !importedSession.IsActive {
		t.Error("Imported session should be active")
	}

	// Verify output variable was set
	output, err := ctx.GetVariable("_output")
	if err != nil {
		t.Fatalf("Failed to get _output variable: %v", err)
	}

	expectedOutput := "Imported session as"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, output)
	}

	// Verify session variables were set
	sessionID, err := ctx.GetVariable("#session_id")
	if err != nil {
		t.Fatalf("Failed to get #session_id variable: %v", err)
	}
	if sessionID != importedSession.ID {
		t.Errorf("Expected session ID '%s', got '%s'", importedSession.ID, sessionID)
	}

	sessionName, err := ctx.GetVariable("#session_name")
	if err != nil {
		t.Fatalf("Failed to get #session_name variable: %v", err)
	}
	if sessionName != importedSession.Name {
		t.Errorf("Expected session name '%s', got '%s'", importedSession.Name, sessionName)
	}

	messageCount, err := ctx.GetVariable("#message_count")
	if err != nil {
		t.Fatalf("Failed to get #message_count variable: %v", err)
	}
	expectedCount := "2"
	if messageCount != expectedCount {
		t.Errorf("Expected message count '%s', got '%s'", expectedCount, messageCount)
	}
}

func TestJSONImportCommand_Execute_MissingFile(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONImportTestRegistry(t, ctx)

	// Create command
	cmd := &JSONImportCommand{}

	// Execute without file parameter
	args := map[string]string{}
	err := cmd.Execute(args, "")
	if err == nil {
		t.Fatal("Expected error for missing file parameter, got nil")
	}

	expectedError := "file path is required"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJSONImportCommand_Execute_FileNotFound(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONImportTestRegistry(t, ctx)

	// Create command
	cmd := &JSONImportCommand{}

	// Execute with non-existent file
	args := map[string]string{
		"file": "/nonexistent/file.json",
	}
	err := cmd.Execute(args, "")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}

	expectedError := "failed to read JSON file"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJSONImportCommand_Execute_InvalidJSON(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONImportTestRegistry(t, ctx)

	// Create file with invalid JSON
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.json")

	err := os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	// Create command
	cmd := &JSONImportCommand{}

	// Execute with invalid JSON file
	args := map[string]string{
		"file": invalidFile,
	}
	err = cmd.Execute(args, "")
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}

	expectedError := "failed to unmarshal JSON"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJSONImportCommand_Execute_MultipleImports(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONImportTestRegistry(t, ctx)

	// Get services
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test session data
	testSession := &neurotypes.ChatSession{
		ID:           "original-id",
		Name:         "original-session",
		SystemPrompt: "You are a coding assistant",
		Messages: []neurotypes.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Write a hello world program",
				Timestamp: testutils.GetCurrentTime(ctx),
			},
		},
		CreatedAt: testutils.GetCurrentTime(ctx),
		UpdatedAt: testutils.GetCurrentTime(ctx),
		IsActive:  false,
	}

	// Write test session to JSON file
	tempDir := t.TempDir()
	importFile := filepath.Join(tempDir, "test-import.json")

	jsonData, err := json.MarshalIndent(testSession, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test session: %v", err)
	}

	err = os.WriteFile(importFile, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	// Create command
	cmd := &JSONImportCommand{}
	args := map[string]string{
		"file": importFile,
	}

	// Import first session
	err = cmd.Execute(args, "")
	if err != nil {
		t.Fatalf("First import failed: %v", err)
	}

	// Import second session (same data, should get different name)
	err = cmd.Execute(args, "")
	if err != nil {
		t.Fatalf("Second import failed: %v", err)
	}

	// Verify both sessions were imported with different names
	sessions := chatService.ListSessions()
	if len(sessions) != 2 {
		t.Fatalf("Expected 2 imported sessions, got %d", len(sessions))
	}

	// Verify sessions have different names
	if sessions[0].Name == sessions[1].Name {
		t.Error("Imported sessions should have different auto-generated names")
	}

	// Verify content is the same
	for _, session := range sessions {
		if session.SystemPrompt != testSession.SystemPrompt {
			t.Errorf("Expected system prompt '%s', got '%s'", testSession.SystemPrompt, session.SystemPrompt)
		}
		if len(session.Messages) != len(testSession.Messages) {
			t.Errorf("Expected %d messages, got %d", len(testSession.Messages), len(session.Messages))
		}
	}

	// Verify the second imported session is active
	activeSession, err := chatService.GetActiveSession()
	if err != nil {
		t.Fatalf("Failed to get active session: %v", err)
	}

	// The active session should be one of the imported sessions
	found := false
	for _, session := range sessions {
		if session.ID == activeSession.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Active session should be one of the imported sessions")
	}
}

func TestJSONImportCommand_HelpInfo(t *testing.T) {
	cmd := &JSONImportCommand{}
	helpInfo := cmd.HelpInfo()

	if helpInfo.Command != "session-json-import" {
		t.Errorf("Expected command 'session-json-import', got '%s'", helpInfo.Command)
	}

	if len(helpInfo.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(helpInfo.Options))
	}

	if helpInfo.Options[0].Name != "file" {
		t.Errorf("Expected option 'file', got '%s'", helpInfo.Options[0].Name)
	}

	if !helpInfo.Options[0].Required {
		t.Error("Expected 'file' option to be required")
	}

	if len(helpInfo.Examples) == 0 {
		t.Error("Expected at least one example")
	}

	if len(helpInfo.StoredVariables) == 0 {
		t.Error("Expected at least one stored variable")
	}

	// Check for specific stored variables
	variableNames := make(map[string]bool)
	for _, variable := range helpInfo.StoredVariables {
		variableNames[variable.Name] = true
	}

	expectedVariables := []string{
		"#session_id", "#session_name", "#message_count", "#system_prompt", "#session_created",
		"#session_original_id", "#session_original_name", "#session_original_created", "_output",
	}
	for _, expectedVar := range expectedVariables {
		if !variableNames[expectedVar] {
			t.Errorf("Expected stored variable '%s' not found", expectedVar)
		}
	}
}
