package services

import (
	"fmt"
	"strconv"

	"neuroshell/internal/context"
)

// StackService provides command stacking functionality for the state machine
type StackService struct {
	initialized bool
}

// NewStackService creates a new stack service instance
func NewStackService() *StackService {
	return &StackService{
		initialized: false,
	}
}

// Name returns the service name for registry
func (ss *StackService) Name() string {
	return "stack"
}

// Initialize initializes the stack service
func (ss *StackService) Initialize() error {
	ss.initialized = true
	return nil
}

// Basic stack operations

// PushCommand adds a single command to the execution stack
func (ss *StackService) PushCommand(command string) {
	if !ss.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)

	// Check stack depth limit before pushing
	if err := ss.checkStackDepthLimit(ctx); err != nil {
		// In case of stack overflow, we cannot push more commands
		// Log the error but don't crash the application
		fmt.Printf("Stack overflow prevented: %v\n", err)
		return
	}

	ctx.PushCommand(command)
}

// PushCommands adds multiple commands to the execution stack
func (ss *StackService) PushCommands(commands []string) {
	if !ss.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)

	// Check if adding all commands would exceed stack depth limit
	currentSize := ctx.GetStackSize()
	maxDepth := ss.getMaxStackDepth(ctx)

	if currentSize+len(commands) > maxDepth {
		fmt.Printf("Stack overflow prevented: would exceed maximum depth of %d\n", maxDepth)
		return
	}

	ctx.PushCommands(commands)
}

// PopCommand removes and returns the next command from the stack
func (ss *StackService) PopCommand() (string, bool) {
	if !ss.initialized {
		return "", false
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.PopCommand()
}

// PeekCommand returns the next command without removing it from the stack
func (ss *StackService) PeekCommand() (string, bool) {
	if !ss.initialized {
		return "", false
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.PeekCommand()
}

// ClearStack removes all commands from the execution stack
func (ss *StackService) ClearStack() {
	if !ss.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	ctx.ClearStack()
}

// GetStackSize returns the number of commands in the execution stack
func (ss *StackService) GetStackSize() int {
	if !ss.initialized {
		return 0
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.GetStackSize()
}

// IsEmpty returns true if the stack is empty
func (ss *StackService) IsEmpty() bool {
	if !ss.initialized {
		return true
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.IsStackEmpty()
}

// PeekStack returns a copy of the execution stack without modifying it
func (ss *StackService) PeekStack() []string {
	if !ss.initialized {
		return []string{}
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.PeekStack()
}

// Try block support methods

// PushErrorBoundary pushes error boundary markers for try blocks
func (ss *StackService) PushErrorBoundary(tryID string) {
	if !ss.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	ctx.PushErrorBoundary(tryID)
}

// PopErrorBoundary removes the most recent try block context
func (ss *StackService) PopErrorBoundary() {
	if !ss.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	ctx.PopErrorBoundary()
}

// IsInTryBlock returns true if currently inside a try block
func (ss *StackService) IsInTryBlock() bool {
	if !ss.initialized {
		return false
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.IsInTryBlock()
}

// GetCurrentTryID returns the ID of the current try block
func (ss *StackService) GetCurrentTryID() string {
	if !ss.initialized {
		return ""
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.GetCurrentTryID()
}

// GetCurrentTryDepth returns the current try block depth
func (ss *StackService) GetCurrentTryDepth() int {
	if !ss.initialized {
		return 0
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.GetCurrentTryDepth()
}

// SetTryErrorCaptured marks the current try block as having captured an error
func (ss *StackService) SetTryErrorCaptured() {
	if !ss.initialized {
		return
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	ctx.SetTryErrorCaptured()
}

// IsTryErrorCaptured returns true if the current try block has captured an error
func (ss *StackService) IsTryErrorCaptured() bool {
	if !ss.initialized {
		return false
	}
	ctx := context.GetGlobalContext().(*context.NeuroContext)
	return ctx.IsTryErrorCaptured()
}

// Helper methods for stack overflow protection

// checkStackDepthLimit checks if the current stack depth is within limits
func (ss *StackService) checkStackDepthLimit(ctx *context.NeuroContext) error {
	currentSize := ctx.GetStackSize()
	maxDepth := ss.getMaxStackDepth(ctx)

	if currentSize >= maxDepth {
		return fmt.Errorf("stack depth limit reached (%d commands)", maxDepth)
	}

	return nil
}

// getMaxStackDepth retrieves the maximum stack depth from user configuration
func (ss *StackService) getMaxStackDepth(ctx *context.NeuroContext) int {
	// Try to get user-configured limit from _max_stack_depth variable
	if maxDepthStr, err := ctx.GetVariable("_max_stack_depth"); err == nil && maxDepthStr != "" {
		if maxDepth, err := strconv.Atoi(maxDepthStr); err == nil && maxDepth > 0 {
			return maxDepth
		}
	}

	// Default to 1000 if not configured or invalid
	return 1000
}
