package services

import (
	"fmt"
	"strings"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// ChatSessionService provides chat session management operations for NeuroShell.
// It handles creation, storage, and retrieval of LLM conversation sessions.
// Session data is stored in the context rather than service instance for persistence.
type ChatSessionService struct {
	initialized bool
	context     neurotypes.Context // Reference to context for session storage
}

// NewChatSessionService creates a new ChatSessionService instance.
func NewChatSessionService() *ChatSessionService {
	return &ChatSessionService{
		initialized: false,
		context:     nil, // Will be set during initialization
	}
}

// Name returns the service name "chat_session" for registration.
func (c *ChatSessionService) Name() string {
	return "chat_session"
}

// Initialize sets up the ChatSessionService for operation.
func (c *ChatSessionService) Initialize(ctx neurotypes.Context) error {
	c.context = ctx
	c.initialized = true
	return nil
}

// ValidateSessionName checks if a session name is valid according to NeuroShell naming rules.
// It performs smart preprocessing including trimming whitespace and removing quotes.
// Returns the processed name and any validation error.
func (c *ChatSessionService) ValidateSessionName(name string) (string, error) {
	// Smart preprocessing
	processed := c.preprocessSessionName(name)

	// After preprocessing, name cannot be empty
	if processed == "" {
		return "", fmt.Errorf("session name cannot be empty")
	}

	// Session name validation rules
	if len(processed) > 64 {
		return "", fmt.Errorf("session name too long (max 64 characters)")
	}

	// Allow spaces and most printable characters, but not control characters
	for _, char := range processed {
		if char < 32 || char == 127 { // Control characters
			return "", fmt.Errorf("session name contains invalid characters")
		}
	}

	// Reserved names that cannot be used
	reservedNames := []string{"new", "list", "active", "current", "default", "temp", "temporary"}
	lowerName := strings.ToLower(processed)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return "", fmt.Errorf("session name '%s' is reserved", processed)
		}
	}

	return processed, nil
}

// preprocessSessionName performs smart preprocessing on session names.
// It trims whitespace, removes surrounding quotes, and handles basic formatting.
func (c *ChatSessionService) preprocessSessionName(name string) string {
	// Trim whitespace from beginning and end
	processed := strings.TrimSpace(name)

	// Remove surrounding quotes if present
	if len(processed) >= 2 {
		if (processed[0] == '"' && processed[len(processed)-1] == '"') ||
			(processed[0] == '\'' && processed[len(processed)-1] == '\'') {
			processed = processed[1 : len(processed)-1]
			// Trim again after removing quotes
			processed = strings.TrimSpace(processed)
		}
	}

	return processed
}

// IsSessionNameAvailable checks if a session name is available for use.
func (c *ChatSessionService) IsSessionNameAvailable(name string) bool {
	if !c.initialized {
		return false
	}

	if name == "" {
		return true // Empty name is always available (auto-generated)
	}

	nameToID := c.context.GetSessionNameToID()
	_, exists := nameToID[name]
	return !exists
}

// CreateSession creates a new chat session with the given parameters.
// Returns the created session and any error encountered.
func (c *ChatSessionService) CreateSession(name, systemPrompt, initialMessage string) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	// Validate and preprocess session name
	processedName, err := c.ValidateSessionName(name)
	if err != nil {
		return nil, fmt.Errorf("invalid session name: %w", err)
	}

	// For CreateSession, name is now required (no auto-generation)
	if processedName == "" {
		return nil, fmt.Errorf("session name is required")
	}

	// Check name availability
	if !c.IsSessionNameAvailable(processedName) {
		return nil, fmt.Errorf("session name '%s' is already in use", processedName)
	}

	// Generate unique session ID (deterministic in test mode)
	sessionID := testutils.GenerateUUID(c.context)

	// Use the processed name
	name = processedName

	// Set default system prompt if not provided
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant."
	}

	// Create new session (deterministic time in test mode)
	now := testutils.GetCurrentTime(c.context)
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
			ID:        testutils.GenerateUUID(c.context),
			Role:      "user",
			Content:   initialMessage,
			Timestamp: now,
		}
		session.Messages = append(session.Messages, userMessage)
	}

	// Get current sessions and mappings from context
	sessions := c.context.GetChatSessions()
	nameToID := c.context.GetSessionNameToID()

	// Deactivate previous active session
	activeID := c.context.GetActiveSessionID()
	if activeID != "" {
		if prevSession, exists := sessions[activeID]; exists {
			prevSession.IsActive = false
		}
	}

	// Store session and update mappings
	sessions[sessionID] = session
	nameToID[name] = sessionID

	// Update context with new state
	c.context.SetChatSessions(sessions)
	c.context.SetSessionNameToID(nameToID)
	c.context.SetActiveSessionID(sessionID)

	return session, nil
}

