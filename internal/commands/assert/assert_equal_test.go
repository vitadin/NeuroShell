package assert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestEqualCommand_Name(t *testing.T) {
	cmd := &EqualCommand{}
	assert.Equal(t, "assert-equal", cmd.Name())
}

func TestEqualCommand_ParseMode(t *testing.T) {
	cmd := &EqualCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestEqualCommand_Description(t *testing.T) {
	cmd := &EqualCommand{}
	assert.Equal(t, "Compare two values for equality", cmd.Description())
}

func TestEqualCommand_Usage(t *testing.T) {
	cmd := &EqualCommand{}
	assert.Equal(t, "\\assert-equal[expect=expected_value, actual=actual_value]", cmd.Usage())
}

// setupAssertTestRegistry sets up a test registry with required services
func setupAssertTestRegistry(t *testing.T, ctx *context.NeuroContext) func() {
	// Save the original registry and context
	originalRegistry := services.GetGlobalRegistry()
	originalContext := context.GetGlobalContext()

	// Create new test registry
	testRegistry := services.NewRegistry()

	// Create and register variable service
	variableService := services.NewVariableService()
	err := testRegistry.RegisterService(variableService)
	require.NoError(t, err)

	// Initialize all services
	err = testRegistry.InitializeAll()
	require.NoError(t, err)

	// Set global registry and context for testing
	services.SetGlobalRegistry(testRegistry)
	context.SetGlobalContext(ctx)

	// Return cleanup function
	return func() {
		services.SetGlobalRegistry(originalRegistry)
		context.SetGlobalContext(originalContext)
	}
}

func TestEqualCommand_Execute_MissingArguments(t *testing.T) {
	cmd := &EqualCommand{}
	ctx := context.New()
	cleanup := setupAssertTestRegistry(t, ctx)
	defer cleanup()

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
			err := cmd.Execute(tt.args, "")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "Usage:")
		})
	}
}

func TestEqualCommand_Execute_EqualValues(t *testing.T) {
	cmd := &EqualCommand{}
	ctx := context.New()
	cleanup := setupAssertTestRegistry(t, ctx)
	defer cleanup()

	args := map[string]string{
		"expect": "test_value",
		"actual": "test_value",
	}

	err := cmd.Execute(args, "")

	// Passing assertions should not return errors
	assert.NoError(t, err)

	// Check assert-specific variables are set correctly for passing assertion
	result, err := ctx.GetVariable("_assert_result")
	assert.NoError(t, err)
	assert.Equal(t, "PASS", result)

	expected, err := ctx.GetVariable("_assert_expected")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", expected)

	actual, err := ctx.GetVariable("_assert_actual")
	assert.NoError(t, err)
	assert.Equal(t, "test_value", actual)

	// Note: _status and _error are now managed by ErrorManagementService at framework level
}

func TestEqualCommand_Execute_UnequalValues(t *testing.T) {
	cmd := &EqualCommand{}
	ctx := context.New()
	cleanup := setupAssertTestRegistry(t, ctx)
	defer cleanup()

	args := map[string]string{
		"expect": "expected_value",
		"actual": "different_value",
	}

	err := cmd.Execute(args, "")

	// Assert commands should return errors when assertions fail
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "assertion failed")

	// Check assert-specific variables are still set correctly
	result, err := ctx.GetVariable("_assert_result")
	assert.NoError(t, err)
	assert.Equal(t, "FAIL", result)

	expected, err := ctx.GetVariable("_assert_expected")
	assert.NoError(t, err)
	assert.Equal(t, "expected_value", expected)

	actual, err := ctx.GetVariable("_assert_actual")
	assert.NoError(t, err)
	assert.Equal(t, "different_value", actual)

	// Note: _status and _error are now managed by ErrorManagementService at framework level
}

func TestEqualCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &EqualCommand{}

	// Don't set up services to test service unavailability
	args := map[string]string{
		"expect": "value1",
		"actual": "value2",
	}

	err := cmd.Execute(args, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestEqualCommand_Execute_EdgeCases(t *testing.T) {
	cmd := &EqualCommand{}
	ctx := context.New()
	cleanup := setupAssertTestRegistry(t, ctx)
	defer cleanup()

	tests := []struct {
		name     string
		expect   string
		actual   string
		wantPass bool
	}{
		{
			name:     "empty strings equal",
			expect:   "",
			actual:   "",
			wantPass: true,
		},
		{
			name:     "whitespace differences",
			expect:   "value",
			actual:   " value ",
			wantPass: false,
		},
		{
			name:     "case sensitive comparison",
			expect:   "Value",
			actual:   "value",
			wantPass: false,
		},
		{
			name:     "numbers as strings",
			expect:   "123",
			actual:   "123",
			wantPass: true,
		},
		{
			name:     "special characters",
			expect:   "test@#$%",
			actual:   "test@#$%",
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]string{
				"expect": tt.expect,
				"actual": tt.actual,
			}

			err := cmd.Execute(args, "")

			// Check if command should succeed or fail
			if tt.wantPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "assertion failed")
			}

			// Check assertion result variable
			result, err := ctx.GetVariable("_assert_result")
			assert.NoError(t, err)
			if tt.wantPass {
				assert.Equal(t, "PASS", result)
			} else {
				assert.Equal(t, "FAIL", result)
			}
		})
	}
}

// BenchmarkEqualCommand_Execute benchmarks the execute method performance
func BenchmarkEqualCommand_Execute(b *testing.B) {
	cmd := &EqualCommand{}
	ctx := context.New()

	// Set up test registry
	testRegistry := services.NewRegistry()
	variableService := services.NewVariableService()
	err := testRegistry.RegisterService(variableService)
	require.NoError(b, err)
	err = testRegistry.InitializeAll()
	require.NoError(b, err)

	originalRegistry := services.GetGlobalRegistry()
	originalContext := context.GetGlobalContext()
	services.SetGlobalRegistry(testRegistry)
	context.SetGlobalContext(ctx)

	defer func() {
		services.SetGlobalRegistry(originalRegistry)
		context.SetGlobalContext(originalContext)
	}()

	args := map[string]string{
		"expect": "benchvalue",
		"actual": "benchvalue",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(args, "")
	}
}

// Interface compliance test
func TestEqualCommand_InterfaceCompliance(_ *testing.T) {
	var _ neurotypes.Command = (*EqualCommand)(nil)
}
