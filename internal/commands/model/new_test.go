package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestNewCommand_Name(t *testing.T) {
	cmd := &NewCommand{}
	assert.Equal(t, "model-new", cmd.Name())
}

func TestNewCommand_ParseMode(t *testing.T) {
	cmd := &NewCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestNewCommand_Description(t *testing.T) {
	cmd := &NewCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "model")
	assert.Contains(t, strings.ToLower(desc), "configuration")
}

func TestNewCommand_Usage(t *testing.T) {
	cmd := &NewCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\model-new")
	assert.Contains(t, usage, "provider=")
	assert.Contains(t, usage, "base_model=")
	assert.Contains(t, usage, "model_name")
}

func TestNewCommand_HelpInfo(t *testing.T) {
	cmd := &NewCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "model-new", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check that required options are present
	requiredOptions := []string{"provider", "base_model"}
	for _, reqOpt := range requiredOptions {
		found := false
		for _, opt := range helpInfo.Options {
			if opt.Name == reqOpt && opt.Required {
				found = true
				break
			}
		}
		assert.True(t, found, "Required option %s should be in help info", reqOpt)
	}

	// Check that examples are provided
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)
}

func TestNewCommand_Execute_BasicFunctionality(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupModelTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
	}{
		{
			name: "create basic OpenAI model",
			args: map[string]string{
				"provider":   "openai",
				"base_model": "gpt-4",
			},
			input:       "my-gpt4",
			expectError: false,
		},
		{
			name: "create Anthropic model",
			args: map[string]string{
				"provider":   "anthropic",
				"base_model": "claude-3-sonnet",
			},
			input:       "claude-work",
			expectError: false,
		},
		{
			name: "create model with temperature",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-3.5-turbo",
				"temperature": "0.7",
			},
			input:       "creative-gpt",
			expectError: false,
		},
		{
			name: "create model with multiple parameters",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"temperature": "0.8",
				"max_tokens":  "2000",
				"top_p":       "0.9",
			},
			input:       "full-config-model",
			expectError: false,
		},
		{
			name: "create model with description",
			args: map[string]string{
				"provider":    "anthropic",
				"base_model":  "claude-3-haiku",
				"description": "Fast model for quick responses",
			},
			input:       "fast-claude",
			expectError: false,
		},
		{
			name: "missing model name",
			args: map[string]string{
				"provider":   "openai",
				"base_model": "gpt-4",
			},
			input:       "",
			expectError: true,
		},
		{
			name: "missing provider",
			args: map[string]string{
				"base_model": "gpt-4",
			},
			input:       "test-model",
			expectError: true,
		},
		{
			name: "missing base_model",
			args: map[string]string{
				"provider": "openai",
			},
			input:       "test-model",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupModelTestRegistry(t, ctx)

			err := cmd.Execute(tt.args, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify model was created by checking system variables
			modelID, err := ctx.GetVariable("#model_id")
			assert.NoError(t, err)
			assert.NotEmpty(t, modelID)

			modelName, err := ctx.GetVariable("#model_name")
			assert.NoError(t, err)
			assert.Equal(t, tt.input, modelName)

			modelProvider, err := ctx.GetVariable("#model_provider")
			assert.NoError(t, err)
			assert.Equal(t, tt.args["provider"], modelProvider)

			modelBase, err := ctx.GetVariable("#model_base")
			assert.NoError(t, err)
			assert.Equal(t, tt.args["base_model"], modelBase)

			// Check parameter count
			paramCount, err := ctx.GetVariable("#model_param_count")
			assert.NoError(t, err)
			assert.NotEmpty(t, paramCount)

			// Check creation timestamp
			modelCreated, err := ctx.GetVariable("#model_created")
			assert.NoError(t, err)
			assert.NotEmpty(t, modelCreated)

			// Check output variable
			output, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "Created model")
			assert.Contains(t, output, tt.input)
		})
	}
}

