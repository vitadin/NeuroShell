package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"neuroshell/pkg/neurotypes"
)

// ChatSessionService provides chat session management operations for NeuroShell.
// It handles creation, storage, and retrieval of LLM conversation sessions.
type ChatSessionService struct {
	initialized bool
	sessions    map[string]*neurotypes.ChatSession // Session storage by ID
	nameToID    map[string]string                  // Name to ID mapping
	activeID    string                             // Currently active session ID
}

// NewChatSessionService creates a new ChatSessionService instance.
func NewChatSessionService() *ChatSessionService {
	return &ChatSessionService{
		initialized: false,
		sessions:    make(map[string]*neurotypes.ChatSession),
		nameToID:    make(map[string]string),
		activeID:    "",
	}
}

// Name returns the service name "chat_session" for registration.
func (c *ChatSessionService) Name() string {
	return "chat_session"
}

// Initialize sets up the ChatSessionService for operation.
func (c *ChatSessionService) Initialize(_ neurotypes.Context) error {
	c.initialized = true
	return nil
}

// ValidateSessionName checks if a session name is valid according to NeuroShell naming rules.
func (c *ChatSessionService) ValidateSessionName(name string) error {
	if name == "" {
		return nil // Empty name is valid (will be auto-generated)
	}

	// Session name validation rules
	if len(name) > 64 {
		return fmt.Errorf("session name too long (max 64 characters)")
	}

	if len(name) < 3 {
		return fmt.Errorf("session name too short (min 3 characters)")
	}

	// Allow alphanumeric, hyphens, underscores, and dots
	validName := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*[a-zA-Z0-9]$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid session name: must start and end with alphanumeric, contain only letters, numbers, hyphens, underscores, and dots")
	}

	// Reserved names that cannot be used
	reservedNames := []string{"new", "list", "active", "current", "default", "temp", "temporary"}
	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return fmt.Errorf("session name '%s' is reserved", name)
		}
	}

	return nil
}

// IsSessionNameAvailable checks if a session name is available for use.
func (c *ChatSessionService) IsSessionNameAvailable(name string) bool {
	if !c.initialized {
		return false
	}

	if name == "" {
		return true // Empty name is always available (auto-generated)
	}

	_, exists := c.nameToID[name]
	return !exists
}

// CreateSession creates a new chat session with the given parameters.
// Returns the created session and any error encountered.
func (c *ChatSessionService) CreateSession(name, systemPrompt, initialMessage string) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	// Validate session name
	if err := c.ValidateSessionName(name); err != nil {
		return nil, fmt.Errorf("invalid session name: %w", err)
	}

	// Check name availability
	if name != "" && !c.IsSessionNameAvailable(name) {
		return nil, fmt.Errorf("session name '%s' is already in use", name)
	}

	// Generate unique session ID
	sessionID := uuid.New().String()

	// Generate name if not provided (use first 8 chars of UUID)
	if name == "" {
		name = sessionID[:8]
	}

	// Set default system prompt if not provided
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant."
	}

	// Create new session
	now := time.Now()
	session := &neurotypes.ChatSession{
		ID:           sessionID,
		Name:         name,
		SystemPrompt: systemPrompt,
		Messages:     make([]neurotypes.Message, 0),
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true, // New session becomes active
	}

	// Add initial user message if provided
	if initialMessage != "" {
		userMessage := neurotypes.Message{
			ID:        uuid.New().String(),
			Role:      "user",
			Content:   initialMessage,
			Timestamp: now,
		}
		session.Messages = append(session.Messages, userMessage)
	}

	// Deactivate previous active session
	if c.activeID != "" {
		if prevSession, exists := c.sessions[c.activeID]; exists {
			prevSession.IsActive = false
		}
	}

	// Store session and update mappings
	c.sessions[sessionID] = session
	c.nameToID[name] = sessionID
	c.activeID = sessionID

	return session, nil
}

