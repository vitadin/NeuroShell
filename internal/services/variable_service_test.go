package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

func TestVariableService_Name(t *testing.T) {
	service := NewVariableService()
	assert.Equal(t, "variable", service.Name())
}

func TestVariableService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		ctx  neurotypes.Context
		want error
	}{
		{
			name: "successful initialization",
			ctx:  context.NewTestContext(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewVariableService()
			err := service.Initialize()

			if tt.want != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.want.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, service.initialized)
			}
		})
	}
}

func TestVariableService_Get(t *testing.T) {
	tests := []struct {
		name       string
		varName    string
		setupVars  map[string]string
		setupError error
		wantValue  string
		wantError  string
	}{
		{
			name:      "get existing variable",
			varName:   "test_var",
			setupVars: map[string]string{"test_var": "test_value"},
			wantValue: "test_value",
		},
		{
			name:      "get non-existing variable",
			varName:   "missing_var",
			setupVars: map[string]string{},
			wantValue: "", // Should return empty string, not error
		},
		{
			name:      "get system variable @user",
			varName:   "@user",
			setupVars: map[string]string{},
			wantValue: "testuser",
		},
		{
			name:      "get system variable @pwd",
			varName:   "@pwd",
			setupVars: map[string]string{},
			wantValue: "/test/pwd",
		},
		{
			name:      "get system variable #test_mode",
			varName:   "#test_mode",
			setupVars: map[string]string{},
			wantValue: "true",
		},
		{
			name:       "context error",
			varName:    "any_var",
			setupVars:  map[string]string{},
			setupError: assert.AnError,
			wantError:  "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewVariableService()
			ctx := context.NewTestContext()
			// Setup variables in context
			for k, v := range tt.setupVars {
				_ = ctx.SetVariable(k, v)
			}

			// Setup context error if needed - real context doesn't have error injection
			// so we skip tests that expect context errors
			if tt.setupError != nil {
				t.Skip("Skipping error injection test with real context")
			}

			// Initialize service
			err := service.Initialize()
			require.NoError(t, err)

			// Setup global context for testing
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			// Test Get
			// Setup global context for testing
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			value, err := service.Get(tt.varName)

			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}

func TestVariableService_Get_NotInitialized(t *testing.T) {
	service := NewVariableService()
	ctx := context.NewTestContext()

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	value, err := service.Get("test_var")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not initialized")
	assert.Empty(t, value)
}

func TestVariableService_Set(t *testing.T) {
	tests := []struct {
		name       string
		varName    string
		varValue   string
		setupError error
		wantError  string
	}{
		{
			name:     "set new variable",
			varName:  "new_var",
			varValue: "new_value",
		},
		{
			name:     "set existing variable",
			varName:  "existing_var",
			varValue: "updated_value",
		},
		{
			name:     "set empty value",
			varName:  "empty_var",
			varValue: "",
		},
		{
			name:     "set variable with special characters",
			varName:  "special_var",
			varValue: "value with spaces & symbols!",
		},
		{
			name:       "context error",
			varName:    "any_var",
			varValue:   "any_value",
			setupError: assert.AnError,
			wantError:  "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewVariableService()
			ctx := context.NewTestContext()

			// Setup context error if needed - real context doesn't have error injection
			// so we skip tests that expect context errors
			if tt.setupError != nil {
				t.Skip("Skipping error injection test with real context")
			}

			// Initialize service
			err := service.Initialize()
			require.NoError(t, err)

			// Test Set
			// Setup global context for testing
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			err = service.Set(tt.varName, tt.varValue)

			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			} else {
				assert.NoError(t, err)

				// Verify the variable was set
				value, err := ctx.GetVariable(tt.varName)
				assert.NoError(t, err)
				assert.Equal(t, tt.varValue, value)
			}
		})
	}
}

func TestVariableService_Set_NotInitialized(t *testing.T) {
	service := NewVariableService()
	ctx := context.NewTestContext()

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	err := service.Set("test_var", "test_value")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not initialized")
}

