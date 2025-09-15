package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSessionSubcontext_CreateSession(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Test creating a session with name and system prompt
	session, err := subctx.CreateSession("test-session", "You are a test assistant")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if session.Name != "test-session" {
		t.Errorf("Expected session name 'test-session', got '%s'", session.Name)
	}

	if session.SystemPrompt != "You are a test assistant" {
		t.Errorf("Expected system prompt 'You are a test assistant', got '%s'", session.SystemPrompt)
	}

	if !session.IsActive {
		t.Error("Expected new session to be active")
	}

	if len(session.Messages) != 0 {
		t.Errorf("Expected empty messages, got %d messages", len(session.Messages))
	}

	// Test creating session with empty system prompt (should get default)
	session2, err := subctx.CreateSession("test-session-2", "")
	if err != nil {
		t.Fatalf("Failed to create session with empty system prompt: %v", err)
	}

	if session2.SystemPrompt != "You are a helpful assistant." {
		t.Errorf("Expected default system prompt, got '%s'", session2.SystemPrompt)
	}
}

func TestSessionSubcontext_GetSession(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session first
	created, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test getting session by ID
	retrieved, err := subctx.GetSession(created.ID)
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected session ID %s, got %s", created.ID, retrieved.ID)
	}

	if retrieved.Name != created.Name {
		t.Errorf("Expected session name %s, got %s", created.Name, retrieved.Name)
	}

	// Test getting non-existent session
	_, err = subctx.GetSession("non-existent-id")
	if err == nil {
		t.Error("Expected error when getting non-existent session")
	}
}

func TestSessionSubcontext_GetSessionByName(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session first
	created, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test getting session by name
	retrieved, err := subctx.GetSessionByName("test-session")
	if err != nil {
		t.Fatalf("Failed to get session by name: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected session ID %s, got %s", created.ID, retrieved.ID)
	}

	// Test getting non-existent session by name
	_, err = subctx.GetSessionByName("non-existent-name")
	if err == nil {
		t.Error("Expected error when getting non-existent session by name")
	}
}

func TestSessionSubcontext_GetActiveSession(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Test with no active session
	_, err := subctx.GetActiveSession()
	if err == nil {
		t.Error("Expected error when no active session")
	}

	// Create a session (should become active)
	created, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test getting active session
	active, err := subctx.GetActiveSession()
	if err != nil {
		t.Fatalf("Failed to get active session: %v", err)
	}

	if active.ID != created.ID {
		t.Errorf("Expected active session ID %s, got %s", created.ID, active.ID)
	}
}

func TestSessionSubcontext_SetActiveSession(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create two sessions
	session1, err := subctx.CreateSession("session1", "Prompt 1")
	if err != nil {
		t.Fatalf("Failed to create session 1: %v", err)
	}

	session2, err := subctx.CreateSession("session2", "Prompt 2")
	if err != nil {
		t.Fatalf("Failed to create session 2: %v", err)
	}

	// Verify session2 is active (newly created sessions are active)
	active, err := subctx.GetActiveSession()
	if err != nil {
		t.Fatalf("Failed to get active session: %v", err)
	}
	if active.ID != session2.ID {
		t.Errorf("Expected session2 to be active, got session %s", active.ID)
	}

	// Set session1 as active
	err = subctx.SetActiveSession(session1.ID)
	if err != nil {
		t.Fatalf("Failed to set active session: %v", err)
	}

	// Verify session1 is now active
	active, err = subctx.GetActiveSession()
	if err != nil {
		t.Fatalf("Failed to get active session: %v", err)
	}
	if active.ID != session1.ID {
		t.Errorf("Expected session1 to be active, got session %s", active.ID)
	}

	// Test setting non-existent session as active
	err = subctx.SetActiveSession("non-existent-id")
	if err == nil {
		t.Error("Expected error when setting non-existent session as active")
	}
}

func TestSessionSubcontext_DeleteSession(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session
	session, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Delete the session
	err = subctx.DeleteSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// Verify session is deleted
	_, err = subctx.GetSession(session.ID)
	if err == nil {
		t.Error("Expected error when getting deleted session")
	}

	// Verify session is removed from name mapping
	_, err = subctx.GetSessionByName("test-session")
	if err == nil {
		t.Error("Expected error when getting deleted session by name")
	}

	// Test deleting non-existent session
	err = subctx.DeleteSession("non-existent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent session")
	}
}

