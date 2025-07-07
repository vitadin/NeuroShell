package services

import (
	"fmt"
	"sync"

	"neuroshell/pkg/neurotypes"
)

// Registry manages service registration and lifecycle for NeuroShell services.
type Registry struct {
	mu       sync.RWMutex
	services map[string]neurotypes.Service
}

// NewRegistry creates a new service registry with an empty service map.
func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]neurotypes.Service),
	}
}

// RegisterService adds a service to the registry, returning an error if already registered.
func (r *Registry) RegisterService(service neurotypes.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := service.Name()
	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	r.services[name] = service
	return nil
}

// GetService retrieves a service by name, returning an error if not found.
func (r *Registry) GetService(name string) (neurotypes.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// HasService checks if a service is registered without returning an error.
func (r *Registry) HasService(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.services[name]
	return exists
}

// InitializeAll initializes all registered services with the provided context.
func (r *Registry) InitializeAll(ctx neurotypes.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, service := range r.services {
		if err := service.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize service %s: %w", name, err)
		}
	}

	return nil
}

// GetAllServices returns a copy of all registered services.
func (r *Registry) GetAllServices() map[string]neurotypes.Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]neurotypes.Service)
	for name, service := range r.services {
		result[name] = service
	}

	return result
}

// Typed service access methods provide type-safe access to common services

// GetVariableService retrieves the variable service with proper type casting.
func (r *Registry) GetVariableService() (*VariableService, error) {
	service, err := r.GetService("variable")
	if err != nil {
		return nil, err
	}

	variableService, ok := service.(*VariableService)
	if !ok {
		return nil, fmt.Errorf("variable service has incorrect type")
	}

	return variableService, nil
}

// GetInterpolationService retrieves the interpolation service with proper type casting.
func (r *Registry) GetInterpolationService() (*InterpolationService, error) {
	service, err := r.GetService("interpolation")
	if err != nil {
		return nil, err
	}

	interpolationService, ok := service.(*InterpolationService)
	if !ok {
		return nil, fmt.Errorf("interpolation service has incorrect type")
	}

	return interpolationService, nil
}

// GetRenderService retrieves the render service with proper type casting.
func (r *Registry) GetRenderService() (*RenderService, error) {
	service, err := r.GetService("render")
	if err != nil {
		return nil, err
	}

	renderService, ok := service.(*RenderService)
	if !ok {
		return nil, fmt.Errorf("render service has incorrect type")
	}

	return renderService, nil
}

// GetBashService retrieves the bash service with proper type casting.
func (r *Registry) GetBashService() (*BashService, error) {
	service, err := r.GetService("bash")
	if err != nil {
		return nil, err
	}

	bashService, ok := service.(*BashService)
	if !ok {
		return nil, fmt.Errorf("bash service has incorrect type")
	}

	return bashService, nil
}

// GetExecutorService retrieves the executor service with proper type casting.
func (r *Registry) GetExecutorService() (*ExecutorService, error) {
	service, err := r.GetService("executor")
	if err != nil {
		return nil, err
	}

	executorService, ok := service.(*ExecutorService)
	if !ok {
		return nil, fmt.Errorf("executor service has incorrect type")
	}

	return executorService, nil
}

// GetScriptService retrieves the script service with proper type casting.
func (r *Registry) GetScriptService() (*ScriptService, error) {
	service, err := r.GetService("script")
	if err != nil {
		return nil, err
	}

	scriptService, ok := service.(*ScriptService)
	if !ok {
		return nil, fmt.Errorf("script service has incorrect type")
	}

	return scriptService, nil
}

// GetHelpService retrieves the help service with proper type casting.
func (r *Registry) GetHelpService() (*HelpService, error) {
	service, err := r.GetService("help")
	if err != nil {
		return nil, err
	}

	helpService, ok := service.(*HelpService)
	if !ok {
		return nil, fmt.Errorf("help service has incorrect type")
	}

	return helpService, nil
}

// GetEditorService retrieves the editor service with proper type casting.
func (r *Registry) GetEditorService() (*EditorService, error) {
	service, err := r.GetService("editor")
	if err != nil {
		return nil, err
	}

	editorService, ok := service.(*EditorService)
	if !ok {
		return nil, fmt.Errorf("editor service has incorrect type")
	}

	return editorService, nil
}

