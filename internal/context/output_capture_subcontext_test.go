package context

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOutputCaptureSubcontext(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()
	assert.NotNil(t, ctx)

	// Initial state should be empty
	assert.Equal(t, "", ctx.GetCurrentOutput())
	assert.Equal(t, "", ctx.GetLastOutput())
}

func TestOutputCaptureSubcontext_CaptureOutput(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()

	// Capture some output
	output := "Hello, World!"
	ctx.CaptureOutput(output)

	// Check current output
	assert.Equal(t, output, ctx.GetCurrentOutput())
	// Last output should still be empty
	assert.Equal(t, "", ctx.GetLastOutput())
}

func TestOutputCaptureSubcontext_ResetOutput(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()

	// Set initial output
	firstOutput := "First command output"
	ctx.CaptureOutput(firstOutput)
	assert.Equal(t, firstOutput, ctx.GetCurrentOutput())

	// Reset output (moves current to last)
	ctx.ResetOutput()

	// Current should be empty, last should have the previous output
	assert.Equal(t, "", ctx.GetCurrentOutput())
	assert.Equal(t, firstOutput, ctx.GetLastOutput())
}

func TestOutputCaptureSubcontext_MultipleCommands(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()

	// First command
	first := "First output"
	ctx.CaptureOutput(first)
	assert.Equal(t, first, ctx.GetCurrentOutput())
	assert.Equal(t, "", ctx.GetLastOutput())

	// Reset for second command
	ctx.ResetOutput()
	assert.Equal(t, "", ctx.GetCurrentOutput())
	assert.Equal(t, first, ctx.GetLastOutput())

	// Second command
	second := "Second output"
	ctx.CaptureOutput(second)
	assert.Equal(t, second, ctx.GetCurrentOutput())
	assert.Equal(t, first, ctx.GetLastOutput())

	// Reset for third command
	ctx.ResetOutput()
	assert.Equal(t, "", ctx.GetCurrentOutput())
	assert.Equal(t, second, ctx.GetLastOutput())

	// Third command
	third := "Third output"
	ctx.CaptureOutput(third)
	assert.Equal(t, third, ctx.GetCurrentOutput())
	assert.Equal(t, second, ctx.GetLastOutput())
}

func TestOutputCaptureSubcontext_EmptyOutput(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()

	// Capture empty output
	ctx.CaptureOutput("")
	assert.Equal(t, "", ctx.GetCurrentOutput())

	// Reset and check
	ctx.ResetOutput()
	assert.Equal(t, "", ctx.GetCurrentOutput())
	assert.Equal(t, "", ctx.GetLastOutput())
}

func TestOutputCaptureSubcontext_ThreadSafety(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()
	var wg sync.WaitGroup

	// Number of goroutines to test with
	numGoroutines := 10
	numOperationsPerGoroutine := 100

	// Test concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(_ int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				// Mix of operations
				switch j % 3 {
				case 0:
					ctx.ResetOutput()
				case 1:
					ctx.CaptureOutput("output")
				default:
					_ = ctx.GetCurrentOutput()
					_ = ctx.GetLastOutput()
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check that the context is still in a valid state
	// We can't predict exact values due to race conditions, but operations should not panic
	current := ctx.GetCurrentOutput()
	last := ctx.GetLastOutput()

	// Should be able to access without panic (string values are always valid)
	assert.NotNil(t, current)
	assert.NotNil(t, last)
}

func TestOutputCaptureSubcontext_LargeOutput(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()

	// Test with large output
	largeOutput := make([]byte, 1024*1024) // 1MB
	for i := range largeOutput {
		largeOutput[i] = byte('A' + (i % 26))
	}
	largeOutputStr := string(largeOutput)

	ctx.CaptureOutput(largeOutputStr)
	assert.Equal(t, largeOutputStr, ctx.GetCurrentOutput())

	ctx.ResetOutput()
	assert.Equal(t, "", ctx.GetCurrentOutput())
	assert.Equal(t, largeOutputStr, ctx.GetLastOutput())
}

func TestNewOutputCaptureSubcontextFromContext(t *testing.T) {
	// Create a NeuroContext first
	neuroCtx := New()
	assert.NotNil(t, neuroCtx)

	// Get OutputCaptureSubcontext from NeuroContext
	outputCtx := NewOutputCaptureSubcontextFromContext(neuroCtx)
	assert.NotNil(t, outputCtx)

	// Should be the same instance as the one in NeuroContext
	assert.Equal(t, neuroCtx.outputCaptureCtx, outputCtx)

	// Test basic functionality
	outputCtx.CaptureOutput("test output")
	assert.Equal(t, "test output", outputCtx.GetCurrentOutput())
	assert.Equal(t, "test output", neuroCtx.GetCurrentOutput())
}

func TestOutputCaptureSubcontext_IntegrationWithNeuroContext(t *testing.T) {
	ctx := New()

	// Test that NeuroContext methods work correctly
	assert.Equal(t, "", ctx.GetCurrentOutput())
	assert.Equal(t, "", ctx.GetLastOutput())

	// Capture output through NeuroContext
	ctx.CaptureOutput("Hello from NeuroContext")
	assert.Equal(t, "Hello from NeuroContext", ctx.GetCurrentOutput())
	assert.Equal(t, "", ctx.GetLastOutput())

	// Reset through NeuroContext
	ctx.ResetOutput()
	assert.Equal(t, "", ctx.GetCurrentOutput())
	assert.Equal(t, "Hello from NeuroContext", ctx.GetLastOutput())
}
