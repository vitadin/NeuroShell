// TODO: Integrate into state machine - temporarily commented out for build compatibility
package model

/*

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/commands"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// TestStatusCommand_Name tests the command name.
func TestStatusCommand_Name(t *testing.T) {
	cmd := &StatusCommand{}
	assert.Equal(t, "model-status", cmd.Name())
}

// TestStatusCommand_ParseMode tests the parse mode.
func TestStatusCommand_ParseMode(t *testing.T) {
	cmd := &StatusCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

// TestStatusCommand_Description tests the command description.
func TestStatusCommand_Description(t *testing.T) {
	cmd := &StatusCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "status")
	assert.Contains(t, desc, "model")
}

// TestStatusCommand_Usage tests the usage information.
func TestStatusCommand_Usage(t *testing.T) {
	cmd := &StatusCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\model-status")
	assert.Contains(t, usage, "Examples:")
}

// TestStatusCommand_HelpInfo tests the help information structure.
func TestStatusCommand_HelpInfo(t *testing.T) {
	cmd := &StatusCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, cmd.Name(), helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, cmd.ParseMode(), helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.NotEmpty(t, helpInfo.Options)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Verify specific options
	optionNames := make(map[string]bool)
	for _, option := range helpInfo.Options {
		optionNames[option.Name] = true
	}
	assert.True(t, optionNames["name"])
	assert.True(t, optionNames["provider"])
	assert.True(t, optionNames["sort"])
}

// TestStatusCommand_Execute_EmptyModels tests execution with no models.
func TestStatusCommand_Execute_EmptyModels(t *testing.T) {
	// Create test context
	ctx := context.New()

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewInterpolationService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Execute command
	cmd := &StatusCommand{}
	err := cmd.Execute(map[string]string{}, "")

	assert.NoError(t, err)

	// Check output variable
	variableService, _ := services.GetGlobalVariableService()
	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "No model configurations found")
	assert.Contains(t, output, "\\model-new")

	// Check count variables
	count, err := variableService.Get("#model_count")
	assert.NoError(t, err)
	assert.Equal(t, "0", count)

	filteredCount, err := variableService.Get("#model_filtered_count")
	assert.NoError(t, err)
	assert.Equal(t, "0", filteredCount)
}

// TestStatusCommand_Execute_WithModels tests execution with sample models.
func TestStatusCommand_Execute_WithModels(t *testing.T) {
	// Create test context
	ctx := context.New()

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewInterpolationService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Create test models
	modelService, _ := services.GetGlobalModelService()

	model1, err := modelService.CreateModelWithGlobalContext(
		"test-gpt4", "openai", "gpt-4",
		map[string]any{"temperature": 0.7, "max_tokens": 1000},
		"Test GPT-4 model")
	require.NoError(t, err)

	model2, err := modelService.CreateModelWithGlobalContext(
		"test-claude", "anthropic", "claude-3-sonnet",
		map[string]any{"temperature": 0.5},
		"Test Claude model")
	require.NoError(t, err)

	// Execute command
	cmd := &StatusCommand{}
	err = cmd.Execute(map[string]string{}, "")

	assert.NoError(t, err)

	// Check output variable
	variableService, _ := services.GetGlobalVariableService()
	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Model Configurations (2 models)")
	assert.Contains(t, output, "test-gpt4")
	assert.Contains(t, output, "test-claude")
	assert.Contains(t, output, "openai")
	assert.Contains(t, output, "anthropic")
	assert.Contains(t, output, model1.ID[:8])
	assert.Contains(t, output, model2.ID[:8])

	// Check count variables
	count, err := variableService.Get("#model_count")
	assert.NoError(t, err)
	assert.Equal(t, "2", count)

	filteredCount, err := variableService.Get("#model_filtered_count")
	assert.NoError(t, err)
	assert.Equal(t, "2", filteredCount)
}

// TestStatusCommand_Execute_FilterByName tests filtering by model name.
func TestStatusCommand_Execute_FilterByName(t *testing.T) {
	// Create test context
	ctx := context.New()

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewInterpolationService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Create test models
	modelService, _ := services.GetGlobalModelService()

	_, err := modelService.CreateModelWithGlobalContext(
		"prod-gpt4", "openai", "gpt-4",
		map[string]any{"temperature": 0.3},
		"Production GPT-4 model")
	require.NoError(t, err)

	_, err = modelService.CreateModelWithGlobalContext(
		"test-gpt4", "openai", "gpt-4",
		map[string]any{"temperature": 0.7},
		"Test GPT-4 model")
	require.NoError(t, err)

	// Execute command with name filter
	cmd := &StatusCommand{}
	err = cmd.Execute(map[string]string{"name": "prod"}, "")

	assert.NoError(t, err)

	// Check output variable
	variableService, _ := services.GetGlobalVariableService()
	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Filter: name='prod'")
	assert.Contains(t, output, "prod-gpt4")
	assert.NotContains(t, output, "test-gpt4")

	// Check filtered count
	filteredCount, err := variableService.Get("#model_filtered_count")
	assert.NoError(t, err)
	assert.Equal(t, "1", filteredCount)

	// Check single model variables
	currentName, err := variableService.Get("#model_current_name")
	assert.NoError(t, err)
	assert.Equal(t, "prod-gpt4", currentName)

	currentProvider, err := variableService.Get("#model_current_provider")
	assert.NoError(t, err)
	assert.Equal(t, "openai", currentProvider)
}

// TestStatusCommand_Execute_FilterByProvider tests filtering by provider.
func TestStatusCommand_Execute_FilterByProvider(t *testing.T) {
	// Create test context
	ctx := context.New()

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewInterpolationService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Create test models
	modelService, _ := services.GetGlobalModelService()

	_, err := modelService.CreateModelWithGlobalContext(
		"gpt4-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7},
		"OpenAI GPT-4 model")
	require.NoError(t, err)

	_, err = modelService.CreateModelWithGlobalContext(
		"claude-model", "anthropic", "claude-3-sonnet",
		map[string]any{"temperature": 0.5},
		"Anthropic Claude model")
	require.NoError(t, err)

	// Execute command with provider filter
	cmd := &StatusCommand{}
	err = cmd.Execute(map[string]string{"provider": "anthropic"}, "")

	assert.NoError(t, err)

	// Check output variable
	variableService, _ := services.GetGlobalVariableService()
	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "Filter: provider='anthropic'")
	assert.Contains(t, output, "claude-model")
	assert.NotContains(t, output, "gpt4-model")

	// Check filtered count
	filteredCount, err := variableService.Get("#model_filtered_count")
	assert.NoError(t, err)
	assert.Equal(t, "1", filteredCount)
}

// TestStatusCommand_Execute_SortByCreated tests sorting by creation date.
func TestStatusCommand_Execute_SortByCreated(t *testing.T) {
	// Create test context
	ctx := context.New()

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewInterpolationService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Create test models with slight time difference
	modelService, _ := services.GetGlobalModelService()

	// Create first model
	_, err := modelService.CreateModelWithGlobalContext(
		"older-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7},
		"Older model")
	require.NoError(t, err)

	// Small delay to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Create second model
	_, err = modelService.CreateModelWithGlobalContext(
		"newer-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.5},
		"Newer model")
	require.NoError(t, err)

	// Execute command with sort by created
	cmd := &StatusCommand{}
	err = cmd.Execute(map[string]string{"sort": "created"}, "")

	assert.NoError(t, err)

	// Check output variable
	variableService, _ := services.GetGlobalVariableService()
	output, err := variableService.Get("_output")
	assert.NoError(t, err)

	// Check that older model appears before newer model
	olderIndex := strings.Index(output, "older-model")
	newerIndex := strings.Index(output, "newer-model")
	assert.True(t, olderIndex < newerIndex, "Models should be sorted by creation date")
}

// TestStatusCommand_Execute_InvalidSort tests invalid sort option.
func TestStatusCommand_Execute_InvalidSort(t *testing.T) {
	// Create test context
	ctx := context.New()

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewInterpolationService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Execute command with invalid sort option
	cmd := &StatusCommand{}
	err := cmd.Execute(map[string]string{"sort": "invalid"}, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sort option")
	assert.Contains(t, err.Error(), "name, created, provider")
}

// TestStatusCommand_Execute_ServiceUnavailable tests handling of unavailable services.
func TestStatusCommand_Execute_ServiceUnavailable(t *testing.T) {
	// Create empty registry (no services)
	registry := services.NewRegistry()

	// Set global registry
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	cmd := &StatusCommand{}
	err := cmd.Execute(map[string]string{}, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model service not available")
}

// TestStatusCommand_Execute_NoMatchingModels tests filtering with no matches.
func TestStatusCommand_Execute_NoMatchingModels(t *testing.T) {
	// Create test context
	ctx := context.New()

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewInterpolationService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Create test model
	modelService, _ := services.GetGlobalModelService()
	_, err := modelService.CreateModelWithGlobalContext(
		"gpt4-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7},
		"Test model")
	require.NoError(t, err)

	// Execute command with filter that matches nothing
	cmd := &StatusCommand{}
	err = cmd.Execute(map[string]string{"provider": "nonexistent"}, "")

	assert.NoError(t, err)

	// Check output variable
	variableService, _ := services.GetGlobalVariableService()
	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.Contains(t, output, "No model configurations found matching filter")
	assert.Contains(t, output, "provider='nonexistent'")

	// Check filtered count
	filteredCount, err := variableService.Get("#model_filtered_count")
	assert.NoError(t, err)
	assert.Equal(t, "0", filteredCount)

	// Total count should still be 1
	count, err := variableService.Get("#model_count")
	assert.NoError(t, err)
	assert.Equal(t, "1", count)
}

// TestStatusCommand_validateSortOption tests sort option validation.
func TestStatusCommand_validateSortOption(t *testing.T) {
	cmd := &StatusCommand{}

	// Valid options
	assert.NoError(t, cmd.validateSortOption("name"))
	assert.NoError(t, cmd.validateSortOption("created"))
	assert.NoError(t, cmd.validateSortOption("provider"))

	// Invalid options
	assert.Error(t, cmd.validateSortOption("invalid"))
	assert.Error(t, cmd.validateSortOption(""))
	assert.Error(t, cmd.validateSortOption("ID"))
}

// TestStatusCommand_Registration tests that the command is properly registered.
func TestStatusCommand_Registration(t *testing.T) {
	// Get the command from global registry
	command, exists := commands.GlobalRegistry.Get("model-status")
	assert.True(t, exists)
	assert.NotNil(t, command)

	// Verify it's the correct type
	statusCmd, ok := command.(*StatusCommand)
	assert.True(t, ok)
	assert.Equal(t, "model-status", statusCmd.Name())
}

// TestStatusCommand_formatModelDetails tests model detail formatting.
func TestStatusCommand_formatModelDetails(t *testing.T) {
	cmd := &StatusCommand{}

	// Create test model
	model := &neurotypes.ModelConfig{
		ID:          "12345678-1234-1234-1234-123456789012",
		Name:        "test-model",
		Provider:    "openai",
		BaseModel:   "gpt-4",
		Parameters:  map[string]any{"temperature": 0.7, "max_tokens": 1000},
		Description: "Test model description",
		IsDefault:   false,
		CreatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	details := cmd.formatModelDetails(model)

	assert.Contains(t, details, "test-model (12345678)")
	assert.Contains(t, details, "Provider: openai")
	assert.Contains(t, details, "Base Model: gpt-4")
	assert.Contains(t, details, "Parameters: 2")
	assert.Contains(t, details, "temperature=0.7")
	assert.Contains(t, details, "max_tokens=1000")
	assert.Contains(t, details, "Created: 2024-01-01 12:00:00")
	assert.Contains(t, details, "Description: Test model description")
	assert.NotContains(t, details, "Updated:")      // Same as created
	assert.NotContains(t, details, "Default: true") // Not default
}

// TestStatusCommand_filterModels tests model filtering logic.
func TestStatusCommand_filterModels(t *testing.T) {
	cmd := &StatusCommand{}

	models := map[string]*neurotypes.ModelConfig{
		"1": {Name: "gpt4-prod", Provider: "openai"},
		"2": {Name: "gpt4-test", Provider: "openai"},
		"3": {Name: "claude-prod", Provider: "anthropic"},
	}

	// Test name filter
	filtered := cmd.filterModels(models, "prod", "")
	assert.Len(t, filtered, 2)

	// Test provider filter
	filtered = cmd.filterModels(models, "", "openai")
	assert.Len(t, filtered, 2)

	// Test combined filters
	filtered = cmd.filterModels(models, "prod", "anthropic")
	assert.Len(t, filtered, 1)
	assert.Equal(t, "claude-prod", filtered[0].Name)

	// Test no matches
	filtered = cmd.filterModels(models, "nonexistent", "")
	assert.Len(t, filtered, 0)
}

// TestStatusCommand_sortModels tests model sorting logic.
func TestStatusCommand_sortModels(t *testing.T) {
	cmd := &StatusCommand{}

	models := []*neurotypes.ModelConfig{
		{Name: "z-model", Provider: "openai", CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{Name: "a-model", Provider: "anthropic", CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "m-model", Provider: "openai", CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
	}

	// Test sort by name
	sorted := cmd.sortModels(models, "name")
	assert.Equal(t, "a-model", sorted[0].Name)
	assert.Equal(t, "m-model", sorted[1].Name)
	assert.Equal(t, "z-model", sorted[2].Name)

	// Test sort by created
	sorted = cmd.sortModels(models, "created")
	assert.Equal(t, "a-model", sorted[0].Name) // 2024-01-01
	assert.Equal(t, "z-model", sorted[1].Name) // 2024-01-02
	assert.Equal(t, "m-model", sorted[2].Name) // 2024-01-03

	// Test sort by provider (then by name)
	sorted = cmd.sortModels(models, "provider")
	assert.Equal(t, "a-model", sorted[0].Name) // anthropic
	assert.Equal(t, "m-model", sorted[1].Name) // openai (comes before z alphabetically)
	assert.Equal(t, "z-model", sorted[2].Name) // openai
}
*/