func TestVariableService_InterpolateString(t *testing.T) {
	service := NewVariableService()
	ctx := context.NewTestContext()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test interpolation - note this will fail since we're using MockContext
	// but we need to test the error handling
	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	result, err := service.InterpolateString("Hello ${name}")

	// TestContext should work properly
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestVariableService_InterpolateString_NotInitialized(t *testing.T) {
	service := NewVariableService()
	ctx := context.NewTestContext()

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	result, err := service.InterpolateString("Hello ${name}")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not initialized")
	assert.Empty(t, result)
}

func TestVariableService_GetAllVariables(t *testing.T) {
	service := NewVariableService()
	ctx := context.NewTestContext()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test GetAllVariables - note this will fail since we're using MockContext
	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	result, err := service.GetAllVariables()

	// TestContext should work properly
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestVariableService_GetAllVariables_NotInitialized(t *testing.T) {
	service := NewVariableService()
	ctx := context.NewTestContext()

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	result, err := service.GetAllVariables()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not initialized")
	assert.Nil(t, result)
}

// Benchmark tests
func BenchmarkVariableService_Get(b *testing.B) {
	service := NewVariableService()
	ctx := context.NewTestContext()
	_ = ctx.SetVariable("test_var", "test_value")

	err := service.Initialize()
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Setup global context for testing
		context.SetGlobalContext(ctx)
		defer context.ResetGlobalContext()

		_, _ = service.Get("test_var")
	}
}

func BenchmarkVariableService_Set(b *testing.B) {
	service := NewVariableService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Setup global context for testing
		context.SetGlobalContext(ctx)
		defer context.ResetGlobalContext()

		_ = service.Set("test_var", "test_value")
	}
}

func BenchmarkVariableService_GetSystemVariable(b *testing.B) {
	service := NewVariableService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Setup global context for testing
		context.SetGlobalContext(ctx)
		defer context.ResetGlobalContext()

		_, _ = service.Get("@user")
	}
}

// Table-driven tests for comprehensive coverage
func TestVariableService_SystemVariables(t *testing.T) {
	testCases := []struct {
		name     string
		variable string
		expected string
	}{
		{"user variable", "@user", "testuser"},
		{"pwd variable", "@pwd", "/test/pwd"},
		{"home variable", "@home", "/test/home"},
		{"date variable", "@date", "2024-01-01"},
		{"os variable", "@os", "test-os"},
		{"test_mode variable", "#test_mode", "true"},
	}

	service := NewVariableService()
	ctx := context.NewTestContext()

	err := service.Initialize()
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup global context for testing
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			value, err := service.Get(tc.variable)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		})
	}
}

// TestVariableService_ConcurrentAccess has been removed as part of context mutex simplification.
// NeuroShell now operates as a sequential, single-threaded shell where concurrent access
// to variables is not an expected use case. See docs/context-mutex-simplification.md for details.

func TestVariableService_SetSystemVariable(t *testing.T) {
	tests := []struct {
		name      string
		varName   string
		varValue  string
		wantError string
	}{
		{
			name:     "set system variable with underscore prefix",
			varName:  "_test_var",
			varValue: "success",
		},
		{
			name:     "set system variable with @ prefix",
			varName:  "@custom",
			varValue: "custom_value",
		},
		{
			name:     "set system variable with # prefix",
			varName:  "#custom_mode",
			varValue: "test",
		},
		{
			name:     "set system variable with empty value",
			varName:  "_empty",
			varValue: "",
		},
		{
			name:      "fail to set regular variable",
			varName:   "regular_var",
			varValue:  "value",
			wantError: "SetSystemVariable can only set system variables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewVariableService()
			ctx := context.New() // Use real NeuroContext instead of mock

			// Initialize service
			err := service.Initialize()
			require.NoError(t, err)

			// Test SetSystemVariable
			// Setup global context for testing
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			err = service.SetSystemVariable(tt.varName, tt.varValue)

			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			} else {
				assert.NoError(t, err)

				// Verify the variable was set by getting it back
				value, err := service.Get(tt.varName)
				assert.NoError(t, err)
				assert.Equal(t, tt.varValue, value)
			}
		})
	}
}

