package builtin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestTimerCommand_Name(t *testing.T) {
	cmd := &TimerCommand{}
	assert.Equal(t, "timer", cmd.Name())
}

func TestTimerCommand_ParseMode(t *testing.T) {
	cmd := &TimerCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestTimerCommand_Description(t *testing.T) {
	cmd := &TimerCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "countdown timer")
}

func TestTimerCommand_Usage(t *testing.T) {
	cmd := &TimerCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\timer")
}

func TestTimerCommand_HelpInfo(t *testing.T) {
	cmd := &TimerCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "timer", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)
	assert.NotEmpty(t, helpInfo.StoredVariables)

	// Check that examples contain valid timer commands
	foundExample := false
	for _, example := range helpInfo.Examples {
		if example.Command == "\\timer 5" {
			foundExample = true
			assert.Contains(t, example.Description, "5-second")
			break
		}
	}
	assert.True(t, foundExample, "Should have timer 5 example")
}

func TestTimerCommand_Execute_ValidInputs(t *testing.T) {
	cmd := &TimerCommand{}
	setupTimerTestRegistry(t)

	tests := []struct {
		name          string
		input         string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid integer",
			input:       "5",
			expectError: false,
		},
		{
			name:        "valid decimal",
			input:       "2.5",
			expectError: false,
		},
		{
			name:        "minimum valid value",
			input:       "0.1",
			expectError: false,
		},
		{
			name:        "maximum valid value",
			input:       "100",
			expectError: false,
		},
		{
			name:        "valid with extra whitespace",
			input:       "  10  ",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(map[string]string{}, tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTimerCommand_Execute_InvalidInputs(t *testing.T) {
	cmd := &TimerCommand{}
	setupTimerTestRegistry(t)

	tests := []struct {
		name          string
		input         string
		errorContains string
	}{
		{
			name:          "empty input",
			input:         "",
			errorContains: "duration is required",
		},
		{
			name:          "non-numeric input",
			input:         "abc",
			errorContains: "not a valid number",
		},
		{
			name:          "zero value",
			input:         "0",
			errorContains: "must be a positive number",
		},
		{
			name:          "negative value",
			input:         "-5",
			errorContains: "must be a positive number",
		},
		{
			name:          "too big value",
			input:         "101",
			errorContains: "too big",
		},
		{
			name:          "way too big value",
			input:         "1000",
			errorContains: "too big",
		},
		{
			name:          "special characters",
			input:         "5!",
			errorContains: "not a valid number",
		},
		{
			name:          "multiple numbers",
			input:         "5 10",
			errorContains: "not a valid number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(map[string]string{}, tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

func TestTimerCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &TimerCommand{}

	// Setup context but not the temporal display service
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	services.SetGlobalRegistry(services.NewRegistry())

	// Register only variable service, not temporal display service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		context.ResetGlobalContext()
	})

	err = cmd.Execute(map[string]string{}, "5")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "temporal display service not available")
}

func TestTimerCommand_Execute_Integration(t *testing.T) {
	cmd := &TimerCommand{}
	setupTimerTestRegistry(t)

	// Start a very short timer for integration testing
	err := cmd.Execute(map[string]string{}, "0.2")
	assert.NoError(t, err)

	// Verify the temporal service has an active timer
	serviceInterface, err := services.GetGlobalRegistry().GetService("temporal-display")
	require.NoError(t, err)

	temporalService, ok := serviceInterface.(*services.TemporalDisplayService)
	require.True(t, ok)

	// Give a moment for the timer to start
	time.Sleep(50 * time.Millisecond)

	// Check that there's at least one active display (we can't easily check the specific ID)
	// Since the timer ID is generated dynamically
	hasActiveTimer := false
	// We'll check by trying to create a timer with a known ID and see if the service is working
	testErr := temporalService.StartTimer("test-timer-check", 1*time.Second)
	if testErr == nil {
		hasActiveTimer = temporalService.IsActive("test-timer-check")
		_ = temporalService.Stop("test-timer-check")
	}
	assert.True(t, hasActiveTimer, "Temporal service should be functional")

	// Verify _output variable was set with initial start message
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Timer started")

	// Wait for the original timer to complete
	time.Sleep(300 * time.Millisecond)

	// After completion, _output should contain completion message
	output, err = variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Timer completed!")
}

func TestTimerCommand_Execute_BoundaryValues(t *testing.T) {
	cmd := &TimerCommand{}
	setupTimerTestRegistry(t)

	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "just above zero",
			input:       "0.001",
			expectError: false,
		},
		{
			name:        "exactly zero",
			input:       "0.0",
			expectError: true,
			errorMsg:    "must be a positive number",
		},
		{
			name:        "exactly 100",
			input:       "100.0",
			expectError: false,
		},
		{
			name:        "just above 100",
			input:       "100.1",
			expectError: true,
			errorMsg:    "too big",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(map[string]string{}, tt.input)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTimerCommand_Execute_VariableStorage(t *testing.T) {
	cmd := &TimerCommand{}
	setupTimerTestRegistry(t)

	err := cmd.Execute(map[string]string{}, "1.5")
	assert.NoError(t, err)

	// Check that _output variable was set with start message
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Timer started for 1.5 seconds")
}

func TestTimerCommand_WaitForTimerCompletion(t *testing.T) {
	cmd := &TimerCommand{}
	setupTimerTestRegistry(t)

	// Get temporal service
	serviceInterface, err := services.GetGlobalRegistry().GetService("temporal-display")
	require.NoError(t, err)

	temporalService, ok := serviceInterface.(*services.TemporalDisplayService)
	require.True(t, ok)

	// Start a very short timer manually to test completion
	timerID := "test-completion-timer"
	duration := 100 * time.Millisecond

	err = temporalService.StartTimer(timerID, duration)
	require.NoError(t, err)

	// Call waitForTimerCompletion in a goroutine
	done := make(chan bool)
	go func() {
		cmd.waitForTimerCompletion(temporalService, timerID, duration, 0.1)
		done <- true
	}()

	// Wait for completion
	select {
	case <-done:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timer completion took too long")
	}

	// Verify _output was updated with completion message
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Equal(t, "Timer completed!", output)
}

// setupTimerTestRegistry sets up a test environment with required services for timer command testing
func setupTimerTestRegistry(t *testing.T) {
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Create a new registry for testing
	services.SetGlobalRegistry(services.NewRegistry())

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewTemporalDisplayService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	t.Cleanup(func() {
		context.ResetGlobalContext()
	})
}

// Interface compliance check
var _ neurotypes.Command = (*TimerCommand)(nil)
