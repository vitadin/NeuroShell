// Package services provides the core business logic services for NeuroShell.
// It implements service interfaces for script execution, variable management, and system operations.
package services

import (
	"fmt"

	"neuroshell/internal/context"
	"neuroshell/internal/parser"
	"neuroshell/pkg/types"
)

// ExecutorService manages command execution queue and orchestrates script processing.
type ExecutorService struct {
	initialized bool
}

// NewExecutorService creates a new ExecutorService instance.
func NewExecutorService() *ExecutorService {
	return &ExecutorService{
		initialized: false,
	}
}

// Name returns the service name "executor" for registration.
func (e *ExecutorService) Name() string {
	return "executor"
}

// Initialize sets up the ExecutorService for operation.
func (e *ExecutorService) Initialize(_ types.Context) error {
	e.initialized = true
	return nil
}

// ParseCommand converts a string command to a parsed Command struct
func (e *ExecutorService) ParseCommand(text string) (*parser.Command, error) {
	if !e.initialized {
		return nil, fmt.Errorf("executor service not initialized")
	}

	cmd := parser.ParseInput(text)
	return cmd, nil
}

// GetNextCommand returns the next command from the queue without executing it
func (e *ExecutorService) GetNextCommand(ctx types.Context) (*parser.Command, error) {
	if !e.initialized {
		return nil, fmt.Errorf("executor service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	commandText, hasMore := neuroCtx.DequeueCommand()
	if !hasMore {
		return nil, nil // No more commands
	}

	return e.ParseCommand(commandText)
}

// GetQueueStatus returns information about the execution queue
func (e *ExecutorService) GetQueueStatus(ctx types.Context) (map[string]interface{}, error) {
	if !e.initialized {
		return nil, fmt.Errorf("executor service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("context is not a NeuroContext")
	}

	status := make(map[string]interface{})
	status["queue_size"] = neuroCtx.GetQueueSize()
	status["pending_commands"] = neuroCtx.PeekQueue()

	if execStatus, exists := neuroCtx.GetScriptMetadata("execution_status"); exists {
		status["execution_status"] = execStatus
	}
	if execCount, exists := neuroCtx.GetScriptMetadata("executed_commands"); exists {
		status["executed_commands"] = execCount
	}
	if execError, exists := neuroCtx.GetScriptMetadata("execution_error"); exists {
		status["execution_error"] = execError
	}

	return status, nil
}

// MarkCommandExecuted updates execution progress in context
func (e *ExecutorService) MarkCommandExecuted(ctx types.Context) error {
	if !e.initialized {
		return fmt.Errorf("executor service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	// Increment executed command count
	executed := 0
	if count, exists := neuroCtx.GetScriptMetadata("executed_commands"); exists {
		if countInt, ok := count.(int); ok {
			executed = countInt
		}
	}
	neuroCtx.SetScriptMetadata("executed_commands", executed+1)

	return nil
}

// MarkExecutionError records an execution error in context
func (e *ExecutorService) MarkExecutionError(ctx types.Context, err error, command string) error {
	if !e.initialized {
		return fmt.Errorf("executor service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	neuroCtx.SetScriptMetadata("execution_error", err.Error())
	neuroCtx.SetScriptMetadata("failed_command", command)
	neuroCtx.SetScriptMetadata("execution_status", "failed")

	return nil
}

// MarkExecutionComplete marks successful completion of all commands
func (e *ExecutorService) MarkExecutionComplete(ctx types.Context) error {
	if !e.initialized {
		return fmt.Errorf("executor service not initialized")
	}

	neuroCtx, ok := ctx.(*context.NeuroContext)
	if !ok {
		return fmt.Errorf("context is not a NeuroContext")
	}

	neuroCtx.SetScriptMetadata("execution_status", "completed")

	return nil
}
