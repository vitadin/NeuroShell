// Package model contains tests for OpenAI-specific model commands.
package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

func TestOpenAIModelNewCommand_Name(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}
	assert.Equal(t, "openai-model-new", cmd.Name())
}

func TestOpenAIModelNewCommand_ParseMode(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestOpenAIModelNewCommand_Description(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}
	assert.Equal(t, "Create OpenAI model configurations with reasoning support", cmd.Description())
}

func TestOpenAIModelNewCommand_Usage(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "openai-model-new")
	assert.Contains(t, usage, "catalog_id")
	assert.Contains(t, usage, "reasoning_effort")
	assert.Contains(t, usage, "max_output_tokens")
}

func TestOpenAIModelNewCommand_HelpInfo(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "openai-model-new", helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check options
	assert.Greater(t, len(helpInfo.Options), 5) // Should have catalog_id, base_model, reasoning_effort, etc.

	// Check examples
	assert.Greater(t, len(helpInfo.Examples), 3)

	// Check stored variables
	assert.Greater(t, len(helpInfo.StoredVariables), 3)

	// Check notes
	assert.Greater(t, len(helpInfo.Notes), 3)
}

func TestOpenAIModelNewCommand_Execute_BasicFunctionality(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		checkFunc   func(t *testing.T, ctx neurotypes.Context)
	}{
		{
			name: "create basic OpenAI model with catalog_id",
			args: map[string]string{
				"catalog_id": "G4O", // GPT-4o
			},
			input:       "my-gpt4o",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				provider, err := ctx.GetVariable("#model_provider")
				assert.NoError(t, err)
				assert.Equal(t, "openai", provider)

				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4o-2024-11-20", baseModel)

				modelName, err := ctx.GetVariable("#model_name")
				assert.NoError(t, err)
				assert.Equal(t, "my-gpt4o", modelName)
			},
		},
		{
			name: "create model with catalog_id GPT-4",
			args: map[string]string{
				"catalog_id": "G41",
			},
			input:       "my-gpt4",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				provider, err := ctx.GetVariable("#model_provider")
				assert.NoError(t, err)
				assert.Equal(t, "openai", provider)

				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4.1-2025-04-14", baseModel)
			},
		},
		{
			name: "create model with reasoning parameters",
			args: map[string]string{
				"catalog_id":        "O3", // O3 reasoning model
				"reasoning_effort":  "high",
				"max_output_tokens": "50000",
				"reasoning_summary": "auto",
			},
			input:       "reasoning-o3",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				provider, err := ctx.GetVariable("#model_provider")
				assert.NoError(t, err)
				assert.Equal(t, "openai", provider)

				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "o3", baseModel)
			},
		},
		{
			name: "create model with standard parameters",
			args: map[string]string{
				"catalog_id":  "G41",
				"temperature": "0.7",
				"max_tokens":  "2000",
				"top_p":       "0.9",
			},
			input:       "configured-gpt4",
			expectError: false,
		},
		{
			name: "missing model name",
			args: map[string]string{
				"catalog_id": "G41",
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
			ctx := context.New()
			setupOpenAIModelTestRegistry(t, ctx)

			var err error
			outputStr := stringprocessing.CaptureOutput(func() {
				err = cmd.Execute(tt.args, tt.input)
			})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, outputStr, "Created model")

				// Verify model was created
				modelID, err := ctx.GetVariable("#model_id")
				assert.NoError(t, err)
				assert.NotEmpty(t, modelID)

				if tt.checkFunc != nil {
					tt.checkFunc(t, ctx)
				}
			}
		})
	}
}

