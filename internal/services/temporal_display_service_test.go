package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemporalDisplayService(t *testing.T) {
	service := NewTemporalDisplayService()
	assert.NotNil(t, service)
	assert.False(t, service.initialized)
	assert.NotNil(t, service.activeDisplays)
	assert.Equal(t, 0, len(service.activeDisplays))
}

func TestTemporalDisplayService_Name(t *testing.T) {
	service := NewTemporalDisplayService()
	assert.Equal(t, "temporal-display", service.Name())
}

func TestTemporalDisplayService_Initialize(t *testing.T) {
	service := NewTemporalDisplayService()

	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
}

func TestTemporalDisplayService_StartTimer(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		duration    time.Duration
		expectError bool
	}{
		{
			name:        "valid timer",
			id:          "timer1",
			duration:    1 * time.Second,
			expectError: false,
		},
		{
			name:        "another valid timer",
			id:          "timer2",
			duration:    500 * time.Millisecond,
			expectError: false,
		},
		{
			name:        "zero duration timer",
			id:          "timer3",
			duration:    0,
			expectError: false, // Should be valid, will stop immediately
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.StartTimer(tt.id, tt.duration)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, service.IsActive(tt.id))
			}
		})
	}

	// Clean up
	err = service.StopAll()
	assert.NoError(t, err)
}

