package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
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

	// Check that essential options are present (but not necessarily required since from_id is alternative)
	essentialOptions := []string{"provider", "base_model", "from_id"}
	for _, essOpt := range essentialOptions {
		found := false
		for _, opt := range helpInfo.Options {
			if opt.Name == essOpt {
				found = true
				break
			}
		}
		assert.True(t, found, "Essential option %s should be in help info", essOpt)
	}

	// Check that examples are provided
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)
}

func TestNewCommand_Execute_BasicFunctionality(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupTestServices(ctx)
	defer cleanupTestServices()

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
			setupTestServices(ctx)
			defer cleanupTestServices()

			err := cmd.Execute(tt.args, tt.input, ctx)

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
	setupTestServices(ctx)
	defer cleanupTestServices()

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
			setupTestServices(ctx)
			defer cleanupTestServices()

			err := cmd.Execute(tt.args, tt.input, ctx)

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
	setupTestServices(ctx)
	defer cleanupTestServices()

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
			setupTestServices(ctx)
			defer cleanupTestServices()

			err := cmd.Execute(baseArgs, tt.modelName, ctx)

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
	setupTestServices(ctx)
	defer cleanupTestServices()

	baseArgs := map[string]string{
		"provider":   "openai",
		"base_model": "gpt-4",
	}

	// Create first model
	err := cmd.Execute(baseArgs, "duplicate-test", ctx)
	assert.NoError(t, err)

	// Try to create second model with same name
	err = cmd.Execute(baseArgs, "duplicate-test", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model name 'duplicate-test' already exists")
}

func TestNewCommand_Execute_VariableInterpolation(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupTestServices(ctx)
	defer cleanupTestServices()

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

	err := cmd.Execute(args, input, ctx)
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
	setupTestServices(ctx)
	defer cleanupTestServices()

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

	err := cmd.Execute(args, "custom-model", ctx)
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
	ctx := context.New()

	// Don't setup services - should fail
	args := map[string]string{
		"provider":   "openai",
		"base_model": "gpt-4",
	}

	err := cmd.Execute(args, "test-model", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not available")
}

func TestNewCommand_Execute_EdgeCases(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupTestServices(ctx)
	defer cleanupTestServices()

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
			setupTestServices(ctx)
			defer cleanupTestServices()

			err := cmd.Execute(tt.args, tt.input, ctx)

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

func TestNewCommand_Execute_FromExisting(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupTestServices(ctx)
	defer cleanupTestServices()

	// First create a base model to clone from
	baseArgs := map[string]string{
		"provider":    "openai",
		"base_model":  "gpt-4",
		"temperature": "0.7",
		"max_tokens":  "1000",
		"description": "Base model for cloning",
	}
	err := cmd.Execute(baseArgs, "base-model", ctx)
	require.NoError(t, err)

	// Get the created model ID
	baseModelID, err := ctx.GetVariable("#model_id")
	require.NoError(t, err)
	require.NotEmpty(t, baseModelID)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		checkFunc   func(t *testing.T, ctx neurotypes.Context)
	}{
		{
			name: "clone basic model",
			args: map[string]string{
				"from_id": baseModelID,
			},
			input:       "cloned-model",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				// Check that cloned model was created
				clonedModelID, err := ctx.GetVariable("#model_id")
				assert.NoError(t, err)
				assert.NotEmpty(t, clonedModelID)
				assert.NotEqual(t, baseModelID, clonedModelID) // Should be different IDs

				// Check model name
				modelName, err := ctx.GetVariable("#model_name")
				assert.NoError(t, err)
				assert.Equal(t, "cloned-model", modelName)

				// Check provider inheritance
				provider, err := ctx.GetVariable("#model_provider")
				assert.NoError(t, err)
				assert.Equal(t, "openai", provider)

				// Check base model inheritance
				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", baseModel)
			},
		},
		{
			name: "clone with parameter overrides",
			args: map[string]string{
				"from_id":     baseModelID,
				"temperature": "0.9",
				"max_tokens":  "2000",
				"description": "Custom cloned model",
			},
			input:       "custom-clone",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				// Verify model was created
				modelID, err := ctx.GetVariable("#model_id")
				assert.NoError(t, err)
				assert.NotEmpty(t, modelID)

				// Check description override
				description, err := ctx.GetVariable("#model_description")
				assert.NoError(t, err)
				assert.Equal(t, "Custom cloned model", description)

				// Verify parameter count includes overrides
				paramCount, err := ctx.GetVariable("#model_param_count")
				assert.NoError(t, err)
				// Should have temperature, max_tokens from overrides
				assert.True(t, paramCount != "0")
			},
		},
		{
			name: "clone with variable interpolation in from_id",
			args: map[string]string{
				"from_id": "${#model_id}", // Should reference the last created model
			},
			input:       "interpolated-clone",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				modelName, err := ctx.GetVariable("#model_name")
				assert.NoError(t, err)
				assert.Equal(t, "interpolated-clone", modelName)
			},
		},
		{
			name: "clone nonexistent model",
			args: map[string]string{
				"from_id": "nonexistent-model-id",
			},
			input:       "should-fail",
			expectError: true,
		},
		{
			name: "clone with conflicting provider options",
			args: map[string]string{
				"from_id":    baseModelID,
				"provider":   "anthropic", // Should conflict with from_id
				"base_model": "claude-3-sonnet",
			},
			input:       "conflict-model",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input, ctx)

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

func TestNewCommand_Execute_MethodValidation(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupTestServices(ctx)
	defer cleanupTestServices()

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing provider and from_id",
			args: map[string]string{
				"base_model": "gpt-4",
			},
			input:       "invalid-model",
			expectError: true,
			errorMsg:    "provider is required",
		},
		{
			name: "missing base_model with provider",
			args: map[string]string{
				"provider": "openai",
			},
			input:       "invalid-model",
			expectError: true,
			errorMsg:    "base_model is required",
		},
		{
			name: "empty model name",
			args: map[string]string{
				"provider":   "openai",
				"base_model": "gpt-4",
			},
			input:       "",
			expectError: true,
			errorMsg:    "model name is required",
		},
		{
			name: "invalid temperature",
			args: map[string]string{
				"provider":    "openai",
				"base_model":  "gpt-4",
				"temperature": "invalid",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid temperature",
		},
		{
			name: "invalid max_tokens",
			args: map[string]string{
				"provider":   "openai",
				"base_model": "gpt-4",
				"max_tokens": "invalid",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid max_tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input, ctx)

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

func TestNewCommand_UpdatedUsage(t *testing.T) {
	cmd := &NewCommand{}
	usage := cmd.Usage()

	// Verify from_id option is documented
	assert.Contains(t, usage, "from_id")
	assert.Contains(t, usage, "Method 1 - From Provider/Base Model")
	assert.Contains(t, usage, "Method 2 - From Existing Model")

	// Verify both creation methods are shown in examples
	assert.Contains(t, usage, "from_id=${_catalog_model_id}")
	assert.Contains(t, usage, "from_id=existing-model")
}

func TestNewCommand_UpdatedHelpInfo(t *testing.T) {
	cmd := &NewCommand{}
	helpInfo := cmd.HelpInfo()

	// Check that from_id option is present
	hasFromIDOption := false
	for _, option := range helpInfo.Options {
		if option.Name == "from_id" {
			hasFromIDOption = true
			assert.Contains(t, option.Description, "clone")
			assert.Equal(t, "string", option.Type)
			assert.False(t, option.Required) // Should be optional
			break
		}
	}
	assert.True(t, hasFromIDOption, "from_id option should be present in help info")

	// Check examples include from_id usage
	hasFromIDExample := false
	for _, example := range helpInfo.Examples {
		if strings.Contains(example.Command, "from_id") {
			hasFromIDExample = true
			break
		}
	}
	assert.True(t, hasFromIDExample, "should have example showing from_id usage")

	// Check notes mention the conflict restriction
	hasConflictNote := false
	for _, note := range helpInfo.Notes {
		if strings.Contains(note, "provider+base_model OR from_id") {
			hasConflictNote = true
			break
		}
	}
	assert.True(t, hasConflictNote, "should have note about conflicting options")
}

// Interface compliance check
var _ neurotypes.Command = (*NewCommand)(nil)
