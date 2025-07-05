package session

import (
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
	"os"
)

// setupTestServices initializes services for command testing
func setupTestServices(ctx neurotypes.Context) {
	// Set EDITOR environment variable for tests
	_ = os.Setenv("EDITOR", "echo")
	// Create a new registry for tests
	registry := services.NewRegistry()

	// Register required services
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewInterpolationService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewCatalogService())
	_ = registry.RegisterService(services.NewBashService())
	_ = registry.RegisterService(services.NewRenderService())
	_ = registry.RegisterService(services.NewChatSessionService())

	// Initialize all services
	_ = registry.InitializeAll(ctx)

	// Set as global registry
	services.SetGlobalRegistryForTesting(registry)
}

// cleanupTestServices resets the global registry
func cleanupTestServices() {
	services.ResetGlobalRegistryForTesting()
}