func TestSessionSubcontext_RenameSession(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session
	session, err := subctx.CreateSession("original-name", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Rename the session
	err = subctx.RenameSession(session.ID, "new-name")
	if err != nil {
		t.Fatalf("Failed to rename session: %v", err)
	}

	// Verify the session can be retrieved by new name
	retrieved, err := subctx.GetSessionByName("new-name")
	if err != nil {
		t.Fatalf("Failed to get session by new name: %v", err)
	}

	if retrieved.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, retrieved.ID)
	}

	// Verify old name no longer works
	_, err = subctx.GetSessionByName("original-name")
	if err == nil {
		t.Error("Expected error when getting session by old name")
	}

	// Test renaming to existing name
	session2, err := subctx.CreateSession("another-session", "Another prompt")
	if err != nil {
		t.Fatalf("Failed to create second session: %v", err)
	}

	err = subctx.RenameSession(session2.ID, "new-name")
	if err == nil {
		t.Error("Expected error when renaming to existing name")
	}
}

func TestSessionSubcontext_CopySession(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session with some messages
	original, err := subctx.CreateSession("original", "Original prompt")
	if err != nil {
		t.Fatalf("Failed to create original session: %v", err)
	}

	// Add some messages
	err = subctx.AddUserMessage(original.ID, "Hello")
	if err != nil {
		t.Fatalf("Failed to add user message: %v", err)
	}

	err = subctx.AddAssistantMessage(original.ID, "Hi there!")
	if err != nil {
		t.Fatalf("Failed to add assistant message: %v", err)
	}

	// Copy the session
	copiedSession, err := subctx.CopySession(original.ID, "copy")
	if err != nil {
		t.Fatalf("Failed to copy session: %v", err)
	}

	// Verify copy has different ID
	if copiedSession.ID == original.ID {
		t.Error("Expected copy to have different ID")
	}

	// Verify copy has correct name
	if copiedSession.Name != "copy" {
		t.Errorf("Expected copy name 'copy', got '%s'", copiedSession.Name)
	}

	// Verify copy has same system prompt
	if copiedSession.SystemPrompt != original.SystemPrompt {
		t.Errorf("Expected copy system prompt '%s', got '%s'", original.SystemPrompt, copiedSession.SystemPrompt)
	}

	// Verify copy has same number of messages
	if len(copiedSession.Messages) != len(original.Messages) {
		t.Errorf("Expected %d messages, got %d", len(original.Messages), len(copiedSession.Messages))
	}

	// Verify copy messages have different IDs but same content
	for i := 0; i < len(original.Messages); i++ {
		if copiedSession.Messages[i].ID == original.Messages[i].ID {
			t.Errorf("Expected message %d to have different ID", i)
		}
		if copiedSession.Messages[i].Content != original.Messages[i].Content {
			t.Errorf("Expected message %d content '%s', got '%s'", i, original.Messages[i].Content, copiedSession.Messages[i].Content)
		}
	}

	// Verify copy is not active
	if copiedSession.IsActive {
		t.Error("Expected copy to not be active")
	}
}

