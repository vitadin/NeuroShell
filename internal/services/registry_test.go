package services

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// Mock service for testing
type MockService struct {
	name              string
	initializeCalled  bool
	initializeError   error
	initializeContext neurotypes.Context
}

func NewMockService(name string) *MockService {
	return &MockService{
		name: name,
	}
}

func (m *MockService) Name() string {
	return m.name
}

func (m *MockService) Initialize(ctx neurotypes.Context) error {
	m.initializeCalled = true
	m.initializeContext = ctx
	return m.initializeError
}

func (m *MockService) SetInitializeError(err error) {
	m.initializeError = err
}

func TestRegistry_NewRegistry(t *testing.T) {
	registry := NewRegistry()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.services)
	assert.Equal(t, 0, len(registry.services))
}

func TestRegistry_RegisterService(t *testing.T) {
	tests := []struct {
		name    string
		service neurotypes.Service
		wantErr bool
	}{
		{
			name:    "register new service",
			service: NewMockService("test1"),
			wantErr: false,
		},
		{
			name:    "register another service",
			service: NewMockService("test2"),
			wantErr: false,
		},
	}

	registry := NewRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterService(tt.service)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify service was registered
				retrieved, err := registry.GetService(tt.service.Name())
				assert.NoError(t, err)
				assert.Equal(t, tt.service, retrieved)
			}
		})
	}
}

func TestRegistry_RegisterService_Duplicate(t *testing.T) {
	registry := NewRegistry()
	service1 := NewMockService("duplicate")
	service2 := NewMockService("duplicate")

	// Register first service
	err := registry.RegisterService(service1)
	assert.NoError(t, err)

	// Try to register service with same name
	err = registry.RegisterService(service2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service duplicate already registered")

	// Verify original service is still registered
	retrieved, err := registry.GetService("duplicate")
	assert.NoError(t, err)
	assert.Equal(t, service1, retrieved)
}

func TestRegistry_GetService(t *testing.T) {
	registry := NewRegistry()
	service := NewMockService("test")

	// Register service
	err := registry.RegisterService(service)
	require.NoError(t, err)

	tests := []struct {
		name        string
		serviceName string
		wantErr     bool
		wantService neurotypes.Service
	}{
		{
			name:        "get existing service",
			serviceName: "test",
			wantErr:     false,
			wantService: service,
		},
		{
			name:        "get non-existing service",
			serviceName: "nonexistent",
			wantErr:     true,
			wantService: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieved, err := registry.GetService(tt.serviceName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
				assert.Nil(t, retrieved)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantService, retrieved)
			}
		})
	}
}

func TestRegistry_InitializeAll(t *testing.T) {
	ctx := testutils.NewMockContext()

	tests := []struct {
		name     string
		services []neurotypes.Service
		wantErr  bool
	}{
		{
			name:     "initialize empty registry",
			services: []neurotypes.Service{},
			wantErr:  false,
		},
		{
			name: "initialize single service",
			services: []neurotypes.Service{
				NewMockService("service1"),
			},
			wantErr: false,
		},
		{
			name: "initialize multiple services",
			services: []neurotypes.Service{
				NewMockService("service1"),
				NewMockService("service2"),
				NewMockService("service3"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewRegistry()

			// Register all services
			for _, service := range tt.services {
				err := registry.RegisterService(service)
				require.NoError(t, err)
			}

			// Initialize all
			err := registry.InitializeAll(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify all services were initialized
				for _, service := range tt.services {
					mockService := service.(*MockService)
					assert.True(t, mockService.initializeCalled)
					assert.Equal(t, ctx, mockService.initializeContext)
				}
			}
		})
	}
}

func TestRegistry_InitializeAll_WithError(t *testing.T) {
	registry := NewRegistry()
	ctx := testutils.NewMockContext()

	service1 := NewMockService("service1")
	service2 := NewMockService("service2")
	service3 := NewMockService("service3")

	// Set service2 to return an error
	service2.SetInitializeError(errors.New("initialization failed"))

	err := registry.RegisterService(service1)
	require.NoError(t, err)
	err = registry.RegisterService(service2)
	require.NoError(t, err)
	err = registry.RegisterService(service3)
	require.NoError(t, err)

	// Initialize all - should fail on service2
	err = registry.InitializeAll(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize service service2")
	assert.Contains(t, err.Error(), "initialization failed")
}

func TestRegistry_GetAllServices(t *testing.T) {
	registry := NewRegistry()

	services := []neurotypes.Service{
		NewMockService("service1"),
		NewMockService("service2"),
		NewMockService("service3"),
	}

	// Register services
	for _, service := range services {
		err := registry.RegisterService(service)
		require.NoError(t, err)
	}

	// Get all services
	allServices := registry.GetAllServices()

	assert.Equal(t, len(services), len(allServices))

	for _, service := range services {
		retrieved, exists := allServices[service.Name()]
		assert.True(t, exists)
		assert.Equal(t, service, retrieved)
	}

	// Verify it's a copy (modifying returned map shouldn't affect registry)
	allServices["new_service"] = NewMockService("new_service")

	// Original registry should not have the new service
	_, err := registry.GetService("new_service")
	assert.Error(t, err)
}

// Test concurrent access
func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	ctx := testutils.NewMockContext()

	// Number of goroutines
	numGoroutines := 10
	servicesPerGoroutine := 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Concurrent registration
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < servicesPerGoroutine; j++ {
				serviceName := fmt.Sprintf("service_%d_%d", id, j)
				service := NewMockService(serviceName)

				err := registry.RegisterService(service)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all services were registered
	allServices := registry.GetAllServices()
	expectedCount := numGoroutines * servicesPerGoroutine
	assert.Equal(t, expectedCount, len(allServices))

	// Test concurrent retrieval
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < servicesPerGoroutine; j++ {
				serviceName := fmt.Sprintf("service_%d_%d", id, j)

				service, err := registry.GetService(serviceName)
				assert.NoError(t, err)
				assert.Equal(t, serviceName, service.Name())
			}
		}(i)
	}

	wg.Wait()

	// Test concurrent initialization
	err := registry.InitializeAll(ctx)
	assert.NoError(t, err)

	// Verify all services were initialized
	for _, service := range allServices {
		mockService := service.(*MockService)
		assert.True(t, mockService.initializeCalled)
	}
}

