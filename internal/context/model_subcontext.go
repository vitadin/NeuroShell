// Package context provides model-specific context operations for NeuroShell.
// This file implements ModelSubcontext, a focused interface for model management
// that eliminates the need for services to know about global context internals.
package context

import (
	"fmt"
	"strings"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// ModelSubcontext provides focused model operations without exposing full context internals.
// This interface is designed to be passed to services that only need model functionality,
// following the Interface Segregation Principle.
type ModelSubcontext interface {
	// Core model operations
	CreateModel(name, provider, baseModel string, parameters map[string]any, description, catalogID string) (*neurotypes.ModelConfig, error)
	GetModel(id string) (*neurotypes.ModelConfig, error)
	GetModelByName(name string) (*neurotypes.ModelConfig, error)
	ListModels() (map[string]*neurotypes.ModelConfig, error)
	DeleteModel(id string) error
	DeleteModelByName(name string) error

	// Model management
	SetActiveModel(modelID string) error
	SetActiveModelByName(name string) error
	ClearActiveModel() error
	GetActiveModelConfig() (*neurotypes.ModelConfig, error)

	// Model naming and validation
	ModelNameExists(name string) bool
	ModelIDExists(id string) bool
	ValidateModelName(name string) error

	// Model mappings
	GetModelNameToID() map[string]string
	GetModelIDToName() map[string]string

	// Provider registry
	GetSupportedProviders() []string
	IsValidProvider(provider string) bool

	// Active model tracking
	GetActiveModelID() string
	SetActiveModelID(modelID string)
}

// modelSubcontextImpl implements ModelSubcontext using a NeuroContext.
// This provides a clean abstraction layer that services can depend on.
type modelSubcontextImpl struct {
	ctx *NeuroContext
}

// NewModelSubcontext creates a new ModelSubcontext from a NeuroContext.
// This is the factory function that services should use to get model functionality.
func NewModelSubcontext(ctx *NeuroContext) ModelSubcontext {
	return &modelSubcontextImpl{ctx: ctx}
}

// CreateModel creates a new model configuration with the given parameters.
func (m *modelSubcontextImpl) CreateModel(name, provider, baseModel string, parameters map[string]any, description, catalogID string) (*neurotypes.ModelConfig, error) {
	// Validate model name
	if err := m.ValidateModelName(name); err != nil {
		return nil, err
	}

	// Check if model name already exists
	if m.ModelNameExists(name) {
		return nil, fmt.Errorf("model name '%s' already exists", name)
	}

	// Validate required fields
	if provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if baseModel == "" {
		return nil, fmt.Errorf("base_model is required")
	}

	// Validate provider
	if !m.IsValidProvider(provider) {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	// Generate unique model ID (deterministic in test mode)
	modelID := testutils.GenerateUUID(m.ctx)

	// Ensure parameters map is not nil
	if parameters == nil {
		parameters = make(map[string]any)
	}

	// Create model configuration (deterministic time in test mode)
	now := testutils.GetCurrentTime(m.ctx)
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
	if err := m.storeModel(model); err != nil {
		return nil, fmt.Errorf("failed to store model: %w", err)
	}

	// Auto-activate the newly created model (following session pattern)
	m.ctx.SetActiveModelID(modelID)

	return model, nil
}

// GetModel retrieves a model configuration by ID.
func (m *modelSubcontextImpl) GetModel(id string) (*neurotypes.ModelConfig, error) {
	model, exists := m.ctx.models[id]
	if !exists {
		return nil, fmt.Errorf("model with ID '%s' not found", id)
	}

	return model, nil
}

// GetModelByName retrieves a model configuration by name.
func (m *modelSubcontextImpl) GetModelByName(name string) (*neurotypes.ModelConfig, error) {
	// Get ID from name mapping
	modelID, exists := m.ctx.modelNameToID[name]
	if !exists {
		return nil, fmt.Errorf("model with name '%s' not found", name)
	}

	// Get model by ID
	return m.GetModel(modelID)
}

// ListModels returns all model configurations.
func (m *modelSubcontextImpl) ListModels() (map[string]*neurotypes.ModelConfig, error) {
	return m.ctx.models, nil
}

// DeleteModel removes a model configuration by ID.
// It maintains bidirectional mapping consistency by removing both name->ID and ID->name mappings.
func (m *modelSubcontextImpl) DeleteModel(id string) error {
	// Check if model exists
	model, exists := m.ctx.models[id]
	if !exists {
		return fmt.Errorf("model with ID '%s' not found", id)
	}

	// Remove from bidirectional mappings
	delete(m.ctx.modelNameToID, model.Name)
	delete(m.ctx.modelIDToName, id)

	// Remove model from storage
	delete(m.ctx.models, id)

	// Clear active model if it was the deleted one
	if m.ctx.activeModelID == id {
		m.ctx.activeModelID = ""
	}

	return nil
}

// DeleteModelByName removes a model configuration by name.
func (m *modelSubcontextImpl) DeleteModelByName(name string) error {
	// Get ID from name mapping
	modelID, exists := m.ctx.modelNameToID[name]
	if !exists {
		return fmt.Errorf("model with name '%s' not found", name)
	}

	// Delete by ID
	return m.DeleteModel(modelID)
}

// SetActiveModel sets the specified model as active by ID.
func (m *modelSubcontextImpl) SetActiveModel(modelID string) error {
	// Validate that the model exists
	_, err := m.GetModel(modelID)
	if err != nil {
		return fmt.Errorf("cannot set active model: %w", err)
	}

	// Set as active in context
	m.ctx.SetActiveModelID(modelID)
	return nil
}

// SetActiveModelByName sets the specified model as active by name.
func (m *modelSubcontextImpl) SetActiveModelByName(name string) error {
	model, err := m.GetModelByName(name)
	if err != nil {
		return fmt.Errorf("cannot set active model: %w", err)
	}

	return m.SetActiveModel(model.ID)
}

// ClearActiveModel clears the active model setting.
func (m *modelSubcontextImpl) ClearActiveModel() error {
	m.ctx.SetActiveModelID("")
	return nil
}

// GetActiveModelConfig returns the model configuration for the active model.
// Uses context-only tracking: checks active model ID, falls back to latest model,
// or returns synthetic default if no models exist.
func (m *modelSubcontextImpl) GetActiveModelConfig() (*neurotypes.ModelConfig, error) {
	// 1. Try to get active model ID from context (single source of truth)
	activeID := m.ctx.GetActiveModelID()
	if activeID != "" {
		if model, err := m.GetModel(activeID); err == nil {
			return model, nil
		}
		// If active ID points to deleted model, clear it
		m.ctx.SetActiveModelID("")
	}

	// 2. If no active model, find latest created/updated model
	models, err := m.ListModels()
	if err != nil {
		return nil, err
	}
	if len(models) > 0 {
		latest := m.findLatestModelByTimestamp(models)
		m.ctx.SetActiveModelID(latest.ID) // Auto-set as active
		return latest, nil
	}

	// 3. Final fallback: synthetic default (no models exist)
	return m.createSyntheticDefault(), nil
}

// ModelNameExists checks if a model name already exists in the context.
func (m *modelSubcontextImpl) ModelNameExists(name string) bool {
	_, exists := m.ctx.modelNameToID[name]
	return exists
}

// ModelIDExists checks if a model ID already exists in the context.
func (m *modelSubcontextImpl) ModelIDExists(id string) bool {
	_, exists := m.ctx.models[id]
	return exists
}

// ValidateModelName validates that a model name meets requirements:
// - Not empty
// - No spaces
// - Reasonable length
func (m *modelSubcontextImpl) ValidateModelName(name string) error {
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

// GetModelNameToID returns the model name to ID mapping.
func (m *modelSubcontextImpl) GetModelNameToID() map[string]string {
	return m.ctx.modelNameToID
}

// GetModelIDToName returns the model ID to name mapping.
func (m *modelSubcontextImpl) GetModelIDToName() map[string]string {
	return m.ctx.modelIDToName
}

// GetSupportedProviders returns the list of supported LLM provider names.
func (m *modelSubcontextImpl) GetSupportedProviders() []string {
	return m.ctx.GetSupportedProviders()
}

// IsValidProvider checks if a given provider name is supported.
func (m *modelSubcontextImpl) IsValidProvider(provider string) bool {
	return m.ctx.IsValidProvider(provider)
}

// GetActiveModelID returns the currently active model ID.
func (m *modelSubcontextImpl) GetActiveModelID() string {
	return m.ctx.activeModelID
}

// SetActiveModelID sets the currently active model ID.
func (m *modelSubcontextImpl) SetActiveModelID(modelID string) {
	m.ctx.activeModelID = modelID
}

// storeModel stores a model configuration in the context with bidirectional mapping.
func (m *modelSubcontextImpl) storeModel(model *neurotypes.ModelConfig) error {
	// Store model
	m.ctx.models[model.ID] = model

	// Update bidirectional mappings
	m.ctx.modelNameToID[model.Name] = model.ID
	m.ctx.modelIDToName[model.ID] = model.Name

	return nil
}

// findLatestModelByTimestamp finds the most recently updated model for auto-activation.
// Uses UpdatedAt timestamp to determine the latest model.
func (m *modelSubcontextImpl) findLatestModelByTimestamp(models map[string]*neurotypes.ModelConfig) *neurotypes.ModelConfig {
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
func (m *modelSubcontextImpl) createSyntheticDefault() *neurotypes.ModelConfig {
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
		CreatedAt:   testutils.GetCurrentTime(m.ctx),
		UpdatedAt:   testutils.GetCurrentTime(m.ctx),
	}
}
