// Package context provides session-specific context operations for NeuroShell.
// This file implements SessionSubcontext, a focused interface for session management
// that eliminates the need for services to know about global context internals.
package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// SessionSubcontext provides focused session operations without exposing full context internals.
// This interface is designed to be passed to services that only need session functionality,
// following the Interface Segregation Principle.
type SessionSubcontext interface {
	// Core session operations
	CreateSession(name, systemPrompt string) (*neurotypes.ChatSession, error)
	GetSession(sessionID string) (*neurotypes.ChatSession, error)
	GetSessionByName(name string) (*neurotypes.ChatSession, error)
	GetActiveSession() (*neurotypes.ChatSession, error)
	SetActiveSession(sessionID string) error
	DeleteSession(sessionID string) error
	RenameSession(sessionID, newName string) error
	CopySession(sourceID, targetName string) (*neurotypes.ChatSession, error)

	// Session listing and query
	ListSessions() []*neurotypes.ChatSession
	GetAllSessions() map[string]*neurotypes.ChatSession
	GetSessionNameToID() map[string]string

	// Message management
	AddUserMessage(sessionID, content string) error
	AddAssistantMessage(sessionID, content string) error
	GetMessages(sessionID string) ([]neurotypes.Message, error)
	UpdateMessage(sessionID, messageID, newContent string) error
	DeleteMessage(sessionID, messageID string) error

	// System prompt management
	GetSystemPrompt(sessionID string) (string, error)
	SetSystemPrompt(sessionID, systemPrompt string) error

	// Session persistence
	SaveSession(sessionID string) error
	LoadSession(sessionID string) error
	ExportSession(sessionID, filePath string) error
	ImportSession(filePath string) (*neurotypes.ChatSession, error)

	// Session metadata
	GetSessionCount() int
	GetMessageCount(sessionID string) int
	GetLastMessage(sessionID string) (*neurotypes.Message, error)

	// Session search and lookup
	FindSessionByNamePrefix(prefix string) (*neurotypes.ChatSession, error)
	FindSessionByIDPrefix(prefix string) (*neurotypes.ChatSession, error)
	GetLatestSession() (*neurotypes.ChatSession, error)

	// Session information
	GetSessionInfo(sessionID string) (*SessionInfo, error)

	// Message access methods (extracted from main context)
	GetNthRecentMessage(n int) (string, error)
	GetNthChronologicalMessage(n int) (string, error)
}

// sessionSubcontextImpl implements SessionSubcontext using a NeuroContext.
// This provides a clean abstraction layer that services can depend on.
type sessionSubcontextImpl struct {
	ctx *NeuroContext
}

// NewSessionSubcontext creates a new SessionSubcontext from a NeuroContext.
// This is the factory function that services should use to get session functionality.
func NewSessionSubcontext(ctx *NeuroContext) SessionSubcontext {
	return &sessionSubcontextImpl{ctx: ctx}
}

// CreateSession creates a new chat session with the given name and system prompt.
func (s *sessionSubcontextImpl) CreateSession(name, systemPrompt string) (*neurotypes.ChatSession, error) {
	// Generate unique session ID (deterministic in test mode)
	sessionID := testutils.GenerateUUID(s.ctx)

	// Set default system prompt if not provided
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant."
	}

	// Create new session (deterministic time in test mode)
	now := time.Now()
	if s.ctx.testMode {
		// In test mode, use a deterministic time
		now = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	session := &neurotypes.ChatSession{
		ID:           sessionID,
		Name:         name,
		SystemPrompt: systemPrompt,
		Messages:     make([]neurotypes.Message, 0),
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true, // New session becomes active
	}

	// Deactivate previous active session
	if s.ctx.activeSessionID != "" {
		if prevSession, exists := s.ctx.chatSessions[s.ctx.activeSessionID]; exists {
			prevSession.IsActive = false
		}
	}

	// Store session and update mappings
	s.ctx.chatSessions[sessionID] = session
	s.ctx.sessionNameToID[name] = sessionID
	s.ctx.activeSessionID = sessionID

	return session, nil
}

// GetSession retrieves a session by its ID.
func (s *sessionSubcontextImpl) GetSession(sessionID string) (*neurotypes.ChatSession, error) {
	session, exists := s.ctx.chatSessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session with ID %s not found", sessionID)
	}
	return session, nil
}

// GetSessionByName retrieves a session by its name.
func (s *sessionSubcontextImpl) GetSessionByName(name string) (*neurotypes.ChatSession, error) {
	sessionID, exists := s.ctx.sessionNameToID[name]
	if !exists {
		return nil, fmt.Errorf("session with name %s not found", name)
	}
	return s.GetSession(sessionID)
}

