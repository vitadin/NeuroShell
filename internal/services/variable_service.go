package services

import (
	"fmt"

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
func (v *VariableService) Initialize(_ neurotypes.Context) error {
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
