package services

import (
	"testing"

	"neuroshell/internal/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellPromptService_Name(t *testing.T) {
	service := NewShellPromptService()
	assert.Equal(t, "shell_prompt", service.Name())
}

func TestShellPromptService_Initialize(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	service := NewShellPromptService()
	assert.False(t, service.initialized)

	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Check that default configuration is set
	linesCount, err := ctx.GetVariable("_prompt_lines_count")
	assert.NoError(t, err)
	assert.Equal(t, "1", linesCount)

	line1, err := ctx.GetVariable("_prompt_line1")
	assert.NoError(t, err)
	assert.Equal(t, "neuro> ", line1)
}

func TestShellPromptService_Initialize_ExistingConfig(t *testing.T) {
	// Setup test context with existing configuration
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Pre-configure prompt
	err := ctx.SetVariable("_prompt_lines_count", "2")
	require.NoError(t, err)
	err = ctx.SetVariable("_prompt_line1", "custom> ")
	require.NoError(t, err)

	service := NewShellPromptService()
	err = service.Initialize()
	assert.NoError(t, err)

	// Should not override existing configuration
	linesCount, err := ctx.GetVariable("_prompt_lines_count")
	assert.NoError(t, err)
	assert.Equal(t, "2", linesCount)

	line1, err := ctx.GetVariable("_prompt_line1")
	assert.NoError(t, err)
	assert.Equal(t, "custom> ", line1)
}

func TestShellPromptService_GetPromptLines_Default(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	service := NewShellPromptService()
	err := service.Initialize()
	require.NoError(t, err)

	lines, err := service.GetPromptLines()
	assert.NoError(t, err)
	assert.Len(t, lines, 1)
	assert.Equal(t, "neuro> ", lines[0])
}

func TestShellPromptService_GetPromptLines_MultiLine(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Configure multi-line prompt
	err := ctx.SetVariable("_prompt_lines_count", "3")
	require.NoError(t, err)
	err = ctx.SetVariable("_prompt_line1", "${@pwd} [${#session_name}]")
	require.NoError(t, err)
	err = ctx.SetVariable("_prompt_line2", "├─ ${@time}")
	require.NoError(t, err)
	err = ctx.SetVariable("_prompt_line3", "└➤ ")
	require.NoError(t, err)

	service := NewShellPromptService()
	err = service.Initialize()
	require.NoError(t, err)

	lines, err := service.GetPromptLines()
	assert.NoError(t, err)
	assert.Len(t, lines, 3)
	assert.Equal(t, "${@pwd} [${#session_name}]", lines[0])
	assert.Equal(t, "├─ ${@time}", lines[1])
	assert.Equal(t, "└➤ ", lines[2])
}

func TestShellPromptService_GetPromptLines_PartialConfig(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Configure prompt with missing lines
	err := ctx.SetVariable("_prompt_lines_count", "3")
	require.NoError(t, err)
	err = ctx.SetVariable("_prompt_line1", "first> ")
	require.NoError(t, err)
	// _prompt_line2 is missing
	err = ctx.SetVariable("_prompt_line3", "third> ")
	require.NoError(t, err)

	service := NewShellPromptService()
	err = service.Initialize()
	require.NoError(t, err)

	lines, err := service.GetPromptLines()
	assert.NoError(t, err)
	// Should only return configured lines
	assert.Len(t, lines, 2)
	assert.Equal(t, "first> ", lines[0])
	assert.Equal(t, "third> ", lines[1])
}

func TestShellPromptService_GetPromptLines_InvalidCount(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	// Configure invalid lines count
	err := ctx.SetVariable("_prompt_lines_count", "10")
	require.NoError(t, err)
	err = ctx.SetVariable("_prompt_line1", "test> ")
	require.NoError(t, err)

	service := NewShellPromptService()
	err = service.Initialize()
	require.NoError(t, err)

	lines, err := service.GetPromptLines()
	assert.NoError(t, err)
	// Should fall back to 1 line when count is invalid
	assert.Len(t, lines, 1)
	assert.Equal(t, "test> ", lines[0])
}

