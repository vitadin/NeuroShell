package assert

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/services"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

func TestAssertEqualCommand_Name(t *testing.T) {
	cmd := &AssertEqualCommand{}
	assert.Equal(t, "assert-equal", cmd.Name())
}

func TestAssertEqualCommand_ParseMode(t *testing.T) {
	cmd := &AssertEqualCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestAssertEqualCommand_Description(t *testing.T) {
	cmd := &AssertEqualCommand{}
	assert.Equal(t, "Compare two values for equality", cmd.Description())
}

func TestAssertEqualCommand_Usage(t *testing.T) {
	cmd := &AssertEqualCommand{}
	assert.Equal(t, "\\assert-equal[expect=expected_value, actual=actual_value]", cmd.Usage())
}

func TestAssertEqualCommand_Execute_MissingArguments(t *testing.T) {
	cmd := &AssertEqualCommand{}

	tests := []struct {
		name string
		args map[string]string
	}{
		{
			name: "missing both arguments",
			args: map[string]string{},
		},
		{
			name: "missing expect argument",
			args: map[string]string{
				"actual": "value",
			},
		},
		{
			name: "missing actual argument",
			args: map[string]string{
				"expect": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testutils.NewMockContext()

			err := cmd.Execute(tt.args, "", ctx)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Usage:")
		})
	}
}

func TestAssertEqualCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &AssertEqualCommand{}
	ctx := testutils.NewMockContext()

	// Don't set up services to test service unavailability

	args := map[string]string{
		"expect": "hello",
		"actual": "hello",
	}

	err := cmd.Execute(args, "", ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interpolation service not available")
}

func TestAssertEqualCommand_Execute_InterpolationServiceError(t *testing.T) {
	cmd := &AssertEqualCommand{}
	ctx := testutils.NewMockContext()

	// Set up services - but MockContext will cause interpolation to fail
	setupTestServices(t, ctx)

	args := map[string]string{
		"expect": "hello",
		"actual": "hello",
	}

	err := cmd.Execute(args, "", ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to interpolate expected value")
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
}

func TestAssertEqualCommand_Execute_WrongServiceType(t *testing.T) {
	cmd := &AssertEqualCommand{}
	ctx := testutils.NewMockContext()

	// Setup registry but register wrong service type under "interpolation" name
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Register a different service under "interpolation" name
	err := services.GetGlobalRegistry().RegisterService(&mockWrongService{})
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})

	args := map[string]string{
		"expect": "hello",
		"actual": "hello",
	}

	err = cmd.Execute(args, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect type")
}

func TestAssertEqualCommand_Execute_VariableServiceError(t *testing.T) {
	cmd := &AssertEqualCommand{}
	ctx := testutils.NewMockContext()

	// Set up only interpolation service, missing variable service
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Register only interpolation service
	interpolationService := services.NewInterpolationService()
	err := services.GetGlobalRegistry().RegisterService(interpolationService)
	require.NoError(t, err)
	err = interpolationService.Initialize(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})

	args := map[string]string{
		"expect": "hello",
		"actual": "hello",
	}

	err = cmd.Execute(args, "", ctx)
	assert.Error(t, err)
	// Should fail at interpolation stage due to MockContext incompatibility
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
}

// Helper functions

// setupTestServices sets up test services in the global registry
func setupTestServices(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Register InterpolationService
	interpolationService := services.NewInterpolationService()
	err := services.GetGlobalRegistry().RegisterService(interpolationService)
	require.NoError(t, err)
	err = interpolationService.Initialize(ctx)
	require.NoError(t, err)

	// Register VariableService
	variableService := services.NewVariableService()
	err = services.GetGlobalRegistry().RegisterService(variableService)
	require.NoError(t, err)
	err = variableService.Initialize(ctx)
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
	})
}

// captureOutput captures stdout during function execution
func captureOutput(fn func()) string {
	// Save original stdout
	oldStdout := os.Stdout

	// Create pipe
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Channel to receive output
	outputChan := make(chan string)

	// Start goroutine to read output
	go func() {
		defer close(outputChan)
		output, _ := io.ReadAll(r)
		outputChan <- string(output)
	}()

	// Execute function
	fn()

	// Restore stdout and close writer
	_ = w.Close()
	os.Stdout = oldStdout

	// Return captured output
	return <-outputChan
}

// mockWrongService is a mock service with wrong type for testing
type mockWrongService struct{}

func (m *mockWrongService) Name() string                          { return "interpolation" }
func (m *mockWrongService) Initialize(_ neurotypes.Context) error { return nil }

// Interface compliance test
func TestAssertEqualCommand_InterfaceCompliance(_ *testing.T) {
	var _ neurotypes.Command = (*AssertEqualCommand)(nil)
}

// Benchmark tests
func BenchmarkAssertEqualCommand_Execute_ServiceError(b *testing.B) {
	cmd := &AssertEqualCommand{}
	ctx := testutils.NewMockContext()

	// Don't setup services to measure error handling overhead
	args := map[string]string{
		"expect": "benchvalue",
		"actual": "benchvalue",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(args, "", ctx)
	}
}

func BenchmarkAssertEqualCommand_Execute_WithServices(b *testing.B) {
	cmd := &AssertEqualCommand{}
	ctx := testutils.NewMockContext()

	// Setup services (will fail at interpolation but measures setup overhead)
	services.SetGlobalRegistry(services.NewRegistry())
	interpolationService := services.NewInterpolationService()
	_ = services.GetGlobalRegistry().RegisterService(interpolationService)
	_ = interpolationService.Initialize(ctx)

	variableService := services.NewVariableService()
	_ = services.GetGlobalRegistry().RegisterService(variableService)
	_ = variableService.Initialize(ctx)

	args := map[string]string{
		"expect": "benchvalue",
		"actual": "benchvalue",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(args, "", ctx)
	}
}
