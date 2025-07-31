package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

func TestChatSessionService_ExportSessionToJSON(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context with test mode enabled
	ctx := neuroshellcontext.NewTestContext()

	// Create service
	service := NewChatSessionService()
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Create a test session
	session, err := service.CreateSession("test-session", "You are a test assistant", "Hello!")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add more messages to make it interesting
	err = service.AddMessageWithContext(session.ID, "assistant", "Hello! How can I help you?", ctx)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Create temporary file for export
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "test-export.json")

	// Test export
	err = service.ExportSessionToJSONWithContext(session.ID, exportFile, ctx)
	if err != nil {
		t.Fatalf("Failed to export session: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(exportFile); os.IsNotExist(err) {
		t.Fatalf("Export file was not created")
	}

	// Read and verify exported content
	exportedData, err := os.ReadFile(exportFile)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	var exportedSession neurotypes.ChatSession
	err = json.Unmarshal(exportedData, &exportedSession)
	if err != nil {
		t.Fatalf("Failed to unmarshal exported JSON: %v", err)
	}

	// Verify exported data matches original session
	if exportedSession.ID != session.ID {
		t.Errorf("Expected ID %s, got %s", session.ID, exportedSession.ID)
	}
	if exportedSession.Name != session.Name {
		t.Errorf("Expected name %s, got %s", session.Name, exportedSession.Name)
	}
	if exportedSession.SystemPrompt != session.SystemPrompt {
		t.Errorf("Expected system prompt %s, got %s", session.SystemPrompt, exportedSession.SystemPrompt)
	}
	if len(exportedSession.Messages) != len(session.Messages) {
		t.Errorf("Expected %d messages, got %d", len(session.Messages), len(exportedSession.Messages))
	}
}

func TestChatSessionService_ExportSessionToJSON_InvalidSession(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()

	// Create service
	service := NewChatSessionService()
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Try to export non-existent session
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "test-export.json")

	err = service.ExportSessionToJSONWithContext("invalid-id", exportFile, ctx)
	if err == nil {
		t.Fatal("Expected error for invalid session ID, got nil")
	}
}

func TestChatSessionService_ExportSessionToJSON_InvalidPath(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context
	ctx := neuroshellcontext.NewTestContext()

	// Create service and session
	service := NewChatSessionService()
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	session, err := service.CreateSession("test-session", "You are a test assistant", "Hello!")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Try to export to invalid path (directory that doesn't exist)
	invalidPath := "/nonexistent/directory/export.json"

	err = service.ExportSessionToJSONWithContext(session.ID, invalidPath, ctx)
	if err == nil {
		t.Fatal("Expected error for invalid file path, got nil")
	}
}

func TestChatSessionService_ImportSessionFromJSON(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context with test mode enabled
	ctx := neuroshellcontext.NewTestContext()

	// Create service
	service := NewChatSessionService()
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Create a test session to export first
	originalSession, err := service.CreateSession("original-session", "You are a helpful assistant", "Test message")
	if err != nil {
		t.Fatalf("Failed to create original session: %v", err)
	}

	// Add another message
	err = service.AddMessageWithContext(originalSession.ID, "assistant", "I'm ready to help!", ctx)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Export the session
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "test-export.json")

	err = service.ExportSessionToJSONWithContext(originalSession.ID, exportFile, ctx)
	if err != nil {
		t.Fatalf("Failed to export session: %v", err)
	}

	// Now import the session
	importedSession, err := service.ImportSessionFromJSONWithContext(exportFile, ctx)
	if err != nil {
		t.Fatalf("Failed to import session: %v", err)
	}

	// Verify imported session has new identity but preserves content
	if importedSession.ID == originalSession.ID {
		t.Error("Imported session should have different ID")
	}
	if importedSession.Name == originalSession.Name {
		t.Error("Imported session should have different name")
	}
	if importedSession.SystemPrompt != originalSession.SystemPrompt {
		t.Errorf("Expected system prompt %s, got %s", originalSession.SystemPrompt, importedSession.SystemPrompt)
	}
	if len(importedSession.Messages) != len(originalSession.Messages) {
		t.Errorf("Expected %d messages, got %d", len(originalSession.Messages), len(importedSession.Messages))
	}

	// Verify messages are preserved exactly
	for i, originalMsg := range originalSession.Messages {
		importedMsg := importedSession.Messages[i]
		if importedMsg.Role != originalMsg.Role {
			t.Errorf("Message %d role mismatch: expected %s, got %s", i, originalMsg.Role, importedMsg.Role)
		}
		if importedMsg.Content != originalMsg.Content {
			t.Errorf("Message %d content mismatch: expected %s, got %s", i, originalMsg.Content, importedMsg.Content)
		}
		// Timestamps should be preserved from original
		if !importedMsg.Timestamp.Equal(originalMsg.Timestamp) {
			t.Errorf("Message %d timestamp should be preserved", i)
		}
	}

	// Verify session is active
	if !importedSession.IsActive {
		t.Error("Imported session should be active")
	}

	// Verify session was properly stored in context
	storedSession, err := service.GetSessionWithContext(importedSession.ID, ctx)
	if err != nil {
		t.Fatalf("Failed to retrieve imported session: %v", err)
	}
	if storedSession.Name != importedSession.Name {
		t.Errorf("Stored session name mismatch: expected %s, got %s", importedSession.Name, storedSession.Name)
	}
}

