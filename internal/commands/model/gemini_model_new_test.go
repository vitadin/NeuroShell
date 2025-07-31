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

func TestGeminiModelNewCommand_Name(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	assert.Equal(t, "gemini-model-new", cmd.Name())
}

func TestGeminiModelNewCommand_ParseMode(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestGeminiModelNewCommand_Description(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, strings.ToLower(desc), "gemini")
	assert.Contains(t, strings.ToLower(desc), "model")
	assert.Contains(t, strings.ToLower(desc), "thinking")
}

func TestGeminiModelNewCommand_Usage(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\gemini-model-new")
	assert.Contains(t, usage, "catalog_id=")
	assert.Contains(t, usage, "thinking_budget=")
	assert.Contains(t, usage, "model_name")
	assert.Contains(t, usage, "GM25F")
	assert.Contains(t, usage, "GM25P")
}

func TestGeminiModelNewCommand_HelpInfo(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "gemini-model-new", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check that catalog_id option is present and marked as required
	catalogIDFound := false
	thinkingBudgetFound := false
	for _, opt := range helpInfo.Options {
		if opt.Name == "catalog_id" {
			catalogIDFound = true
			assert.True(t, opt.Required, "catalog_id should be marked as required")
		}
		if opt.Name == "thinking_budget" {
			thinkingBudgetFound = true
			assert.False(t, opt.Required, "thinking_budget should be optional")
			assert.Equal(t, "-1", opt.Default, "thinking_budget should default to -1")
		}
	}
	assert.True(t, catalogIDFound, "catalog_id option should be in help info")
	assert.True(t, thinkingBudgetFound, "thinking_budget option should be in help info")

	// Check that examples are provided
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Check stored variables
	assert.NotEmpty(t, helpInfo.StoredVariables)
	foundModelID := false
	foundActiveProvider := false
	for _, variable := range helpInfo.StoredVariables {
		if variable.Name == "_model_id" {
			foundModelID = true
		}
		if variable.Name == "#active_model_provider" {
			foundActiveProvider = true
			assert.Contains(t, variable.Example, "gemini")
		}
	}
	assert.True(t, foundModelID, "_model_id variable should be documented")
	assert.True(t, foundActiveProvider, "#active_model_provider variable should be documented")
}

func TestGeminiModelNewCommand_Execute_BasicFunctionality(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	ctx := context.New()

	// Setup test registry with required services
	setupGeminiModelTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
	}{
		{
			name: "create Gemini Flash model from catalog",
			args: map[string]string{
				"catalog_id": "GM25F",
			},
			input:       "gemini-flash",
			expectError: false,
		},
		{
			name: "create Gemini Pro model from catalog",
			args: map[string]string{
				"catalog_id": "GM25P",
			},
			input:       "gemini-pro",
			expectError: false,
		},
		{
			name: "create model with description",
			args: map[string]string{
				"catalog_id":  "GM25F",
				"description": "Fast Gemini model for quick responses",
			},
			input:       "fast-gemini",
			expectError: false,
		},
		{
			name: "missing model name",
			args: map[string]string{
				"catalog_id": "GM25F",
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
		{
			name: "non-Gemini catalog_id should fail",
			args: map[string]string{
				"catalog_id": "CS4", // Anthropic model
			},
			input:       "wrong-provider",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupGeminiModelTestRegistry(t, ctx)

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
			assert.Equal(t, "gemini", modelProvider)

			// Note: Active model variables are set by \model-activate command pushed to stack

			// Check output variable
			output, err := ctx.GetVariable("_output")
			assert.NoError(t, err)
			assert.Contains(t, output, "Created model")
			assert.Contains(t, output, tt.input)
			assert.Contains(t, output, "gemini")
		})
	}
}