func TestSessionSubcontext_MessageManagement(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session
	session, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test adding user message
	err = subctx.AddUserMessage(session.ID, "Hello world")
	if err != nil {
		t.Fatalf("Failed to add user message: %v", err)
	}

	// Test adding assistant message
	err = subctx.AddAssistantMessage(session.ID, "Hi there!")
	if err != nil {
		t.Fatalf("Failed to add assistant message: %v", err)
	}

	// Test getting messages
	messages, err := subctx.GetMessages(session.ID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != "user" || messages[0].Content != "Hello world" {
		t.Errorf("Expected first message to be user message 'Hello world', got role '%s', content '%s'", messages[0].Role, messages[0].Content)
	}

	if messages[1].Role != "assistant" || messages[1].Content != "Hi there!" {
		t.Errorf("Expected second message to be assistant message 'Hi there!', got role '%s', content '%s'", messages[1].Role, messages[1].Content)
	}

	// Test updating message
	messageID := messages[0].ID
	err = subctx.UpdateMessage(session.ID, messageID, "Updated message")
	if err != nil {
		t.Fatalf("Failed to update message: %v", err)
	}

	// Verify message was updated
	updatedMessages, err := subctx.GetMessages(session.ID)
	if err != nil {
		t.Fatalf("Failed to get updated messages: %v", err)
	}

	if updatedMessages[0].Content != "Updated message" {
		t.Errorf("Expected updated message content 'Updated message', got '%s'", updatedMessages[0].Content)
	}

	// Test deleting message
	err = subctx.DeleteMessage(session.ID, messageID)
	if err != nil {
		t.Fatalf("Failed to delete message: %v", err)
	}

	// Verify message was deleted
	finalMessages, err := subctx.GetMessages(session.ID)
	if err != nil {
		t.Fatalf("Failed to get final messages: %v", err)
	}

	if len(finalMessages) != 1 {
		t.Errorf("Expected 1 message after deletion, got %d", len(finalMessages))
	}

	if finalMessages[0].Content != "Hi there!" {
		t.Errorf("Expected remaining message to be 'Hi there!', got '%s'", finalMessages[0].Content)
	}
}

func TestSessionSubcontext_SystemPromptManagement(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session
	session, err := subctx.CreateSession("test-session", "Original prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test getting system prompt
	prompt, err := subctx.GetSystemPrompt(session.ID)
	if err != nil {
		t.Fatalf("Failed to get system prompt: %v", err)
	}

	if prompt != "Original prompt" {
		t.Errorf("Expected system prompt 'Original prompt', got '%s'", prompt)
	}

	// Test setting system prompt
	err = subctx.SetSystemPrompt(session.ID, "New prompt")
	if err != nil {
		t.Fatalf("Failed to set system prompt: %v", err)
	}

	// Verify system prompt was updated
	updatedPrompt, err := subctx.GetSystemPrompt(session.ID)
	if err != nil {
		t.Fatalf("Failed to get updated system prompt: %v", err)
	}

	if updatedPrompt != "New prompt" {
		t.Errorf("Expected updated system prompt 'New prompt', got '%s'", updatedPrompt)
	}
}

func TestSessionSubcontext_SessionPersistence(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session
	session, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add a message
	err = subctx.AddUserMessage(session.ID, "Test message")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Test saving session
	err = subctx.SaveSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Test exporting session
	tempDir := t.TempDir()
	exportPath := filepath.Join(tempDir, "exported_session.json")
	err = subctx.ExportSession(session.ID, exportPath)
	if err != nil {
		t.Fatalf("Failed to export session: %v", err)
	}

	// Verify export file exists
	if _, err := os.Stat(exportPath); err != nil {
		t.Errorf("Export file does not exist: %v", err)
	}

	// Test importing session
	imported, err := subctx.ImportSession(exportPath)
	if err != nil {
		t.Fatalf("Failed to import session: %v", err)
	}

	// Verify imported session has correct data (name might have :v1 suffix due to conflicts)
	expectedName := session.Name
	if imported.Name != expectedName && imported.Name != expectedName+":v1" {
		t.Errorf("Expected imported session name '%s' or '%s:v1', got '%s'", expectedName, expectedName, imported.Name)
	}

	if imported.SystemPrompt != session.SystemPrompt {
		t.Errorf("Expected imported system prompt '%s', got '%s'", session.SystemPrompt, imported.SystemPrompt)
	}

	if len(imported.Messages) != len(session.Messages) {
		t.Errorf("Expected %d messages in imported session, got %d", len(session.Messages), len(imported.Messages))
	}

	// Verify imported session has different ID
	if imported.ID == session.ID {
		t.Error("Expected imported session to have different ID")
	}
}

func TestSessionSubcontext_SessionMetadata(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Test with no sessions
	count := subctx.GetSessionCount()
	if count != 0 {
		t.Errorf("Expected 0 sessions, got %d", count)
	}

	// Create a session
	session, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test session count
	count = subctx.GetSessionCount()
	if count != 1 {
		t.Errorf("Expected 1 session, got %d", count)
	}

	// Test message count (should be 0 initially)
	msgCount := subctx.GetMessageCount(session.ID)
	if msgCount != 0 {
		t.Errorf("Expected 0 messages, got %d", msgCount)
	}

	// Add a message
	err = subctx.AddUserMessage(session.ID, "Test message")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Test message count again
	msgCount = subctx.GetMessageCount(session.ID)
	if msgCount != 1 {
		t.Errorf("Expected 1 message, got %d", msgCount)
	}

	// Test getting last message
	lastMsg, err := subctx.GetLastMessage(session.ID)
	if err != nil {
		t.Fatalf("Failed to get last message: %v", err)
	}

	if lastMsg.Content != "Test message" {
		t.Errorf("Expected last message content 'Test message', got '%s'", lastMsg.Content)
	}

	if lastMsg.Role != "user" {
		t.Errorf("Expected last message role 'user', got '%s'", lastMsg.Role)
	}
}

func TestSessionSubcontext_SessionSearch(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create sessions with similar names
	session1, err := subctx.CreateSession("project-alpha", "Prompt 1")
	if err != nil {
		t.Fatalf("Failed to create session 1: %v", err)
	}

	session2, err := subctx.CreateSession("project-beta", "Prompt 2")
	if err != nil {
		t.Fatalf("Failed to create session 2: %v", err)
	}

	session3, err := subctx.CreateSession("other-project", "Prompt 3")
	if err != nil {
		t.Fatalf("Failed to create session 3: %v", err)
	}

	// Test finding session by name prefix
	found, err := subctx.FindSessionByNamePrefix("project")
	if err != nil {
		t.Fatalf("Failed to find session by name prefix: %v", err)
	}

	// Should find a session with "project" prefix (any of them is fine)
	if !strings.HasPrefix(found.Name, "project") {
		t.Errorf("Expected to find session with 'project' prefix, found '%s'", found.Name)
	}

	// Verify that at least one of our sessions was found
	if found.ID != session1.ID && found.ID != session2.ID {
		t.Errorf("Expected to find one of our project sessions, found session with ID %s", found.ID)
	}

	// Test finding session by ID prefix
	found, err = subctx.FindSessionByIDPrefix(session2.ID[:8]) // Use first 8 characters
	if err != nil {
		t.Fatalf("Failed to find session by ID prefix: %v", err)
	}

	if found.ID != session2.ID {
		t.Errorf("Expected to find session2, found session with ID %s", found.ID)
	}

	// Test finding non-existent session
	_, err = subctx.FindSessionByNamePrefix("non-existent")
	if err == nil {
		t.Error("Expected error when finding non-existent session")
	}

	// Test getting latest session
	latest, err := subctx.GetLatestSession()
	if err != nil {
		t.Fatalf("Failed to get latest session: %v", err)
	}

	if latest.ID != session3.ID {
		t.Errorf("Expected session3 to be latest, got session with ID %s", latest.ID)
	}
}

func TestSessionSubcontext_SessionValidation(t *testing.T) {
	// Test session name validation
	tests := []struct {
		name  string
		valid bool
	}{
		{"valid-name", true},
		{"valid name", true},
		{"valid_name", true},
		{"valid-name-123", true},
		{"", false},
		{"name\nwith\nnewlines", false},
		{"name\twith\ttabs", false},
		{" name-with-leading-space", false},
		{"name-with-trailing-space ", false},
	}

	for _, test := range tests {
		err := ValidateSessionName(test.name)
		if test.valid && err != nil {
			t.Errorf("Expected name '%s' to be valid, got error: %v", test.name, err)
		}
		if !test.valid && err == nil {
			t.Errorf("Expected name '%s' to be invalid, got no error", test.name)
		}
	}
}

func TestSessionSubcontext_SessionInfo(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session
	session, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add a message
	err = subctx.AddUserMessage(session.ID, "Test message")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Get session info
	info, err := subctx.GetSessionInfo(session.ID)
	if err != nil {
		t.Fatalf("Failed to get session info: %v", err)
	}

	// Verify session info
	if info.ID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, info.ID)
	}

	if info.Name != session.Name {
		t.Errorf("Expected session name %s, got %s", session.Name, info.Name)
	}

	if info.MessageCount != 1 {
		t.Errorf("Expected 1 message, got %d", info.MessageCount)
	}

	if !info.IsActive {
		t.Error("Expected session to be active")
	}

	if info.SystemPrompt != session.SystemPrompt {
		t.Errorf("Expected system prompt %s, got %s", session.SystemPrompt, info.SystemPrompt)
	}
}

