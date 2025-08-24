package services

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
)

func TestNewShortcutService(t *testing.T) {
	service := NewShortcutService()
	assert.NotNil(t, service)
	assert.Equal(t, "shortcut", service.Name())
	assert.False(t, service.initialized)
	assert.NotNil(t, service.shortcuts)
	assert.Empty(t, service.shortcuts)
}

func TestShortcutService_Name(t *testing.T) {
	service := NewShortcutService()
	assert.Equal(t, "shortcut", service.Name())
}

func TestShortcutService_Initialize(t *testing.T) {
	service := NewShortcutService()

	// Initialize service
	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Check that Ctrl+S shortcut is registered by default
	assert.Contains(t, service.shortcuts, rune(19)) // ASCII 19 = Ctrl+S
	shortcut := service.shortcuts[19]
	assert.Equal(t, rune(19), shortcut.KeyCode)
	assert.Equal(t, "Ctrl+S", shortcut.Name)
	assert.Equal(t, "Save all sessions", shortcut.Description)
	assert.NotNil(t, shortcut.Handler)
}

func TestShortcutService_RegisterShortcut(t *testing.T) {
	service := NewShortcutService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name         string
		keyCode      rune
		shortcutName string
		description  string
		handler      ShortcutHandler
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "successful registration",
			keyCode:      rune(20), // Ctrl+T
			shortcutName: "Ctrl+T",
			description:  "Test shortcut",
			handler:      func() error { return nil },
			wantErr:      false,
		},
		{
			name:         "duplicate key code registration",
			keyCode:      rune(19), // Ctrl+S already registered
			shortcutName: "Ctrl+S",
			description:  "Duplicate shortcut",
			handler:      func() error { return nil },
			wantErr:      true,
			errMsg:       "shortcut for key code 19 already registered",
		},
		{
			name:         "another successful registration",
			keyCode:      rune(21), // Ctrl+U
			shortcutName: "Ctrl+U",
			description:  "Another test shortcut",
			handler:      func() error { return nil },
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RegisterShortcut(tt.keyCode, tt.shortcutName, tt.description, tt.handler)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, service.shortcuts, tt.keyCode)
				shortcut := service.shortcuts[tt.keyCode]
				assert.Equal(t, tt.keyCode, shortcut.KeyCode)
				assert.Equal(t, tt.shortcutName, shortcut.Name)
				assert.Equal(t, tt.description, shortcut.Description)
				assert.NotNil(t, shortcut.Handler)
			}
		})
	}
}

func TestShortcutService_RegisterShortcut_NotInitialized(t *testing.T) {
	service := NewShortcutService()

	// Try to register without initializing
	err := service.RegisterShortcut(rune(20), "Ctrl+T", "Test", func() error { return nil })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shortcut service not initialized")
}

func TestShortcutService_ExecuteShortcut(t *testing.T) {
	service := NewShortcutService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test variables to track handler execution
	var executedKeyCode rune
	var handlerError error

	// Register a test shortcut
	testHandler := func() error {
		executedKeyCode = rune(20)
		return handlerError
	}

	err = service.RegisterShortcut(rune(20), "Ctrl+T", "Test shortcut", testHandler)
	require.NoError(t, err)

	tests := []struct {
		name         string
		keyCode      rune
		handlerError error
		wantExecuted bool
	}{
		{
			name:         "execute existing shortcut successfully",
			keyCode:      rune(20),
			handlerError: nil,
			wantExecuted: true,
		},
		{
			name:         "execute non-existent shortcut",
			keyCode:      rune(99), // Not registered
			handlerError: nil,
			wantExecuted: false,
		},
		{
			name:         "execute existing shortcut with handler error",
			keyCode:      rune(20),
			handlerError: errors.New("handler failed"),
			wantExecuted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset test variables
			executedKeyCode = 0
			handlerError = tt.handlerError

			executed := service.ExecuteShortcut(tt.keyCode)

			if tt.wantExecuted {
				assert.True(t, executed)
				// Give a small amount of time for the goroutine to execute
				// Note: In real usage, we can't easily test the async execution
				// but we can verify that ExecuteShortcut returns true
			} else {
				assert.False(t, executed)
				assert.Equal(t, rune(0), executedKeyCode) // Handler should not execute
			}
		})
	}
}

func TestShortcutService_ExecuteShortcut_NotInitialized(t *testing.T) {
	service := NewShortcutService()

	// Try to execute without initializing
	executed := service.ExecuteShortcut(rune(19))
	assert.False(t, executed)
}

