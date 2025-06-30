package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/testutils"
)

func TestVariableService_Name(t *testing.T) {
	service := NewVariableService()
	assert.Equal(t, "variable", service.Name())
}

func TestVariableService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		ctx  *testutils.MockContext
		want error
	}{
		{
			name: "successful initialization",
			ctx:  testutils.NewMockContext(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewVariableService()
			err := service.Initialize(tt.ctx)

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
			wantError: "variable 'missing_var' not found",
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
			ctx := testutils.NewMockContextWithVars(tt.setupVars)

			// Setup context error if needed
			if tt.setupError != nil {
				ctx.SetGetVariableError(tt.setupError)
			}

			// Initialize service
			err := service.Initialize(ctx)
			require.NoError(t, err)

			// Test Get
			value, err := service.Get(tt.varName, ctx)

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
	ctx := testutils.NewMockContext()

	value, err := service.Get("test_var", ctx)

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
			ctx := testutils.NewMockContext()

			// Setup context error if needed
			if tt.setupError != nil {
				ctx.SetSetVariableError(tt.setupError)
			}

			// Initialize service
			err := service.Initialize(ctx)
			require.NoError(t, err)

			// Test Set
			err = service.Set(tt.varName, tt.varValue, ctx)

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
	ctx := testutils.NewMockContext()

	err := service.Set("test_var", "test_value", ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not initialized")
}

func TestVariableService_InterpolateString(t *testing.T) {
	service := NewVariableService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test interpolation - note this will fail since we're using MockContext
	// but we need to test the error handling
	result, err := service.InterpolateString("Hello ${name}", ctx)

	// Should fail because MockContext is not a NeuroContext
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
	assert.Empty(t, result)
}

func TestVariableService_InterpolateString_NotInitialized(t *testing.T) {
	service := NewVariableService()
	ctx := testutils.NewMockContext()

	result, err := service.InterpolateString("Hello ${name}", ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not initialized")
	assert.Empty(t, result)
}

func TestVariableService_GetAllVariables(t *testing.T) {
	service := NewVariableService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test GetAllVariables - note this will fail since we're using MockContext
	result, err := service.GetAllVariables(ctx)

	// Should fail because MockContext is not a NeuroContext
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
	assert.Nil(t, result)
}

func TestVariableService_GetAllVariables_NotInitialized(t *testing.T) {
	service := NewVariableService()
	ctx := testutils.NewMockContext()

	result, err := service.GetAllVariables(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not initialized")
	assert.Nil(t, result)
}

// Benchmark tests
func BenchmarkVariableService_Get(b *testing.B) {
	service := NewVariableService()
	ctx := testutils.NewMockContextWithVars(map[string]string{
		"test_var": "test_value",
	})

	err := service.Initialize(ctx)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Get("test_var", ctx)
	}
}

func BenchmarkVariableService_Set(b *testing.B) {
	service := NewVariableService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.Set("test_var", "test_value", ctx)
	}
}

func BenchmarkVariableService_GetSystemVariable(b *testing.B) {
	service := NewVariableService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Get("@user", ctx)
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
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := service.Get(tc.variable, ctx)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, value)
		})
	}
}

func TestVariableService_ConcurrentAccess(t *testing.T) {
	service := NewVariableService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool)

	// Start multiple goroutines for concurrent access
	for i := 0; i < 10; i++ {
		go func(id int) {
			// Set a variable
			varName := fmt.Sprintf("var_%d", id)
			varValue := fmt.Sprintf("value_%d", id)

			err := service.Set(varName, varValue, ctx)
			assert.NoError(t, err)

			// Get the variable
			value, err := service.Get(varName, ctx)
			assert.NoError(t, err)
			assert.Equal(t, varValue, value)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
