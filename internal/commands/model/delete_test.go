package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestDeleteCommand_Name(t *testing.T) {
	cmd := &DeleteCommand{}
	assert.Equal(t, "model-delete", cmd.Name())
}

func TestDeleteCommand_ParseMode(t *testing.T) {
	cmd := &DeleteCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestDeleteCommand_Description(t *testing.T) {
	cmd := &DeleteCommand{}
	assert.Contains(t, cmd.Description(), "Delete model configuration")
}

func TestDeleteCommand_Execute_DeleteByName(t *testing.T) {
	// Create test context
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services
	modelService, err := services.GetGlobalModelService()
	require.NoError(t, err)

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	// Create test models
	model1, err := modelService.CreateModel("test-model-1", "openai", "gpt-4", nil, "Test model 1", ctx)
	require.NoError(t, err)

	model2, err := modelService.CreateModel("test-model-2", "openai", "gpt-3.5-turbo", nil, "Test model 2", ctx)
	require.NoError(t, err)

	// Create delete command
	cmd := &DeleteCommand{}

	// Test deleting by exact name match
	err = cmd.Execute(map[string]string{}, "test-model-1")
	require.NoError(t, err)

	// Verify model was deleted
	_, err = modelService.GetModel(model1.ID, ctx)
	assert.Error(t, err, "Model should be deleted")

	// Verify other model still exists
	_, err = modelService.GetModel(model2.ID, ctx)
	assert.NoError(t, err, "Other model should still exist")

	// Verify deletion variables were set
	deletedName, err := variableService.Get("#deleted_model_name")
	require.NoError(t, err)
	assert.Equal(t, "test-model-1", deletedName)

	deletedID, err := variableService.Get("#deleted_model_id")
	require.NoError(t, err)
	assert.Equal(t, model1.ID, deletedID)
}

func TestDeleteCommand_Execute_DeleteByPartialName(t *testing.T) {
	// Create test context
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services
	modelService, err := services.GetGlobalModelService()
	require.NoError(t, err)

	// Create test model
	model, err := modelService.CreateModel("my-claude-model", "anthropic", "claude-3-sonnet", nil, "Test Claude model", ctx)
	require.NoError(t, err)

	// Create delete command
	cmd := &DeleteCommand{}

	// Test deleting by partial name match
	err = cmd.Execute(map[string]string{}, "claude")
	require.NoError(t, err)

	// Verify model was deleted
	_, err = modelService.GetModel(model.ID, ctx)
	assert.Error(t, err, "Model should be deleted")
}

func TestDeleteCommand_Execute_DeleteByID(t *testing.T) {
	// Create test context
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services
	modelService, err := services.GetGlobalModelService()
	require.NoError(t, err)

	// Create test model
	model, err := modelService.CreateModel("test-model", "openai", "gpt-4", nil, "Test model", ctx)
	require.NoError(t, err)

	// Create delete command
	cmd := &DeleteCommand{}

	// Test deleting by ID prefix
	idPrefix := model.ID[:8]
	args := map[string]string{"id": "true"}
	err = cmd.Execute(args, idPrefix)
	require.NoError(t, err)

	// Verify model was deleted
	_, err = modelService.GetModel(model.ID, ctx)
	assert.Error(t, err, "Model should be deleted")
}

func TestDeleteCommand_Execute_NoMatches(t *testing.T) {
	// Create test context
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services
	modelService, err := services.GetGlobalModelService()
	require.NoError(t, err)

	// Create test model
	_, err = modelService.CreateModel("test-model", "openai", "gpt-4", nil, "Test model", ctx)
	require.NoError(t, err)

	// Create delete command
	cmd := &DeleteCommand{}

	// Test with non-matching name
	err = cmd.Execute(map[string]string{}, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "No models found matching name")
	assert.Contains(t, err.Error(), "Available models:")
}

func TestDeleteCommand_Execute_MultipleMatches(t *testing.T) {
	// Create test context
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services
	modelService, err := services.GetGlobalModelService()
	require.NoError(t, err)

	// Create test models with similar names
	_, err = modelService.CreateModel("gpt-model-1", "openai", "gpt-4", nil, "GPT model 1", ctx)
	require.NoError(t, err)

	_, err = modelService.CreateModel("gpt-model-2", "openai", "gpt-3.5-turbo", nil, "GPT model 2", ctx)
	require.NoError(t, err)

	// Create delete command
	cmd := &DeleteCommand{}

	// Test with ambiguous name
	err = cmd.Execute(map[string]string{}, "gpt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Multiple models match name")
	assert.Contains(t, err.Error(), "Please be more specific")
}

func TestDeleteCommand_Execute_EmptyInput(t *testing.T) {
	// Create test context
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	cmd := &DeleteCommand{}

	err := cmd.Execute(map[string]string{}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "model name or ID prefix is required")
}

func TestDeleteCommand_Execute_NoModels(t *testing.T) {
	// Create test context
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Create delete command
	cmd := &DeleteCommand{}

	// Test with no models in system
	err := cmd.Execute(map[string]string{}, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no models found")
}

func TestDeleteCommand_Execute_ActiveModelCleanup(t *testing.T) {
	// Create test context
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services
	modelService, err := services.GetGlobalModelService()
	require.NoError(t, err)

	// Create and set active model
	model, err := modelService.CreateModel("active-model", "openai", "gpt-4", nil, "Active model", ctx)
	require.NoError(t, err)

	// Verify it's active (should be auto-activated)
	assert.Equal(t, model.ID, ctx.GetActiveModelID())

	// Create delete command
	cmd := &DeleteCommand{}

	// Delete the active model
	err = cmd.Execute(map[string]string{}, "active-model")
	require.NoError(t, err)

	// Verify active model was cleared
	assert.Equal(t, "", ctx.GetActiveModelID())
}

func TestDeleteCommand_findModelsByName(t *testing.T) {
	cmd := &DeleteCommand{}

	// Create test models
	models := map[string]*neurotypes.ModelConfig{
		"id1": {
			ID:   "id1",
			Name: "my-gpt-model",
		},
		"id2": {
			ID:   "id2",
			Name: "claude-model",
		},
		"id3": {
			ID:   "id3",
			Name: "another-gpt-config",
		},
	}

	// Test exact match
	matches := cmd.findModelsByName(models, "my-gpt-model")
	assert.Len(t, matches, 1)
	assert.Equal(t, "my-gpt-model", matches[0].Name)

	// Test partial match
	matches = cmd.findModelsByName(models, "gpt")
	assert.Len(t, matches, 2)

	// Test case insensitive match
	matches = cmd.findModelsByName(models, "CLAUDE")
	assert.Len(t, matches, 1)
	assert.Equal(t, "claude-model", matches[0].Name)

	// Test no match
	matches = cmd.findModelsByName(models, "nonexistent")
	assert.Len(t, matches, 0)
}

func TestDeleteCommand_findModelsByIDPrefix(t *testing.T) {
	cmd := &DeleteCommand{}

	// Create test models
	models := map[string]*neurotypes.ModelConfig{
		"abc12345": {
			ID:   "abc12345",
			Name: "model1",
		},
		"abc67890": {
			ID:   "abc67890",
			Name: "model2",
		},
		"def12345": {
			ID:   "def12345",
			Name: "model3",
		},
	}

	// Test unique prefix match
	matches := cmd.findModelsByIDPrefix(models, "abc1")
	assert.Len(t, matches, 1)
	assert.Equal(t, "abc12345", matches[0].ID)

	// Test multiple prefix matches
	matches = cmd.findModelsByIDPrefix(models, "abc")
	assert.Len(t, matches, 2)

	// Test case insensitive match
	matches = cmd.findModelsByIDPrefix(models, "ABC1")
	assert.Len(t, matches, 1)
	assert.Equal(t, "abc12345", matches[0].ID)

	// Test no match
	matches = cmd.findModelsByIDPrefix(models, "xyz")
	assert.Len(t, matches, 0)
}