func TestGeminiModelNewCommand_Execute_ThinkingBudgetValidation(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	ctx := context.New()
	setupGeminiModelTestRegistry(t, ctx)

	tests := []struct {
		name           string
		catalogID      string
		thinkingBudget string
		expectError    bool
		errorMsg       string
	}{
		// Gemini Flash (GM25F) tests - range 0-24576, can disable
		{
			name:           "Flash - valid positive thinking_budget",
			catalogID:      "GM25F",
			thinkingBudget: "2048",
			expectError:    false,
		},
		{
			name:           "Flash - valid minimum thinking_budget",
			catalogID:      "GM25F",
			thinkingBudget: "0",
			expectError:    false,
		},
		{
			name:           "Flash - valid maximum thinking_budget",
			catalogID:      "GM25F",
			thinkingBudget: "24576",
			expectError:    false,
		},
		{
			name:           "Flash - dynamic thinking_budget",
			catalogID:      "GM25F",
			thinkingBudget: "-1",
			expectError:    false,
		},
		{
			name:           "Flash - thinking_budget above range",
			catalogID:      "GM25F",
			thinkingBudget: "30000",
			expectError:    true,
			errorMsg:       "outside valid range",
		},
		{
			name:           "Flash - negative thinking_budget (not -1)",
			catalogID:      "GM25F",
			thinkingBudget: "-5",
			expectError:    true,
			errorMsg:       "outside valid range",
		},
		// Gemini Pro (GM25P) tests - range 128-32768, cannot disable
		{
			name:           "Pro - valid thinking_budget",
			catalogID:      "GM25P",
			thinkingBudget: "16384",
			expectError:    false,
		},
		{
			name:           "Pro - valid minimum thinking_budget",
			catalogID:      "GM25P",
			thinkingBudget: "128",
			expectError:    false,
		},
		{
			name:           "Pro - valid maximum thinking_budget",
			catalogID:      "GM25P",
			thinkingBudget: "32768",
			expectError:    false,
		},
		{
			name:           "Pro - dynamic thinking_budget",
			catalogID:      "GM25P",
			thinkingBudget: "-1",
			expectError:    false,
		},
		{
			name:           "Pro - cannot disable thinking_budget",
			catalogID:      "GM25P",
			thinkingBudget: "0",
			expectError:    true,
			errorMsg:       "thinking cannot be disabled",
		},
		{
			name:           "Pro - thinking_budget below minimum",
			catalogID:      "GM25P",
			thinkingBudget: "100",
			expectError:    true,
			errorMsg:       "outside valid range",
		},
		{
			name:           "Pro - thinking_budget above maximum",
			catalogID:      "GM25P",
			thinkingBudget: "40000",
			expectError:    true,
			errorMsg:       "outside valid range",
		},
		// Invalid thinking_budget format
		{
			name:           "Flash - invalid thinking_budget format",
			catalogID:      "GM25F",
			thinkingBudget: "abc",
			expectError:    true,
			errorMsg:       "invalid thinking_budget value",
		},
		{
			name:           "Flash - empty thinking_budget",
			catalogID:      "GM25F",
			thinkingBudget: "",
			expectError:    true,
			errorMsg:       "invalid thinking_budget value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupGeminiModelTestRegistry(t, ctx)

			args := map[string]string{
				"catalog_id":      tt.catalogID,
				"thinking_budget": tt.thinkingBudget,
			}

			err := cmd.Execute(args, "test-thinking-model")

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

func TestGeminiModelNewCommand_Execute_StandardParameters(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	ctx := context.New()
	setupGeminiModelTestRegistry(t, ctx)

	tests := []struct {
		name        string
		args        map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid temperature",
			args: map[string]string{
				"catalog_id":  "GM25F",
				"temperature": "0.7",
			},
			expectError: false,
		},
		{
			name: "temperature at minimum",
			args: map[string]string{
				"catalog_id":  "GM25F",
				"temperature": "0.0",
			},
			expectError: false,
		},
		{
			name: "temperature at maximum",
			args: map[string]string{
				"catalog_id":  "GM25F",
				"temperature": "1.0",
			},
			expectError: false,
		},
		{
			name: "temperature above maximum",
			args: map[string]string{
				"catalog_id":  "GM25F",
				"temperature": "1.5",
			},
			expectError: true,
			errorMsg:    "temperature must be between 0.0 and 1.0",
		},
		{
			name: "temperature below minimum",
			args: map[string]string{
				"catalog_id":  "GM25F",
				"temperature": "-0.1",
			},
			expectError: true,
			errorMsg:    "temperature must be between 0.0 and 1.0",
		},
		{
			name: "valid max_tokens",
			args: map[string]string{
				"catalog_id": "GM25F",
				"max_tokens": "2000",
			},
			expectError: false,
		},
		{
			name: "zero max_tokens",
			args: map[string]string{
				"catalog_id": "GM25F",
				"max_tokens": "0",
			},
			expectError: true,
			errorMsg:    "max_tokens must be positive",
		},
		{
			name: "negative max_tokens",
			args: map[string]string{
				"catalog_id": "GM25F",
				"max_tokens": "-100",
			},
			expectError: true,
			errorMsg:    "max_tokens must be positive",
		},
		{
			name: "valid top_p",
			args: map[string]string{
				"catalog_id": "GM25F",
				"top_p":      "0.9",
			},
			expectError: false,
		},
		{
			name: "top_p above maximum",
			args: map[string]string{
				"catalog_id": "GM25F",
				"top_p":      "1.1",
			},
			expectError: true,
			errorMsg:    "top_p must be between 0.0 and 1.0",
		},
		{
			name: "valid top_k",
			args: map[string]string{
				"catalog_id": "GM25F",
				"top_k":      "40",
			},
			expectError: false,
		},
		{
			name: "zero top_k",
			args: map[string]string{
				"catalog_id": "GM25F",
				"top_k":      "0",
			},
			expectError: true,
			errorMsg:    "top_k must be positive",
		},
		{
			name: "valid presence_penalty",
			args: map[string]string{
				"catalog_id":       "GM25F",
				"presence_penalty": "1.0",
			},
			expectError: false,
		},
		{
			name: "presence_penalty above maximum",
			args: map[string]string{
				"catalog_id":       "GM25F",
				"presence_penalty": "2.1",
			},
			expectError: true,
			errorMsg:    "presence_penalty must be between -2.0 and 2.0",
		},
		{
			name: "valid frequency_penalty",
			args: map[string]string{
				"catalog_id":        "GM25F",
				"frequency_penalty": "-1.5",
			},
			expectError: false,
		},
		{
			name: "frequency_penalty below minimum",
			args: map[string]string{
				"catalog_id":        "GM25F",
				"frequency_penalty": "-2.1",
			},
			expectError: true,
			errorMsg:    "frequency_penalty must be between -2.0 and 2.0",
		},
		{
			name: "multiple valid parameters",
			args: map[string]string{
				"catalog_id":        "GM25F",
				"temperature":       "0.8",
				"max_tokens":        "1500",
				"top_p":             "0.95",
				"top_k":             "50",
				"thinking_budget":   "4096",
				"presence_penalty":  "0.5",
				"frequency_penalty": "-0.5",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupGeminiModelTestRegistry(t, ctx)

			err := cmd.Execute(tt.args, "test-params-model")

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

				// Check parameter count (excluding catalog_id and description)
				paramCount, err := ctx.GetVariable("#model_param_count")
				assert.NoError(t, err)
				assert.NotEqual(t, "0", paramCount)
			}
		})
	}
}

func TestGeminiModelNewCommand_Execute_CustomParameters(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	ctx := context.New()
	setupGeminiModelTestRegistry(t, ctx)

	// Test with custom provider-specific parameters
	args := map[string]string{
		"catalog_id":        "GM25F",
		"temperature":       "0.7",
		"thinking_budget":   "2048",
		"custom_param1":     "value1",
		"custom_param2":     "value2",
		"google_search_api": "enabled",
	}

	err := cmd.Execute(args, "custom-gemini-model")
	assert.NoError(t, err)

	// Verify model was created
	modelID, err := ctx.GetVariable("#model_id")
	assert.NoError(t, err)
	assert.NotEmpty(t, modelID)

	// Verify parameter count includes custom parameters
	paramCount, err := ctx.GetVariable("#model_param_count")
	assert.NoError(t, err)
	assert.Equal(t, "5", paramCount) // temperature, thinking_budget, custom_param1, custom_param2, google_search_api
}

func TestGeminiModelNewCommand_Execute_EdgeCases(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	ctx := context.New()
	setupGeminiModelTestRegistry(t, ctx)

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
				"catalog_id":  "GM25F",
				"temperature": "0",
			},
			input:       "zero-temp-gemini",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				paramCount, err := ctx.GetVariable("#model_param_count")
				assert.NoError(t, err)
				assert.Equal(t, "1", paramCount)
			},
		},
		{
			name: "model with thinking disabled",
			args: map[string]string{
				"catalog_id":      "GM25F", // Flash can disable thinking
				"thinking_budget": "0",
			},
			input:       "no-thinking-gemini",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				paramCount, err := ctx.GetVariable("#model_param_count")
				assert.NoError(t, err)
				assert.Equal(t, "1", paramCount) // Only thinking_budget parameter
			},
		},
		{
			name: "model with dynamic thinking",
			args: map[string]string{
				"catalog_id":      "GM25P",
				"thinking_budget": "-1",
			},
			input:       "dynamic-thinking-pro",
			expectError: false,
		},
		{
			name: "model with empty description",
			args: map[string]string{
				"catalog_id":  "GM25F",
				"description": "",
			},
			input:       "empty-desc-gemini",
			expectError: false,
			checkFunc: func(t *testing.T, ctx neurotypes.Context) {
				// Should not set description variable if empty
				value, err := ctx.GetVariable("#model_description")
				assert.NoError(t, err)
				assert.Equal(t, "", value)
			},
		},
		{
			name: "model with only required parameters",
			args: map[string]string{
				"catalog_id": "GM25F",
			},
			input:       "minimal-gemini",
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
			setupGeminiModelTestRegistry(t, ctx)

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

func TestGeminiModelNewCommand_Execute_ServiceNotAvailable(t *testing.T) {
	cmd := &GeminiModelNewCommand{}

	// Don't setup services - should fail
	args := map[string]string{
		"catalog_id": "GM25F",
	}

	err := cmd.Execute(args, "test-model")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service not available")
}

func TestGeminiModelNewCommand_Execute_InvalidCatalogID(t *testing.T) {
	cmd := &GeminiModelNewCommand{}
	ctx := context.New()
	setupGeminiModelTestRegistry(t, ctx)

	tests := []struct {
		name      string
		catalogID string
		errorMsg  string
	}{
		{
			name:      "invalid catalog ID",
			catalogID: "INVALID123",
			errorMsg:  "failed to find model with catalog_id",
		},
		{
			name:      "empty catalog ID",
			catalogID: "",
			errorMsg:  "catalog_id is required",
		},
		{
			name:      "non-Gemini catalog ID",
			catalogID: "CS4", // Anthropic model
			errorMsg:  "is not a Gemini model",
		},
		{
			name:      "OpenAI catalog ID",
			catalogID: "O3R", // OpenAI model
			errorMsg:  "is not a Gemini model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset context for each test
			ctx = context.New()
			setupGeminiModelTestRegistry(t, ctx)

			args := map[string]string{
				"catalog_id": tt.catalogID,
			}

			err := cmd.Execute(args, "test-model")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

// setupGeminiModelTestRegistry sets up a test environment with required services for Gemini model commands
func setupGeminiModelTestRegistry(t *testing.T, ctx neurotypes.Context) {
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
var _ neurotypes.Command = (*GeminiModelNewCommand)(nil)
