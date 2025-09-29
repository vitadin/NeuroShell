package context

import (
	"sync"
)

// OutputCaptureSubcontext defines the interface for automatic command output capture functionality.
// This manages the current and last command outputs for automatic capture of any command's console output.
type OutputCaptureSubcontext interface {
	// Output capture operations
	CaptureOutput(output string)
	GetCurrentOutput() string
	GetLastOutput() string
	ResetOutput() // Called before command execution to move current to last
}

// outputCaptureSubcontext implements the OutputCaptureSubcontext interface.
type outputCaptureSubcontext struct {
	// Output capture management
	lastOutput    string       // Last command's captured output
	currentOutput string       // Current command's captured output
	outputMutex   sync.RWMutex // Protects output fields
}

// NewOutputCaptureSubcontext creates a new OutputCaptureSubcontext instance.
func NewOutputCaptureSubcontext() OutputCaptureSubcontext {
	return &outputCaptureSubcontext{
		lastOutput:    "",
		currentOutput: "",
	}
}

// NewOutputCaptureSubcontextFromContext creates an OutputCaptureSubcontext from an existing NeuroContext.
// This is used by services to get a reference to the context's output capture subcontext.
func NewOutputCaptureSubcontextFromContext(ctx *NeuroContext) OutputCaptureSubcontext {
	return ctx.outputCaptureCtx
}

// ResetOutput resets the current output to empty and moves current to last.
// This should be called before executing a new command.
func (o *outputCaptureSubcontext) ResetOutput() {
	o.outputMutex.Lock()
	defer o.outputMutex.Unlock()

	// Move current output to last
	o.lastOutput = o.currentOutput

	// Reset current output to empty
	o.currentOutput = ""
}

// CaptureOutput captures the output from command execution.
// This should be called during or after command execution with the captured output.
func (o *outputCaptureSubcontext) CaptureOutput(output string) {
	o.outputMutex.Lock()
	defer o.outputMutex.Unlock()

	o.currentOutput = output
}

// GetCurrentOutput returns the current command's captured output (thread-safe read).
func (o *outputCaptureSubcontext) GetCurrentOutput() string {
	o.outputMutex.RLock()
	defer o.outputMutex.RUnlock()

	return o.currentOutput
}

// GetLastOutput returns the last command's captured output (thread-safe read).
func (o *outputCaptureSubcontext) GetLastOutput() string {
	o.outputMutex.RLock()
	defer o.outputMutex.RUnlock()

	return o.lastOutput
}