// Test real services
func TestRegistry_RealServices(t *testing.T) {
	registry := NewRegistry()
	ctx := testutils.NewMockContext()

	// Register actual services
	services := []neurotypes.Service{
		NewVariableService(),
		NewScriptService(),
		NewExecutorService(),
		NewInterpolationService(),
		NewEditorService(),
	}

	for _, service := range services {
		err := registry.RegisterService(service)
		assert.NoError(t, err)
	}

	// Verify all services are registered
	allServices := registry.GetAllServices()
	assert.Equal(t, len(services), len(allServices))

	// Test retrieving each service
	for _, service := range services {
		retrieved, err := registry.GetService(service.Name())
		assert.NoError(t, err)
		assert.Equal(t, service, retrieved)
	}

	// Initialize all services
	err := registry.InitializeAll(ctx)
	assert.NoError(t, err)

	// Cleanup EditorService temp directory
	editorService, err := registry.GetService("editor")
	if err == nil {
		if es, ok := editorService.(*EditorService); ok {
			_ = es.Cleanup()
		}
	}
}

// Test registry state consistency
func TestRegistry_StateConsistency(t *testing.T) {
	registry := NewRegistry()

	// Test empty state
	allServices := registry.GetAllServices()
	assert.Equal(t, 0, len(allServices))

	_, err := registry.GetService("nonexistent")
	assert.Error(t, err)

	// Register service
	service := NewMockService("test")
	err = registry.RegisterService(service)
	assert.NoError(t, err)

	// Test state after registration
	allServices = registry.GetAllServices()
	assert.Equal(t, 1, len(allServices))

	retrieved, err := registry.GetService("test")
	assert.NoError(t, err)
	assert.Equal(t, service, retrieved)

	// Test duplicate registration fails
	err = registry.RegisterService(NewMockService("test"))
	assert.Error(t, err)

	// State should remain unchanged
	allServices = registry.GetAllServices()
	assert.Equal(t, 1, len(allServices))

	retrieved, err = registry.GetService("test")
	assert.NoError(t, err)
	assert.Equal(t, service, retrieved)
}

// Benchmark tests
func BenchmarkRegistry_RegisterService(b *testing.B) {
	registry := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service := NewMockService(fmt.Sprintf("service_%d", i))
		_ = registry.RegisterService(service)
	}
}

func BenchmarkRegistry_GetService(b *testing.B) {
	registry := NewRegistry()

	// Pre-register some services
	for i := 0; i < 100; i++ {
		service := NewMockService(fmt.Sprintf("service_%d", i))
		_ = registry.RegisterService(service)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serviceName := fmt.Sprintf("service_%d", i%100)
		_, _ = registry.GetService(serviceName)
	}
}

func BenchmarkRegistry_GetAllServices(b *testing.B) {
	registry := NewRegistry()

	// Pre-register services
	for i := 0; i < 100; i++ {
		service := NewMockService(fmt.Sprintf("service_%d", i))
		_ = registry.RegisterService(service)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetAllServices()
	}
}

// Test GlobalRegistry
func TestGlobalRegistry(t *testing.T) {
	// Note: This test modifies the global registry, so it might affect other tests
	// In a real scenario, you'd want to reset or use a separate instance

	assert.NotNil(t, GlobalRegistry)

	// Test basic functionality
	service := NewMockService("global_test")
	err := GlobalRegistry.RegisterService(service)
	assert.NoError(t, err)

	retrieved, err := GlobalRegistry.GetService("global_test")
	assert.NoError(t, err)
	assert.Equal(t, service, retrieved)
}
