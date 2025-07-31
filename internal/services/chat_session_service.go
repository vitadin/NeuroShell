package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// ChatSessionService provides chat session management operations for NeuroShell.
// It handles creation, storage, and retrieval of LLM conversation sessions.
// Session data is stored in the global context singleton for persistence.
type ChatSessionService struct {
	initialized bool
}

// NewChatSessionService creates a new ChatSessionService instance.
func NewChatSessionService() *ChatSessionService {
	return &ChatSessionService{
		initialized: false,
	}
}

// Name returns the service name "chat_session" for registration.
func (c *ChatSessionService) Name() string {
	return "chat_session"
}

// Initialize sets up the ChatSessionService for operation.
func (c *ChatSessionService) Initialize() error {
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

	// Note: Reserved names are now handled in CreateSession with auto-versioning
	// No longer rejecting reserved names here - they will be auto-versioned

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

	ctx := neuroshellcontext.GetGlobalContext()
	nameToID := ctx.GetSessionNameToID()
	_, exists := nameToID[name]
	return !exists
}

// isNameReservedOrConflicted checks if a name is reserved or already in use.
func (c *ChatSessionService) isNameReservedOrConflicted(name string) bool {
	// Check if name is reserved
	reservedNames := []string{"new", "list", "active", "current", "default", "temp", "temporary"}
	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return true
		}
	}

	// Check if name is already in use
	return !c.IsSessionNameAvailable(name)
}

// generateAvailableName generates an available session name using versioning.
// If the preferred name conflicts (reserved or already exists), it tries name:v1, name:v2, etc.
// Returns the available name.
func (c *ChatSessionService) generateAvailableName(preferredName string) string {
	if !c.isNameReservedOrConflicted(preferredName) {
		return preferredName
	}

	// Try versioned names: name:v1, name:v2, etc.
	version := 1
	for {
		versionedName := preferredName + ":v" + strconv.Itoa(version)
		if !c.isNameReservedOrConflicted(versionedName) {
			return versionedName
		}
		version++

		// Safety check to avoid infinite loop (highly unlikely but good practice)
		if version > 1000 {
			// Fallback to timestamp-based name
			return preferredName + ":v" + strconv.FormatInt(testutils.GetCurrentTime(neuroshellcontext.GetGlobalContext()).Unix(), 10)
		}
	}
}

// CreateSession creates a new chat session with the given parameters.
// Returns the created session and any error encountered.
func (c *ChatSessionService) CreateSession(name, systemPrompt, initialMessage string) (*neurotypes.ChatSession, error) {
	ctx := neuroshellcontext.GetGlobalContext()
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

	// Generate available name with auto-versioning if needed
	originalName := processedName
	availableName := c.generateAvailableName(processedName)

	// Show warning if name was changed
	if availableName != originalName {
		fmt.Printf("⚠️  Session name '%s' conflicts, created as '%s'\n", originalName, availableName)
		logger.Info("Session name auto-versioned due to conflict", "original", originalName, "created", availableName)
	}

	// Generate unique session ID (deterministic in test mode)
	sessionID := testutils.GenerateUUID(ctx)

	// Use the available name
	name = availableName

	// Set default system prompt if not provided
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant."
	}

	// Create new session (deterministic time in test mode)
	now := testutils.GetCurrentTime(ctx)
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
			ID:        testutils.GenerateUUID(ctx),
			Role:      "user",
			Content:   initialMessage,
			Timestamp: now,
		}
		session.Messages = append(session.Messages, userMessage)
	}

	// Get current sessions and mappings from context
	sessions := ctx.GetChatSessions()
	nameToID := ctx.GetSessionNameToID()

	// Deactivate previous active session
	activeID := ctx.GetActiveSessionID()
	if activeID != "" {
		if prevSession, exists := sessions[activeID]; exists {
			prevSession.IsActive = false
		}
	}

	// Store session and update mappings
	sessions[sessionID] = session
	nameToID[name] = sessionID

	// Update context with new state
	ctx.SetChatSessions(sessions)
	ctx.SetSessionNameToID(nameToID)
	ctx.SetActiveSessionID(sessionID)

	return session, nil
}

// GetSession retrieves a session by ID.
func (c *ChatSessionService) GetSession(sessionID string) (*neurotypes.ChatSession, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.GetSessionWithContext(sessionID, ctx)
}