// GetChatSessionService retrieves the chat session service with proper type casting.
func (r *Registry) GetChatSessionService() (*ChatSessionService, error) {
	service, err := r.GetService("chat_session")
	if err != nil {
		return nil, err
	}

	chatSessionService, ok := service.(*ChatSessionService)
	if !ok {
		return nil, fmt.Errorf("chat session service has incorrect type")
	}

	return chatSessionService, nil
}

// GetModelService retrieves the model service with proper type casting.
func (r *Registry) GetModelService() (*ModelService, error) {
	service, err := r.GetService("model")
	if err != nil {
		return nil, err
	}

	modelService, ok := service.(*ModelService)
	if !ok {
		return nil, fmt.Errorf("model service has incorrect type")
	}

	return modelService, nil
}

// GetModelCatalogService retrieves the model catalog service with proper type casting.
func (r *Registry) GetModelCatalogService() (*ModelCatalogService, error) {
	service, err := r.GetService("model_catalog")
	if err != nil {
		return nil, err
	}

	modelCatalogService, ok := service.(*ModelCatalogService)
	if !ok {
		return nil, fmt.Errorf("model catalog service has incorrect type")
	}

	return modelCatalogService, nil
}

// GetAPIService retrieves the API service with proper type casting.
func (r *Registry) GetAPIService() (*APIService, error) {
	service, err := r.GetService("api")
	if err != nil {
		return nil, err
	}

	apiService, ok := service.(*APIService)
	if !ok {
		return nil, fmt.Errorf("api service has incorrect type")
	}

	return apiService, nil
}

// GlobalRegistry is the global service registry instance used throughout NeuroShell.
var GlobalRegistry = NewRegistry()

// globalRegistryMu protects access to the GlobalRegistry variable itself
var globalRegistryMu sync.RWMutex

// GetGlobalRegistry returns the global service registry instance in a thread-safe manner
func GetGlobalRegistry() *Registry {
	globalRegistryMu.RLock()
	defer globalRegistryMu.RUnlock()
	return GlobalRegistry
}

// SetGlobalRegistry sets the global service registry instance in a thread-safe manner
func SetGlobalRegistry(registry *Registry) {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()
	GlobalRegistry = registry
}

// Global service access convenience functions

// GetGlobalVariableService returns the variable service from the global registry.
func GetGlobalVariableService() (*VariableService, error) {
	return GetGlobalRegistry().GetVariableService()
}

// GetGlobalInterpolationService returns the interpolation service from the global registry.
func GetGlobalInterpolationService() (*InterpolationService, error) {
	return GetGlobalRegistry().GetInterpolationService()
}

// GetGlobalRenderService returns the render service from the global registry.
func GetGlobalRenderService() (*RenderService, error) {
	return GetGlobalRegistry().GetRenderService()
}

// GetGlobalBashService returns the bash service from the global registry.
func GetGlobalBashService() (*BashService, error) {
	return GetGlobalRegistry().GetBashService()
}

// GetGlobalExecutorService returns the executor service from the global registry.
func GetGlobalExecutorService() (*ExecutorService, error) {
	return GetGlobalRegistry().GetExecutorService()
}

// GetGlobalScriptService returns the script service from the global registry.
func GetGlobalScriptService() (*ScriptService, error) {
	return GetGlobalRegistry().GetScriptService()
}

// GetGlobalHelpService returns the help service from the global registry.
func GetGlobalHelpService() (*HelpService, error) {
	return GetGlobalRegistry().GetHelpService()
}

// GetGlobalEditorService returns the editor service from the global registry.
func GetGlobalEditorService() (*EditorService, error) {
	return GetGlobalRegistry().GetEditorService()
}

// GetGlobalChatSessionService returns the chat session service from the global registry.
func GetGlobalChatSessionService() (*ChatSessionService, error) {
	return GetGlobalRegistry().GetChatSessionService()
}

// GetGlobalModelService returns the model service from the global registry.
func GetGlobalModelService() (*ModelService, error) {
	return GetGlobalRegistry().GetModelService()
}

// GetGlobalModelCatalogService returns the model catalog service from the global registry.
func GetGlobalModelCatalogService() (*ModelCatalogService, error) {
	return GetGlobalRegistry().GetModelCatalogService()
}

// GetGlobalAPIService returns the API service from the global registry.
func GetGlobalAPIService() (*APIService, error) {
	return GetGlobalRegistry().GetAPIService()
}
