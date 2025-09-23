// Package services provides keyboard shortcut management for NeuroShell.
package services

import (
	"fmt"
	"os"
	"sync"

	"neuroshell/internal/logger"
)

// ShortcutHandler defines the function signature for shortcut handlers
type ShortcutHandler func() error

// Shortcut represents a keyboard shortcut with its metadata and handler
type Shortcut struct {
	KeyCode     rune            // ASCII code for the key combination
	Name        string          // Human-readable name (e.g., "Ctrl+S")
	Description string          // What the shortcut does
	Handler     ShortcutHandler // Function to execute when shortcut is triggered
}

// ShortcutService manages keyboard shortcuts for NeuroShell
type ShortcutService struct {
	initialized bool
	shortcuts   map[rune]*Shortcut
	mutex       sync.RWMutex
}

// NewShortcutService creates a new shortcut service instance
func NewShortcutService() *ShortcutService {
	return &ShortcutService{
		initialized: false,
		shortcuts:   make(map[rune]*Shortcut),
	}
}

// Name returns the service name for registry
func (s *ShortcutService) Name() string {
	return "shortcut"
}

// Initialize initializes the shortcut service and registers default shortcuts
func (s *ShortcutService) Initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.initialized = true

	// Register Ctrl+S shortcut for saving all sessions
	ctrlS := &Shortcut{
		KeyCode:     19, // ASCII code for Ctrl+S
		Name:        "Ctrl+S",
		Description: "Save all sessions",
		Handler:     s.handleSaveAllSessions,
	}

	s.shortcuts[19] = ctrlS

	logger.Debug("ShortcutService initialized with default shortcuts")
	return nil
}

// RegisterShortcut adds a new shortcut to the service
func (s *ShortcutService) RegisterShortcut(keyCode rune, name, description string, handler ShortcutHandler) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		return fmt.Errorf("shortcut service not initialized")
	}

	if _, exists := s.shortcuts[keyCode]; exists {
		return fmt.Errorf("shortcut for key code %d already registered", keyCode)
	}

	s.shortcuts[keyCode] = &Shortcut{
		KeyCode:     keyCode,
		Name:        name,
		Description: description,
		Handler:     handler,
	}

	return nil
}

// ExecuteShortcut executes the handler for the given key code
func (s *ShortcutService) ExecuteShortcut(keyCode rune) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return false
	}

	shortcut, exists := s.shortcuts[keyCode]
	if !exists {
		return false
	}

	// Execute the shortcut handler in a safe way
	go func() {
		if err := shortcut.Handler(); err != nil {
			logger.Error("Shortcut execution failed", "shortcut", shortcut.Name, "error", err)
		}
	}()

	return true
}

// GetShortcuts returns a list of all registered shortcuts
func (s *ShortcutService) GetShortcuts() []*Shortcut {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	shortcuts := make([]*Shortcut, 0, len(s.shortcuts))
	for _, shortcut := range s.shortcuts {
		shortcuts = append(shortcuts, shortcut)
	}

	return shortcuts
}

// handleSaveAllSessions is the default handler for Ctrl+S shortcut
func (s *ShortcutService) handleSaveAllSessions() error {
	// Get the chat session service
	sessionService, err := GetGlobalRegistry().GetService("chat_session")
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	chatService, ok := sessionService.(*ChatSessionService)
	if !ok {
		return fmt.Errorf("chat session service type assertion failed")
	}

	// Save all sessions
	count, err := chatService.SaveAllSessions()
	if err != nil {
		// Print error message to stderr
		fmt.Fprintf(os.Stderr, "Failed to save sessions: %v\n", err)
		return err
	}

	// Print success message to stdout
	switch count {
	case 0:
		fmt.Println("No sessions to save")
	case 1:
		fmt.Println("Saved 1 session")
	default:
		fmt.Printf("Saved %d sessions\n", count)
	}

	return nil
}

// GetGlobalShortcutService returns the global shortcut service instance
func GetGlobalShortcutService() (*ShortcutService, error) {
	service, err := GetGlobalRegistry().GetService("shortcut")
	if err != nil {
		return nil, fmt.Errorf("shortcut service not registered: %w", err)
	}
	shortcutService, ok := service.(*ShortcutService)
	if !ok {
		return nil, fmt.Errorf("shortcut service type assertion failed")
	}
	return shortcutService, nil
}
