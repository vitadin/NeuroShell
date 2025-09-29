// Package context provides Claude Code-specific context operations for NeuroShell.
// This file implements ClaudeCodeSubcontext, a focused interface for Claude Code CLI integration
// that eliminates the need for commands to know about global context internals.
package context

import (
	"fmt"

	"neuroshell/pkg/neurotypes"
)

// ClaudeCodeSubcontext provides focused Claude Code operations without exposing full context internals.
// This interface is designed to be passed to commands that only need Claude Code functionality,
// following the Interface Segregation Principle.
type ClaudeCodeSubcontext interface {
	// Core Claude Code operations
	InitializeClaudeCode() error
	IsClaudeCodeInitialized() bool
	ExecuteCommand(command string) (string, error)
	ExecuteCommandAsync(command string) (string, error) // Returns job ID

	// Session management
	CreateSession(sessionName string) error
	SwitchSession(sessionName string) error
	GetCurrentSession() (string, error)
	ListSessions() ([]string, error)
	DeleteSession(sessionName string) error

	// Job management
	WaitForJob(jobID string) (map[string]interface{}, error)
	GetJobStatus(jobID string) (map[string]interface{}, error)
	ListActiveJobs() ([]string, error)
	CancelJob(jobID string) error

	// Authentication and configuration
	CheckAuthentication() (bool, error)
	SetWorkingDirectory(path string) error
	GetWorkingDirectory() (string, error)
	GetClaudeCodeVersion() (string, error)

	// Variable integration
	GetClaudeCodeVariable(name string) (string, error)
	SetClaudeCodeVariable(name string, value string) error
	GetAllClaudeCodeVariables() map[string]string

	// Output handling
	GetJobOutput(jobID string) (string, error)
	GetJobError(jobID string) (string, error)
	StreamJobOutput(jobID string) (<-chan string, error)

	// Status and introspection
	GetServiceStatus() (map[string]interface{}, error)
	GetActiveSessionInfo() (map[string]interface{}, error)
}

// claudeCodeSubcontextImpl implements ClaudeCodeSubcontext using the global service registry.
// This provides a clean abstraction layer that commands can depend on.
type claudeCodeSubcontextImpl struct {
	ctx *NeuroContext
}

// NewClaudeCodeSubcontext creates a new ClaudeCodeSubcontext from a NeuroContext.
// This is the factory function that commands should use to get Claude Code functionality.
func NewClaudeCodeSubcontext(ctx *NeuroContext) ClaudeCodeSubcontext {
	return &claudeCodeSubcontextImpl{ctx: ctx}
}

// InitializeClaudeCode initializes the Claude Code CLI process.
func (c *claudeCodeSubcontextImpl) InitializeClaudeCode() error {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return err
	}
	// For now, just return success since service lookup worked
	return nil
}

// IsClaudeCodeInitialized checks if Claude Code is initialized and ready.
func (c *claudeCodeSubcontextImpl) IsClaudeCodeInitialized() bool {
	_, err := c.getClaudeCodeService()
	return err == nil
}

// ExecuteCommand executes a Claude Code command synchronously.
func (c *claudeCodeSubcontextImpl) ExecuteCommand(command string) (string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return "", err
	}
	// Simplified implementation - return placeholder
	return fmt.Sprintf("Command executed: %s", command), nil
}

// ExecuteCommandAsync executes a Claude Code command asynchronously and returns a job ID.
func (c *claudeCodeSubcontextImpl) ExecuteCommandAsync(_ string) (string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return "", err
	}
	// Simplified implementation - return placeholder job ID
	return "job_123", nil
}

// CreateSession creates a new Claude Code session.
func (c *claudeCodeSubcontextImpl) CreateSession(_ string) error {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return err
	}
	// Simplified implementation
	return nil
}

// SwitchSession switches to an existing Claude Code session.
func (c *claudeCodeSubcontextImpl) SwitchSession(_ string) error {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return err
	}
	// Simplified implementation
	return nil
}

// GetCurrentSession returns the name of the current Claude Code session.
func (c *claudeCodeSubcontextImpl) GetCurrentSession() (string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return "", err
	}
	// Simplified implementation
	return "default", nil
}

// ListSessions returns a list of all Claude Code sessions.
func (c *claudeCodeSubcontextImpl) ListSessions() ([]string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return nil, err
	}
	// Simplified implementation
	return []string{"default"}, nil
}

// DeleteSession deletes a Claude Code session.
func (c *claudeCodeSubcontextImpl) DeleteSession(_ string) error {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return err
	}
	// Simplified implementation
	return nil
}

// WaitForJob waits for a job to complete and returns the result.
func (c *claudeCodeSubcontextImpl) WaitForJob(jobID string) (map[string]interface{}, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return nil, err
	}
	// Simplified implementation
	result := make(map[string]interface{})
	result["job_id"] = jobID
	result["status"] = "completed"
	return result, nil
}

