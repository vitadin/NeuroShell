package services

import (
	"fmt"
	"strings"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/internal/testutils"
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
func (m *ModelService) Initialize(_ neurotypes.Context) error {
	m.initialized = true
	return nil
}

// CreateModel creates a new model configuration with the given parameters.
// It validates the model name (no spaces, unique), generates a unique ID,
// and maintains bidirectional mapping between names and IDs.
func (m *ModelService) CreateModel(name, provider, baseModel string, parameters map[string]any, description string, ctx neurotypes.Context) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	// Validate model name
	if err := m.validateModelName(name); err != nil {
		return nil, err
	}

	// Check if model name already exists
	if ctx.ModelNameExists(name) {
		return nil, fmt.Errorf("model name '%s' already exists", name)
	}

	// Validate required fields
	if provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if baseModel == "" {
		return nil, fmt.Errorf("base_model is required")
	}

	// Generate unique model ID (deterministic in test mode)
	modelID := testutils.GenerateUUID(ctx)

	// Ensure parameters map is not nil
	if parameters == nil {
		parameters = make(map[string]any)
	}

	// Create model configuration (deterministic time in test mode)
	now := testutils.GetCurrentTime(ctx)
	model := &neurotypes.ModelConfig{
		ID:          modelID,
		Name:        name,
		Provider:    provider,
		BaseModel:   baseModel,
		Parameters:  parameters,
		Description: description,
		IsDefault:   false, // New models are not default by default
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Store model in context with bidirectional mapping
	if err := m.storeModel(model, ctx); err != nil {
		return nil, fmt.Errorf("failed to store model: %w", err)
	}

	return model, nil
}

// CreateModelWithGlobalContext creates a new model configuration using the global context singleton.
func (m *ModelService) CreateModelWithGlobalContext(name, provider, baseModel string, parameters map[string]any, description string) (*neurotypes.ModelConfig, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.CreateModel(name, provider, baseModel, parameters, description, ctx)
}

// GetModel retrieves a model configuration by ID.
func (m *ModelService) GetModel(id string, ctx neurotypes.Context) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	models := ctx.GetModels()
	model, exists := models[id]
	if !exists {
		return nil, fmt.Errorf("model with ID '%s' not found", id)
	}

	return model, nil
}

// GetModelWithGlobalContext retrieves a model configuration by ID using the global context singleton.
func (m *ModelService) GetModelWithGlobalContext(id string) (*neurotypes.ModelConfig, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.GetModel(id, ctx)
}

// GetModelByName retrieves a model configuration by name.
func (m *ModelService) GetModelByName(name string, ctx neurotypes.Context) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	// Get ID from name mapping
	nameToID := ctx.GetModelNameToID()
	modelID, exists := nameToID[name]
	if !exists {
		return nil, fmt.Errorf("model with name '%s' not found", name)
	}

	// Get model by ID
	return m.GetModel(modelID, ctx)
}

// GetModelByNameWithGlobalContext retrieves a model configuration by name using the global context singleton.
func (m *ModelService) GetModelByNameWithGlobalContext(name string) (*neurotypes.ModelConfig, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.GetModelByName(name, ctx)
}

// ListModels returns all model configurations.
func (m *ModelService) ListModels(ctx neurotypes.Context) (map[string]*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	return ctx.GetModels(), nil
}

// ListModelsWithGlobalContext returns all model configurations using the global context singleton.
func (m *ModelService) ListModelsWithGlobalContext() (map[string]*neurotypes.ModelConfig, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.ListModels(ctx)
}

// DeleteModel removes a model configuration by ID.
// It maintains bidirectional mapping consistency by removing both name->ID and ID->name mappings.
func (m *ModelService) DeleteModel(id string, ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	// Check if model exists
	models := ctx.GetModels()
	model, exists := models[id]
	if !exists {
		return fmt.Errorf("model with ID '%s' not found", id)
	}

	// Remove from bidirectional mappings
	nameToID := ctx.GetModelNameToID()
	idToName := ctx.GetModelIDToName()

	// Remove name->ID mapping
	delete(nameToID, model.Name)
	ctx.SetModelNameToID(nameToID)

	// Remove ID->name mapping
	delete(idToName, id)
	ctx.SetModelIDToName(idToName)

	// Remove model from storage
	delete(models, id)
	ctx.SetModels(models)

	return nil
}

// DeleteModelWithGlobalContext removes a model configuration by ID using the global context singleton.
func (m *ModelService) DeleteModelWithGlobalContext(id string) error {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.DeleteModel(id, ctx)
}

// DeleteModelByName removes a model configuration by name.
func (m *ModelService) DeleteModelByName(name string, ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	// Get ID from name mapping
	nameToID := ctx.GetModelNameToID()
	modelID, exists := nameToID[name]
	if !exists {
		return fmt.Errorf("model with name '%s' not found", name)
	}

	// Delete by ID
	return m.DeleteModel(modelID, ctx)
}

// DeleteModelByNameWithGlobalContext removes a model configuration by name using the global context singleton.
func (m *ModelService) DeleteModelByNameWithGlobalContext(name string) error {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.DeleteModelByName(name, ctx)
}

// validateModelName validates that a model name meets requirements:
// - Not empty
// - No spaces
// - Reasonable length
func (m *ModelService) validateModelName(name string) error {
	if name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	if strings.Contains(name, " ") {
		return fmt.Errorf("model name cannot contain spaces")
	}

	if len(name) > 100 {
		return fmt.Errorf("model name cannot exceed 100 characters")
	}

	// Check for invalid characters (basic validation)
	if strings.ContainsAny(name, "\n\t\r") {
		return fmt.Errorf("model name cannot contain newlines or tabs")
	}

	return nil
}

// storeModel stores a model configuration in the context with bidirectional mapping.
func (m *ModelService) storeModel(model *neurotypes.ModelConfig, ctx neurotypes.Context) error {
	// Get current mappings
	models := ctx.GetModels()
	nameToID := ctx.GetModelNameToID()
	idToName := ctx.GetModelIDToName()

	// Store model
	models[model.ID] = model
	ctx.SetModels(models)

	// Update bidirectional mappings
	nameToID[model.Name] = model.ID
	ctx.SetModelNameToID(nameToID)

	idToName[model.ID] = model.Name
	ctx.SetModelIDToName(idToName)

	return nil
}

// ValidateModelParameters validates model parameters against standard constraints.
// This is a basic validation - more sophisticated validation could be added later.
func (m *ModelService) ValidateModelParameters(parameters map[string]any) error {
	if parameters == nil {
		return nil
	}

	// Validate common parameters
	if temp, ok := parameters["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			if tempFloat < 0.0 || tempFloat > 1.0 {
				return fmt.Errorf("temperature must be between 0.0 and 1.0")
			}
		}
	}

	if maxTokens, ok := parameters["max_tokens"]; ok {
		if maxTokensInt, ok := maxTokens.(int); ok {
			if maxTokensInt <= 0 {
				return fmt.Errorf("max_tokens must be positive")
			}
		}
	}

	if topP, ok := parameters["top_p"]; ok {
		if topPFloat, ok := topP.(float64); ok {
			if topPFloat < 0.0 || topPFloat > 1.0 {
				return fmt.Errorf("top_p must be between 0.0 and 1.0")
			}
		}
	}

	return nil
}
