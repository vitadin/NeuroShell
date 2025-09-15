package context

import (
	"sync"
)

// TryBlockContext represents the context for a try block with error boundaries
type TryBlockContext struct {
	ID            string // Unique identifier for this try block
	StartDepth    int    // Stack depth when try block started
	ErrorCaptured bool   // Whether an error has been captured
}

// SilentBlockContext represents the context for a silent block with output suppression
type SilentBlockContext struct {
	ID         string // Unique identifier for this silent block
	StartDepth int    // Stack depth when silent block started
}

// StackSubcontext defines the interface for execution stack management functionality.
// This includes command stacking, try blocks, and silent blocks for control flow.
type StackSubcontext interface {
	// Basic stack operations
	PushCommand(command string)
	PushCommands(commands []string)
	PopCommand() (string, bool)
	PeekCommand() (string, bool)
	ClearStack()
	GetStackSize() int
	IsStackEmpty() bool
	PeekStack() []string

	// Try block support methods
	PushErrorBoundary(tryID string)
	PopErrorBoundary()
	IsInTryBlock() bool
	GetCurrentTryID() string
	GetCurrentTryDepth() int
	SetTryErrorCaptured()
	IsTryErrorCaptured() bool

	// Silent block support methods
	PushSilentBoundary(silentID string)
	PopSilentBoundary()
	IsInSilentBlock() bool
	GetCurrentSilentID() string
	GetCurrentSilentDepth() int
}

// stackSubcontext implements the StackSubcontext interface.
type stackSubcontext struct {
	// Stack-based execution support
	executionStack     []string             // Execution stack (LIFO order)
	tryBlocks          []TryBlockContext    // Try block management
	currentTryDepth    int                  // Current try block depth
	silentBlocks       []SilentBlockContext // Silent block management
	currentSilentDepth int                  // Current silent block depth
	stackMutex         sync.RWMutex         // Protects executionStack, tryBlocks, and silentBlocks
}

// NewStackSubcontext creates a new StackSubcontext instance.
func NewStackSubcontext() StackSubcontext {
	return &stackSubcontext{
		executionStack:     make([]string, 0),
		tryBlocks:          make([]TryBlockContext, 0),
		currentTryDepth:    0,
		silentBlocks:       make([]SilentBlockContext, 0),
		currentSilentDepth: 0,
	}
}

// NewStackSubcontextFromContext creates a StackSubcontext from an existing NeuroContext.
// This is used by services to get a reference to the context's stack subcontext.
func NewStackSubcontextFromContext(ctx *NeuroContext) StackSubcontext {
	return ctx.stackCtx
}

// Basic stack operations

// PushCommand adds a single command to the execution stack
func (s *stackSubcontext) PushCommand(command string) {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()
	s.executionStack = append(s.executionStack, command)
}

// PushCommands adds multiple commands to the execution stack
func (s *stackSubcontext) PushCommands(commands []string) {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()
	s.executionStack = append(s.executionStack, commands...)
}

// PopCommand removes and returns the last command from the stack (LIFO)
func (s *stackSubcontext) PopCommand() (string, bool) {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()

	if len(s.executionStack) == 0 {
		return "", false
	}

	lastIndex := len(s.executionStack) - 1
	command := s.executionStack[lastIndex]
	s.executionStack = s.executionStack[:lastIndex]
	return command, true
}

// PeekCommand returns the next command without removing it from the stack
func (s *stackSubcontext) PeekCommand() (string, bool) {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()

	if len(s.executionStack) == 0 {
		return "", false
	}

	return s.executionStack[len(s.executionStack)-1], true
}

// ClearStack removes all commands from the execution stack
func (s *stackSubcontext) ClearStack() {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()
	s.executionStack = make([]string, 0)
}

// GetStackSize returns the number of commands in the execution stack
func (s *stackSubcontext) GetStackSize() int {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()
	return len(s.executionStack)
}

