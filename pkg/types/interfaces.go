package types

import "time"

type Context interface {
	GetVariable(name string) (string, error)
	SetVariable(name string, value string) error
	GetMessageHistory(n int) []Message
	GetSessionState() SessionState
}

type Service interface {
	Name() string
	Initialize(ctx Context) error
}

type Command interface {
	Name() string
	Execute(args []string, input string, services ServiceRegistry) error
}

type ServiceRegistry interface {
	GetService(name string) (Service, error)
	RegisterService(service Service) error
}

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type SessionState struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Variables map[string]string `json:"variables"`
	History   []Message         `json:"history"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type CommandArgs struct {
	Options map[string]string
	Message string
}