// GetJobStatus returns the current status of a job.
func (c *claudeCodeSubcontextImpl) GetJobStatus(jobID string) (map[string]interface{}, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return nil, err
	}
	// Simplified implementation
	result := make(map[string]interface{})
	result["job_id"] = jobID
	result["status"] = "unknown"
	return result, nil
}

// ListActiveJobs returns a list of active job IDs.
func (c *claudeCodeSubcontextImpl) ListActiveJobs() ([]string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return nil, err
	}
	// Simplified implementation
	return []string{}, nil
}

// CancelJob cancels a running job.
func (c *claudeCodeSubcontextImpl) CancelJob(_ string) error {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return err
	}
	// Simplified implementation
	return nil
}

// CheckAuthentication checks if Claude Code is properly authenticated.
func (c *claudeCodeSubcontextImpl) CheckAuthentication() (bool, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return false, err
	}
	// Simplified implementation
	return true, nil
}

// SetWorkingDirectory sets the working directory for Claude Code operations.
func (c *claudeCodeSubcontextImpl) SetWorkingDirectory(_ string) error {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return err
	}
	// Simplified implementation
	return nil
}

// GetWorkingDirectory returns the current working directory for Claude Code operations.
func (c *claudeCodeSubcontextImpl) GetWorkingDirectory() (string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return "", err
	}
	// Simplified implementation
	return "/current/working/directory", nil
}

// GetClaudeCodeVersion returns the version of the Claude Code CLI.
func (c *claudeCodeSubcontextImpl) GetClaudeCodeVersion() (string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return "", err
	}
	// Simplified implementation
	return "1.5.0", nil
}

// GetClaudeCodeVariable retrieves a Claude Code-specific variable.
func (c *claudeCodeSubcontextImpl) GetClaudeCodeVariable(name string) (string, error) {
	// Claude Code variables are stored with cc_ prefix in the variable system
	variableSubcontext := NewVariableSubcontext(c.ctx)
	return variableSubcontext.GetVariable("cc_" + name)
}

// SetClaudeCodeVariable sets a Claude Code-specific variable.
func (c *claudeCodeSubcontextImpl) SetClaudeCodeVariable(name string, value string) error {
	// Claude Code variables are stored with cc_ prefix in the variable system
	variableSubcontext := NewVariableSubcontext(c.ctx)
	return variableSubcontext.SetSystemVariable("cc_"+name, value)
}

// GetAllClaudeCodeVariables returns all Claude Code-specific variables.
func (c *claudeCodeSubcontextImpl) GetAllClaudeCodeVariables() map[string]string {
	variableSubcontext := NewVariableSubcontext(c.ctx)
	allVars := variableSubcontext.GetAllVariables()

	claudeCodeVars := make(map[string]string)
	for name, value := range allVars {
		if len(name) > 3 && name[:3] == "cc_" {
			// Remove the cc_ prefix for the returned map
			claudeCodeVars[name[3:]] = value
		}
	}

	return claudeCodeVars
}

// GetJobOutput returns the output of a completed or running job.
func (c *claudeCodeSubcontextImpl) GetJobOutput(jobID string) (string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return "", err
	}
	// Simplified implementation
	return fmt.Sprintf("Output for job %s", jobID), nil
}

// GetJobError returns the error output of a job.
func (c *claudeCodeSubcontextImpl) GetJobError(_ string) (string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return "", err
	}
	// Simplified implementation
	return "", nil
}

// StreamJobOutput returns a channel for streaming job output.
func (c *claudeCodeSubcontextImpl) StreamJobOutput(_ string) (<-chan string, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return nil, err
	}
	// Simplified implementation - return a closed channel
	ch := make(chan string)
	close(ch)
	return ch, nil
}

// GetServiceStatus returns the current status of the Claude Code service.
func (c *claudeCodeSubcontextImpl) GetServiceStatus() (map[string]interface{}, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return nil, err
	}
	// Return generic status information
	result := make(map[string]interface{})
	result["initialized"] = true
	result["service_name"] = "claudecode"
	return result, nil
}

// GetActiveSessionInfo returns information about the current Claude Code session.
func (c *claudeCodeSubcontextImpl) GetActiveSessionInfo() (map[string]interface{}, error) {
	_, err := c.getClaudeCodeService()
	if err != nil {
		return nil, err
	}
	// Return generic session information
	result := make(map[string]interface{})
	result["session_name"] = "default"
	result["active"] = true
	return result, nil
}

// getClaudeCodeService retrieves the Claude Code service from the global registry.
// This is a helper method that handles service lookup.
func (c *claudeCodeSubcontextImpl) getClaudeCodeService() (neurotypes.Service, error) {
	// For now, just return a basic error check since this is a stub implementation
	// The actual service interaction will be implemented when commands are added
	return nil, fmt.Errorf("claude code service lookup is not yet implemented")
}
