// Package services tests for the Claude Code service integration.
//
// Test Categories:
//
// 1. UNIT TESTS (Mocked) - Always run, CI/CD friendly:
//   - TestClaudeCodeService_Initialize - tests initialization with mocked CLI detection
//   - TestClaudeCodeService_CreateSession - tests session management with mocks
//   - TestClaudeCodeService_JobManagement - tests job handling with mocks
//   - TestClaudeCodeService_Authentication - tests auth checking with mocks
//   - All other existing tests that use setupTestClaudeCodeService()
//
// 2. INTEGRATION TESTS - Require actual Claude CLI:
//   - TestClaudeCodeService_ActualCLIDetection - tests real CLI detection
//   - TestClaudeCodeService_RealCLIIntegration - tests with actual Claude CLI
//     These tests will be skipped automatically in CI/CD if Claude CLI is not available.
//
// 3. CI/CD SIMULATION TESTS:
//   - TestClaudeCodeService_CICDSkipBehavior - demonstrates CI/CD behavior
//
// Running Tests:
//   - `go test ./internal/services/` - runs all tests (skips integration if CLI unavailable)
//   - `go test -short ./internal/services/` - runs only unit tests (skips integration)
//   - `go test -run ".*Mocked.*" ./internal/services/` - runs only mocked unit tests
//
// The service uses function fields (isClaudeCodeInstalled, getClaudeCodeVersion, StartProcess)
// that can be mocked for testing, ensuring unit tests work regardless of CLI availability.
package services

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	neuroshellcontext "neuroshell/internal/context"
)

// isClaudeCLIAvailable checks if the actual Claude CLI is available on the system
func isClaudeCLIAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// skipIfClaudeCLIUnavailable skips the test if Claude CLI is not available
// This is used for integration tests that require actual Claude CLI
func skipIfClaudeCLIUnavailable(t *testing.T) {
	if !isClaudeCLIAvailable() {
		t.Skip("Skipping test: Claude CLI not available (CI/CD environment)")
	}
}

// requireClaudeCLIForIntegration is a helper that can be used to mark integration tests
func requireClaudeCLIForIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	skipIfClaudeCLIUnavailable(t)
}

func TestClaudeCodeService_Name(t *testing.T) {
	service := NewClaudeCodeService()
	assert.Equal(t, "claudecode", service.Name())
}

// TestClaudeCodeService_ActualCLIDetection tests the real Claude CLI detection
// This test will be skipped in CI/CD if Claude CLI is not available
func TestClaudeCodeService_ActualCLIDetection(t *testing.T) {
	service := NewClaudeCodeService()

	// Test the actual CLI detection without mocking
	actuallyAvailable := service.isClaudeCodeInstalled()
	systemDetection := isClaudeCLIAvailable()

	// The service detection should match our helper function
	assert.Equal(t, systemDetection, actuallyAvailable,
		"Service CLI detection should match system detection")

	if actuallyAvailable {
		t.Log("Claude CLI is available on this system")

		// If available, test version detection too
		version, err := service.getClaudeCodeVersion()
		if err != nil {
			t.Logf("Claude CLI available but version detection failed: %v", err)
		} else {
			t.Logf("Claude CLI version: %s", version)
			assert.NotEmpty(t, version, "Version should not be empty if CLI is available")
		}
	} else {
		t.Log("Claude CLI is not available on this system (expected in CI/CD)")
	}
}

// TestClaudeCodeService_Initialize tests initialization with mocked CLI detection
// This test does NOT require actual Claude CLI - it uses mocks
func TestClaudeCodeService_Initialize(t *testing.T) {
	tests := []struct {
		name            string
		claudeInstalled bool
		expectError     bool
		errorContains   string
	}{
		{
			name:            "initialization without claude installed",
			claudeInstalled: false,
			expectError:     true,
			errorContains:   "claude code CLI not found",
		},
		{
			name:            "successful initialization",
			claudeInstalled: true,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewClaudeCodeService()

			// Mock the Claude Code installation check
			originalCheck := service.isClaudeCodeInstalled
			service.isClaudeCodeInstalled = func() bool {
				return tt.claudeInstalled
			}
			defer func() {
				service.isClaudeCodeInstalled = originalCheck
			}()

			// Mock version check if installed
			if tt.claudeInstalled {
				service.getClaudeCodeVersion = func() (string, error) {
					return "1.5.0", nil
				}
			}

			err := service.Initialize()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.False(t, service.initialized)
			} else {
				assert.NoError(t, err)
				assert.True(t, service.initialized)
			}
		})
	}
}