// GetSessionWithContext retrieves a session by ID using provided context.
func (c *ChatSessionService) GetSessionWithContext(sessionID string, ctx neurotypes.Context) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	sessions := ctx.GetChatSessions()
	session, exists := sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session with ID '%s' not found", sessionID)
	}

	return session, nil
}

// GetSessionByName retrieves a session by name.
func (c *ChatSessionService) GetSessionByName(name string) (*neurotypes.ChatSession, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.GetSessionByNameWithContext(name, ctx)
}

// GetSessionByNameWithContext retrieves a session by name using provided context.
func (c *ChatSessionService) GetSessionByNameWithContext(name string, ctx neurotypes.Context) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	nameToID := ctx.GetSessionNameToID()
	sessionID, exists := nameToID[name]
	if !exists {
		return nil, fmt.Errorf("session with name '%s' not found", name)
	}

	return c.GetSessionWithContext(sessionID, ctx)
}

// GetSessionByNameOrID retrieves a session by name or ID.
// This is the primary method commands should use for session lookup.
func (c *ChatSessionService) GetSessionByNameOrID(nameOrID string) (*neurotypes.ChatSession, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.GetSessionByNameOrIDWithContext(nameOrID, ctx)
}

// GetSessionByNameOrIDWithContext retrieves a session by name or ID using provided context.
func (c *ChatSessionService) GetSessionByNameOrIDWithContext(nameOrID string, ctx neurotypes.Context) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	nameToID := ctx.GetSessionNameToID()
	sessions := ctx.GetChatSessions()

	// First try by name
	if sessionID, exists := nameToID[nameOrID]; exists {
		return c.GetSessionWithContext(sessionID, ctx)
	}

	// Then try by ID
	if session, exists := sessions[nameOrID]; exists {
		return session, nil
	}

	return nil, fmt.Errorf("session '%s' not found (tried both name and ID)", nameOrID)
}

// GetActiveSession returns the currently active session.
func (c *ChatSessionService) GetActiveSession() (*neurotypes.ChatSession, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.GetActiveSessionWithContext(ctx)
}

// GetActiveSessionWithContext returns the currently active session using provided context.
func (c *ChatSessionService) GetActiveSessionWithContext(ctx neurotypes.Context) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	activeID := ctx.GetActiveSessionID()
	if activeID == "" {
		return nil, fmt.Errorf("no active session")
	}

	return c.GetSessionWithContext(activeID, ctx)
}

// SetActiveSession sets the specified session as active by name or ID.
func (c *ChatSessionService) SetActiveSession(nameOrID string) error {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.SetActiveSessionWithContext(nameOrID, ctx)
}

// SetActiveSessionWithContext sets the specified session as active by name or ID using provided context.
func (c *ChatSessionService) SetActiveSessionWithContext(nameOrID string, ctx neurotypes.Context) error {
	if !c.initialized {
		return fmt.Errorf("chat session service not initialized")
	}

	session, err := c.GetSessionByNameOrIDWithContext(nameOrID, ctx)
	if err != nil {
		return err
	}

	return c.setActiveSessionInternal(session.ID, ctx)
}

// setActiveSessionInternal sets a session as active by ID, handling deactivation of previous session.
// This is an internal method used by both SetActiveSession and AddMessage operations.
func (c *ChatSessionService) setActiveSessionInternal(sessionID string, ctx neurotypes.Context) error {
	sessions := ctx.GetChatSessions()

	// Deactivate previous active session
	activeID := ctx.GetActiveSessionID()
	if activeID != "" && activeID != sessionID {
		if prevSession, exists := sessions[activeID]; exists {
			prevSession.IsActive = false
		}
	}

	// Set new active session
	if session, exists := sessions[sessionID]; exists {
		session.IsActive = true
	}
	ctx.SetActiveSessionID(sessionID)

	// Update sessions in context
	ctx.SetChatSessions(sessions)

	return nil
}

// DeleteSession removes a session by name or ID.
func (c *ChatSessionService) DeleteSession(nameOrID string) error {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.DeleteSessionWithContext(nameOrID, ctx)
}

// DeleteSessionWithContext removes a session by name or ID using provided context.
func (c *ChatSessionService) DeleteSessionWithContext(nameOrID string, ctx neurotypes.Context) error {
	if !c.initialized {
		return fmt.Errorf("chat session service not initialized")
	}

	session, err := c.GetSessionByNameOrIDWithContext(nameOrID, ctx)
	if err != nil {
		return err
	}

	sessions := ctx.GetChatSessions()
	nameToID := ctx.GetSessionNameToID()

	// Remove from mappings
	delete(sessions, session.ID)
	delete(nameToID, session.Name)

	// Clear active session if it was the deleted one
	if ctx.GetActiveSessionID() == session.ID {
		ctx.SetActiveSessionID("")
	}

	// Update context with modified mappings
	ctx.SetChatSessions(sessions)
	ctx.SetSessionNameToID(nameToID)

	return nil
}

