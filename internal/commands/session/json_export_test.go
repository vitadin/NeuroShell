package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

func TestJSONExportCommand_Name(t *testing.T) {
	cmd := &JSONExportCommand{}
	if cmd.Name() != "session-json-export" {
		t.Errorf("Expected name 'session-json-export', got '%s'", cmd.Name())
	}
}

func setupJSONExportTestRegistry(t *testing.T, ctx neurotypes.Context) {
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

func TestJSONExportCommand_Execute_Success(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONExportTestRegistry(t, ctx)

	// Get services
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create a test session
	_, err = chatService.CreateSession("test-session", "You are a test assistant", "Hello!")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create command
	cmd := &JSONExportCommand{}

	// Setup test file
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "export.json")

	// Execute command
	args := map[string]string{
		"file": exportFile,
	}
	err = cmd.Execute(args, "test-session")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Fatal("Export file was not created")
	}

	// Verify output variable was set
	output, err := ctx.GetVariable("_output")
	if err != nil {
		t.Fatalf("Failed to get _output variable: %v", err)
	}

	expectedOutput := "Exported session 'test-session'"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, output)
	}
}

func TestJSONExportCommand_Execute_PrefixMatching(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONExportTestRegistry(t, ctx)

	// Get services
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create test sessions
	_, err = chatService.CreateSession("project-alpha", "Assistant for project alpha", "")
	if err != nil {
		t.Fatalf("Failed to create session 1: %v", err)
	}

	_, err = chatService.CreateSession("project-beta", "Assistant for project beta", "")
	if err != nil {
		t.Fatalf("Failed to create session 2: %v", err)
	}

	// Create command
	cmd := &JSONExportCommand{}

	// Setup test file
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "export.json")

	// Test prefix matching - "project-a" should uniquely match "project-alpha"
	args := map[string]string{
		"file": exportFile,
	}
	err = cmd.Execute(args, "project-a")
	if err != nil {
		t.Fatalf("Execute with prefix matching failed: %v", err)
	}

	// Verify correct session was exported
	output, err := ctx.GetVariable("_output")
	if err != nil {
		t.Fatalf("Failed to get _output variable: %v", err)
	}

	if !strings.Contains(output, "project-alpha") {
		t.Errorf("Expected output to contain 'project-alpha', got '%s'", output)
	}
}

func TestJSONExportCommand_Execute_MissingFile(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONExportTestRegistry(t, ctx)

	// Create command
	cmd := &JSONExportCommand{}

	// Execute without file parameter
	args := map[string]string{}
	err := cmd.Execute(args, "test-session")
	if err == nil {
		t.Fatal("Expected error for missing file parameter, got nil")
	}

	expectedError := "file path is required"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJSONExportCommand_Execute_MissingSession(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONExportTestRegistry(t, ctx)

	// Create command
	cmd := &JSONExportCommand{}

	// Execute without session identifier
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "export.json")

	args := map[string]string{
		"file": exportFile,
	}
	err := cmd.Execute(args, "")
	if err == nil {
		t.Fatal("Expected error for missing session identifier, got nil")
	}

	expectedError := "session identifier is required"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJSONExportCommand_Execute_SessionNotFound(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONExportTestRegistry(t, ctx)

	// Create command
	cmd := &JSONExportCommand{}

	// Try to export non-existent session
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "export.json")

	args := map[string]string{
		"file": exportFile,
	}
	err := cmd.Execute(args, "nonexistent-session")
	if err == nil {
		t.Fatal("Expected error for non-existent session, got nil")
	}

	expectedError := "session lookup failed"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJSONExportCommand_Execute_AmbiguousPrefix(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context
	ctx := neuroshellcontext.NewTestContext()
	setupJSONExportTestRegistry(t, ctx)

	// Get services
	chatService, err := services.GetGlobalChatSessionService()
	require.NoError(t, err)

	// Create sessions with ambiguous prefixes
	_, err = chatService.CreateSession("test-alpha", "Assistant alpha", "")
	if err != nil {
		t.Fatalf("Failed to create session 1: %v", err)
	}

	_, err = chatService.CreateSession("test-beta", "Assistant beta", "")
	if err != nil {
		t.Fatalf("Failed to create session 2: %v", err)
	}

	// Create command
	cmd := &JSONExportCommand{}

	// Try to export with ambiguous prefix
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "export.json")

	args := map[string]string{
		"file": exportFile,
	}
	err = cmd.Execute(args, "test")
	if err == nil {
		t.Fatal("Expected error for ambiguous prefix, got nil")
	}

	expectedError := "multiple sessions match"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestJSONExportCommand_HelpInfo(t *testing.T) {
	cmd := &JSONExportCommand{}
	helpInfo := cmd.HelpInfo()

	if helpInfo.Command != "session-json-export" {
		t.Errorf("Expected command 'session-json-export', got '%s'", helpInfo.Command)
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
}
