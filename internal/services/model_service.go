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
		CatalogID:   catalogID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Store model in context with bidirectional mapping
	if err := m.storeModel(model, ctx); err != nil {
		return nil, fmt.Errorf("failed to store model: %w", err)
	}

	// Auto-activate the newly created model (following session pattern)
	ctx.SetActiveModelID(modelID)

	return model, nil
}

// CreateModelWithGlobalContext creates a new model configuration using the global context singleton.
func (m *ModelService) CreateModelWithGlobalContext(name, provider, baseModel string, parameters map[string]any, description, catalogID string) (*neurotypes.ModelConfig, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.CreateModel(name, provider, baseModel, parameters, description, catalogID, ctx)
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

	// Clear active model if it was the deleted one
	if ctx.GetActiveModelID() == id {
		ctx.SetActiveModelID("")
	}

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

// GetActiveModelConfig returns the model configuration for the active model.
// Uses context-only tracking: checks active model ID, falls back to latest model,
// or returns synthetic default if no models exist.
func (m *ModelService) GetActiveModelConfig(ctx neurotypes.Context) (*neurotypes.ModelConfig, error) {
	if !m.initialized {
		return nil, fmt.Errorf("model service not initialized")
	}

	// 1. Try to get active model ID from context (single source of truth)
	activeID := ctx.GetActiveModelID()
	if activeID != "" {
		if model, err := m.GetModel(activeID, ctx); err == nil {
			return model, nil
		}
		// If active ID points to deleted model, clear it
		ctx.SetActiveModelID("")
	}

	// 2. If no active model, find latest created/updated model
	models := ctx.GetModels()
	if len(models) > 0 {
		latest := m.findLatestModelByTimestamp(models)
		ctx.SetActiveModelID(latest.ID) // Auto-set as active
		return latest, nil
	}

	// 3. Final fallback: synthetic default (no models exist)
	return m.createSyntheticDefault(ctx), nil
}

// GetActiveModelConfigWithGlobalContext returns the active model configuration using the global context singleton.
func (m *ModelService) GetActiveModelConfigWithGlobalContext() (*neurotypes.ModelConfig, error) {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.GetActiveModelConfig(ctx)
}

// findLatestModelByTimestamp finds the most recently updated model for auto-activation.
// Uses UpdatedAt timestamp to determine the latest model.
func (m *ModelService) findLatestModelByTimestamp(models map[string]*neurotypes.ModelConfig) *neurotypes.ModelConfig {
	var latest *neurotypes.ModelConfig
	for _, model := range models {
		if latest == nil || model.UpdatedAt.After(latest.UpdatedAt) {
			latest = model
		}
	}
	return latest
}

// createSyntheticDefault creates a synthetic default GPT-4 configuration as fallback.
// This is used when no models exist in the system.
func (m *ModelService) createSyntheticDefault(ctx neurotypes.Context) *neurotypes.ModelConfig {
	return &neurotypes.ModelConfig{
		ID:        "default-gpt-4",
		Name:      "default-gpt-4",
		Provider:  "openai",
		BaseModel: "gpt-4",
		Parameters: map[string]any{
			"temperature": 0.7,
			"max_tokens":  1000,
		},
		Description: "Default GPT-4 configuration (synthetic)",
		CreatedAt:   testutils.GetCurrentTime(ctx),
		UpdatedAt:   testutils.GetCurrentTime(ctx),
	}
}

// SetActiveModel sets the specified model as active by ID.
func (m *ModelService) SetActiveModel(modelID string, ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	// Validate that the model exists
	_, err := m.GetModel(modelID, ctx)
	if err != nil {
		return fmt.Errorf("cannot set active model: %w", err)
	}

	// Set as active in context
	ctx.SetActiveModelID(modelID)
	return nil
}

// SetActiveModelWithGlobalContext sets the specified model as active by ID using the global context singleton.
func (m *ModelService) SetActiveModelWithGlobalContext(modelID string) error {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.SetActiveModel(modelID, ctx)
}

// SetActiveModelByName sets the specified model as active by name.
func (m *ModelService) SetActiveModelByName(name string, ctx neurotypes.Context) error {
	model, err := m.GetModelByName(name, ctx)
	if err != nil {
		return fmt.Errorf("cannot set active model: %w", err)
	}

	return m.SetActiveModel(model.ID, ctx)
}

// SetActiveModelByNameWithGlobalContext sets the specified model as active by name using the global context singleton.
func (m *ModelService) SetActiveModelByNameWithGlobalContext(name string) error {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.SetActiveModelByName(name, ctx)
}

// ClearActiveModel clears the active model setting.
func (m *ModelService) ClearActiveModel(ctx neurotypes.Context) error {
	if !m.initialized {
		return fmt.Errorf("model service not initialized")
	}

	ctx.SetActiveModelID("")
	return nil
}

// ClearActiveModelWithGlobalContext clears the active model setting using the global context singleton.
func (m *ModelService) ClearActiveModelWithGlobalContext() error {
	ctx := neuroshellcontext.GetGlobalContext()
	return m.ClearActiveModel(ctx)
}
