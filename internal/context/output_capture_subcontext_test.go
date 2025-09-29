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
	assert.Equal(t, "", ctx.GetLastOutput())
}

func TestOutputCaptureSubcontext_CaptureOutput(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()

	// Capture some output
	output := "Hello, World!"
	ctx.CaptureOutput(output)

	// Check last output
	assert.Equal(t, output, ctx.GetLastOutput())
}

func TestOutputCaptureSubcontext_MultipleCaptures(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()

	// First capture
	first := "First output"
	ctx.CaptureOutput(first)
	assert.Equal(t, first, ctx.GetLastOutput())

	// Second capture (overwrites the first)
	second := "Second output"
	ctx.CaptureOutput(second)
	assert.Equal(t, second, ctx.GetLastOutput())

	// Third capture (overwrites the second)
	third := "Third output"
	ctx.CaptureOutput(third)
	assert.Equal(t, third, ctx.GetLastOutput())
}

func TestOutputCaptureSubcontext_EmptyOutput(t *testing.T) {
	ctx := NewOutputCaptureSubcontext()

	// Capture empty output
	ctx.CaptureOutput("")
	assert.Equal(t, "", ctx.GetLastOutput())

	// Capture non-empty, then empty again
	ctx.CaptureOutput("some output")
	assert.Equal(t, "some output", ctx.GetLastOutput())

	ctx.CaptureOutput("")
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
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				// Mix of operations
				if j%2 == 0 {
					ctx.CaptureOutput("output from goroutine " + string(rune(id+'0')))
				} else {
					_ = ctx.GetLastOutput()
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check that the context is still in a valid state
	// We can't predict exact values due to race conditions, but operations should not panic
	last := ctx.GetLastOutput()

	// Should be able to access without panic (string values are always valid)
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
	assert.Equal(t, "test output", outputCtx.GetLastOutput())
	assert.Equal(t, "test output", neuroCtx.GetLastOutput())
}

func TestOutputCaptureSubcontext_IntegrationWithNeuroContext(t *testing.T) {
	ctx := New()

	// Test that NeuroContext methods work correctly
	assert.Equal(t, "", ctx.GetLastOutput())

	// Capture output through NeuroContext
	ctx.CaptureOutput("Hello from NeuroContext")
	assert.Equal(t, "Hello from NeuroContext", ctx.GetLastOutput())

	// Capture new output
	ctx.CaptureOutput("Second output from NeuroContext")
	assert.Equal(t, "Second output from NeuroContext", ctx.GetLastOutput())
}