// ListSessions returns all stored sessions.
func (c *ChatSessionService) ListSessions() []*neurotypes.ChatSession {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.ListSessionsWithContext(ctx)
}

// ListSessionsWithContext returns all stored sessions using provided context.
func (c *ChatSessionService) ListSessionsWithContext(ctx neurotypes.Context) []*neurotypes.ChatSession {
	if !c.initialized {
		return make([]*neurotypes.ChatSession, 0)
	}

	sessionMap := ctx.GetChatSessions()
	sessions := make([]*neurotypes.ChatSession, 0, len(sessionMap))

	// Collect session IDs and sort them for deterministic order
	sessionIDs := make([]string, 0, len(sessionMap))
	for sessionID := range sessionMap {
		sessionIDs = append(sessionIDs, sessionID)
	}

	// Sort IDs to ensure deterministic base order before any command-level sorting
	sort.Strings(sessionIDs)

	// Add sessions in deterministic order
	for _, sessionID := range sessionIDs {
		if session, exists := sessionMap[sessionID]; exists {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// GetSessionsWithPrefix returns all sessions whose names start with the given prefix.
func (c *ChatSessionService) GetSessionsWithPrefix(prefix string) []*neurotypes.ChatSession {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.GetSessionsWithPrefixWithContext(prefix, ctx)
}

// GetSessionsWithPrefixWithContext returns all sessions whose names start with the given prefix using provided context.
func (c *ChatSessionService) GetSessionsWithPrefixWithContext(prefix string, ctx neurotypes.Context) []*neurotypes.ChatSession {
	if !c.initialized {
		return make([]*neurotypes.ChatSession, 0)
	}

	nameToID := ctx.GetSessionNameToID()
	sessions := ctx.GetChatSessions()

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
	ctx := neuroshellcontext.GetGlobalContext()
	return c.FindSessionByPrefixWithContext(identifier, ctx)
}

// FindSessionByPrefixWithContext performs smart session lookup using provided context.
func (c *ChatSessionService) FindSessionByPrefixWithContext(identifier string, ctx neurotypes.Context) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	if identifier == "" {
		return nil, fmt.Errorf("session identifier cannot be empty")
	}

	nameToID := ctx.GetSessionNameToID()
	sessions := ctx.GetChatSessions()

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
	matches := c.GetSessionsWithPrefixWithContext(identifier, ctx)

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
	ctx := neuroshellcontext.GetGlobalContext()
	return c.AddMessageWithContext(nameOrID, role, content, ctx)
}

// AddMessageWithContext adds a message to the specified session using provided context.
func (c *ChatSessionService) AddMessageWithContext(nameOrID string, role, content string, ctx neurotypes.Context) error {
	if !c.initialized {
		return fmt.Errorf("chat session service not initialized")
	}

	session, err := c.GetSessionByNameOrIDWithContext(nameOrID, ctx)
	if err != nil {
		return err
	}

	message := neurotypes.Message{
		ID:        testutils.GenerateUUID(ctx),
		Role:      role,
		Content:   content,
		Timestamp: testutils.GetCurrentTime(ctx),
	}

	session.Messages = append(session.Messages, message)
	session.UpdatedAt = testutils.GetCurrentTime(ctx)

	// Auto-activate the session when a message is added
	// This ensures the most recently active session is always the current active session
	err = c.setActiveSessionInternal(session.ID, ctx)
	if err != nil {
		return fmt.Errorf("failed to activate session after adding message: %w", err)
	}

	// Update session in context
	sessions := ctx.GetChatSessions()
	sessions[session.ID] = session
	ctx.SetChatSessions(sessions)

	return nil
}

// GetMessageCount returns the number of messages in the specified session.
func (c *ChatSessionService) GetMessageCount(nameOrID string) (int, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.GetMessageCountWithContext(nameOrID, ctx)
}

// GetMessageCountWithContext returns the number of messages in the specified session using provided context.
func (c *ChatSessionService) GetMessageCountWithContext(nameOrID string, ctx neurotypes.Context) (int, error) {
	if !c.initialized {
		return 0, fmt.Errorf("chat session service not initialized")
	}

	session, err := c.GetSessionByNameOrIDWithContext(nameOrID, ctx)
	if err != nil {
		return 0, err
	}

	return len(session.Messages), nil
}

// GetSessionNames returns all session names for easy listing.
func (c *ChatSessionService) GetSessionNames() []string {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.GetSessionNamesWithContext(ctx)
}

// GetSessionNamesWithContext returns all session names for easy listing using provided context.
func (c *ChatSessionService) GetSessionNamesWithContext(ctx neurotypes.Context) []string {
	if !c.initialized {
		return make([]string, 0)
	}

	nameToID := ctx.GetSessionNameToID()
	names := make([]string, 0, len(nameToID))
	for name := range nameToID {
		names = append(names, name)
	}

	return names
}

// HasSessions returns true if any sessions exist.
func (c *ChatSessionService) HasSessions() bool {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.HasSessionsWithContext(ctx)
}

// HasSessionsWithContext returns true if any sessions exist using provided context.
func (c *ChatSessionService) HasSessionsWithContext(ctx neurotypes.Context) bool {
	if !c.initialized {
		return false
	}

	sessions := ctx.GetChatSessions()
	return len(sessions) > 0
}

// GenerateDefaultSessionName creates an auto-generated session name following OS folder naming conventions.
// It tries patterns like "Session 1", "Session 2", etc., ensuring uniqueness.
// Fallback patterns include "Chat N", "Work N", "Project N", and finally timestamp-based names.
func (c *ChatSessionService) GenerateDefaultSessionName() string {
	if !c.initialized {
		return "Session 1" // Safe fallback if service not initialized
	}

	// Primary naming patterns (similar to OS "New Folder" conventions)
	baseNames := []string{"Session", "Chat", "Work", "Project"}

	// Try each base name with numbered suffixes
	for _, baseName := range baseNames {
		for i := 1; i <= 999; i++ {
			candidateName := fmt.Sprintf("%s %d", baseName, i)
			if c.IsSessionNameAvailable(candidateName) {
				return candidateName
			}
		}
	}

	// Final fallback: timestamp-based name (guaranteed unique)
	now := testutils.GetCurrentTime(neuroshellcontext.GetGlobalContext())
	return fmt.Sprintf("Session %d", now.Unix())
}

// CopySession creates a deep copy of an existing session with a new identity.
// It preserves all content (system prompt, messages, metadata) but generates
// a new UUID, name (custom or auto-generated), and fresh timestamps.
// The copied session automatically becomes the active session.
func (c *ChatSessionService) CopySession(sourceIdentifier, targetName string) (*neurotypes.ChatSession, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.CopySessionWithContext(sourceIdentifier, targetName, ctx)
}

// CopySessionWithContext creates a deep copy of an existing session using provided context.
func (c *ChatSessionService) CopySessionWithContext(sourceIdentifier, targetName string, ctx neurotypes.Context) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	// Find the source session using smart matching
	sourceSession, err := c.FindSessionByPrefixWithContext(sourceIdentifier, ctx)
	if err != nil {
		return nil, fmt.Errorf("source session lookup failed: %w", err)
	}

	// Validate and process target name
	var processedTargetName string
	if targetName == "" {
		// Auto-generate name if not provided
		processedTargetName = c.GenerateDefaultSessionName()
	} else {
		// Validate custom target name
		processedTargetName, err = c.ValidateSessionName(targetName)
		if err != nil {
			return nil, fmt.Errorf("invalid target session name: %w", err)
		}
	}

	// Check if target name is available
	if !c.IsSessionNameAvailable(processedTargetName) {
		return nil, fmt.Errorf("target session name '%s' is already in use", processedTargetName)
	}

	// Create deep copy with new identity
	newID := testutils.GenerateUUID(ctx)
	now := testutils.GetCurrentTime(ctx)

	// Deep copy all messages
	copiedMessages := make([]neurotypes.Message, len(sourceSession.Messages))
	for i, msg := range sourceSession.Messages {
		copiedMessages[i] = neurotypes.Message{
			ID:        testutils.GenerateUUID(ctx), // Generate new ID for each message
			Role:      msg.Role,                    // Preserve role
			Content:   msg.Content,                 // Preserve content
			Timestamp: msg.Timestamp,               // Preserve original timestamp
		}
	}

	// Create the copied session
	copiedSession := &neurotypes.ChatSession{
		ID:           newID,
		Name:         processedTargetName,
		SystemPrompt: sourceSession.SystemPrompt, // Preserve system prompt
		Messages:     copiedMessages,
		CreatedAt:    now,   // New creation timestamp
		UpdatedAt:    now,   // New update timestamp
		IsActive:     false, // Will be activated below
	}

	// Store the copied session
	sessions := ctx.GetChatSessions()
	nameToID := ctx.GetSessionNameToID()

	sessions[copiedSession.ID] = copiedSession
	nameToID[copiedSession.Name] = copiedSession.ID

	// Update context with new session data
	ctx.SetChatSessions(sessions)
	ctx.SetSessionNameToID(nameToID)

	// Note: The copied session is not automatically activated
	// This allows users to create multiple copies without changing the active session

	logger.Debug("Session copied successfully",
		"source_id", sourceSession.ID,
		"source_name", sourceSession.Name,
		"target_id", copiedSession.ID,
		"target_name", copiedSession.Name,
		"message_count", len(copiedMessages))

	return copiedSession, nil
}

// ExportSessionToJSON exports a session by ID to a JSON file.
// The session must exist and the file path must be valid and writable.
func (c *ChatSessionService) ExportSessionToJSON(sessionID, filepath string) error {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.ExportSessionToJSONWithContext(sessionID, filepath, ctx)
}

// ExportSessionToJSONWithContext exports a session by ID to a JSON file using provided context.
func (c *ChatSessionService) ExportSessionToJSONWithContext(sessionID, filepath string, ctx neurotypes.Context) error {
	if !c.initialized {
		return fmt.Errorf("chat session service not initialized")
	}

	// Get the session by ID
	session, err := c.GetSessionWithContext(sessionID, ctx)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Marshal session to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session to JSON: %w", err)
	}

	// Write JSON data to file
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// ImportSessionFromJSON imports a session from a JSON file and reconstructs it with new identity.
// The imported session gets a new ID, auto-generated name, and current timestamps.
func (c *ChatSessionService) ImportSessionFromJSON(filepath string) (*neurotypes.ChatSession, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return c.ImportSessionFromJSONWithContext(filepath, ctx)
}

// ImportSessionFromJSONWithContext imports a session from a JSON file using provided context.
func (c *ChatSessionService) ImportSessionFromJSONWithContext(filepath string, ctx neurotypes.Context) (*neurotypes.ChatSession, error) {
	if !c.initialized {
		return nil, fmt.Errorf("chat session service not initialized")
	}

	// Read JSON file
	jsonData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	// Unmarshal JSON to session struct
	var originalSession neurotypes.ChatSession
	if err := json.Unmarshal(jsonData, &originalSession); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Reconstruct session with new identity
	reconstructedSession, err := c.reconstructImportedSessionWithContext(&originalSession, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct session: %w", err)
	}

	// Store the reconstructed session
	sessions := ctx.GetChatSessions()
	nameToID := ctx.GetSessionNameToID()

	// Deactivate previous active session
	activeID := ctx.GetActiveSessionID()
	if activeID != "" {
		if prevSession, exists := sessions[activeID]; exists {
			prevSession.IsActive = false
		}
	}

	// Store session and update mappings
	sessions[reconstructedSession.ID] = reconstructedSession
	nameToID[reconstructedSession.Name] = reconstructedSession.ID

	// Update context with new state
	ctx.SetChatSessions(sessions)
	ctx.SetSessionNameToID(nameToID)
	ctx.SetActiveSessionID(reconstructedSession.ID)

	return reconstructedSession, nil
}

// reconstructImportedSessionWithContext creates a new session from imported data with fresh identity.
// Preserves all content but assigns new ID, auto-generated name, and current timestamps.
func (c *ChatSessionService) reconstructImportedSessionWithContext(originalSession *neurotypes.ChatSession, ctx neurotypes.Context) (*neurotypes.ChatSession, error) {
	// Generate new identity
	newID := testutils.GenerateUUID(ctx)
	newName := c.GenerateDefaultSessionName()
	now := testutils.GetCurrentTime(ctx)

	// Create reconstructed session preserving content but with new identity
	reconstructedSession := &neurotypes.ChatSession{
		ID:           newID,
		Name:         newName,
		SystemPrompt: originalSession.SystemPrompt,
		Messages:     originalSession.Messages, // Preserve conversation history with original timestamps
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true, // Imported session becomes active
	}

	return reconstructedSession, nil
}
