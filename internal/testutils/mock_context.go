package testutils

import (
	"fmt"
	"sync"
	"time"

	"neuroshell/pkg/types"
)

// MockContext implements the Context interface for testing
type MockContext struct {
	mu           sync.RWMutex
	variables    map[string]string
	history      []types.Message
	sessionState types.SessionState
	testMode     bool

	// For testing error scenarios
	getVariableError error
	setVariableError error
}

// NewMockContext creates a new mock context with default values
func NewMockContext() *MockContext {
	return &MockContext{
		variables: make(map[string]string),
		history:   []types.Message{},
		sessionState: types.SessionState{
			ID:        "test-session-123",
			Name:      "test-session",
			Variables: make(map[string]string),
			History:   []types.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		testMode: true,
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
func (m *MockContext) GetMessageHistory(n int) []types.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n <= 0 {
		return []types.Message{}
	}

	if n >= len(m.history) {
		return m.history
	}

	return m.history[len(m.history)-n:]
}

// GetSessionState implements Context.GetSessionState
func (m *MockContext) GetSessionState() types.SessionState {
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

	msg := types.Message{
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

	m.history = []types.Message{}
	m.sessionState.History = []types.Message{}
}
