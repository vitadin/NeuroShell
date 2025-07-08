// Package neurotypes defines session and conversation management types for NeuroShell.
// This file contains the core types for managing conversation history, session state,
// and chat sessions with LLM agents.
package neurotypes

import "time"

// Message represents a single message in the conversation history.
// Messages track the role (user/assistant), content, and timestamp for each interaction.
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionState represents the complete state of a NeuroShell session.
// It includes all variables, message history, and metadata that can be saved and restored.
type SessionState struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
	History   []Message         `json:"history"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// ChatSession represents a conversation session with an LLM agent.
// It maintains conversation history, system context, and metadata for LLM interactions.
type ChatSession struct {
	ID           string    `json:"id"`            // Unique session identifier
	Name         string    `json:"name"`          // User-friendly session name
	SystemPrompt string    `json:"system_prompt"` // System message for LLM context
	Messages     []Message `json:"messages"`      // Ordered conversation history
	CreatedAt    time.Time `json:"created_at"`    // Session creation timestamp
	UpdatedAt    time.Time `json:"updated_at"`    // Last modification timestamp
	IsActive     bool      `json:"is_active"`     // Whether this is the current active session
}
