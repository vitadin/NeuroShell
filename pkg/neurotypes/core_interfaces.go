// Package neurotypes defines core architectural interfaces for NeuroShell.
// This file contains the fundamental interfaces that define the system's structure
// and enable the modular architecture, including context management, service registration,
// and command handling.
package neurotypes

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

	// LLM client storage methods
	GetLLMClient(apiKey string) (LLMClient, bool)
	SetLLMClient(apiKey string, client LLMClient)

	// Testing and debugging methods
	GetAllVariables() map[string]string
	SetVariableWithValidation(name string, value string) error
}

// Service defines the interface for NeuroShell services that provide specific functionality.
// Services are initialized at startup and can be accessed by commands during execution.
// Services use the global context singleton for all state access.
type Service interface {
	Name() string
	Initialize() error
}

// Command defines the interface that all NeuroShell commands must implement.
// Commands handle user input and perform specific actions within the shell environment.
// Commands should only interact with services, not context directly.
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