// GetSession retrieves a session by ID.
func (c *ChatSessionService) GetSession(sessionID string) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	sessions := c.context.GetChatSessions()
	session, exists := sessions[sessionID]
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

	nameToID := c.context.GetSessionNameToID()
	sessionID, exists := nameToID[name]
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

	nameToID := c.context.GetSessionNameToID()
	sessions := c.context.GetChatSessions()

	// First try by name
	if sessionID, exists := nameToID[nameOrID]; exists {
		return c.GetSession(sessionID)
	}

	// Then try by ID
	if session, exists := sessions[nameOrID]; exists {
		return session, nil
	}

	return nil, fmt.Errorf("session '%s' not found (tried both name and ID)", nameOrID)
}

// GetActiveSession returns the currently active session.
func (c *ChatSessionService) GetActiveSession() (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	activeID := c.context.GetActiveSessionID()
	if activeID == "" {
		return nil, fmt.Errorf("no active session")
	}

	return c.GetSession(activeID)
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

	sessions := c.context.GetChatSessions()

	// Deactivate previous active session
	activeID := c.context.GetActiveSessionID()
	if activeID != "" {
		if prevSession, exists := sessions[activeID]; exists {
			prevSession.IsActive = false
		}
	}

	// Set new active session
	session.IsActive = true
	c.context.SetActiveSessionID(session.ID)

	// Update sessions in context
	c.context.SetChatSessions(sessions)

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

	sessions := c.context.GetChatSessions()
	nameToID := c.context.GetSessionNameToID()

	// Remove from mappings
	delete(sessions, session.ID)
	delete(nameToID, session.Name)

	// Clear active session if it was the deleted one
	if c.context.GetActiveSessionID() == session.ID {
		c.context.SetActiveSessionID("")
	}

	// Update context with modified mappings
	c.context.SetChatSessions(sessions)
	c.context.SetSessionNameToID(nameToID)

	return nil
}

// ListSessions returns all stored sessions.
func (c *ChatSessionService) ListSessions() []*neurotypes.ChatSession {
	if !c.initialized {
		return make([]*neurotypes.ChatSession, 0)
	}

	sessionMap := c.context.GetChatSessions()
	sessions := make([]*neurotypes.ChatSession, 0, len(sessionMap))
	for _, session := range sessionMap {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetSessionsWithPrefix returns all sessions whose names start with the given prefix.
func (c *ChatSessionService) GetSessionsWithPrefix(prefix string) []*neurotypes.ChatSession {
	if !c.initialized {
		return make([]*neurotypes.ChatSession, 0)
	}

	nameToID := c.context.GetSessionNameToID()
	sessions := c.context.GetChatSessions()

	var matches []*neurotypes.ChatSession

	for sessionName, sessionID := range nameToID {
		if strings.HasPrefix(strings.ToLower(sessionName), strings.ToLower(prefix)) {
			if session, exists := sessions[sessionID]; exists {
				matches = append(matches, session)
			}
		}
	}

	return matches
}

// FindSessionByPrefix performs smart session lookup with the following priority:
// 1. Exact name match
// 2. Exact ID match
// 3. Prefix matching (must be unique)
func (c *ChatSessionService) FindSessionByPrefix(identifier string) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	if identifier == "" {
		return nil, fmt.Errorf("session identifier cannot be empty")
	}

	nameToID := c.context.GetSessionNameToID()
	sessions := c.context.GetChatSessions()

	// 1. Try exact name match first
	if sessionID, exists := nameToID[identifier]; exists {
		if session, exists := sessions[sessionID]; exists {
			return session, nil
		}
	}

	// 2. Try exact ID match
	if session, exists := sessions[identifier]; exists {
		return session, nil
	}

	// 3. Try prefix matching
	matches := c.GetSessionsWithPrefix(identifier)

	if len(matches) == 0 {
		return nil, fmt.Errorf("no session found for '%s' (tried exact name, exact ID, prefix match)", identifier)
	}

	if len(matches) > 1 {
		var matchNames []string
		for _, match := range matches {
			matchNames = append(matchNames, match.Name)
		}
		return nil, fmt.Errorf("multiple sessions match prefix '%s': %s", identifier, strings.Join(matchNames, ", "))
	}

	// Exactly one match
	return matches[0], nil
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
		ID:        testutils.GenerateUUID(c.context),
		Role:      role,
		Content:   content,
		Timestamp: testutils.GetCurrentTime(c.context),
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = testutils.GetCurrentTime(c.context)

	// Update session in context
	sessions := c.context.GetChatSessions()
	sessions[session.ID] = session
	c.context.SetChatSessions(sessions)

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

	nameToID := c.context.GetSessionNameToID()
	names := make([]string, 0, len(nameToID))
	for name := range nameToID {
		names = append(names, name)
	}

	return names
}

// HasSessions returns true if any sessions exist.
func (c *ChatSessionService) HasSessions() bool {
	if !c.initialized {
		return false
	}

	sessions := c.context.GetChatSessions()
	return len(sessions) > 0
}
