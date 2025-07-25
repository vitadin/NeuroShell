package services

import (
	"fmt"
	"strconv"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// VariableService provides variable management operations for NeuroShell contexts.
type VariableService struct {
	initialized bool
}

// NewVariableService creates a new VariableService instance.
func NewVariableService() *VariableService {
	return &VariableService{
		initialized: false,
	}
}

// Name returns the service name "variable" for registration.
func (v *VariableService) Name() string {
	return "variable"
}

// Initialize sets up the VariableService for operation.
func (v *VariableService) Initialize() error {
	v.initialized = true
	return nil
}

// Get retrieves a variable value from the global context
func (v *VariableService) Get(name string) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	return ctx.GetVariable(name)
}

// Set stores a variable value in the global context
func (v *VariableService) Set(name, value string) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	return ctx.SetVariable(name, value)
}

// SetSystemVariable sets a system variable in the global context (for internal app use only)
func (v *VariableService) SetSystemVariable(name, value string) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	// Cast to NeuroContext to access SetSystemVariable method
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.SetSystemVariable(name, value)
}

// InterpolateString processes ${var} replacements in a string using the global context
func (v *VariableService) InterpolateString(text string) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	// Cast to NeuroContext to access InterpolateVariables method
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return "", fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.InterpolateVariables(text), nil
}

// GetAllVariables returns all variables from the global context (useful for debugging and listing)
func (v *VariableService) GetAllVariables() (map[string]string, error) {
	if !v.initialized {
		return nil, fmt.Errorf("variable service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.GetAllVariables(), nil
}

// UpdateMessageHistoryVariables updates the message history variables (${1}, ${2}, etc.)
// based on the latest messages in the active chat session. ${1} is the latest assistant response,
// ${2} is the latest user message, and so on.
func (v *VariableService) UpdateMessageHistoryVariables(session *neurotypes.ChatSession) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()

	// Clear existing message history variables (1-10)
	for i := 1; i <= 10; i++ {
		varName := strconv.Itoa(i)
		_ = ctx.SetVariable(varName, "")
	}

	// Get messages in reverse order (latest first)
	messages := session.Messages
	if len(messages) == 0 {
		return nil
	}

	// Update variables with recent messages
	// ${1} = latest assistant response
	// ${2} = latest user message
	// ${3} = previous assistant response
	// ${4} = previous user message
	// etc.

	varIndex := 1
	for i := len(messages) - 1; i >= 0 && varIndex <= 10; i-- {
		message := messages[i]
		varName := strconv.Itoa(varIndex)

		// Set the variable to the message content
		err := ctx.SetVariable(varName, message.Content)
		if err != nil {
			return fmt.Errorf("failed to set variable %s: %w", varName, err)
		}

		varIndex++
	}

	return nil
}

// GetEnv retrieves an environment variable value through the context layer.
// This follows the proper architecture: Command -> Service -> Context -> OS.
// Returns empty string if the environment variable doesn't exist (no error).
func (v *VariableService) GetEnv(name string) string {
	if !v.initialized {
		return "" // Return empty string instead of error for graceful handling
	}

	ctx := neuroshellcontext.GetGlobalContext()
	return ctx.GetEnv(name)
}

// SetEnvVariable sets an environment variable through the context layer.
// In test mode, this sets a test environment override.
// In production mode, this sets an actual OS environment variable.
func (v *VariableService) SetEnvVariable(name, value string) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	// Cast to NeuroContext to access environment variable methods
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.SetEnvVariable(name, value)
}