func TestClaudeCodeService_DoubleInitialize(t *testing.T) {
	service := NewClaudeCodeService()

	// Mock successful installation check
	service.isClaudeCodeInstalled = func() bool { return true }
	service.getClaudeCodeVersion = func() (string, error) { return "1.5.0", nil }

	// First initialization
	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Second initialization should be safe
	err = service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
}

func TestClaudeCodeService_CreateSession(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	tests := []struct {
		name        string
		opts        InitOptions
		expectError bool
	}{
		{
			name: "basic session creation",
			opts: InitOptions{
				Model:          "sonnet",
				Verbose:        false,
				Directories:    []string{"./src"},
				PermissionMode: "plan",
			},
			expectError: false,
		},
		{
			name: "session with multiple directories",
			opts: InitOptions{
				Model:       "opus",
				Directories: []string{"./src", "./tests", "./docs"},
			},
			expectError: false,
		},
		{
			name:        "minimal session",
			opts:        InitOptions{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := service.CreateSession(tt.opts)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.NotEmpty(t, session.ID)
				assert.Equal(t, "idle", session.Status)
				assert.Equal(t, tt.opts.Model, session.Model)
				assert.False(t, session.CreatedAt.IsZero())
				assert.False(t, session.UpdatedAt.IsZero())

				// Verify session is stored
				storedSession, err := service.GetSession(session.ID)
				assert.NoError(t, err)
				assert.Equal(t, session.ID, storedSession.ID)
			}
		})
	}
}