// IsStackEmpty returns true if the stack is empty
func (s *stackSubcontext) IsStackEmpty() bool {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()
	return len(s.executionStack) == 0
}

// PeekStack returns a copy of the execution stack without modifying it
// Returns the stack in reverse order (top to bottom, LIFO order)
func (s *stackSubcontext) PeekStack() []string {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()

	result := make([]string, len(s.executionStack))
	// Copy in reverse order to show stack from top to bottom
	for i, cmd := range s.executionStack {
		result[len(s.executionStack)-1-i] = cmd
	}
	return result
}

// Try block support methods

// PushErrorBoundary pushes error boundary markers for try blocks
func (s *stackSubcontext) PushErrorBoundary(tryID string) {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()

	// Create try block context
	tryBlock := TryBlockContext{
		ID:            tryID,
		StartDepth:    len(s.executionStack),
		ErrorCaptured: false,
	}

	s.tryBlocks = append(s.tryBlocks, tryBlock)
	s.currentTryDepth++
}

// PopErrorBoundary removes the most recent try block context
func (s *stackSubcontext) PopErrorBoundary() {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()

	if len(s.tryBlocks) > 0 {
		s.tryBlocks = s.tryBlocks[:len(s.tryBlocks)-1]
		s.currentTryDepth--
	}
}

// IsInTryBlock returns true if currently inside a try block
func (s *stackSubcontext) IsInTryBlock() bool {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()
	return len(s.tryBlocks) > 0
}

// GetCurrentTryID returns the ID of the current try block
func (s *stackSubcontext) GetCurrentTryID() string {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()

	if len(s.tryBlocks) == 0 {
		return ""
	}

	return s.tryBlocks[len(s.tryBlocks)-1].ID
}

// GetCurrentTryDepth returns the current try block depth
func (s *stackSubcontext) GetCurrentTryDepth() int {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()
	return s.currentTryDepth
}

// SetTryErrorCaptured marks the current try block as having captured an error
func (s *stackSubcontext) SetTryErrorCaptured() {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()

	if len(s.tryBlocks) > 0 {
		s.tryBlocks[len(s.tryBlocks)-1].ErrorCaptured = true
	}
}

// IsTryErrorCaptured returns true if the current try block has captured an error
func (s *stackSubcontext) IsTryErrorCaptured() bool {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()

	if len(s.tryBlocks) == 0 {
		return false
	}

	return s.tryBlocks[len(s.tryBlocks)-1].ErrorCaptured
}

// Silent block support methods

// PushSilentBoundary pushes silent boundary markers for silent blocks
func (s *stackSubcontext) PushSilentBoundary(silentID string) {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()

	// Create silent block context
	silentBlock := SilentBlockContext{
		ID:         silentID,
		StartDepth: len(s.executionStack),
	}

	s.silentBlocks = append(s.silentBlocks, silentBlock)
	s.currentSilentDepth++
}

// PopSilentBoundary removes the most recent silent block context
func (s *stackSubcontext) PopSilentBoundary() {
	s.stackMutex.Lock()
	defer s.stackMutex.Unlock()

	if len(s.silentBlocks) > 0 {
		s.silentBlocks = s.silentBlocks[:len(s.silentBlocks)-1]
		s.currentSilentDepth--
	}
}

// IsInSilentBlock returns true if currently inside a silent block
func (s *stackSubcontext) IsInSilentBlock() bool {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()
	return len(s.silentBlocks) > 0
}

// GetCurrentSilentID returns the ID of the current silent block
func (s *stackSubcontext) GetCurrentSilentID() string {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()

	if len(s.silentBlocks) == 0 {
		return ""
	}

	return s.silentBlocks[len(s.silentBlocks)-1].ID
}

// GetCurrentSilentDepth returns the current silent block depth
func (s *stackSubcontext) GetCurrentSilentDepth() int {
	s.stackMutex.RLock()
	defer s.stackMutex.RUnlock()
	return s.currentSilentDepth
}
