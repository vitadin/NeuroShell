package services

import (
	"fmt"

	neuroshellcontext "neuroshell/internal/context"
)

// VariableService provides variable management operations for NeuroShell contexts.
// This refactored version uses dependency injection for better testability and architecture.
type VariableService struct {
	initialized bool
	varCtx      neuroshellcontext.VariableSubcontext
}

// NewVariableService creates a new VariableService instance.
func NewVariableService() *VariableService {
	return &VariableService{
		initialized: false,
		varCtx:      nil, // Will be set during initialization
	}
}

// Name returns the service name "variable" for registration.
func (v *VariableService) Name() string {
	return "variable"
}

// Initialize sets up the VariableService for operation.
// It uses the global context for backward compatibility but injects a VariableSubcontext.
func (v *VariableService) Initialize() error {
	// For backward compatibility, get the global context and create subcontext
	ctx := neuroshellcontext.GetGlobalContext()
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return fmt.Errorf("global context is not a NeuroContext")
	}

	v.varCtx = neuroshellcontext.NewVariableSubcontext(neuroCtx)
	v.initialized = true
	return nil
}

// Get retrieves a variable value from the variable subcontext
func (v *VariableService) Get(name string) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	if v.varCtx == nil {
		return "", fmt.Errorf("variable subcontext not available")
	}

	return v.varCtx.GetVariable(name)
}

// Set stores a variable value in the variable subcontext
func (v *VariableService) Set(name, value string) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	if v.varCtx == nil {
		return fmt.Errorf("variable subcontext not available")
	}

	// Validate variable name before setting
	if err := v.ValidateVariableName(name); err != nil {
		return err
	}

	return v.varCtx.SetVariable(name, value)
}

// SetSystemVariable sets a system variable in the variable subcontext (for internal app use only)
func (v *VariableService) SetSystemVariable(name, value string) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	if v.varCtx == nil {
		return fmt.Errorf("variable subcontext not available")
	}

	return v.varCtx.SetSystemVariable(name, value)
}

// InterpolateString processes ${var} replacements in a string using the variable subcontext
func (v *VariableService) InterpolateString(text string) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	if v.varCtx == nil {
		return "", fmt.Errorf("variable subcontext not available")
	}

	return v.varCtx.InterpolateVariables(text), nil
}

// GetAllVariables returns all variables from the variable subcontext (useful for debugging and listing)
func (v *VariableService) GetAllVariables() (map[string]string, error) {
	if !v.initialized {
		return nil, fmt.Errorf("variable service not initialized")
	}

	if v.varCtx == nil {
		return nil, fmt.Errorf("variable subcontext not available")
	}

	return v.varCtx.GetAllVariables(), nil
}

// GetEnv retrieves an environment variable value through the variable subcontext.
// This follows the proper architecture: Command -> Service -> Context -> OS.
// Returns empty string if the environment variable doesn't exist (no error).
func (v *VariableService) GetEnv(name string) string {
	if !v.initialized {
		return "" // Return empty string instead of error for graceful handling
	}

	if v.varCtx == nil {
		return ""
	}

	return v.varCtx.GetEnv(name)
}

// SetEnvVariable sets an environment variable through the variable subcontext.
// In test mode, this sets a test environment override.
// In production mode, this sets an actual OS environment variable.
func (v *VariableService) SetEnvVariable(name, value string) error {
	if !v.initialized {
		return fmt.Errorf("variable service not initialized")
	}

	if v.varCtx == nil {
		return fmt.Errorf("variable subcontext not available")
	}

	return v.varCtx.SetEnvVariable(name, value)
}

// GetEnvVariable retrieves an environment variable value through the variable subcontext.
// This is a pure function that only gets the environment variable without side effects.
// Respects test mode for environment variable retrieval.
func (v *VariableService) GetEnvVariable(name string) (string, error) {
	if !v.initialized {
		return "", fmt.Errorf("variable service not initialized")
	}

	if v.varCtx == nil {
		return "", fmt.Errorf("variable subcontext not available")
	}

	value := v.varCtx.GetEnvVariable(name)
	return value, nil
}

// ValidateVariableName checks if a variable name follows NeuroShell naming conventions.
// This delegates to the context package's validation function.
func (v *VariableService) ValidateVariableName(name string) error {
	return neuroshellcontext.ValidateVariableName(name)
}

// AnalyzeVariable analyzes a variable name and provides metadata about it.
// This delegates to the context package's analysis function.
func (v *VariableService) AnalyzeVariable(name string) neuroshellcontext.VariableInfo {
	return neuroshellcontext.AnalyzeVariable(name)
}
