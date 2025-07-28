package context

import (
	"sync"

	"neuroshell/pkg/neurotypes"
)

// globalContext holds the singleton instance of the global context
var globalContext neurotypes.Context

// globalContextMu protects access to the global context instance
var globalContextMu sync.RWMutex

// globalContextOnce ensures singleton initialization happens only once
var globalContextOnce sync.Once

// GetGlobalContext returns the global context singleton instance in a thread-safe manner.
// If no global context has been set, it creates a new NeuroContext instance.
func GetGlobalContext() neurotypes.Context {
	globalContextOnce.Do(func() {
		if globalContext == nil {
			globalContext = New()
		}
	})

	globalContextMu.RLock()
	defer globalContextMu.RUnlock()
	return globalContext
}

// SetGlobalContext sets the global context instance in a thread-safe manner.
// This is useful for testing or when you need to replace the global context.
func SetGlobalContext(ctx neurotypes.Context) {
	globalContextMu.Lock()
	defer globalContextMu.Unlock()
	globalContext = ctx
}

// ResetGlobalContext resets the global context singleton to nil and resets the sync.Once.
// This is primarily for testing purposes to ensure clean state between tests.
func ResetGlobalContext() {
	globalContextMu.Lock()
	defer globalContextMu.Unlock()
	globalContext = nil
	// Reset the sync.Once by creating a new instance
	globalContextOnce = sync.Once{}
}