// GetSession retrieves a session by ID.
func (c *ChatSessionService) GetSession(sessionID string) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	session, exists := c.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session with ID '%s' not found", sessionID)
	}

	return session, nil
}

// GetSessionByName retrieves a session by name.
func (c *ChatSessionService) GetSessionByName(name string) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	sessionID, exists := c.nameToID[name]
	if !exists {
		return nil, fmt.Errorf("session with name '%s' not found", name)
	}

	return c.GetSession(sessionID)
}

// GetSessionByNameOrID retrieves a session by name or ID.
// This is the primary method commands should use for session lookup.
func (c *ChatSessionService) GetSessionByNameOrID(nameOrID string) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	// First try by name
	if sessionID, exists := c.nameToID[nameOrID]; exists {
		return c.GetSession(sessionID)
	}

	// Then try by ID
	if session, exists := c.sessions[nameOrID]; exists {
		return session, nil
	}

	return nil, fmt.Errorf("session '%s' not found (tried both name and ID)", nameOrID)
}

// GetActiveSession returns the currently active session.
func (c *ChatSessionService) GetActiveSession() (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	if c.activeID == "" {
		return nil, fmt.Errorf("no active session")
	}

	return c.GetSession(c.activeID)
}

// SetActiveSession sets the specified session as active by name or ID.
func (c *ChatSessionService) SetActiveSession(nameOrID string) error {
	if !c.initialized {
		return fmt.Errorf("chat session service not initialized")
	}

	session, err := c.GetSessionByNameOrID(nameOrID)
	if err != nil {
		return err
	}

	// Deactivate previous active session
	if c.activeID != "" {
		if prevSession, exists := c.sessions[c.activeID]; exists {
			prevSession.IsActive = false
		}
	}

	// Set new active session
	session.IsActive = true
	c.activeID = session.ID

	return nil
}

// DeleteSession removes a session by name or ID.
func (c *ChatSessionService) DeleteSession(nameOrID string) error {
	if !c.initialized {
		return fmt.Errorf("chat session service not initialized")
	}

	session, err := c.GetSessionByNameOrID(nameOrID)
	if err != nil {
		return err
	}

	// Remove from mappings
	delete(c.sessions, session.ID)
	delete(c.nameToID, session.Name)

	// Clear active session if it was the deleted one
	if c.activeID == session.ID {
		c.activeID = ""
	}

	return nil
}

// ListSessions returns all stored sessions.
func (c *ChatSessionService) ListSessions() []*neurotypes.ChatSession {
	if !c.initialized {
		return make([]*neurotypes.ChatSession, 0)
	}

	sessions := make([]*neurotypes.ChatSession, 0, len(c.sessions))
	for _, session := range c.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// AddMessage adds a message to the specified session.
func (c *ChatSessionService) AddMessage(nameOrID string, role, content string) error {
	if !c.initialized {
		return fmt.Errorf("chat session service not initialized")
	}

	session, err := c.GetSessionByNameOrID(nameOrID)
	if err != nil {
		return err
	}

	message := neurotypes.Message{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = time.Now()

	return nil
}

// GetMessageCount returns the number of messages in the specified session.
func (c *ChatSessionService) GetMessageCount(nameOrID string) (int, error) {
	if !c.initialized {
		return 0, fmt.Errorf("chat session service not initialized")
	}

	session, err := c.GetSessionByNameOrID(nameOrID)
	if err != nil {
		return 0, err
	}

	return len(session.Messages), nil
}

// GetSessionNames returns all session names for easy listing.
func (c *ChatSessionService) GetSessionNames() []string {
	if !c.initialized {
		return make([]string, 0)
	}

	names := make([]string, 0, len(c.nameToID))
	for name := range c.nameToID {
		names = append(names, name)
	}

	return names
}

// HasSessions returns true if any sessions exist.
func (c *ChatSessionService) HasSessions() bool {
	return c.initialized && len(c.sessions) > 0
}