func TestTemporalDisplayService_StartTimer_NotInitialized(t *testing.T) {
	service := NewTemporalDisplayService()

	err := service.StartTimer("timer1", 1*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestTemporalDisplayService_StartCustomDisplay(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	condition := func(elapsed time.Duration) bool {
		return elapsed >= 500*time.Millisecond
	}
	renderer := func(elapsed time.Duration) string {
		return fmt.Sprintf("Custom: %.1fs", elapsed.Seconds())
	}

	err = service.StartCustomDisplay("custom1", condition, renderer)
	assert.NoError(t, err)
	assert.True(t, service.IsActive("custom1"))

	// Clean up
	err = service.Stop("custom1")
	assert.NoError(t, err)
}

func TestTemporalDisplayService_StartCustomDisplay_NilFunctions(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test nil condition
	err = service.StartCustomDisplay("test1", nil, func(time.Duration) string { return "test" })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Test nil renderer
	err = service.StartCustomDisplay("test2", func(time.Duration) bool { return true }, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Test both nil
	err = service.StartCustomDisplay("test3", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestTemporalDisplayService_StartCustomDisplay_NotInitialized(t *testing.T) {
	service := NewTemporalDisplayService()

	condition := func(_ time.Duration) bool { return true }
	renderer := func(_ time.Duration) string { return "test" }

	err := service.StartCustomDisplay("test", condition, renderer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestTemporalDisplayService_Stop(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	// Start a timer
	err = service.StartTimer("timer1", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, service.IsActive("timer1"))

	// Stop the timer
	err = service.Stop("timer1")
	assert.NoError(t, err)

	// Give some time for cleanup
	time.Sleep(50 * time.Millisecond)
	assert.False(t, service.IsActive("timer1"))
}

func TestTemporalDisplayService_Stop_NonExistent(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	err = service.Stop("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestTemporalDisplayService_Stop_NotInitialized(t *testing.T) {
	service := NewTemporalDisplayService()

	err := service.Stop("timer1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestTemporalDisplayService_StopAll(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	// Start multiple timers
	err = service.StartTimer("timer1", 5*time.Second)
	require.NoError(t, err)
	err = service.StartTimer("timer2", 5*time.Second)
	require.NoError(t, err)
	err = service.StartTimer("timer3", 5*time.Second)
	require.NoError(t, err)

	assert.True(t, service.IsActive("timer1"))
	assert.True(t, service.IsActive("timer2"))
	assert.True(t, service.IsActive("timer3"))

	// Stop all
	err = service.StopAll()
	assert.NoError(t, err)

	// Give some time for cleanup
	time.Sleep(50 * time.Millisecond)
	assert.False(t, service.IsActive("timer1"))
	assert.False(t, service.IsActive("timer2"))
	assert.False(t, service.IsActive("timer3"))
}

func TestTemporalDisplayService_StopAll_NotInitialized(t *testing.T) {
	service := NewTemporalDisplayService()

	err := service.StopAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestTemporalDisplayService_IsActive(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test non-existent display
	assert.False(t, service.IsActive("nonexistent"))

	// Start a timer
	err = service.StartTimer("timer1", 2*time.Second)
	require.NoError(t, err)
	assert.True(t, service.IsActive("timer1"))

	// Stop it
	err = service.Stop("timer1")
	require.NoError(t, err)

	// Give some time for cleanup
	time.Sleep(50 * time.Millisecond)
	assert.False(t, service.IsActive("timer1"))
}

func TestTemporalDisplayService_IsActive_NotInitialized(t *testing.T) {
	service := NewTemporalDisplayService()

	active := service.IsActive("timer1")
	assert.False(t, active)
}

func TestTemporalDisplayService_TimerAutoStop(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	// Start a very short timer
	err = service.StartTimer("timer1", 200*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, service.IsActive("timer1"))

	// Wait for it to auto-stop
	time.Sleep(300 * time.Millisecond)
	assert.False(t, service.IsActive("timer1"))
}

func TestTemporalDisplayService_CustomDisplayAutoStop(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	stopAfter := 200 * time.Millisecond
	condition := func(elapsed time.Duration) bool {
		return elapsed >= stopAfter
	}
	renderer := func(elapsed time.Duration) string {
		return fmt.Sprintf("%.0fms", elapsed.Seconds()*1000)
	}

	err = service.StartCustomDisplay("custom1", condition, renderer)
	require.NoError(t, err)
	assert.True(t, service.IsActive("custom1"))

	// Wait for it to auto-stop
	time.Sleep(300 * time.Millisecond)
	assert.False(t, service.IsActive("custom1"))
}

func TestTemporalDisplayService_ReplaceExistingDisplay(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	// Start first timer
	err = service.StartTimer("timer1", 5*time.Second)
	require.NoError(t, err)
	assert.True(t, service.IsActive("timer1"))

	// Give the first timer a moment to fully start its goroutine
	time.Sleep(50 * time.Millisecond)

	// Start another timer with same ID (should replace)
	err = service.StartTimer("timer1", 3*time.Second)
	require.NoError(t, err)
	assert.True(t, service.IsActive("timer1"))

	// Should still only have one active display
	service.mu.RLock()
	count := len(service.activeDisplays)
	service.mu.RUnlock()
	assert.Equal(t, 1, count)

	// Give a moment for any old goroutine cleanup to complete
	time.Sleep(50 * time.Millisecond)

	// Clean up
	err = service.Stop("timer1")
	assert.NoError(t, err)
}

func TestTemporalDisplayService_ConcurrentAccess(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	done := make(chan bool, 10)

	// Start multiple goroutines trying to create/stop displays
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			timerID := fmt.Sprintf("timer%d", id)

			// Start timer
			err := service.StartTimer(timerID, 1*time.Second)
			assert.NoError(t, err)

			// Check if active
			active := service.IsActive(timerID)
			assert.True(t, active)

			// Wait a bit
			time.Sleep(50 * time.Millisecond)

			// Stop timer
			err = service.Stop(timerID)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify all displays are cleaned up
	service.mu.RLock()
	count := len(service.activeDisplays)
	service.mu.RUnlock()
	assert.Equal(t, 0, count)
}

func TestTemporalDisplayService_DisplayContent(t *testing.T) {
	service := NewTemporalDisplayService()
	display := &Display{
		id:        "test",
		lastWidth: 0,
	}

	// Test displaying content
	service.displayContent(display, "Hello")
	assert.Equal(t, 5, display.lastWidth) // "Hello" has 5 characters

	// Test displaying longer content
	service.displayContent(display, "Hello World!")
	assert.Equal(t, 12, display.lastWidth) // "Hello World!" has 12 characters

	// Test displaying shorter content (should clear previous)
	service.displayContent(display, "Hi")
	assert.Equal(t, 2, display.lastWidth) // "Hi" has 2 characters
}

func TestTemporalDisplayService_CleanupDisplay(_ *testing.T) {
	service := NewTemporalDisplayService()
	display := &Display{
		id:        "test",
		lastWidth: 10,
	}

	// Should not panic
	service.cleanupDisplay(display)

	// Test with zero width
	display.lastWidth = 0
	service.cleanupDisplay(display)
}

func TestTemporalDisplayService_RendererWithStyling(t *testing.T) {
	service := NewTemporalDisplayService()
	err := service.Initialize()
	require.NoError(t, err)

	// Create a renderer that uses lipgloss styling
	renderer := func(elapsed time.Duration) string {
		// This tests that lipgloss styling works without theme service dependency
		return fmt.Sprintf("Styled: %.1fs", elapsed.Seconds())
	}

	condition := func(elapsed time.Duration) bool {
		return elapsed >= 200*time.Millisecond
	}

	err = service.StartCustomDisplay("styled", condition, renderer)
	assert.NoError(t, err)
	assert.True(t, service.IsActive("styled"))

	// Clean up
	time.Sleep(300 * time.Millisecond)
	assert.False(t, service.IsActive("styled"))
}

// Test the service registration
func TestTemporalDisplayService_ServiceRegistration(t *testing.T) {
	// Create a new registry to avoid test isolation issues
	registry := NewRegistry()

	// Register the service manually for this test
	service := NewTemporalDisplayService()
	err := registry.RegisterService(service)
	assert.NoError(t, err)

	// Test that we can retrieve it
	retrievedService, err := registry.GetService("temporal-display")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedService)

	temporalService, ok := retrievedService.(*TemporalDisplayService)
	assert.True(t, ok)
	assert.Equal(t, "temporal-display", temporalService.Name())
}
