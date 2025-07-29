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
	assert.Contains(t, usage, "catalog_id=")
	assert.Contains(t, usage, "model_name")
}

func TestNewCommand_HelpInfo(t *testing.T) {
	cmd := &NewCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "model-new", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check that catalog_id option is present and marked as required
	catalogIDFound := false
	for _, opt := range helpInfo.Options {
		if opt.Name == "catalog_id" {
			catalogIDFound = true
			assert.True(t, opt.Required, "catalog_id should be marked as required")
			break
		}
	}
	assert.True(t, catalogIDFound, "catalog_id option should be in help info")

	// Ensure provider and base_model options are not present
	for _, opt := range helpInfo.Options {
		assert.NotEqual(t, "provider", opt.Name, "provider option should not be present")
		assert.NotEqual(t, "base_model", opt.Name, "base_model option should not be present")
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
			name: "create Anthropic model from catalog",
			args: map[string]string{
				"catalog_id": "CS4",
			},
			input:       "claude-work",
			expectError: false,
		},
		{
			name: "create model with description",
			args: map[string]string{
				"catalog_id":  "CS4",
				"description": "Fast model for quick responses",
			},
			input:       "fast-claude",
			expectError: false,
		},
		{
			name: "missing model name",
			args: map[string]string{
				"catalog_id": "CS4",
			},
			input:       "",
			expectError: true,
		},
		{
			name:        "missing catalog_id",
			args:        map[string]string{},
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
			assert.Equal(t, "anthropic", modelProvider) // All test cases use CS4 catalog_id

			modelBase, err := ctx.GetVariable("#model_base")
			assert.NoError(t, err)
			assert.Equal(t, "claude-sonnet-4-20250514", modelBase) // CS4 base model

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
			name: "valid parameters with catalog_id",
			args: map[string]string{
				"catalog_id":  "CS4",
				"temperature": "0.7",
				"max_tokens":  "1000",
			},
			input:       "parameter-test-model",
			expectError: false,
		},
		{
			name: "invalid catalog_id",
			args: map[string]string{
				"catalog_id": "INVALID_ID",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "failed to find model with catalog_id",
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
		"catalog_id": "CS4",
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
		"catalog_id": "CS4",
	}

	// Create first model
	err := cmd.Execute(baseArgs, "duplicate-test")
	assert.NoError(t, err)

	// Try to create second model with same name
	err = cmd.Execute(baseArgs, "duplicate-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model name 'duplicate-test' already exists")
}

// TestNewCommand_Execute_VariableInterpolation removed - interpolation is now handled by state machine

func TestNewCommand_Execute_CustomParameters(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupModelTestRegistry(t, ctx)

	// Test with custom provider-specific parameters using catalog_id
	args := map[string]string{
		"catalog_id":       "CS4",
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
		"catalog_id": "CS4",
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
				"catalog_id":  "CS4",
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
				"catalog_id":  "CS4",
				"temperature": "1.0",
			},
			input:       "max-temp-model",
			expectError: false,
		},
		{
			name: "model with empty description",
			args: map[string]string{
				"catalog_id":  "CS4",
				"description": "",
			},
			input:       "empty-desc-model",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				// Should not set description variable if empty
				value, err := ctx.GetVariable("#model_description")
				assert.NoError(t, err)
				assert.Equal(t, "", value) // Variable should return empty string if not set
			},
		},
		{
			name: "model with only required parameters",
			args: map[string]string{
				"catalog_id": "CS4",
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

func TestNewCommand_Execute_CatalogID(t *testing.T) {
	cmd := &NewCommand{}
	ctx := context.New()
	setupModelTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		errorMsg    string
		checkFunc   func(t *testing.T, ctx neurotypes.Context)
	}{
		{
			name: "create model from catalog ID - CS4",
			args: map[string]string{
				"catalog_id": "CS4",
			},
			input:       "my-claude",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				provider, err := ctx.GetVariable("#model_provider")
				assert.NoError(t, err)
				assert.Equal(t, "anthropic", provider)

				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "claude-sonnet-4-20250514", baseModel)
			},
		},
		// OpenAI O3 catalog test case removed - now covered in openai_model_new_test.go
		{
			name: "create model from catalog ID with case insensitive - cs4",
			args: map[string]string{
				"catalog_id": "cs4",
			},
			input:       "my-claude-lower",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				provider, err := ctx.GetVariable("#model_provider")
				assert.NoError(t, err)
				assert.Equal(t, "anthropic", provider)

				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "claude-sonnet-4-20250514", baseModel)
			},
		},
		// OpenAI O3 with parameters test case removed - now covered in openai_model_new_test.go
		{
			name: "catalog_id overrides manual provider/base_model",
			args: map[string]string{
				"catalog_id": "CS4",
				"provider":   "openai", // Should be ignored
				"base_model": "gpt-4",  // Should be ignored
			},
			input:       "override-test",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				provider, err := ctx.GetVariable("#model_provider")
				assert.NoError(t, err)
				assert.Equal(t, "anthropic", provider) // From catalog, not "openai"

				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "claude-sonnet-4-20250514", baseModel) // From catalog, not "gpt-4"
			},
		},
		{
			name: "invalid catalog ID",
			args: map[string]string{
				"catalog_id": "INVALID123",
			},
			input:       "invalid-model",
			expectError: true,
			errorMsg:    "not found in catalog",
		},
		{
			name: "empty catalog ID",
			args: map[string]string{
				"catalog_id": "",
			},
			input:       "empty-catalog-id",
			expectError: true, // catalog_id is now required
			errorMsg:    "catalog_id is required",
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
				return
			}

			assert.NoError(t, err)

			// Verify model was created
			modelID, err := ctx.GetVariable("#model_id")
			assert.NoError(t, err)
			assert.NotEmpty(t, modelID)

			modelName, err := ctx.GetVariable("#model_name")
			assert.NoError(t, err)
			assert.Equal(t, tt.input, modelName)

			// Check output variable
			output, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "Created model")
			assert.Contains(t, output, tt.input)

			// Run custom checks if provided
			if tt.checkFunc != nil {
				tt.checkFunc(t, ctx)
			}
		})
	}
}

