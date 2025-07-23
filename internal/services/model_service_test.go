package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

func TestModelService_Name(t *testing.T) {
	service := NewModelService()
	assert.Equal(t, "model", service.Name())
}

func TestModelService_Initialize(t *testing.T) {
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
			service := NewModelService()
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

func TestModelService_CreateModel_AutoActivation(t *testing.T) {
	// Clear global context for test isolation
	context.ResetGlobalContext()

	service := NewModelService()
	err := service.Initialize()
	require.NoError(t, err)

	ctx := context.GetGlobalContext()

	// Initially no active model
	assert.Equal(t, "", ctx.GetActiveModelID())

	// Create first model - should become active automatically
	model1, err := service.CreateModel("test-model-1", "openai", "gpt-4",
		map[string]any{"temperature": 0.7}, "Test model 1", ctx)
	require.NoError(t, err)
	assert.Equal(t, model1.ID, ctx.GetActiveModelID())

	// Create second model - should become active automatically
	model2, err := service.CreateModel("test-model-2", "openai", "gpt-3.5-turbo",
		map[string]any{"temperature": 0.5}, "Test model 2", ctx)
	require.NoError(t, err)
	assert.Equal(t, model2.ID, ctx.GetActiveModelID())
}

func TestModelService_GetActiveModelConfig_ContextTracking(t *testing.T) {
	// Clear global context for test isolation
	context.ResetGlobalContext()

	service := NewModelService()
	err := service.Initialize()
	require.NoError(t, err)

	ctx := context.GetGlobalContext()

	// When no models exist, should return synthetic default
	activeModel, err := service.GetActiveModelConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, "default-gpt-4", activeModel.ID)
	assert.Equal(t, "Default GPT-4 configuration (synthetic)", activeModel.Description)

	// Create a model - should become active
	model1, err := service.CreateModel("test-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.8}, "Test model", ctx)
	require.NoError(t, err)

	// GetActiveModelConfig should return the created model
	activeModel, err = service.GetActiveModelConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, model1.ID, activeModel.ID)
	assert.Equal(t, "test-model", activeModel.Name)
}

func TestModelService_SetActiveModel(t *testing.T) {
	// Clear global context for test isolation
	context.ResetGlobalContext()

	service := NewModelService()
	err := service.Initialize()
	require.NoError(t, err)

	ctx := context.GetGlobalContext()

	// Create two models
	model1, err := service.CreateModel("model-1", "openai", "gpt-4", nil, "Model 1", ctx)
	require.NoError(t, err)

	model2, err := service.CreateModel("model-2", "openai", "gpt-3.5-turbo", nil, "Model 2", ctx)
	require.NoError(t, err)

	// model2 should be active (last created)
	assert.Equal(t, model2.ID, ctx.GetActiveModelID())

	// Set model1 as active
	err = service.SetActiveModel(model1.ID, ctx)
	require.NoError(t, err)
	assert.Equal(t, model1.ID, ctx.GetActiveModelID())

	// Verify GetActiveModelConfig returns model1
	activeModel, err := service.GetActiveModelConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, model1.ID, activeModel.ID)

	// Test setting active model by name
	err = service.SetActiveModelByName("model-2", ctx)
	require.NoError(t, err)
	assert.Equal(t, model2.ID, ctx.GetActiveModelID())
}

func TestModelService_DeleteModel_ActiveCleanup(t *testing.T) {
	// Clear global context for test isolation
	context.ResetGlobalContext()

	service := NewModelService()
	err := service.Initialize()
	require.NoError(t, err)

	ctx := context.GetGlobalContext()

	// Create a model
	model, err := service.CreateModel("test-model", "openai", "gpt-4", nil, "Test model", ctx)
	require.NoError(t, err)

	// Verify it's active
	assert.Equal(t, model.ID, ctx.GetActiveModelID())

	// Delete the model
	err = service.DeleteModel(model.ID, ctx)
	require.NoError(t, err)

	// Active model ID should be cleared
	assert.Equal(t, "", ctx.GetActiveModelID())

	// GetActiveModelConfig should fall back to synthetic default
	activeModel, err := service.GetActiveModelConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, "default-gpt-4", activeModel.ID)
}

func TestModelService_GetActiveModelConfig_LatestFallback(t *testing.T) {
	// Clear global context for test isolation
	context.ResetGlobalContext()

	service := NewModelService()
	err := service.Initialize()
	require.NoError(t, err)

	ctx := context.GetGlobalContext()

	// Create two models
	_, err = service.CreateModel("model-1", "openai", "gpt-4", nil, "Model 1", ctx)
	require.NoError(t, err)

	// Wait a bit to ensure different timestamps
	model2, err := service.CreateModel("model-2", "openai", "gpt-3.5-turbo", nil, "Model 2", ctx)
	require.NoError(t, err)

	// Clear active model ID to test fallback
	ctx.SetActiveModelID("")

	// GetActiveModelConfig should find the latest model (model2) and set it as active
	activeModel, err := service.GetActiveModelConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, model2.ID, activeModel.ID)
	assert.Equal(t, model2.ID, ctx.GetActiveModelID()) // Should auto-set as active
}

func TestModelService_ClearActiveModel(t *testing.T) {
	// Clear global context for test isolation
	context.ResetGlobalContext()

	service := NewModelService()
	err := service.Initialize()
	require.NoError(t, err)

	ctx := context.GetGlobalContext()

	// Create a model
	model, err := service.CreateModel("test-model", "openai", "gpt-4", nil, "Test model", ctx)
	require.NoError(t, err)

	// Verify it's active
	assert.Equal(t, model.ID, ctx.GetActiveModelID())

	// Clear active model
	err = service.ClearActiveModel(ctx)
	require.NoError(t, err)

	// Active model should be cleared
	assert.Equal(t, "", ctx.GetActiveModelID())
}