func TestOpenAIModelNewCommand_Execute_ParameterValidation(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}

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
				"catalog_id":  "G4O",
				"temperature": "2.5",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "temperature must be between 0.0 and 2.0",
		},
		{
			name: "invalid temperature - negative",
			args: map[string]string{
				"catalog_id":  "G4O",
				"temperature": "-0.1",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "temperature must be between 0.0 and 2.0",
		},
		{
			name: "invalid temperature - not a number",
			args: map[string]string{
				"catalog_id":  "G4O",
				"temperature": "not-a-number",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid temperature value",
		},
		{
			name: "invalid max_tokens - not a number",
			args: map[string]string{
				"catalog_id": "G4O",
				"max_tokens": "abc",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid max_tokens value",
		},
		{
			name: "invalid max_tokens - negative",
			args: map[string]string{
				"catalog_id": "G4O",
				"max_tokens": "-100",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "max_tokens must be positive",
		},
		{
			name: "invalid top_p - too high",
			args: map[string]string{
				"catalog_id": "G4O",
				"top_p":      "1.5",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "top_p must be between 0.0 and 1.0",
		},
		{
			name: "invalid reasoning_effort",
			args: map[string]string{
				"catalog_id":       "O3",
				"reasoning_effort": "invalid",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid reasoning_effort value",
		},
		{
			name: "invalid max_output_tokens - negative",
			args: map[string]string{
				"catalog_id":        "O3",
				"max_output_tokens": "-1000",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "max_output_tokens must be positive",
		},
		{
			name: "invalid reasoning_summary",
			args: map[string]string{
				"catalog_id":        "O3",
				"reasoning_summary": "invalid",
			},
			input:       "test-model",
			expectError: true,
			errorMsg:    "invalid reasoning_summary value",
		},
		{
			name: "valid edge case parameters",
			args: map[string]string{
				"catalog_id":  "G4O",
				"temperature": "0.0",
				"top_p":       "1.0",
				"max_tokens":  "1",
			},
			input:       "edge-case-model",
			expectError: false,
		},
		{
			name: "valid reasoning parameters",
			args: map[string]string{
				"catalog_id":        "O3",
				"reasoning_effort":  "low",
				"max_output_tokens": "10000",
				"reasoning_summary": "auto",
			},
			input:       "reasoning-model",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.New()
			setupOpenAIModelTestRegistry(t, ctx)

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

func TestOpenAIModelNewCommand_Execute_CatalogIntegration(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		checkFunc   func(t *testing.T, ctx neurotypes.Context)
	}{
		{
			name: "O3 catalog model",
			args: map[string]string{
				"catalog_id": "O3",
			},
			input:       "my-o3",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				provider, err := ctx.GetVariable("#model_provider")
				assert.NoError(t, err)
				assert.Equal(t, "openai", provider)

				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "o3", baseModel)
			},
		},
		{
			name: "O4M catalog model",
			args: map[string]string{
				"catalog_id": "O4M",
			},
			input:       "my-o4-mini",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "o4-mini", baseModel)
			},
		},
		{
			name: "GPT-4o catalog model",
			args: map[string]string{
				"catalog_id": "G4O",
			},
			input:       "my-gpt4o",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				baseModel, err := ctx.GetVariable("#model_base")
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4o-2024-11-20", baseModel)
			},
		},
		{
			name: "invalid catalog_id",
			args: map[string]string{
				"catalog_id": "INVALID",
			},
			input:       "test-model",
			expectError: true,
		},
		{
			name: "non-OpenAI catalog model",
			args: map[string]string{
				"catalog_id": "CS4", // Claude Sonnet 4
			},
			input:       "test-model",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.New()
			setupOpenAIModelTestRegistry(t, ctx)

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

func TestOpenAIModelNewCommand_Execute_ServiceErrors(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T)
		expectError string
	}{
		{
			name: "model service unavailable",
			setupFunc: func(_ *testing.T) {
				services.SetGlobalRegistry(services.NewRegistry())
			},
			expectError: "model service not available",
		},
		{
			name: "variable service unavailable",
			setupFunc: func(t *testing.T) {
				registry := services.NewRegistry()
				err := registry.RegisterService(services.NewModelService())
				require.NoError(t, err)
				err = registry.InitializeAll()
				require.NoError(t, err)
				services.SetGlobalRegistry(registry)
			},
			expectError: "variable service not available",
		},
		{
			name: "model catalog service unavailable",
			setupFunc: func(t *testing.T) {
				registry := services.NewRegistry()
				err := registry.RegisterService(services.NewModelService())
				require.NoError(t, err)
				err = registry.RegisterService(services.NewVariableService())
				require.NoError(t, err)
				err = registry.InitializeAll()
				require.NoError(t, err)
				services.SetGlobalRegistry(registry)
			},
			expectError: "model catalog service not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original registry
			originalRegistry := services.GetGlobalRegistry()
			defer func() {
				services.SetGlobalRegistry(originalRegistry)
			}()

			tt.setupFunc(t)

			args := map[string]string{"catalog_id": "G4O"}
			err := cmd.Execute(args, "test-model")

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestOpenAIModelNewCommand_Execute_ActivationBehavior(t *testing.T) {
	cmd := &OpenAIModelNewCommand{}
	ctx := context.New()
	setupOpenAIModelTestRegistry(t, ctx)

	// Create a model and verify it gets activated
	args := map[string]string{
		"catalog_id": "G4O",
	}

	var err error
	outputStr := stringprocessing.CaptureOutput(func() {
		err = cmd.Execute(args, "test-activation")
	})

	assert.NoError(t, err)
	assert.Contains(t, outputStr, "Created model")

	// Verify model was created
	modelID, err := ctx.GetVariable("#model_id")
	assert.NoError(t, err)
	assert.NotEmpty(t, modelID)

	// Note: The activation happens via stack service, so in tests we just verify
	// the model creation succeeded. Full activation testing would require
	// integration tests with the stack service.
}

// setupOpenAIModelTestRegistry creates a clean test registry for openai-model-new command tests
func setupOpenAIModelTestRegistry(t *testing.T, ctx neurotypes.Context) {
	// Create a new service registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
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

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}

// Interface compliance check
var _ neurotypes.Command = (*OpenAIModelNewCommand)(nil)