func TestNewCommand_Execute_CatalogIDEdgeCases(t *testing.T) {
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
			name: "case insensitive catalog IDs - all variations",
			args: map[string]string{
				"catalog_id": "co37", // Claude Opus 3.7 in lowercase
			},
			input:       "opus-lowercase",
			expectError: false,
		},
		{
			name: "mixed case catalog IDs",
			args: map[string]string{
				"catalog_id": "Cs4", // Mixed case
			},
			input:       "claude-mixed-case",
			expectError: false,
		},
		{
			name: "catalog_id with variable interpolation",
			args: map[string]string{
				"catalog_id": "${model_id_var}",
			},
			input:       "interpolated-model",
			expectError: true, // Will fail because variable doesn't exist
			errorMsg:    "not found in catalog",
		},
		// OpenAI O4M catalog test case removed - now covered in openai_model_new_test.go
	}

	// Set up variable for interpolation test
	require.NoError(t, ctx.SetVariable("model_id_var", "CS4"))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Don't reset context for interpolation test
			if !strings.Contains(tt.name, "interpolation") {
				ctx = context.New()
				setupModelTestRegistry(t, ctx)
			}

			err := cmd.Execute(tt.args, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify model was created
				modelID, err := ctx.GetVariable("#model_id")
				assert.NoError(t, err)
				assert.NotEmpty(t, modelID)
			}
		})
	}
}

// setupModelTestRegistry sets up a test environment with required services for model commands
func setupModelTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewModelService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewModelCatalogService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewStackService())
	require.NoError(t, err)

	// Note: InterpolationService removed - state machine handles interpolation

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

// Interface compliance check
var _ neurotypes.Command = (*NewCommand)(nil)
