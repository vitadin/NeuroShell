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
