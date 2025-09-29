package context

import (
	"sync"
)

// OutputCaptureSubcontext defines the interface for automatic command output capture functionality.
// This manages the last command output for automatic capture of any command's console output.
type OutputCaptureSubcontext interface {
	// Output capture operations
	CaptureOutput(output string)
	GetLastOutput() string
}

// outputCaptureSubcontext implements the OutputCaptureSubcontext interface.
type outputCaptureSubcontext struct {
	// Output capture management
	lastOutput  string       // Last command's captured output
	outputMutex sync.RWMutex // Protects output fields
}

// NewOutputCaptureSubcontext creates a new OutputCaptureSubcontext instance.
func NewOutputCaptureSubcontext() OutputCaptureSubcontext {
	return &outputCaptureSubcontext{
		lastOutput: "",
	}
}

// NewOutputCaptureSubcontextFromContext creates an OutputCaptureSubcontext from an existing NeuroContext.
// This is used by services to get a reference to the context's output capture subcontext.
func NewOutputCaptureSubcontextFromContext(ctx *NeuroContext) OutputCaptureSubcontext {
	return ctx.outputCaptureCtx
}

// CaptureOutput captures the output from command execution.
// This should be called during or after command execution with the captured output.
func (o *outputCaptureSubcontext) CaptureOutput(output string) {
	o.outputMutex.Lock()
	defer o.outputMutex.Unlock()

	o.lastOutput = output
}

// GetLastOutput returns the last command's captured output (thread-safe read).
func (o *outputCaptureSubcontext) GetLastOutput() string {
	o.outputMutex.RLock()
	defer o.outputMutex.RUnlock()

	return o.lastOutput
}
