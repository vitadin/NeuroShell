package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

func TestErrorManagementService_Name(t *testing.T) {
	service := NewErrorManagementService()
	assert.Equal(t, "error_management", service.Name())
}

func TestErrorManagementService_NewErrorManagementService(t *testing.T) {
	service := NewErrorManagementService()

	assert.NotNil(t, service)
	assert.False(t, service.initialized)
	assert.Equal(t, "error_management", service.Name())
}

func TestErrorManagementService_Initialize(t *testing.T) {
	service := NewErrorManagementService()

	// Initially not initialized
	assert.False(t, service.initialized)

	// Initialize the service
	err := service.Initialize()

	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Should be idempotent - can initialize multiple times
	err = service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
}

func TestErrorManagementService_GetServiceInfo(t *testing.T) {
	service := NewErrorManagementService()

	// Test before initialization
	info := service.GetServiceInfo()
	expected := map[string]interface{}{
		"name":        "error_management",
		"initialized": false,
		"description": "Centralized error state management service",
	}
	assert.Equal(t, expected, info)

	// Test after initialization
	err := service.Initialize()
	require.NoError(t, err)

	info = service.GetServiceInfo()
	expected["initialized"] = true
	assert.Equal(t, expected, info)
}

func TestErrorManagementService_ResetErrorState_NotInitialized(t *testing.T) {
	service := NewErrorManagementService()

	err := service.ResetErrorState()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error service not initialized")
}

func TestErrorManagementService_SetErrorState_NotInitialized(t *testing.T) {
	service := NewErrorManagementService()

	err := service.SetErrorState("1", "test error")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error service not initialized")
}

func TestErrorManagementService_GetCurrentErrorState_NotInitialized(t *testing.T) {
	service := NewErrorManagementService()

	status, errorMsg, err := service.GetCurrentErrorState()

	assert.Error(t, err)
	assert.Empty(t, status)
	assert.Empty(t, errorMsg)
	assert.Contains(t, err.Error(), "error service not initialized")
}

func TestErrorManagementService_GetLastErrorState_NotInitialized(t *testing.T) {
	service := NewErrorManagementService()

	status, errorMsg, err := service.GetLastErrorState()

	assert.Error(t, err)
	assert.Empty(t, status)
	assert.Empty(t, errorMsg)
	assert.Contains(t, err.Error(), "error service not initialized")
}

func TestErrorManagementService_SetErrorStateFromCommandResult_NotInitialized(t *testing.T) {
	service := NewErrorManagementService()

	err := service.SetErrorStateFromCommandResult(errors.New("test error"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error service not initialized")
}

func TestErrorManagementService_IsErrorState_NotInitialized(t *testing.T) {
	service := NewErrorManagementService()

	isError, err := service.IsErrorState()

	assert.Error(t, err)
	assert.False(t, isError)
	assert.Contains(t, err.Error(), "error service not initialized")
}

// setupErrorManagementTestContext sets up a clean test context for error management testing
func setupErrorManagementTestContext(_ *testing.T) func() {
	// Create a new test context (this is a NeuroContext in test mode)
	testCtx := context.New()
	testCtx.SetTestMode(true)
	context.SetGlobalContext(testCtx)

	// Return cleanup function
	return func() {
		context.ResetGlobalContext()
	}
}

func TestErrorManagementService_ResetErrorState_Success(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	// Set some error state first
	err = service.SetErrorState("1", "initial error")
	require.NoError(t, err)

	// Reset error state
	err = service.ResetErrorState()
	assert.NoError(t, err)

	// Verify current state is reset
	status, errorMsg, err := service.GetCurrentErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "0", status)
	assert.Equal(t, "", errorMsg)
}

func TestErrorManagementService_SetErrorState_Success(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name     string
		status   string
		errorMsg string
	}{
		{
			name:     "success state",
			status:   "0",
			errorMsg: "",
		},
		{
			name:     "error state",
			status:   "1",
			errorMsg: "command failed",
		},
		{
			name:     "custom exit code",
			status:   "42",
			errorMsg: "custom error message",
		},
		{
			name:     "empty error message",
			status:   "1",
			errorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SetErrorState(tt.status, tt.errorMsg)
			assert.NoError(t, err)

			// Verify the state was set correctly
			status, errorMsg, err := service.GetCurrentErrorState()
			assert.NoError(t, err)
			assert.Equal(t, tt.status, status)
			assert.Equal(t, tt.errorMsg, errorMsg)
		})
	}
}

func TestErrorManagementService_GetCurrentErrorState_Success(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	// Initially should be empty/success state
	status, errorMsg, err := service.GetCurrentErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "0", status)
	assert.Equal(t, "", errorMsg)

	// Set an error state and verify
	err = service.SetErrorState("1", "test error")
	require.NoError(t, err)

	status, errorMsg, err = service.GetCurrentErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "1", status)
	assert.Equal(t, "test error", errorMsg)
}

func TestErrorManagementService_GetLastErrorState_Success(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	// Initially last state should be the default initial values
	status, errorMsg, err := service.GetLastErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "0", status)
	assert.Equal(t, "", errorMsg)

	// Set error state, then reset to move current to last
	err = service.SetErrorState("42", "previous error")
	require.NoError(t, err)

	err = service.ResetErrorState()
	require.NoError(t, err)

	// Now last state should contain the previous error
	status, errorMsg, err = service.GetLastErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "42", status)
	assert.Equal(t, "previous error", errorMsg)
}