// GetActiveSession returns the currently active session.
func (s *sessionSubcontextImpl) GetActiveSession() (*neurotypes.ChatSession, error) {
	if s.ctx.activeSessionID == "" {
		return nil, errors.New("no active session")
	}
	return s.GetSession(s.ctx.activeSessionID)
}

// SetActiveSession sets the active session by ID.
func (s *sessionSubcontextImpl) SetActiveSession(sessionID string) error {
	if _, exists := s.ctx.chatSessions[sessionID]; !exists {
		return fmt.Errorf("session with ID %s not found", sessionID)
	}
	s.ctx.activeSessionID = sessionID
	return nil
}

// DeleteSession removes a session by ID.
func (s *sessionSubcontextImpl) DeleteSession(sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Remove from maps
	delete(s.ctx.chatSessions, sessionID)
	delete(s.ctx.sessionNameToID, session.Name)

	// If this was the active session, clear it
	if s.ctx.activeSessionID == sessionID {
		s.ctx.activeSessionID = ""
	}

	return nil
}

// RenameSession changes the name of a session.
func (s *sessionSubcontextImpl) RenameSession(sessionID, newName string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Check if new name already exists
	if _, exists := s.ctx.sessionNameToID[newName]; exists {
		return fmt.Errorf("session with name %s already exists", newName)
	}

	// Update name and mappings
	oldName := session.Name
	session.Name = newName
	delete(s.ctx.sessionNameToID, oldName)
	s.ctx.sessionNameToID[newName] = sessionID
	session.UpdatedAt = time.Now()

	return nil
}