func TestShellPromptService_GetPromptLines_NotInitialized(t *testing.T) {
	service := NewShellPromptService()

	lines, err := service.GetPromptLines()
	assert.Error(t, err)
	assert.Nil(t, lines)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestShellPromptService_GetPromptLinesCount(t *testing.T) {
	// Setup test context
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	defer context.ResetGlobalContext()

	tests := []struct {
		name          string
		configValue   string
		expectedCount int
	}{
		{
			name:          "valid count 1",
			configValue:   "1",
			expectedCount: 1,
		},
		{
			name:          "valid count 3",
			configValue:   "3",
			expectedCount: 3,
		},
		{
			name:          "valid count 5",
			configValue:   "5",
			expectedCount: 5,
		},
		{
			name:          "invalid count too low",
			configValue:   "0",
			expectedCount: 1,
		},
		{
			name:          "invalid count too high",
			configValue:   "10",
			expectedCount: 1,
		},
		{
			name:          "invalid format",
			configValue:   "abc",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx := context.NewTestContext()
			context.SetGlobalContext(ctx)

			err := ctx.SetVariable("_prompt_lines_count", tt.configValue)
			require.NoError(t, err)

			service := NewShellPromptService()
			err = service.Initialize()
			require.NoError(t, err)

			count, err := service.GetPromptLinesCount()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestShellPromptService_ValidatePromptConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(*context.NeuroContext)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid single line",
			setupFunc: func(ctx *context.NeuroContext) {
				_ = ctx.SetVariable("_prompt_lines_count", "1")
				_ = ctx.SetVariable("_prompt_line1", "test> ")
			},
			expectError: false,
		},
		{
			name: "valid multi line",
			setupFunc: func(ctx *context.NeuroContext) {
				_ = ctx.SetVariable("_prompt_lines_count", "2")
				_ = ctx.SetVariable("_prompt_line1", "line1> ")
				_ = ctx.SetVariable("_prompt_line2", "line2> ")
			},
			expectError: false,
		},
		{
			name: "missing lines count",
			setupFunc: func(_ *context.NeuroContext) {
				// Don't set _prompt_lines_count
			},
			expectError:   true,
			errorContains: "not configured",
		},
		{
			name: "invalid lines count format",
			setupFunc: func(ctx *context.NeuroContext) {
				_ = ctx.SetVariable("_prompt_lines_count", "invalid")
				_ = ctx.SetVariable("_prompt_line1", "test> ")
			},
			expectError:   true,
			errorContains: "invalid prompt lines count format",
		},
		{
			name: "lines count too low",
			setupFunc: func(ctx *context.NeuroContext) {
				_ = ctx.SetVariable("_prompt_lines_count", "0")
				_ = ctx.SetVariable("_prompt_line1", "test> ")
			},
			expectError:   true,
			errorContains: "must be between 1 and 5",
		},
		{
			name: "lines count too high",
			setupFunc: func(ctx *context.NeuroContext) {
				_ = ctx.SetVariable("_prompt_lines_count", "6")
				_ = ctx.SetVariable("_prompt_line1", "test> ")
			},
			expectError:   true,
			errorContains: "must be between 1 and 5",
		},
		{
			name: "missing first line",
			setupFunc: func(ctx *context.NeuroContext) {
				_ = ctx.SetVariable("_prompt_lines_count", "1")
				// Don't set _prompt_line1
			},
			expectError:   true,
			errorContains: "first prompt line must be configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test context
			ctx := context.NewTestContext()
			context.SetGlobalContext(ctx)
			defer context.ResetGlobalContext()

			// Apply test setup
			tt.setupFunc(ctx.(*context.NeuroContext))

			service := NewShellPromptService()
			// Don't initialize the service for validation tests - we want to test validation without defaults
			service.initialized = true // Manually set initialized to allow ValidatePromptConfiguration to run

			err := service.ValidatePromptConfiguration()
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

func TestShellPromptService_GetServiceInfo(t *testing.T) {
	service := NewShellPromptService()

	info := service.GetServiceInfo()
	assert.Equal(t, "shell_prompt", info["name"])
	assert.Equal(t, false, info["initialized"])
	assert.Equal(t, "shell_prompt", info["type"])
	assert.Equal(t, "Shell prompt configuration management", info["description"])

	// After initialization
	err := service.Initialize()
	require.NoError(t, err)

	info = service.GetServiceInfo()
	assert.Equal(t, true, info["initialized"])
}
