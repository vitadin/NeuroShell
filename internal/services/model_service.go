package services

import (
	"fmt"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// ModelService provides model configuration management operations for NeuroShell.
// It handles creation, storage, validation, and retrieval of LLM model configurations
// with bidirectional mapping support for unique model names and IDs.
type ModelService struct {
	initialized bool
}

// NewModelService creates a new ModelService instance.
func NewModelService() *ModelService {
	return &ModelService{
		initialized: false,
	}
}

// Name returns the service name "model" for registration.
func (m *ModelService) Name() string {
	return "model"
}

// Initialize sets up the ModelService for operation.
func (m *ModelService) Initialize() error {
	m.initialized = true
	return nil
}

// CreateModel creates a new model configuration with the given parameters.
// It validates the model name (no spaces, unique), generates a unique ID,
// and maintains bidirectional mapping between names and IDs.
func (m *ModelService) CreateModel(name, provider, baseModel string, parameters map[string]any, description, catalogID string, ctx neurotypes.Context) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.CreateModel(name, provider, baseModel, parameters, description, catalogID)
}

// CreateModelWithGlobalContext creates a new model configuration using the global context singleton.
func (m *ModelService) CreateModelWithGlobalContext(name, provider, baseModel string, parameters map[string]any, description, catalogID string) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.CreateModel(name, provider, baseModel, parameters, description, catalogID)
}

// GetModel retrieves a model configuration by ID.
func (m *ModelService) GetModel(id string, ctx neurotypes.Context) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.GetModel(id)
}

// GetModelWithGlobalContext retrieves a model configuration by ID using the global context singleton.
func (m *ModelService) GetModelWithGlobalContext(id string) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.GetModel(id)
}

// GetModelByName retrieves a model configuration by name.
func (m *ModelService) GetModelByName(name string, ctx neurotypes.Context) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.GetModelByName(name)
}

// GetModelByNameWithGlobalContext retrieves a model configuration by name using the global context singleton.
func (m *ModelService) GetModelByNameWithGlobalContext(name string) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.GetModelByName(name)
}

// ListModels returns all model configurations.
func (m *ModelService) ListModels(ctx neurotypes.Context) (map[string]*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.ListModels()
}

// ListModelsWithGlobalContext returns all model configurations using the global context singleton.
func (m *ModelService) ListModelsWithGlobalContext() (map[string]*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.ListModels()
}

// DeleteModel removes a model configuration by ID.
// It maintains bidirectional mapping consistency by removing both name->ID and ID->name mappings.
func (m *ModelService) DeleteModel(id string, ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.DeleteModel(id)
}

// DeleteModelWithGlobalContext removes a model configuration by ID using the global context singleton.
func (m *ModelService) DeleteModelWithGlobalContext(id string) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.DeleteModel(id)
}

// DeleteModelByName removes a model configuration by name.
func (m *ModelService) DeleteModelByName(name string, ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.DeleteModelByName(name)
}

// DeleteModelByNameWithGlobalContext removes a model configuration by name using the global context singleton.
func (m *ModelService) DeleteModelByNameWithGlobalContext(name string) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.DeleteModelByName(name)
}

// ValidateModelParameters validates model parameters against standard constraints.
// NOTE: Parameter validation is now handled by the ParameterValidatorService
// using model catalog parameter definitions. This method is kept for backward
// compatibility but performs minimal validation.
func (m *ModelService) ValidateModelParameters(parameters map[string]any) error {
	if parameters == nil {
		return nil
	}

	// Parameter validation is now handled by ParameterValidatorService
	// using model catalog parameter definitions. This provides proper
	// type checking, constraint validation, and model-specific rules.
	//
	// This method is kept for any future model service-specific validation
	// that might not belong in the parameter validator.

	return nil
}

// GetActiveModelConfig returns the model configuration for the active model.
// Uses context-only tracking: checks active model ID, falls back to latest model,
// or returns synthetic default if no models exist.
func (m *ModelService) GetActiveModelConfig(ctx neurotypes.Context) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.GetActiveModelConfig()
}

// GetActiveModelConfigWithGlobalContext returns the active model configuration using the global context singleton.
func (m *ModelService) GetActiveModelConfigWithGlobalContext() (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.GetActiveModelConfig()
}

// SetActiveModel sets the specified model as active by ID.
func (m *ModelService) SetActiveModel(modelID string, ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.SetActiveModel(modelID)
}

// SetActiveModelWithGlobalContext sets the specified model as active by ID using the global context singleton.
func (m *ModelService) SetActiveModelWithGlobalContext(modelID string) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.SetActiveModel(modelID)
}

// SetActiveModelByName sets the specified model as active by name.
func (m *ModelService) SetActiveModelByName(name string, ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.SetActiveModelByName(name)
}

// SetActiveModelByNameWithGlobalContext sets the specified model as active by name using the global context singleton.
func (m *ModelService) SetActiveModelByNameWithGlobalContext(name string) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.SetActiveModelByName(name)
}

// ClearActiveModel clears the active model setting.
func (m *ModelService) ClearActiveModel(ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	// Use the model subcontext instead of direct context access
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.ClearActiveModel()
}

// ClearActiveModelWithGlobalContext clears the active model setting using the global context singleton.
func (m *ModelService) ClearActiveModelWithGlobalContext() error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}
	ctx := neuroshellcontext.GetGlobalContext()
	modelCtx := neuroshellcontext.NewModelSubcontext(ctx.(*neuroshellcontext.NeuroContext))
	return modelCtx.ClearActiveModel()
}