func TestNewCommand_Execute_ParameterValidation(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupModelTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid temperature - too high",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"temperature": "1.5",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "temperature must be between 0.0 and 1.0",
		},
		{
			name: "invalid temperature - negative",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"temperature": "-0.1",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "temperature must be between 0.0 and 1.0",
		},
		{
			name: "invalid temperature - not a number",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"temperature": "not-a-number",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid temperature value",
		},
		{
			name: "invalid max_tokens - not a number",
			args: map[string]string{
				"provider":   "openai",
				"base_model": "gpt-4",
				"max_tokens": "abc",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid max_tokens value",
		},
		{
			name: "invalid max_tokens - negative",
			args: map[string]string{
				"provider":   "openai",
				"base_model": "gpt-4",
				"max_tokens": "-100",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "max_tokens must be positive",
		},
		{
			name: "invalid top_p - too high",
			args: map[string]string{
				"provider":   "openai",
				"base_model": "gpt-4",
				"top_p":      "1.5",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "top_p must be between 0.0 and 1.0",
		},
		{
			name: "invalid top_k - not a number",
			args: map[string]string{
				"provider":   "openai",
				"base_model": "gpt-4",
				"top_k":      "not-a-number",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid top_k value",
		},
		{
			name: "valid edge case parameters",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"temperature": "0.0",
				"top_p":       "1.0",
				"max_tokens":  "1",
			},
			input:       "edge-case-model",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupModelTestRegistry(t, ctx)

			err := cmd.Execute(tt.args, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewCommand_Execute_ModelNameValidation(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupModelTestRegistry(t, ctx)

	tests := []struct {
		name        string
		modelName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty model name",
			modelName:   "",
			expectError: true,
			errorMsg:    "model name is required",
		},
		{
			name:        "model name with spaces",
			modelName:   "my test model",
			expectError: true,
			errorMsg:    "model name cannot contain spaces",
		},
		{
			name:        "model name with tabs",
			modelName:   "my\tmodel",
			expectError: true,
			errorMsg:    "model name cannot contain newlines or tabs",
		},
		{
			name:        "model name with newlines",
			modelName:   "my\nmodel",
			expectError: true,
			errorMsg:    "model name cannot contain newlines or tabs",
		},
		{
			name:        "very long model name",
			modelName:   strings.Repeat("a", 101),
			expectError: true,
			errorMsg:    "model name cannot exceed 100 characters",
		},
		{
			name:        "valid model name with hyphens",
			modelName:   "my-test-model",
			expectError: false,
		},
		{
			name:        "valid model name with underscores",
			modelName:   "my_test_model",
			expectError: false,
		},
		{
			name:        "valid model name with numbers",
			modelName:   "model123",
			expectError: false,
		},
		{
			name:        "valid single character name",
			modelName:   "a",
			expectError: false,
		},
		{
			name:        "valid model name at length limit",
			modelName:   strings.Repeat("a", 100),
			expectError: false,
		},
	}

	baseArgs := map[string]string{
		"provider":   "openai",
		"base_model": "gpt-4",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupModelTestRegistry(t, ctx)

			err := cmd.Execute(baseArgs, tt.modelName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewCommand_Execute_DuplicateModelNames(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupModelTestRegistry(t, ctx)

	baseArgs := map[string]string{
		"provider":   "openai",
		"base_model": "gpt-4",
	}

	// Create first model
	err := cmd.Execute(baseArgs, "duplicate-test")
	assert.NoError(t, err)

	// Try to create second model with same name
	err = cmd.Execute(baseArgs, "duplicate-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model name 'duplicate-test' already exists")
}

func TestNewCommand_Execute_VariableInterpolation(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupModelTestRegistry(t, ctx)

	// Set up test variables
	require.NoError(t, ctx.SetVariable("model_prefix", "test"))
	require.NoError(t, ctx.SetVariable("provider_name", "openai"))
	require.NoError(t, ctx.SetVariable("model_desc", "Test model for experiments"))

	// Create model with variable interpolation
	args := map[string]string{
		"provider":    "${provider_name}",
		"base_model":  "gpt-4",
		"description": "${model_desc}",
	}
	input := "${model_prefix}-model"

	err := cmd.Execute(args, input)
	assert.NoError(t, err)

	// Check that variables were interpolated
	modelName, err := ctx.GetVariable("#model_name")
	assert.NoError(t, err)
	assert.Equal(t, "test-model", modelName)

	modelProvider, err := ctx.GetVariable("#model_provider")
	assert.NoError(t, err)
	assert.Equal(t, "openai", modelProvider)

	modelDesc, err := ctx.GetVariable("#model_description")
	assert.NoError(t, err)
	assert.Equal(t, "Test model for experiments", modelDesc)
}

func TestNewCommand_Execute_CustomParameters(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupModelTestRegistry(t, ctx)

	// Test with custom provider-specific parameters
	args := map[string]string{
		"provider":         "custom",
		"base_model":       "custom-llm",
		"temperature":      "0.7",
		"max_tokens":       "1500",
		"custom_param1":    "value1",
		"custom_param2":    "value2",
		"presence_penalty": "0.5",
	}

	err := cmd.Execute(args, "custom-model")
	assert.NoError(t, err)

	// Verify model was created
	modelID, err := ctx.GetVariable("#model_id")
	assert.NoError(t, err)
	assert.NotEmpty(t, modelID)

	// Verify parameter count includes custom parameters
	paramCount, err := ctx.GetVariable("#model_param_count")
	assert.NoError(t, err)
	assert.Equal(t, "5", paramCount) // temperature, max_tokens, custom_param1, custom_param2, presence_penalty
}

func TestNewCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &NewCommand{}

	// Don't setup services - should fail
	args := map[string]string{
		"provider":   "openai",
		"base_model": "gpt-4",
	}

	err := cmd.Execute(args, "test-model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not available")
}

func TestNewCommand_Execute_EdgeCases(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupModelTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		checkFunc   func(t *testing.T, ctx neurotypes.Context)
	}{
		{
			name: "model with zero temperature",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"temperature": "0",
			},
			input:       "zero-temp-model",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				paramCount, err := ctx.GetVariable("#model_param_count")
				assert.NoError(t, err)
				assert.Equal(t, "1", paramCount)
			},
		},
		{
			name: "model with maximum valid temperature",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"temperature": "1.0",
			},
			input:       "max-temp-model",
			expectError: false,
		},
		{
			name: "model with empty description",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"description": "",
			},
			input:       "empty-desc-model",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				// Should not set description variable if empty
				_, err := ctx.GetVariable("#model_description")
				assert.Error(t, err) // Variable should not exist
			},
		},
		{
			name: "model with only required parameters",
			args: map[string]string{
				"provider":   "local",
				"base_model": "llama-2",
			},
			input:       "minimal-model",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				paramCount, err := ctx.GetVariable("#model_param_count")
				assert.NoError(t, err)
				assert.Equal(t, "0", paramCount) // No optional parameters
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupModelTestRegistry(t, ctx)

			err := cmd.Execute(tt.args, tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, ctx)
				}
			}
		})
	}
}

// setupModelTestRegistry sets up a test environment with required services
func setupModelTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewInterpolationService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewModelService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll(ctx)
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

// Interface compliance check
var _ neurotypes.Command = (*NewCommand)(nil)
