package context

import (
	"sync"
)

// ErrorStateSubcontext defines the interface for error state management functionality.
// This manages the current and last error states for command execution tracking.
type ErrorStateSubcontext interface {
	// Error state operations
	ResetErrorState()
	SetErrorState(status string, errorMsg string)
	GetCurrentErrorState() (status string, errorMsg string)
	GetLastErrorState() (status string, errorMsg string)
}

// errorStateSubcontext implements the ErrorStateSubcontext interface.
type errorStateSubcontext struct {
	// Error state management
	lastStatus      string       // Last command's exit status
	lastError       string       // Last command's error message
	currentStatus   string       // Current command's exit status (0 = success, non-zero = error)
	currentError    string       // Current command's error message
	errorStateMutex sync.RWMutex // Protects error state fields
}

// NewErrorStateSubcontext creates a new ErrorStateSubcontext instance.
func NewErrorStateSubcontext() ErrorStateSubcontext {
	return &errorStateSubcontext{
		lastStatus:    "0",
		lastError:     "",
		currentStatus: "0",
		currentError:  "",
	}
}

// NewErrorStateSubcontextFromContext creates an ErrorStateSubcontext from an existing NeuroContext.
// This is used by services to get a reference to the context's error state subcontext.
func NewErrorStateSubcontextFromContext(ctx *NeuroContext) ErrorStateSubcontext {
	return ctx.errorStateCtx
}

// ResetErrorState resets the current error state to success (0/"") and moves current to last.
// This should be called before executing a new command.
func (e *errorStateSubcontext) ResetErrorState() {
	e.errorStateMutex.Lock()
	defer e.errorStateMutex.Unlock()

	// Move current error state to last
	e.lastStatus = e.currentStatus
	e.lastError = e.currentError

	// Reset current state to success
	e.currentStatus = "0"
	e.currentError = ""
}

// SetErrorState sets the current error state based on command execution results.
// This should be called after command execution with the results.
func (e *errorStateSubcontext) SetErrorState(status string, errorMsg string) {
	e.errorStateMutex.Lock()
	defer e.errorStateMutex.Unlock()

	e.currentStatus = status
	e.currentError = errorMsg
}

// GetCurrentErrorState returns the current error state (thread-safe read).
func (e *errorStateSubcontext) GetCurrentErrorState() (status string, errorMsg string) {
	e.errorStateMutex.RLock()
	defer e.errorStateMutex.RUnlock()

	return e.currentStatus, e.currentError
}

// GetLastErrorState returns the last error state (thread-safe read).
func (e *errorStateSubcontext) GetLastErrorState() (status string, errorMsg string) {
	e.errorStateMutex.RLock()
	defer e.errorStateMutex.RUnlock()

	return e.lastStatus, e.lastError
}