func TestSessionSubcontext_ListSessions(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create multiple sessions
	session1, err := subctx.CreateSession("session1", "Prompt 1")
	if err != nil {
		t.Fatalf("Failed to create session 1: %v", err)
	}

	session2, err := subctx.CreateSession("session2", "Prompt 2")
	if err != nil {
		t.Fatalf("Failed to create session 2: %v", err)
	}

	// Test listing sessions as slice
	sessions := subctx.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}

	// Test listing sessions as map
	sessionMap := subctx.GetAllSessions()
	if len(sessionMap) != 2 {
		t.Errorf("Expected 2 sessions in map, got %d", len(sessionMap))
	}

	// Verify all sessions are present
	if _, exists := sessionMap[session1.ID]; !exists {
		t.Errorf("Expected session1 to be in map")
	}

	if _, exists := sessionMap[session2.ID]; !exists {
		t.Errorf("Expected session2 to be in map")
	}

	// Test session name to ID mapping
	nameToID := subctx.GetSessionNameToID()
	if len(nameToID) != 2 {
		t.Errorf("Expected 2 name-to-ID mappings, got %d", len(nameToID))
	}

	if nameToID["session1"] != session1.ID {
		t.Errorf("Expected session1 ID %s, got %s", session1.ID, nameToID["session1"])
	}

	if nameToID["session2"] != session2.ID {
		t.Errorf("Expected session2 ID %s, got %s", session2.ID, nameToID["session2"])
	}
}

