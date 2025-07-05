// Package neurotypes defines core interfaces and data structures used throughout NeuroShell.
// This package contains the fundamental neurotypes that enable the modular architecture,
// including command interfaces, context management, and service registration.
package neurotypes

import "time"

// ParseMode defines how command arguments are parsed from user input.
type ParseMode int

const (
	// ParseModeKeyValue parses arguments as key=value pairs within brackets
	ParseModeKeyValue ParseMode = iota
	// ParseModeRaw treats the entire input as raw text without parsing
	ParseModeRaw
	// ParseModeWithOptions parses arguments with options support
	ParseModeWithOptions
)

// Context provides session state management and variable interpolation for NeuroShell.
// It maintains variables, message history, session metadata, and chat sessions across command executions.
type Context interface {
	GetVariable(name string) (string, error)
	SetVariable(name string, value string) error
	GetMessageHistory(n int) []Message
	GetSessionState() SessionState
	SetTestMode(testMode bool)
	IsTestMode() bool

	// Chat session storage methods
	GetChatSessions() map[string]*ChatSession
	SetChatSessions(sessions map[string]*ChatSession)
	GetSessionNameToID() map[string]string
	SetSessionNameToID(nameToID map[string]string)
	GetActiveSessionID() string
	SetActiveSessionID(sessionID string)

	// Model storage methods (bidirectional mapping)
	GetModels() map[string]*ModelConfig
	SetModels(models map[string]*ModelConfig)
	GetModelNameToID() map[string]string
	SetModelNameToID(nameToID map[string]string)
	GetModelIDToName() map[string]string
	SetModelIDToName(idToName map[string]string)
	ModelNameExists(name string) bool
	ModelIDExists(id string) bool
}

// Service defines the interface for NeuroShell services that provide specific functionality.
// Services are initialized at startup and can be accessed by commands during execution.
type Service interface {
	Name() string
	Initialize(ctx Context) error
}

// Command defines the interface that all NeuroShell commands must implement.
// Commands handle user input and perform specific actions within the shell environment.
// Commands access services through the global service registry, not through direct context access.
type Command interface {
	Name() string
	ParseMode() ParseMode
	Description() string
	Usage() string
	HelpInfo() HelpInfo
	Execute(args map[string]string, input string) error
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

// HelpInfo represents structured help information for a command.
// It provides rich help data that can be rendered in both plain text and styled formats.
type HelpInfo struct {
	Command     string        `json:"command"`            // Command name
	Description string        `json:"description"`        // Brief description of what the command does
	Usage       string        `json:"usage"`              // Usage syntax
	ParseMode   ParseMode     `json:"parse_mode"`         // How the command parses arguments
	Options     []HelpOption  `json:"options,omitempty"`  // Command options/parameters
	Examples    []HelpExample `json:"examples,omitempty"` // Usage examples
	Notes       []string      `json:"notes,omitempty"`    // Additional notes or warnings
}

// HelpOption represents a command option/parameter with detailed information.
type HelpOption struct {
	Name        string `json:"name"`              // Option name
	Description string `json:"description"`       // What this option does
	Required    bool   `json:"required"`          // Whether this option is required
	Type        string `json:"type"`              // Data type (string, bool, int, etc.)
	Default     string `json:"default,omitempty"` // Default value if not specified
}

// HelpExample represents a usage example with explanation.
type HelpExample struct {
	Command     string `json:"command"`     // Example command
	Description string `json:"description"` // What this example demonstrates
}