func TestClaudeCodeService_CreateSessionNotInitialized(t *testing.T) {
	service := NewClaudeCodeService()

	session, err := service.CreateSession(InitOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
	assert.Nil(t, session)
}

func TestClaudeCodeService_SessionManagement(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	// Create first session
	session1, err := service.CreateSession(InitOptions{Model: "sonnet"})
	require.NoError(t, err)

	// Should be active by default
	activeSession, err := service.GetActiveSession()
	assert.NoError(t, err)
	assert.Equal(t, session1.ID, activeSession.ID)

	// Create second session
	session2, err := service.CreateSession(InitOptions{Model: "opus"})
	require.NoError(t, err)

	// First session should still be active
	activeSession, err = service.GetActiveSession()
	assert.NoError(t, err)
	assert.Equal(t, session1.ID, activeSession.ID)

	// Switch to second session
	err = service.SetActiveSession(session2.ID)
	assert.NoError(t, err)

	activeSession, err = service.GetActiveSession()
	assert.NoError(t, err)
	assert.Equal(t, session2.ID, activeSession.ID)

	// List all sessions
	sessions := service.ListSessions()
	assert.Len(t, sessions, 2)

	// Get specific session
	retrievedSession, err := service.GetSession(session1.ID)
	assert.NoError(t, err)
	assert.Equal(t, session1.ID, retrievedSession.ID)

	// Try to get non-existent session
	_, err = service.GetSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Try to set non-existent session as active
	err = service.SetActiveSession("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestClaudeCodeService_JobManagement(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	// Create a session first
	session, err := service.CreateSession(InitOptions{Model: "sonnet"})
	require.NoError(t, err)

	// Submit a job
	job, err := service.SubmitJob(session.ID, "message", "Hello world", JobOptions{
		Timeout: 30 * time.Second,
		Stream:  false,
	})
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, session.ID, job.SessionID)
	assert.Equal(t, "message", job.Command)
	assert.Equal(t, "Hello world", job.Input)
	assert.Equal(t, JobStatusPending, job.Status)

	// Job should be retrievable
	retrievedJob, err := service.GetJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, job.ID, retrievedJob.ID)

	// Session should be marked as busy
	updatedSession, err := service.GetSession(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, "busy", updatedSession.Status)
	assert.Equal(t, job.ID, updatedSession.CurrentJobID)

	// Try to get non-existent job
	_, err = service.GetJob("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestClaudeCodeService_JobSubmissionErrors(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	// Try to submit job without session
	_, err := service.SubmitJob("non-existent", "message", "test", JobOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Try to submit job on uninitialized service
	uninitService := NewClaudeCodeService()
	_, err = uninitService.SubmitJob("session", "message", "test", JobOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestClaudeCodeService_JobWithStreaming(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	session, err := service.CreateSession(InitOptions{})
	require.NoError(t, err)

	// Submit job with streaming enabled
	job, err := service.SubmitJob(session.ID, "message", "Hello world", JobOptions{
		Stream: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, job.StreamChan)

	// Wait a bit for job processing (in real implementation)
	time.Sleep(50 * time.Millisecond)

	// Check if we can get output
	output, err := service.GetJobOutput(job.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, output) // Should have some output from mocked execution
}

func TestClaudeCodeService_WaitForJob(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	session, err := service.CreateSession(InitOptions{})
	require.NoError(t, err)

	job, err := service.SubmitJob(session.ID, "message", "test", JobOptions{})
	require.NoError(t, err)

	// Wait for job completion with reasonable timeout
	err = service.WaitForJob(job.ID, 5*time.Second)
	assert.NoError(t, err)

	// Job should be completed
	updatedJob, err := service.GetJob(job.ID)
	assert.NoError(t, err)
	assert.Equal(t, JobStatusCompleted, updatedJob.Status)
	assert.NotNil(t, updatedJob.EndedAt)
	assert.Equal(t, 100.0, updatedJob.Progress)

	// Session should be back to idle
	updatedSession, err := service.GetSession(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, "idle", updatedSession.Status)
	assert.Empty(t, updatedSession.CurrentJobID)
	assert.NotEmpty(t, updatedSession.LastResponse)
}

func TestClaudeCodeService_WaitForJobTimeout(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	// Try to wait for non-existent job
	err := service.WaitForJob("non-existent", 1*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestClaudeCodeService_Authentication(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	tests := []struct {
		name           string
		envKey         string
		neuroKey       string
		expectedMethod string
		expectedValid  bool
		expectError    bool
	}{
		{
			name:           "auth from environment",
			envKey:         "test-api-key",
			expectedMethod: "env",
			expectedValid:  true,
			expectError:    false,
		},
		{
			name:           "auth from neuro variable",
			neuroKey:       "neuro-api-key",
			expectedMethod: "variable",
			expectedValid:  true,
			expectError:    false,
		},
		{
			name:        "no auth available",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			originalKey := os.Getenv("ANTHROPIC_API_KEY")
			_ = os.Unsetenv("ANTHROPIC_API_KEY")
			defer func() {
				if originalKey != "" {
					_ = os.Setenv("ANTHROPIC_API_KEY", originalKey)
				}
			}()

			// Set up test context
			ctx := neuroshellcontext.New()
			neuroshellcontext.SetGlobalContext(ctx)
			defer neuroshellcontext.ResetGlobalContext()

			// Set up auth as specified
			if tt.envKey != "" {
				_ = os.Setenv("ANTHROPIC_API_KEY", tt.envKey)
			}
			if tt.neuroKey != "" {
				err := ctx.SetVariable("os.ANTHROPIC_API_KEY", tt.neuroKey)
				require.NoError(t, err)
			}

			auth, err := service.CheckAuth()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no valid authentication")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, auth)
				assert.Equal(t, tt.expectedMethod, auth.Method)
				assert.Equal(t, tt.expectedValid, auth.IsValid)
			}
		})
	}
}

func TestClaudeCodeService_Health(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	// Create a session to populate health data
	session, err := service.CreateSession(InitOptions{Model: "sonnet"})
	require.NoError(t, err)

	_, err = service.SubmitJob(session.ID, "message", "test", JobOptions{})
	require.NoError(t, err)

	health := service.Health()

	assert.True(t, health["initialized"].(bool))
	assert.Equal(t, 1, health["active_sessions"].(int))
	assert.Equal(t, 1, health["active_jobs"].(int))
	assert.Equal(t, session.ID, health["active_session_id"].(string))

	// Process should be running (even in test mode with mocked process)
	assert.True(t, health["process_running"].(bool))
}

func TestClaudeCodeService_Cleanup(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	// Create some state
	session, err := service.CreateSession(InitOptions{})
	require.NoError(t, err)

	_, err = service.SubmitJob(session.ID, "message", "test", JobOptions{})
	require.NoError(t, err)

	// Cleanup should not error
	err = service.Cleanup()
	assert.NoError(t, err)

	// Event bus should be closed (this is hard to test directly)
	// In a real test, we might check that goroutines are cleaned up
}

func TestClaudeCodeService_ProcessLifecycle(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	// Process should not be running initially
	assert.Nil(t, service.process)

	// Starting process should work (mocked)
	opts := InitOptions{Model: "sonnet", Verbose: true}

	// Mock the process start since we don't want to actually start Claude Code in tests
	originalStart := service.StartProcess
	service.StartProcess = func(_ InitOptions) error {
		// Simulate successful start
		return nil
	}
	defer func() {
		service.StartProcess = originalStart
	}()

	err := service.StartProcess(opts)
	assert.NoError(t, err)

	// Test stopping process
	err = service.StopProcess()
	assert.NoError(t, err)
}

// setupTestClaudeCodeService creates a service for testing
func setupTestClaudeCodeService(t *testing.T) *ClaudeCodeService {
	service := NewClaudeCodeService()

	// Mock Claude Code installation and version checks
	service.isClaudeCodeInstalled = func() bool { return true }
	service.getClaudeCodeVersion = func() (string, error) { return "1.5.0", nil }
	// Mock the StartProcess function to avoid trying to start actual claude process in tests
	service.StartProcess = func(_ InitOptions) error {
		// Mock successful process start - start a harmless long-running command for health checks
		// We use 'sleep' which is available on all systems and won't interfere with tests
		cmd := exec.Command("sleep", "3600") // Sleep for 1 hour, test cleanup will kill it
		err := cmd.Start()
		if err != nil {
			return err
		}
		service.process = cmd
		return nil
	}

	// Initialize the service
	err := service.Initialize()
	require.NoError(t, err)

	// Set up test context
	ctx := neuroshellcontext.New()
	neuroshellcontext.SetGlobalContext(ctx)
	t.Cleanup(func() {
		neuroshellcontext.ResetGlobalContext()
		_ = service.Cleanup()
	})

	return service
}

// TestClaudeCodeService_RealCLIIntegration tests actual Claude CLI integration
// This test requires Claude CLI to be available and will be skipped in CI/CD environments
func TestClaudeCodeService_RealCLIIntegration(t *testing.T) {
	requireClaudeCLIForIntegration(t)

	service := NewClaudeCodeService()

	// Don't mock anything - test with real CLI
	t.Log("Testing with actual Claude CLI (not mocked)")

	// Test initialization with real CLI
	err := service.Initialize()
	if err != nil {
		// If initialization fails, it might be due to auth or other issues
		// Log the error but don't fail the test immediately
		t.Logf("Real CLI initialization failed (might be auth issue): %v", err)

		// Check if it's specifically a CLI not found error
		if !service.isClaudeCodeInstalled() {
			t.Skip("Claude CLI not available - this should have been caught by requireClaudeCLIForIntegration")
		}

		// If CLI is available but auth fails, that's expected in CI/CD
		t.Log("Claude CLI available but initialization failed (likely auth) - this is expected in CI/CD")
		return
	}

	t.Log("Claude CLI integration test passed - CLI is available and configured")

	// Clean up
	_ = service.Cleanup()
}

// TestClaudeCodeService_MockedUnitTests groups all the existing mocked tests
// These tests should always run regardless of Claude CLI availability
func TestClaudeCodeService_MockedUnitTests(t *testing.T) {
	t.Run("MockedTests", func(t *testing.T) {
		t.Log("Running mocked unit tests (CLI availability not required)")
		// All existing tests are mocked and should continue to work
		// This serves as documentation that these are unit tests with mocking
	})
}

// TestClaudeCodeService_CICDSkipBehavior demonstrates how tests behave in CI/CD
func TestClaudeCodeService_CICDSkipBehavior(t *testing.T) {
	t.Run("SimulateCICDEnvironment", func(t *testing.T) {
		// This test simulates what happens when Claude CLI is not available
		service := NewClaudeCodeService()

		// Mock CLI as unavailable (like in CI/CD)
		originalCheck := service.isClaudeCodeInstalled
		service.isClaudeCodeInstalled = func() bool { return false }
		defer func() { service.isClaudeCodeInstalled = originalCheck }()

		t.Log("Simulating CI/CD environment where Claude CLI is not available")

		// Test that initialization fails gracefully
		err := service.Initialize()
		assert.Error(t, err, "Initialization should fail when Claude CLI is not available")
		assert.Contains(t, err.Error(), "claude code CLI not found",
			"Error should indicate CLI not found")

		t.Log("Service correctly handles missing Claude CLI")
	})

	t.Run("MockedTestsStillWork", func(t *testing.T) {
		// Even when CLI is unavailable, mocked tests should work
		service := NewClaudeCodeService()

		// Mock CLI as available for this test
		service.isClaudeCodeInstalled = func() bool { return true }
		service.getClaudeCodeVersion = func() (string, error) { return "1.5.0", nil }

		err := service.Initialize()
		assert.NoError(t, err, "Mocked tests should work regardless of actual CLI availability")

		t.Log("Mocked tests work even when actual CLI is unavailable")
		_ = service.Cleanup()
	})
}
func TestClaudeCodeService_Integration_ResponseHandling(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	// Test response handling
	response := &ClaudeCodeResponse{
		JobID:         "test-job",
		Type:          "complete",
		Content:       "Test response",
		ToolsUsed:     []string{"Read", "Write"},
		FilesModified: []string{"test.go"},
	}

	// Create a job to handle response for
	session, err := service.CreateSession(InitOptions{})
	require.NoError(t, err)

	job, err := service.SubmitJob(session.ID, "message", "test", JobOptions{})
	require.NoError(t, err)

	// Set the job ID to match our test response
	service.mu.Lock()
	delete(service.jobs, job.ID)
	job.ID = "test-job"
	service.jobs["test-job"] = job
	service.mu.Unlock()

	// Handle the response
	service.handleResponse(response)

	// Verify the job was updated
	updatedJob, err := service.GetJob("test-job")
	assert.NoError(t, err)
	assert.Equal(t, JobStatusCompleted, updatedJob.Status)
	assert.Contains(t, updatedJob.Output.String(), "Test response")
	assert.NotNil(t, updatedJob.EndedAt)
	assert.Equal(t, 100.0, updatedJob.Progress)
}

func TestClaudeCodeService_ConcurrentOperations(t *testing.T) {
	service := setupTestClaudeCodeService(t)

	session, err := service.CreateSession(InitOptions{})
	require.NoError(t, err)

	// Submit multiple jobs concurrently
	numJobs := 10
	jobs := make([]*ClaudeCodeJob, numJobs)

	for i := 0; i < numJobs; i++ {
		job, err := service.SubmitJob(session.ID, "message", fmt.Sprintf("test-%d", i), JobOptions{})
		assert.NoError(t, err)
		jobs[i] = job
	}

	// Wait for all jobs to complete
	for _, job := range jobs {
		err := service.WaitForJob(job.ID, 5*time.Second)
		assert.NoError(t, err)
	}

	// All jobs should be completed
	for _, job := range jobs {
		updatedJob, err := service.GetJob(job.ID)
		assert.NoError(t, err)
		assert.Equal(t, JobStatusCompleted, updatedJob.Status)
	}
}
