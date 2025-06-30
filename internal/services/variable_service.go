package services

import (
	"fmt"

	"neuroshell/internal/context"
	"neuroshell/pkg/types"
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
func (v *VariableService) Initialize(_ types.Context) error {
	v.initialized = true
	return nil
}

// Get retrieves a variable value from context
func (v *VariableService) Get(name string, ctx types.Context) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	return ctx.GetVariable(name)
}

// Set stores a variable value in context
func (v *VariableService) Set(name, value string, ctx types.Context) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	return ctx.SetVariable(name, value)
}

// InterpolateString processes ${var} replacements in a string
func (v *VariableService) InterpolateString(text string, ctx types.Context) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	// Cast to NeuroContext to access InterpolateVariables method
	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return "", fmt.Errorf("context is not a NeuroContext")
	}

	return neuroCtx.InterpolateVariables(text), nil
}

// GetAllVariables returns all variables from context (useful for debugging)
func (v *VariableService) GetAllVariables(ctx types.Context) (map[string]string, error) {
	if !v.initialized {
		return nil, fmt.Errorf("variable service not initialized")
	}

	_, ok := ctx.(*context.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	// This would require adding a method to NeuroContext to expose all variables
	// For now, return empty map as placeholder
	result := make(map[string]string)

	// TODO: Add GetAllVariables method to NeuroContext if needed
	return result, nil
}
