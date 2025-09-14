package services

import (
	"fmt"

	neuroshellcontext "neuroshell/internal/context"
)

// StackService provides command stacking functionality for the state machine
type StackService struct {
	initialized bool
	stackCtx    neuroshellcontext.StackSubcontext
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
	ctx := neuroshellcontext.GetGlobalContext()
	neuroCtx, ok := ctx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return fmt.Errorf("global context is not a NeuroContext")
	}
	ss.stackCtx = neuroshellcontext.NewStackSubcontextFromContext(neuroCtx)
	ss.initialized = true
	return nil
}

// Basic stack operations

// PushCommand adds a single command to the execution stack
func (ss *StackService) PushCommand(command string) {
	if !ss.initialized {
		return
	}
	ss.stackCtx.PushCommand(command)
}

// PushCommands adds multiple commands to the execution stack
func (ss *StackService) PushCommands(commands []string) {
	if !ss.initialized {
		return
	}
	ss.stackCtx.PushCommands(commands)
}

// PopCommand removes and returns the next command from the stack
func (ss *StackService) PopCommand() (string, bool) {
	if !ss.initialized {
		return "", false
	}
	return ss.stackCtx.PopCommand()
}

// PeekCommand returns the next command without removing it from the stack
func (ss *StackService) PeekCommand() (string, bool) {
	if !ss.initialized {
		return "", false
	}
	return ss.stackCtx.PeekCommand()
}

// ClearStack removes all commands from the execution stack
func (ss *StackService) ClearStack() {
	if !ss.initialized {
		return
	}
	ss.stackCtx.ClearStack()
}

// GetStackSize returns the number of commands in the execution stack
func (ss *StackService) GetStackSize() int {
	if !ss.initialized {
		return 0
	}
	return ss.stackCtx.GetStackSize()
}

// IsEmpty returns true if the stack is empty
func (ss *StackService) IsEmpty() bool {
	if !ss.initialized {
		return true
	}
	return ss.stackCtx.IsStackEmpty()
}

// PeekStack returns a copy of the execution stack without modifying it
func (ss *StackService) PeekStack() []string {
	if !ss.initialized {
		return []string{}
	}
	return ss.stackCtx.PeekStack()
}

// Try block support methods

// PushErrorBoundary pushes error boundary markers for try blocks
func (ss *StackService) PushErrorBoundary(tryID string) {
	if !ss.initialized {
		return
	}
	ss.stackCtx.PushErrorBoundary(tryID)
}

// PopErrorBoundary removes the most recent try block context
func (ss *StackService) PopErrorBoundary() {
	if !ss.initialized {
		return
	}
	ss.stackCtx.PopErrorBoundary()
}

// IsInTryBlock returns true if currently inside a try block
func (ss *StackService) IsInTryBlock() bool {
	if !ss.initialized {
		return false
	}
	return ss.stackCtx.IsInTryBlock()
}

// GetCurrentTryID returns the ID of the current try block
func (ss *StackService) GetCurrentTryID() string {
	if !ss.initialized {
		return ""
	}
	return ss.stackCtx.GetCurrentTryID()
}

// GetCurrentTryDepth returns the current try block depth
func (ss *StackService) GetCurrentTryDepth() int {
	if !ss.initialized {
		return 0
	}
	return ss.stackCtx.GetCurrentTryDepth()
}

// SetTryErrorCaptured marks the current try block as having captured an error
func (ss *StackService) SetTryErrorCaptured() {
	if !ss.initialized {
		return
	}
	ss.stackCtx.SetTryErrorCaptured()
}

// IsTryErrorCaptured returns true if the current try block has captured an error
func (ss *StackService) IsTryErrorCaptured() bool {
	if !ss.initialized {
		return false
	}
	return ss.stackCtx.IsTryErrorCaptured()
}

// Silent block support methods

// PushSilentBoundary pushes silent boundary markers for silent blocks
func (ss *StackService) PushSilentBoundary(silentID string) {
	if !ss.initialized {
		return
	}
	ss.stackCtx.PushSilentBoundary(silentID)
}

// PopSilentBoundary removes the most recent silent block context
func (ss *StackService) PopSilentBoundary() {
	if !ss.initialized {
		return
	}
	ss.stackCtx.PopSilentBoundary()
}

// IsInSilentBlock returns true if currently inside a silent block
func (ss *StackService) IsInSilentBlock() bool {
	if !ss.initialized {
		return false
	}
	return ss.stackCtx.IsInSilentBlock()
}

// GetCurrentSilentID returns the ID of the current silent block
func (ss *StackService) GetCurrentSilentID() string {
	if !ss.initialized {
		return ""
	}
	return ss.stackCtx.GetCurrentSilentID()
}

// GetCurrentSilentDepth returns the current silent block depth
func (ss *StackService) GetCurrentSilentDepth() int {
	if !ss.initialized {
		return 0
	}
	return ss.stackCtx.GetCurrentSilentDepth()
}
