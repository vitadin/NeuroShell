package services

import (
	"fmt"
	"sync"

	"neuroshell/pkg/types"
)

type Registry struct {
	mu       sync.RWMutex
	services map[string]types.Service
}

func NewRegistry() *Registry {
	return &Registry{
		services: make(map[string]types.Service),
	}
}

func (r *Registry) RegisterService(service types.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := service.Name()
	if _, exists := r.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	r.services[name] = service
	return nil
}

func (r *Registry) GetService(name string) (types.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

func (r *Registry) InitializeAll(ctx types.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, service := range r.services {
		if err := service.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize service %s: %w", name, err)
		}
	}

	return nil
}

func (r *Registry) GetAllServices() map[string]types.Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]types.Service)
	for name, service := range r.services {
		result[name] = service
	}

	return result
}

// Global registry instance
var GlobalRegistry = NewRegistry()