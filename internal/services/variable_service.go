package services

import (
	"fmt"

	neuroshellcontext "neuroshell/internal/context"
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

// GetEnvVariable retrieves an environment variable value through the context layer.
// This is a pure function that only gets the environment variable without side effects.
// Respects test mode for environment variable retrieval.
func (v *VariableService) GetEnvVariable(name string) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	// Cast to NeuroContext to access GetEnvVariable method
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return "", fmt.Errorf("context is not a NeuroContext")
	}

	value := neuroCtx.GetEnvVariable(name)
	return value, nil
}