func TestChatSessionService_ImportSessionFromJSON_InvalidFile(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()

	// Create service
	service := NewChatSessionService()
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Try to import non-existent file
	_, err = service.ImportSessionFromJSONWithContext("/nonexistent/file.json", ctx)
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
}

func TestChatSessionService_ImportSessionFromJSON_InvalidJSON(t *testing.T) {
	// Setup test context
	ctx := neuroshellcontext.NewTestContext()

	// Create service
	service := NewChatSessionService()
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Create a file with invalid JSON
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.json")

	err = os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	// Try to import invalid JSON
	_, err = service.ImportSessionFromJSONWithContext(invalidFile, ctx)
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestChatSessionService_ExportImportRoundtrip(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context
	ctx := neuroshellcontext.NewTestContext()

	// Create service
	service := NewChatSessionService()
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Create a comprehensive test session
	originalSession, err := service.CreateSession("complex-session", "You are a coding assistant specialized in Go", "")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add multiple messages with different roles
	messages := []struct {
		role    string
		content string
	}{
		{"user", "Help me write a function to reverse a string"},
		{"assistant", "I'll help you write a string reversal function in Go."},
		{"user", "Can you make it more efficient?"},
		{"assistant", "Here's an optimized version using byte manipulation."},
	}

	for _, msg := range messages {
		err = service.AddMessageWithContext(originalSession.ID, msg.role, msg.content, ctx)
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}
	}

	// Export session
	tempDir := t.TempDir()
	exportFile := filepath.Join(tempDir, "roundtrip-export.json")

	err = service.ExportSessionToJSONWithContext(originalSession.ID, exportFile, ctx)
	if err != nil {
		t.Fatalf("Failed to export session: %v", err)
	}

	// Import session
	importedSession, err := service.ImportSessionFromJSONWithContext(exportFile, ctx)
	if err != nil {
		t.Fatalf("Failed to import session: %v", err)
	}

	// Verify all content is preserved (except identity)
	if importedSession.SystemPrompt != originalSession.SystemPrompt {
		t.Errorf("System prompt not preserved: expected %s, got %s", originalSession.SystemPrompt, importedSession.SystemPrompt)
	}

	if len(importedSession.Messages) != len(originalSession.Messages) {
		t.Errorf("Message count mismatch: expected %d, got %d", len(originalSession.Messages), len(importedSession.Messages))
	}

	// Check each message in detail
	for i, originalMsg := range originalSession.Messages {
		if i >= len(importedSession.Messages) {
			t.Fatalf("Imported session missing message %d", i)
		}

		importedMsg := importedSession.Messages[i]

		if importedMsg.Role != originalMsg.Role {
			t.Errorf("Message %d role mismatch: expected %s, got %s", i, originalMsg.Role, importedMsg.Role)
		}
		if importedMsg.Content != originalMsg.Content {
			t.Errorf("Message %d content mismatch: expected %s, got %s", i, originalMsg.Content, importedMsg.Content)
		}
		if !importedMsg.Timestamp.Equal(originalMsg.Timestamp) {
			t.Errorf("Message %d timestamp not preserved", i)
		}
	}

	// Verify new identity was assigned
	if importedSession.ID == originalSession.ID {
		t.Error("Imported session should have different ID")
	}
	if importedSession.Name == originalSession.Name {
		t.Error("Imported session should have different name")
	}
	if importedSession.CreatedAt.Equal(originalSession.CreatedAt) {
		t.Error("Imported session should have new creation timestamp")
	}
	if importedSession.UpdatedAt.Equal(originalSession.UpdatedAt) {
		t.Error("Imported session should have new update timestamp")
	}

	// Verify session is active and properly stored
	if !importedSession.IsActive {
		t.Error("Imported session should be active")
	}

	// Check that both sessions can coexist
	storedOriginal, err := service.GetSessionWithContext(originalSession.ID, ctx)
	if err != nil {
		t.Fatalf("Original session should still exist: %v", err)
	}
	if storedOriginal.Name != originalSession.Name {
		t.Error("Original session should be unchanged")
	}

	storedImported, err := service.GetSessionWithContext(importedSession.ID, ctx)
	if err != nil {
		t.Fatalf("Imported session should be stored: %v", err)
	}
	if storedImported.Name != importedSession.Name {
		t.Error("Imported session should be properly stored")
	}
}