// CopySession creates a copy of a session with a new name.
func (s *sessionSubcontextImpl) CopySession(sourceID, targetName string) (*neurotypes.ChatSession, error) {
	source, err := s.GetSession(sourceID)
	if err != nil {
		return nil, err
	}

	// Generate new session ID
	newSessionID := testutils.GenerateUUID(s.ctx)

	// Create copied session with new identity
	now := time.Now()
	if s.ctx.testMode {
		now = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	copiedSession := &neurotypes.ChatSession{
		ID:           newSessionID,
		Name:         targetName,
		SystemPrompt: source.SystemPrompt,
		Messages:     make([]neurotypes.Message, len(source.Messages)),
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     false,
	}

	// Deep copy messages
	for i, msg := range source.Messages {
		copiedSession.Messages[i] = neurotypes.Message{
			ID:        testutils.GenerateUUID(s.ctx),
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp, // Preserve original timestamp
		}
	}

	// Add to session storage
	s.ctx.chatSessions[newSessionID] = copiedSession
	s.ctx.sessionNameToID[targetName] = newSessionID

	return copiedSession, nil
}

// ListSessions returns all sessions as a slice.
func (s *sessionSubcontextImpl) ListSessions() []*neurotypes.ChatSession {
	sessions := make([]*neurotypes.ChatSession, 0, len(s.ctx.chatSessions))
	for _, session := range s.ctx.chatSessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// GetAllSessions returns all sessions as a map.
func (s *sessionSubcontextImpl) GetAllSessions() map[string]*neurotypes.ChatSession {
	return s.ctx.chatSessions
}

// GetSessionNameToID returns the name-to-ID mapping.
func (s *sessionSubcontextImpl) GetSessionNameToID() map[string]string {
	return s.ctx.sessionNameToID
}

// AddUserMessage adds a user message to a session.
func (s *sessionSubcontextImpl) AddUserMessage(sessionID, content string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	message := neurotypes.Message{
		ID:        testutils.GenerateUUID(s.ctx),
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()
	return nil
}

// AddAssistantMessage adds an assistant message to a session.
func (s *sessionSubcontextImpl) AddAssistantMessage(sessionID, content string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	message := neurotypes.Message{
		ID:        testutils.GenerateUUID(s.ctx),
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()
	return nil
}

// GetMessages returns all messages for a session.
func (s *sessionSubcontextImpl) GetMessages(sessionID string) ([]neurotypes.Message, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.Messages, nil
}

// UpdateMessage updates the content of a message.
func (s *sessionSubcontextImpl) UpdateMessage(sessionID, messageID, newContent string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	for i, msg := range session.Messages {
		if msg.ID == messageID {
			session.Messages[i].Content = newContent
			session.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("message with ID %s not found in session %s", messageID, sessionID)
}

// DeleteMessage removes a message from a session.
func (s *sessionSubcontextImpl) DeleteMessage(sessionID, messageID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	for i, msg := range session.Messages {
		if msg.ID == messageID {
			// Remove message
			session.Messages = append(session.Messages[:i], session.Messages[i+1:]...)
			session.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("message with ID %s not found in session %s", messageID, sessionID)
}

// GetSystemPrompt returns the system prompt for a session.
func (s *sessionSubcontextImpl) GetSystemPrompt(sessionID string) (string, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return "", err
	}
	return session.SystemPrompt, nil
}

// SetSystemPrompt updates the system prompt for a session.
func (s *sessionSubcontextImpl) SetSystemPrompt(sessionID, systemPrompt string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	session.SystemPrompt = systemPrompt
	session.UpdatedAt = time.Now()
	return nil
}

// SaveSession persists a session to disk.
func (s *sessionSubcontextImpl) SaveSession(sessionID string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Get session directory
	configDir, err := s.ctx.GetUserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	sessionsDir := filepath.Join(configDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	// Save session to file
	sessionFile := filepath.Join(sessionsDir, sessionID+".json")
	return s.saveSessionToFile(session, sessionFile)
}

// LoadSession loads a session from disk.
func (s *sessionSubcontextImpl) LoadSession(sessionID string) error {
	// Get session directory
	configDir, err := s.ctx.GetUserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	sessionsDir := filepath.Join(configDir, "sessions")
	sessionFile := filepath.Join(sessionsDir, sessionID+".json")

	// Check if file exists
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return fmt.Errorf("session file not found: %s", sessionFile)
	}

	// Load session from file
	session, err := s.loadSessionFromFile(sessionFile)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	// Add to session storage
	s.ctx.chatSessions[sessionID] = session
	s.ctx.sessionNameToID[session.Name] = sessionID

	return nil
}

// ExportSession exports a session to a specific file path.
func (s *sessionSubcontextImpl) ExportSession(sessionID, filePath string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	return s.saveSessionToFile(session, filePath)
}

// ImportSession imports a session from a file.
func (s *sessionSubcontextImpl) ImportSession(filePath string) (*neurotypes.ChatSession, error) {
	session, err := s.loadSessionFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Generate new ID to avoid conflicts
	session.ID = testutils.GenerateUUID(s.ctx)
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	// Handle name conflicts
	originalName := session.Name
	counter := 1
	for _, exists := s.ctx.sessionNameToID[session.Name]; exists; _, exists = s.ctx.sessionNameToID[session.Name] {
		session.Name = fmt.Sprintf("%s:v%d", originalName, counter)
		counter++
	}

	// Add to session storage
	s.ctx.chatSessions[session.ID] = session
	s.ctx.sessionNameToID[session.Name] = session.ID

	return session, nil
}

// GetSessionCount returns the total number of sessions.
func (s *sessionSubcontextImpl) GetSessionCount() int {
	return len(s.ctx.chatSessions)
}

// GetMessageCount returns the number of messages in a session.
func (s *sessionSubcontextImpl) GetMessageCount(sessionID string) int {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return 0
	}
	return len(session.Messages)
}

// GetLastMessage returns the last message in a session.
func (s *sessionSubcontextImpl) GetLastMessage(sessionID string) (*neurotypes.Message, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if len(session.Messages) == 0 {
		return nil, errors.New("no messages in session")
	}

	return &session.Messages[len(session.Messages)-1], nil
}

// FindSessionByNamePrefix finds a session by name prefix.
func (s *sessionSubcontextImpl) FindSessionByNamePrefix(prefix string) (*neurotypes.ChatSession, error) {
	// Try exact match first
	if session, err := s.GetSessionByName(prefix); err == nil {
		return session, nil
	}

	// Try prefix match
	for name, sessionID := range s.ctx.sessionNameToID {
		if strings.HasPrefix(name, prefix) {
			return s.GetSession(sessionID)
		}
	}

	return nil, fmt.Errorf("no session found with name prefix: %s", prefix)
}

// FindSessionByIDPrefix finds a session by ID prefix.
func (s *sessionSubcontextImpl) FindSessionByIDPrefix(prefix string) (*neurotypes.ChatSession, error) {
	// Try exact match first
	if session, err := s.GetSession(prefix); err == nil {
		return session, nil
	}

	// Try prefix match
	for sessionID := range s.ctx.chatSessions {
		if strings.HasPrefix(sessionID, prefix) {
			return s.GetSession(sessionID)
		}
	}

	return nil, fmt.Errorf("no session found with ID prefix: %s", prefix)
}

// GetLatestSession returns the most recently updated session.
func (s *sessionSubcontextImpl) GetLatestSession() (*neurotypes.ChatSession, error) {
	if len(s.ctx.chatSessions) == 0 {
		return nil, errors.New("no sessions available")
	}

	var latestSession *neurotypes.ChatSession
	var latestTime time.Time

	for _, session := range s.ctx.chatSessions {
		if session.UpdatedAt.After(latestTime) {
			latestTime = session.UpdatedAt
			latestSession = session
		}
	}

	return latestSession, nil
}

// ValidateSessionName checks if a session name is valid.
func ValidateSessionName(name string) error {
	if name == "" {
		return errors.New("session name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, "\n\r\t") {
		return errors.New("session name cannot contain newlines or tabs")
	}

	// Check for reserved names
	if strings.HasPrefix(name, " ") || strings.HasSuffix(name, " ") {
		return errors.New("session name cannot start or end with spaces")
	}

	return nil
}

// SessionInfo provides metadata about a session for debugging and introspection.
type SessionInfo struct {
	ID           string
	Name         string
	MessageCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsActive     bool
	SystemPrompt string
}

// GetNthRecentMessage returns the Nth most recent message from the active session.
// N=1 is the most recent message, N=2 is the previous message, etc.
// Returns the message content and an error if the message cannot be retrieved.
func (s *sessionSubcontextImpl) GetNthRecentMessage(n int) (string, error) {
	// Handle invalid input
	if n < 1 {
		return "", fmt.Errorf("invalid message index %d: must be >= 1", n)
	}

	// Check if there's an active session
	if s.ctx.activeSessionID == "" {
		return "", fmt.Errorf("no active session")
	}

	// Get the active session
	session, exists := s.ctx.chatSessions[s.ctx.activeSessionID]
	if !exists {
		return "", fmt.Errorf("active session %s not found", s.ctx.activeSessionID)
	}

	// Check if we have enough messages
	messageCount := len(session.Messages)
	if messageCount == 0 {
		return "", fmt.Errorf("session has no messages")
	}
	if n > messageCount {
		return "", fmt.Errorf("message index %d out of bounds: session has only %d messages", n, messageCount)
	}

	// Get the Nth most recent message (1-based indexing)
	// messages[len-1] is most recent, messages[len-2] is 2nd most recent, etc.
	messageIndex := messageCount - n
	return session.Messages[messageIndex].Content, nil
}

// GetNthChronologicalMessage returns the Nth message from the active session in chronological order.
// N=1 is the first message, N=2 is the second message, etc.
// Returns the message content and an error if the message cannot be retrieved.
func (s *sessionSubcontextImpl) GetNthChronologicalMessage(n int) (string, error) {
	// Handle invalid input
	if n < 1 {
		return "", fmt.Errorf("invalid message index %d: must be >= 1", n)
	}

	// Check if there's an active session
	if s.ctx.activeSessionID == "" {
		return "", fmt.Errorf("no active session")
	}

	// Get the active session
	session, exists := s.ctx.chatSessions[s.ctx.activeSessionID]
	if !exists {
		return "", fmt.Errorf("active session %s not found", s.ctx.activeSessionID)
	}

	// Check if we have enough messages
	messageCount := len(session.Messages)
	if messageCount == 0 {
		return "", fmt.Errorf("session has no messages")
	}
	if n > messageCount {
		return "", fmt.Errorf("message index %d out of bounds: session has only %d messages", n, messageCount)
	}

	// Get the Nth chronological message (1-based indexing)
	// messages[0] is first, messages[1] is second, etc.
	messageIndex := n - 1
	return session.Messages[messageIndex].Content, nil
}

// GetSessionInfo returns detailed information about a session.
func (s *sessionSubcontextImpl) GetSessionInfo(sessionID string) (*SessionInfo, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return &SessionInfo{
		ID:           session.ID,
		Name:         session.Name,
		MessageCount: len(session.Messages),
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    session.UpdatedAt,
		IsActive:     session.ID == s.ctx.activeSessionID,
		SystemPrompt: session.SystemPrompt,
	}, nil
}

// saveSessionToFile saves a session to a JSON file.
func (s *sessionSubcontextImpl) saveSessionToFile(session *neurotypes.ChatSession, filePath string) error {
	// Marshal session to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session to JSON: %w", err)
	}

	// Create parent directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Write JSON data to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// loadSessionFromFile loads a session from a JSON file.
func (s *sessionSubcontextImpl) loadSessionFromFile(filePath string) (*neurotypes.ChatSession, error) {
	// Read JSON data from file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Unmarshal JSON data
	var session neurotypes.ChatSession
	if err := json.Unmarshal(jsonData, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session JSON: %w", err)
	}

	return &session, nil
}
