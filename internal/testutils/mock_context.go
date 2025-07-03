// Package testutils provides testing utilities and mock implementations for NeuroShell.
// It includes mock contexts and helpers for unit testing the command and service layers.
package testutils

import (
	"fmt"
	"os"
	"sync"
	"time"

	"neuroshell/pkg/neurotypes"
)

// MockContext implements the Context interface for testing
type MockContext struct {
	mu           sync.RWMutex
	variables    map[string]string
	history      []neurotypes.Message
	sessionState neurotypes.SessionState
	testMode     bool

	// Chat session storage for testing
	chatSessions    map[string]*neurotypes.ChatSession
	sessionNameToID map[string]string
	activeSessionID string

	// For testing error scenarios
	getVariableError error
	setVariableError error
}

// NewMockContext creates a new mock context with default values
func NewMockContext() *MockContext {
	return &MockContext{
		variables: make(map[string]string),
		history:   []neurotypes.Message{},
		sessionState: neurotypes.SessionState{
			ID:        "test-session-123",
			Name:      "test-session",
			Variables: make(map[string]string),
			History:   []neurotypes.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		testMode: true,

		// Initialize chat session storage
		chatSessions:    make(map[string]*neurotypes.ChatSession),
		sessionNameToID: make(map[string]string),
		activeSessionID: "",
	}
}

// NewMockContextWithVars creates a mock context with predefined variables
func NewMockContextWithVars(vars map[string]string) *MockContext {
	ctx := NewMockContext()
	for k, v := range vars {
		ctx.variables[k] = v
	}
	return ctx
}

// GetVariable implements Context.GetVariable
func (m *MockContext) GetVariable(name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getVariableError != nil {
		return "", m.getVariableError
	}

	// Handle system variables like the real context
	switch name {
	case "@user":
		return "testuser", nil
	case "@pwd":
		return "/test/pwd", nil
	case "@home":
		return "/test/home", nil
	case "@date":
		return "2024-01-01", nil
	case "@os":
		return "test-os", nil
	case "#test_mode":
		if m.testMode {
			return "true", nil
		}
		return "false", nil
	case "#session_id":
		return m.sessionState.ID, nil
	case "#message_count":
		return fmt.Sprintf("%d", len(m.history)), nil
	}

	// Regular variables
	if value, exists := m.variables[name]; exists {
		return value, nil
	}

	return "", fmt.Errorf("variable '%s' not found", name)
}

// SetVariable implements Context.SetVariable
func (m *MockContext) SetVariable(name string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.setVariableError != nil {
		return m.setVariableError
	}

	m.variables[name] = value
	m.sessionState.Variables[name] = value
	return nil
}

// GetMessageHistory implements Context.GetMessageHistory
func (m *MockContext) GetMessageHistory(n int) []neurotypes.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n <= 0 {
		return []neurotypes.Message{}
	}

	if n >= len(m.history) {
		return m.history
	}

	return m.history[len(m.history)-n:]
}

// GetSessionState implements Context.GetSessionState
func (m *MockContext) GetSessionState() neurotypes.SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.sessionState
}

// SetTestMode implements Context.SetTestMode
func (m *MockContext) SetTestMode(testMode bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.testMode = testMode
}

// IsTestMode implements Context.IsTestMode
func (m *MockContext) IsTestMode() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.testMode
}

// Test helper methods

// SetGetVariableError sets an error to be returned by GetVariable
func (m *MockContext) SetGetVariableError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getVariableError = err
}

// SetSetVariableError sets an error to be returned by SetVariable
func (m *MockContext) SetSetVariableError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setVariableError = err
}

// AddMessage adds a message to the history for testing
func (m *MockContext) AddMessage(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := neurotypes.Message{
		ID:        fmt.Sprintf("msg-%d", len(m.history)+1),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	m.history = append(m.history, msg)
	m.sessionState.History = m.history
}

// SetSessionID sets the session ID for testing
func (m *MockContext) SetSessionID(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessionState.ID = id
}

// GetAllVariables returns all variables for testing
func (m *MockContext) GetAllVariables() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range m.variables {
		result[k] = v
	}
	return result
}

// ClearVariables clears all variables for testing
func (m *MockContext) ClearVariables() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.variables = make(map[string]string)
	m.sessionState.Variables = make(map[string]string)
}

// ClearHistory clears message history for testing
func (m *MockContext) ClearHistory() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.history = []neurotypes.Message{}
	m.sessionState.History = []neurotypes.Message{}
}

// SetSystemVariable sets a system variable (for testing bash service)
func (m *MockContext) SetSystemVariable(name string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.setVariableError != nil {
		return m.setVariableError
	}

	// System variables are prefixed with _, @, or #
	if len(name) > 0 && (name[0] == '_' || name[0] == '@' || name[0] == '#') {
		m.variables[name] = value
		m.sessionState.Variables[name] = value
		return nil
	}

	return fmt.Errorf("variable '%s' is not a system variable", name)
}

// EditorTestHelper provides utilities for testing editor functionality
type EditorTestHelper struct {
	originalEditor string
	originalPath   string
}

// SetupMockEditor configures the environment for fast, non-hanging editor tests
func SetupMockEditor() *EditorTestHelper {
	helper := &EditorTestHelper{
		originalEditor: os.Getenv("EDITOR"),
		originalPath:   os.Getenv("PATH"),
	}

	// Set EDITOR to echo for fast, predictable testing
	_ = os.Setenv("EDITOR", "echo")

	return helper
}

// SetupNoEditor configures the environment to simulate no editor available
func SetupNoEditor() *EditorTestHelper {
	helper := &EditorTestHelper{
		originalEditor: os.Getenv("EDITOR"),
		originalPath:   os.Getenv("PATH"),
	}

	// Remove editor and PATH to simulate no editor found
	_ = os.Unsetenv("EDITOR")
	_ = os.Setenv("PATH", "")

	return helper
}

// Cleanup restores the original environment variables
func (h *EditorTestHelper) Cleanup() {
	_ = os.Setenv("EDITOR", h.originalEditor)
	_ = os.Setenv("PATH", h.originalPath)
}

// Chat session storage methods for Context interface compliance

// GetChatSessions returns all chat sessions stored in the mock context.
func (m *MockContext) GetChatSessions() map[string]*neurotypes.ChatSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.chatSessions
}

// SetChatSessions sets the chat sessions map in the mock context.
func (m *MockContext) SetChatSessions(sessions map[string]*neurotypes.ChatSession) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chatSessions = sessions
}

// GetSessionNameToID returns the session name to ID mapping.
func (m *MockContext) GetSessionNameToID() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessionNameToID
}

// SetSessionNameToID sets the session name to ID mapping in the mock context.
func (m *MockContext) SetSessionNameToID(nameToID map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessionNameToID = nameToID
}

// GetActiveSessionID returns the currently active session ID.
func (m *MockContext) GetActiveSessionID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeSessionID
}

// SetActiveSessionID sets the currently active session ID.
func (m *MockContext) SetActiveSessionID(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeSessionID = sessionID
}