func TestShortcutService_GetShortcuts(t *testing.T) {
	service := NewShortcutService()
	err := service.Initialize()
	require.NoError(t, err)

	// Initially should have the default Ctrl+S shortcut
	shortcuts := service.GetShortcuts()
	assert.Len(t, shortcuts, 1)
	assert.Equal(t, rune(19), shortcuts[0].KeyCode)
	assert.Equal(t, "Ctrl+S", shortcuts[0].Name)

	// Add another shortcut
	err = service.RegisterShortcut(rune(20), "Ctrl+T", "Test shortcut", func() error { return nil })
	require.NoError(t, err)

	shortcuts = service.GetShortcuts()
	assert.Len(t, shortcuts, 2)

	// Verify both shortcuts are present (order may vary due to map iteration)
	keyCodeSet := make(map[rune]bool)
	for _, shortcut := range shortcuts {
		keyCodeSet[shortcut.KeyCode] = true
	}
	assert.True(t, keyCodeSet[rune(19)]) // Ctrl+S
	assert.True(t, keyCodeSet[rune(20)]) // Ctrl+T
}

func TestShortcutService_handleSaveAllSessions_ServiceNotAvailable(t *testing.T) {
	// Setup context with no chat session service registered
	neuroCtx := context.NewTestContext()
	concreteCtx := neuroCtx.(*context.NeuroContext)
	context.SetGlobalContext(concreteCtx)

	// Create registry without chat session service
	registry := NewRegistry()
	SetGlobalRegistry(registry)

	// Create shortcut service
	service := NewShortcutService()
	err := service.Initialize()
	require.NoError(t, err)

	// Execute the save all sessions handler
	err = service.handleSaveAllSessions()

	// Should fail with error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat session service not available")
}

func TestGetGlobalShortcutService(t *testing.T) {
	// Create and register shortcut service
	registry := NewRegistry()
	service := NewShortcutService()
	err := registry.RegisterService(service)
	require.NoError(t, err)
	SetGlobalRegistry(registry)

	// Test getting global service
	globalService, err := GetGlobalShortcutService()
	assert.NoError(t, err)
	assert.NotNil(t, globalService)
	assert.Equal(t, service, globalService)

	// Test second call should return same instance
	globalService2, err := GetGlobalShortcutService()
	assert.NoError(t, err)
	assert.Equal(t, globalService, globalService2)
}

func TestGetGlobalShortcutService_NotRegistered(t *testing.T) {
	// Create registry without shortcut service
	registry := NewRegistry()
	SetGlobalRegistry(registry)

	// Test getting global service when not registered
	globalService, err := GetGlobalShortcutService()
	assert.Error(t, err)
	assert.Nil(t, globalService)
	assert.Contains(t, err.Error(), "shortcut service not registered")
}

func TestShortcutService_ThreadSafety(t *testing.T) {
	service := NewShortcutService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test concurrent registration and execution
	done := make(chan bool, 2)

	// Goroutine 1: Register shortcuts
	go func() {
		defer func() { done <- true }()
		for i := 50; i < 60; i++ {
			err := service.RegisterShortcut(rune(i), fmt.Sprintf("Key%d", i), "Test", func() error { return nil })
			if err != nil {
				// Expected for some duplicates, continue
				continue
			}
		}
	}()

	// Goroutine 2: Execute shortcuts and get shortcuts list
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			service.ExecuteShortcut(rune(i))
			service.GetShortcuts()
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Should not panic and should have some shortcuts registered
	shortcuts := service.GetShortcuts()
	assert.True(t, len(shortcuts) >= 1) // At least the default Ctrl+S
}

// TestShortcutService_handleSaveAllSessions_TypeAssertion tests the type assertion failure case
func TestShortcutService_handleSaveAllSessions_TypeAssertion(t *testing.T) {
	// Create a mock service that has the wrong type
	registry := NewRegistry()

	// Register a service with the right name but wrong type
	mockService := &MockGenericService{name: "chatsession"}
	err := registry.RegisterService(mockService)
	require.NoError(t, err)
	SetGlobalRegistry(registry)

	service := NewShortcutService()
	err = service.Initialize()
	require.NoError(t, err)

	// Execute the save all sessions handler
	err = service.handleSaveAllSessions()

	// Should fail with type assertion error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "chat session service type assertion failed")
}

// MockGenericService for testing type assertion failures
type MockGenericService struct {
	name string
}

func (m *MockGenericService) Name() string      { return m.name }
func (m *MockGenericService) Initialize() error { return nil }
