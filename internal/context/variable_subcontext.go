// Package context provides variable-specific context operations for NeuroShell.
// This file implements VariableSubcontext, a focused interface for variable management
// that eliminates the need for services to know about global context internals.
package context

import (
	"fmt"
	"strings"
)

// VariableSubcontext provides focused variable operations without exposing full context internals.
// This interface is designed to be passed to services that only need variable functionality,
// following the Interface Segregation Principle.
type VariableSubcontext interface {
	// Core variable operations
	GetVariable(name string) (string, error)
	SetVariable(name string, value string) error
	GetAllVariables() map[string]string

	// Variable interpolation
	InterpolateVariables(text string) string

	// System variable management (for internal use)
	SetSystemVariable(name string, value string) error

	// Environment variable access
	GetEnv(name string) string
	SetEnvVariable(name string, value string) error
	GetEnvVariable(name string) string
}

// variableSubcontextImpl implements VariableSubcontext using a NeuroContext.
// This provides a clean abstraction layer that services can depend on.
type variableSubcontextImpl struct {
	ctx *NeuroContext
}

// NewVariableSubcontext creates a new VariableSubcontext from a NeuroContext.
// This is the factory function that services should use to get variable functionality.
func NewVariableSubcontext(ctx *NeuroContext) VariableSubcontext {
	return &variableSubcontextImpl{ctx: ctx}
}

// GetVariable retrieves a variable value by name, supporting both user and system variables.
func (v *variableSubcontextImpl) GetVariable(name string) (string, error) {
	return v.ctx.GetVariable(name)
}

// SetVariable sets a user variable, preventing modification of system variables.
func (v *variableSubcontextImpl) SetVariable(name string, value string) error {
	return v.ctx.SetVariable(name, value)
}

// GetAllVariables returns all variables including both user variables and computed system variables.
func (v *variableSubcontextImpl) GetAllVariables() map[string]string {
	return v.ctx.GetAllVariables()
}

// InterpolateVariables replaces ${variable} placeholders in text with their values.
func (v *variableSubcontextImpl) InterpolateVariables(text string) string {
	return v.ctx.InterpolateVariables(text)
}

// SetSystemVariable sets a system variable, allowing internal app use only.
func (v *variableSubcontextImpl) SetSystemVariable(name string, value string) error {
	return v.ctx.SetSystemVariable(name, value)
}

// GetEnv retrieves an environment variable value.
func (v *variableSubcontextImpl) GetEnv(name string) string {
	return v.ctx.GetEnv(name)
}

// SetEnvVariable sets an environment variable through the context layer.
func (v *variableSubcontextImpl) SetEnvVariable(name string, value string) error {
	return v.ctx.SetEnvVariable(name, value)
}

// GetEnvVariable retrieves an environment variable value (pure function variant).
func (v *variableSubcontextImpl) GetEnvVariable(name string) string {
	return v.ctx.GetEnvVariable(name)
}

// ValidateVariableName checks if a variable name follows NeuroShell naming conventions.
// This is a utility function that can be used by services and commands.
func ValidateVariableName(name string) error {
	if name == "" {
		return fmt.Errorf("variable name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, " \t\n\r") {
		return fmt.Errorf("variable name cannot contain whitespace")
	}

	// Validate prefix-based naming conventions
	if strings.HasPrefix(name, "@") || strings.HasPrefix(name, "#") {
		return fmt.Errorf("variable name cannot start with system prefixes @ or #")
	}

	if strings.HasPrefix(name, "_") {
		// Check if it's a whitelisted global variable using a default context
		// Note: This function is used in validation contexts where we may not have a specific context
		// We use a temporary configuration subcontext to get the allowed variables list
		tempConfig := NewConfigurationSubcontext()
		if !tempConfig.IsAllowedGlobalVariable(name) {
			return fmt.Errorf("variable name cannot start with _ unless whitelisted")
		}
	}

	return nil
}

// IsSystemVariable checks if a variable name is a system variable.
func IsSystemVariable(name string) bool {
	return strings.HasPrefix(name, "@") || strings.HasPrefix(name, "#") || strings.HasPrefix(name, "_")
}

// VariableInfo provides metadata about a variable for debugging and introspection.
type VariableInfo struct {
	Name        string
	Value       string
	Type        VariableType
	IsSystem    bool
	IsReadOnly  bool
	Description string
}

// VariableType represents the different types of variables in NeuroShell.
type VariableType string

const (
	// TypeUser represents user-defined variables
	TypeUser VariableType = "user"
	// TypeSystem represents system variables (e.g., @pwd, @user)
	TypeSystem VariableType = "system"
	// TypeCommand represents command output or configuration variables
	TypeCommand VariableType = "command"
	// TypeMetadata represents metadata variables (e.g., #session_id, #message_count)
	TypeMetadata VariableType = "metadata"
)

// AnalyzeVariable analyzes a variable name and provides metadata about it.
func AnalyzeVariable(name string) VariableInfo {
	info := VariableInfo{
		Name:       name,
		IsSystem:   IsSystemVariable(name),
		IsReadOnly: false,
	}

	switch {
	case strings.HasPrefix(name, "@"):
		info.Type = TypeSystem
		info.Description = "System variable (e.g., @pwd, @user, @date)"
		info.IsReadOnly = true
	case strings.HasPrefix(name, "#"):
		info.Type = TypeMetadata
		info.Description = "Metadata variable (e.g., #session_id, #message_count)"
		info.IsReadOnly = true
	case strings.HasPrefix(name, "_"):
		info.Type = TypeCommand
		info.Description = "Command output or configuration variable"
		// Check if it's read-only based on whitelist
		// Use a temporary configuration subcontext to check allowed variables
		tempConfig := NewConfigurationSubcontext()
		info.IsReadOnly = !tempConfig.IsAllowedGlobalVariable(name)
	default:
		info.Type = TypeUser
		info.Description = "User-defined variable"
		info.IsReadOnly = false
	}

	return info
}