func TestErrorManagementService_SetErrorStateFromCommandResult_Success(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name           string
		commandErr     error
		expectedStatus string
		expectedError  string
	}{
		{
			name:           "nil error (success)",
			commandErr:     nil,
			expectedStatus: "0",
			expectedError:  "",
		},
		{
			name:           "command error",
			commandErr:     errors.New("command failed"),
			expectedStatus: "1",
			expectedError:  "command failed",
		},
		{
			name:           "empty error message",
			commandErr:     errors.New(""),
			expectedStatus: "1",
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SetErrorStateFromCommandResult(tt.commandErr)
			assert.NoError(t, err)

			// Verify the state was set correctly
			status, errorMsg, err := service.GetCurrentErrorState()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedError, errorMsg)
		})
	}
}

func TestErrorManagementService_IsErrorState_Success(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name        string
		status      string
		errorMsg    string
		expectError bool
	}{
		{
			name:        "success state (0)",
			status:      "0",
			errorMsg:    "",
			expectError: false,
		},
		{
			name:        "error state (1)",
			status:      "1",
			errorMsg:    "error",
			expectError: true,
		},
		{
			name:        "custom error code",
			status:      "42",
			errorMsg:    "custom error",
			expectError: true,
		},
		{
			name:        "empty string status",
			status:      "",
			errorMsg:    "",
			expectError: true, // Empty string is not "0", so it's an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the error state
			err := service.SetErrorState(tt.status, tt.errorMsg)
			require.NoError(t, err)

			// Check if it's an error state
			isError, err := service.IsErrorState()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectError, isError)
		})
	}
}

func TestErrorManagementService_ErrorStateTransitions(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test the complete lifecycle of error state transitions

	// 1. Initial state should be success
	status, errorMsg, err := service.GetCurrentErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "0", status)
	assert.Equal(t, "", errorMsg)

	lastStatus, lastErrorMsg, err := service.GetLastErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "0", lastStatus)
	assert.Equal(t, "", lastErrorMsg)

	// 2. Set first error
	err = service.SetErrorState("1", "first error")
	require.NoError(t, err)

	status, errorMsg, err = service.GetCurrentErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "1", status)
	assert.Equal(t, "first error", errorMsg)

	// 3. Reset (moves current to last)
	err = service.ResetErrorState()
	require.NoError(t, err)

	// Current should be reset
	status, errorMsg, err = service.GetCurrentErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "0", status)
	assert.Equal(t, "", errorMsg)

	// Last should contain previous error
	lastStatus, lastErrorMsg, err = service.GetLastErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "1", lastStatus)
	assert.Equal(t, "first error", lastErrorMsg)

	// 4. Set second error
	err = service.SetErrorState("2", "second error")
	require.NoError(t, err)

	// 5. Reset again
	err = service.ResetErrorState()
	require.NoError(t, err)

	// Last should now contain the second error
	lastStatus, lastErrorMsg, err = service.GetLastErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "2", lastStatus)
	assert.Equal(t, "second error", lastErrorMsg)
}

func TestErrorManagementService_ConcurrentAccess(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test concurrent access to the service
	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(_ int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// Perform various operations
				_ = service.SetErrorState("1", "concurrent error")
				_, _, _ = service.GetCurrentErrorState()
				_, _, _ = service.GetLastErrorState()
				_ = service.ResetErrorState()
				_, _ = service.IsErrorState()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Service should still be functional after concurrent access
	err = service.SetErrorState("final", "final test")
	assert.NoError(t, err)

	status, errorMsg, err := service.GetCurrentErrorState()
	assert.NoError(t, err)
	assert.Equal(t, "final", status)
	assert.Equal(t, "final test", errorMsg)
}

func TestErrorManagementService_EdgeCases(t *testing.T) {
	cleanup := setupErrorManagementTestContext(t)
	defer cleanup()

	service := NewErrorManagementService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name     string
		status   string
		errorMsg string
	}{
		{
			name:   "very long error message",
			status: "1",
			errorMsg: "This is a very long error message that might test the limits of the error handling system. " +
				"It contains multiple sentences and should be handled gracefully by the error management service. " +
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		},
		{
			name:     "unicode characters",
			status:   "1",
			errorMsg: "Error with unicode: 🚨 错误 エラー",
		},
		{
			name:     "newlines in error message",
			status:   "1",
			errorMsg: "Multi-line\nerror\nmessage",
		},
		{
			name:     "special characters",
			status:   "1",
			errorMsg: "Error: <>&\"'`${}[]()\\|*?",
		},
		{
			name:     "empty status with error message",
			status:   "",
			errorMsg: "error with empty status",
		},
		{
			name:     "numeric string status",
			status:   "127",
			errorMsg: "command not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SetErrorState(tt.status, tt.errorMsg)
			assert.NoError(t, err)

			// Verify the state was set correctly
			status, errorMsg, err := service.GetCurrentErrorState()
			assert.NoError(t, err)
			assert.Equal(t, tt.status, status)
			assert.Equal(t, tt.errorMsg, errorMsg)
		})
	}
}

// Interface compliance test
func TestErrorManagementService_InterfaceCompliance(_ *testing.T) {
	var _ neurotypes.Service = (*ErrorManagementService)(nil)
}