func TestChatSessionService_ReconstructImportedSession(t *testing.T) {
	// Reset test state
	testutils.ResetTestCounters()

	// Setup test context
	ctx := neuroshellcontext.NewTestContext()

	// Create service
	service := NewChatSessionService()
	err := service.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize service: %v", err)
	}

	// Create original session data
	originalSession := &neurotypes.ChatSession{
		ID:           "original-id",
		Name:         "original-name",
		SystemPrompt: "You are a helpful assistant",
		Messages: []neurotypes.Message{
			{
				ID:        "msg-1",
				Role:      "user",
				Content:   "Hello",
				Timestamp: testutils.GetCurrentTime(ctx),
			},
		},
		CreatedAt: testutils.GetCurrentTime(ctx),
		UpdatedAt: testutils.GetCurrentTime(ctx),
		IsActive:  false,
	}

	// Reconstruct session
	reconstructed, err := service.reconstructImportedSessionWithContext(originalSession, ctx)
	if err != nil {
		t.Fatalf("Failed to reconstruct session: %v", err)
	}

	// Verify new identity
	if reconstructed.ID == originalSession.ID {
		t.Error("Reconstructed session should have different ID")
	}
	if reconstructed.Name == originalSession.Name {
		t.Error("Reconstructed session should have different name")
	}
	if reconstructed.CreatedAt.Equal(originalSession.CreatedAt) {
		t.Error("Reconstructed session should have new creation timestamp")
	}
	if reconstructed.UpdatedAt.Equal(originalSession.UpdatedAt) {
		t.Error("Reconstructed session should have new update timestamp")
	}

	// Verify preserved content
	if reconstructed.SystemPrompt != originalSession.SystemPrompt {
		t.Errorf("System prompt not preserved: expected %s, got %s", originalSession.SystemPrompt, reconstructed.SystemPrompt)
	}
	if len(reconstructed.Messages) != len(originalSession.Messages) {
		t.Errorf("Message count mismatch: expected %d, got %d", len(originalSession.Messages), len(reconstructed.Messages))
	}

	// Verify reconstructed session is active
	if !reconstructed.IsActive {
		t.Error("Reconstructed session should be active")
	}

	// Verify name follows naming convention
	if reconstructed.Name != "Session 1" {
		t.Errorf("Expected auto-generated name 'Session 1', got %s", reconstructed.Name)
	}
}

func TestChatSessionService_NotInitialized_ExportImport(t *testing.T) {
	// Create uninitialized service
	service := NewChatSessionService()

	// Test export with uninitialized service
	ctx := neuroshellcontext.NewTestContext()
	err := service.ExportSessionToJSONWithContext("test-id", "/tmp/test.json", ctx)
	if err == nil {
		t.Fatal("Expected error for uninitialized service on export, got nil")
	}

	// Test import with uninitialized service
	_, err = service.ImportSessionFromJSONWithContext("/tmp/test.json", ctx)
	if err == nil {
		t.Fatal("Expected error for uninitialized service on import, got nil")
	}
}
