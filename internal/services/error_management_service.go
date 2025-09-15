package services

import (
	"fmt"

	neuroshellcontext "neuroshell/internal/context"
)

// ErrorManagementService provides centralized error state management for NeuroShell commands.
// It manages the @status/@error system variables and provides a clean interface
// for error state operations across the entire application.
type ErrorManagementService struct {
	initialized   bool
	errorStateCtx neuroshellcontext.ErrorStateSubcontext
}

// NewErrorManagementService creates a new ErrorManagementService instance.
func NewErrorManagementService() *ErrorManagementService {
	return &ErrorManagementService{
		initialized: false,
	}
}

// Name returns the service name "error_management" for registration.
func (e *ErrorManagementService) Name() string {
	return "error_management"
}

// Initialize sets up the ErrorManagementService for operation.
func (e *ErrorManagementService) Initialize() error {
	ctx := neuroshellcontext.GetGlobalContext()
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return fmt.Errorf("global context is not a NeuroContext")
	}
	e.errorStateCtx = neuroshellcontext.NewErrorStateSubcontextFromContext(neuroCtx)
	e.initialized = true
	return nil
}

// GetServiceInfo returns information about the ErrorManagementService for debugging.
func (e *ErrorManagementService) GetServiceInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        e.Name(),
		"initialized": e.initialized,
		"description": "Centralized error state management service",
	}
}

// ResetErrorState resets the current error state to success (0/"") and moves current to last.
// This should be called before executing a new command.
func (e *ErrorManagementService) ResetErrorState() error {
	if !e.initialized {
		return fmt.Errorf("error service not initialized")
	}

	e.errorStateCtx.ResetErrorState()
	return nil
}

// SetErrorState sets the current error state based on command execution results.
// This should be called after command execution with the results.
func (e *ErrorManagementService) SetErrorState(status string, errorMsg string) error {
	if !e.initialized {
		return fmt.Errorf("error service not initialized")
	}

	e.errorStateCtx.SetErrorState(status, errorMsg)
	return nil
}

// GetCurrentErrorState returns the current error state (thread-safe read).
func (e *ErrorManagementService) GetCurrentErrorState() (status string, errorMsg string, err error) {
	if !e.initialized {
		return "", "", fmt.Errorf("error service not initialized")
	}

	status, errorMsg = e.errorStateCtx.GetCurrentErrorState()
	return status, errorMsg, nil
}

// GetLastErrorState returns the last error state (thread-safe read).
func (e *ErrorManagementService) GetLastErrorState() (status string, errorMsg string, err error) {
	if !e.initialized {
		return "", "", fmt.Errorf("error service not initialized")
	}

	status, errorMsg = e.errorStateCtx.GetLastErrorState()
	return status, errorMsg, nil
}

// SetErrorStateFromCommandResult is a convenience method that sets error state based on command execution results.
// If err is not nil, it sets status to "1" and errorMsg to err.Error().
// If err is nil, it sets status to "0" and errorMsg to "".
func (e *ErrorManagementService) SetErrorStateFromCommandResult(err error) error {
	if err != nil {
		return e.SetErrorState("1", err.Error())
	}
	return e.SetErrorState("0", "")
}

// IsErrorState returns true if the current status indicates an error (non-zero).
func (e *ErrorManagementService) IsErrorState() (bool, error) {
	status, _, err := e.GetCurrentErrorState()
	if err != nil {
		return false, err
	}
	return status != "0", nil
}

// Note: GetGlobalErrorManagementService is now defined in registry.go
// to follow the established pattern of the service registry system.