func TestVariableService_SetSystemVariable_NotInitialized(t *testing.T) {
	service := NewVariableService()
	ctx := context.New()

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	err := service.SetSystemVariable("_test", "value")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not initialized")
}

func TestVariableService_SetSystemVariable_WrongContextType(t *testing.T) {
	service := NewVariableService()
	ctx := context.NewTestContext() // Use TestContext

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Test SetSystemVariable with TestContext
	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	err = service.SetSystemVariable("_test", "value")

	assert.NoError(t, err)
}

// TestVariableService_ValidateVariableName tests variable name validation
func TestVariableService_ValidateVariableName(t *testing.T) {
	service := &VariableService{initialized: true} // Just for validation functions

	tests := []struct {
		name        string
		varName     string
		expectError bool
	}{
		{"Valid user variable", "myvar", false},
		{"Valid underscore variable", "_style", false},
		{"Invalid system prefix @", "@system", true},
		{"Invalid system prefix #", "#meta", true},
		{"Invalid whitespace", "my var", true},
		{"Empty name", "", true},
		{"Invalid underscore", "_invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateVariableName(tt.varName)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for variable name '%s', got nil", tt.varName)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for variable name '%s', got: %v", tt.varName, err)
			}
		})
	}
}

// TestVariableService_AnalyzeVariable tests the variable analysis function
func TestVariableService_AnalyzeVariable(t *testing.T) {
	service := &VariableService{initialized: true}

	tests := []struct {
		name         string
		varName      string
		expectedType context.VariableType
		isSystem     bool
		isReadOnly   bool
	}{
		{"User variable", "myvar", context.TypeUser, false, false},
		{"System variable @", "@pwd", context.TypeSystem, true, true},
		{"Metadata variable #", "#session_id", context.TypeMetadata, true, true},
		{"Command variable _", "_output", context.TypeCommand, true, true},
		{"Allowed command variable", "_style", context.TypeCommand, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := service.AnalyzeVariable(tt.varName)

			if info.Type != tt.expectedType {
				t.Errorf("Expected type %v, got %v", tt.expectedType, info.Type)
			}
			if info.IsSystem != tt.isSystem {
				t.Errorf("Expected IsSystem=%v, got %v", tt.isSystem, info.IsSystem)
			}
			if info.IsReadOnly != tt.isReadOnly {
				t.Errorf("Expected IsReadOnly=%v, got %v", tt.isReadOnly, info.IsReadOnly)
			}
		})
	}
}

// TestVariableService_GetEnvVariable tests environment variable operations
func TestVariableService_GetEnvVariable(t *testing.T) {
	service := NewVariableService()
	ctx := context.New()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Test getting environment variable (using the context's mock env)
	value := service.GetEnv("NONEXISTENT_VAR")
	// Should return empty string for non-existent variable
	assert.Empty(t, value)
}

// TestVariableService_SetEnvVariable tests setting environment variables
func TestVariableService_SetEnvVariable(t *testing.T) {
	service := NewVariableService()
	ctx := context.New()

	// Initialize service
	err := service.Initialize()
	require.NoError(t, err)

	// Setup global context for testing
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Test setting environment variable
	err = service.SetEnvVariable("TEST_ENV", "test_value")
	assert.NoError(t, err)

	// Verify it was set
	value := service.GetEnv("TEST_ENV")
	assert.Equal(t, "test_value", value)
}

// TestVariableService_ErrorHandling tests comprehensive error handling
func TestVariableService_ErrorHandling(t *testing.T) {
	t.Run("Uninitialized service", func(t *testing.T) {
		service := NewVariableService()

		_, err := service.Get("test")
		if err == nil || err.Error() != "variable service not initialized" {
			t.Errorf("Expected 'variable service not initialized' error, got: %v", err)
		}
	})

	t.Run("Service with nil subcontext", func(t *testing.T) {
		service := &VariableService{initialized: true, varCtx: nil}

		_, err := service.Get("test")
		if err == nil || err.Error() != "variable subcontext not available" {
			t.Errorf("Expected 'variable subcontext not available' error, got: %v", err)
		}
	})
}
