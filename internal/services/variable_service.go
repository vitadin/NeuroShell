package services

import (
	"fmt"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// VariableService provides variable management operations for NeuroShell contexts.
type VariableService struct {
	initialized bool
	ctx         neurotypes.Context
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
func (v *VariableService) Initialize(ctx neurotypes.Context) error {
	v.ctx = ctx
	v.initialized = true
	return nil
}

// Get retrieves a variable value from context
func (v *VariableService) Get(name string) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	ctx := v.ctx
	return ctx.GetVariable(name)
}

// Set stores a variable value in context
func (v *VariableService) Set(name, value string) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	ctx := v.ctx
	return ctx.SetVariable(name, value)
}

// SetSystemVariable sets a system variable in context (for internal app use only)
func (v *VariableService) SetSystemVariable(name, value string) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	ctx := v.ctx
	// Cast to NeuroContext to access SetSystemVariable method
	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.SetSystemVariable(name, value)
}

// InterpolateString processes ${var} replacements in a string
func (v *VariableService) InterpolateString(text string) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	ctx := v.ctx
	// Cast to NeuroContext to access InterpolateVariables method
	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return "", fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.InterpolateVariables(text), nil
}

// GetAllVariables returns all variables from context (useful for debugging and listing)
func (v *VariableService) GetAllVariables() (map[string]string, error) {
	if !v.initialized {
		return nil, fmt.Errorf("variable service not initialized")
	}

	ctx := v.ctx
	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.GetAllVariables(), nil
}
