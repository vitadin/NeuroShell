// Package types defines core interfaces and data structures used throughout NeuroShell.
// This package contains the fundamental types that enable the modular architecture,
// including command interfaces, context management, and service registration.
package types

import "time"

// ParseMode defines how command arguments are parsed from user input.
type ParseMode int

const (
	// ParseModeKeyValue parses arguments as key=value pairs within brackets
	ParseModeKeyValue ParseMode = iota
	// ParseModeRaw treats the entire input as raw text without parsing
	ParseModeRaw
)

// Context provides session state management and variable interpolation for NeuroShell.
// It maintains variables, message history, and session metadata across command executions.
type Context interface {
	GetVariable(name string) (string, error)
	SetVariable(name string, value string) error
	GetMessageHistory(n int) []Message
	GetSessionState() SessionState
	SetTestMode(testMode bool)
	IsTestMode() bool
}

// Service defines the interface for NeuroShell services that provide specific functionality.
// Services are initialized at startup and can be accessed by commands during execution.
type Service interface {
	Name() string
	Initialize(ctx Context) error
}

// Command defines the interface that all NeuroShell commands must implement.
// Commands handle user input and perform specific actions within the shell environment.
type Command interface {
	Name() string
	ParseMode() ParseMode
	Description() string
	Usage() string
	Execute(args map[string]string, input string, ctx Context) error
}

// ServiceRegistry manages the registration and retrieval of services within NeuroShell.
// It provides a centralized way to access services across the application.
type ServiceRegistry interface {
	GetService(name string) (Service, error)
	RegisterService(service Service) error
}

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

// CommandArgs contains the parsed arguments and message content for command execution.
// It provides a structured way to pass user input to command implementations.
type CommandArgs struct {
	Options map[string]string
	Message string
}