func TestSessionSubcontext_MultipleActiveSessions(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create first session
	session1, err := subctx.CreateSession("session1", "Prompt 1")
	if err != nil {
		t.Fatalf("Failed to create session 1: %v", err)
	}

	// Verify session1 is active
	if !session1.IsActive {
		t.Error("Expected session1 to be active")
	}

	// Create second session
	session2, err := subctx.CreateSession("session2", "Prompt 2")
	if err != nil {
		t.Fatalf("Failed to create session 2: %v", err)
	}

	// Verify session2 is active and session1 is not
	if !session2.IsActive {
		t.Error("Expected session2 to be active")
	}

	// Re-get session1 to verify it's no longer active
	retrievedSession1, err := subctx.GetSession(session1.ID)
	if err != nil {
		t.Fatalf("Failed to get session1: %v", err)
	}

	if retrievedSession1.IsActive {
		t.Error("Expected session1 to be inactive after session2 was created")
	}
}

func TestSessionSubcontext_Timestamps(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session
	session, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify timestamps are set
	if session.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if session.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	// Add a message
	err = subctx.AddUserMessage(session.ID, "Test message")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Get updated session
	updatedSession, err := subctx.GetSession(session.ID)
	if err != nil {
		t.Fatalf("Failed to get updated session: %v", err)
	}

	// Verify UpdatedAt was updated
	if !updatedSession.UpdatedAt.After(session.CreatedAt) {
		t.Error("Expected UpdatedAt to be after CreatedAt")
	}

	// Verify CreatedAt hasn't changed
	if updatedSession.CreatedAt != session.CreatedAt {
		t.Error("Expected CreatedAt to remain unchanged")
	}
}

func TestSessionSubcontext_MessageTimestamps(t *testing.T) {
	ctx := New()
	subctx := NewSessionSubcontext(ctx)

	// Create a session
	session, err := subctx.CreateSession("test-session", "Test prompt")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add a message
	err = subctx.AddUserMessage(session.ID, "Test message")
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	// Get messages
	messages, err := subctx.GetMessages(session.ID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Verify message timestamp
	if messages[0].Timestamp.IsZero() {
		t.Error("Expected message timestamp to be set")
	}

	// Verify message timestamp is recent (within 1 second)
	now := time.Now()
	if messages[0].Timestamp.After(now.Add(time.Second)) || messages[0].Timestamp.Before(now.Add(-time.Second)) {
		t.Errorf("Expected message timestamp to be recent, got %v", messages[0].Timestamp)
	}
}
