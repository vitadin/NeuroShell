// Package commands provides the enhanced command resolution service.
// This service manages the priority-based command resolution system that
// supports builtin, stdlib, and user-defined commands.
package commands

import (
	"fmt"

	"neuroshell/internal/data/embedded"
	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// EnhancedCommandService manages the enhanced command resolution system.
// It provides priority-based command resolution and script loading capabilities.
type EnhancedCommandService struct {
	initialized bool
	registry    *EnhancedCommandRegistry
	stdlibSvc   *embedded.StdlibLoaderService
}

// NewEnhancedCommandService creates a new enhanced command service.
func NewEnhancedCommandService() *EnhancedCommandService {
	return &EnhancedCommandService{
		initialized: false,
		registry:    nil,
		stdlibSvc:   embedded.NewStdlibLoaderService(),
	}
}

// Name returns the service name for registration.
func (e *EnhancedCommandService) Name() string {
	return "enhanced-command"
}

// Initialize sets up the enhanced command resolution system.
// This loads all stdlib scripts and registers them as commands.
func (e *EnhancedCommandService) Initialize() error {
	logger.Debug("Initializing enhanced command service")

	// Initialize stdlib loader service
	err := e.stdlibSvc.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize stdlib loader: %w", err)
	}

	// Create enhanced registry with existing builtin registry
	e.registry = NewEnhancedCommandRegistry(GetGlobalRegistry())

	// Set the stdlib loader
	e.registry.SetStdlibLoader(e.stdlibSvc.GetLoader())

	// Load all stdlib scripts
	err = e.loadStdlibScripts()
	if err != nil {
		return fmt.Errorf("failed to load stdlib scripts: %w", err)
	}

	e.initialized = true
	logger.Debug("Enhanced command service initialized successfully")
	return nil
}

// GetRegistry returns the enhanced command registry.
func (e *EnhancedCommandService) GetRegistry() *EnhancedCommandRegistry {
	if !e.initialized {
		return nil
	}
	return e.registry
}

// GetResolver returns the command resolver interface.
func (e *EnhancedCommandService) GetResolver() neurotypes.CommandResolver {
	if !e.initialized {
		return nil
	}
	return e.registry
}

// loadStdlibScripts loads all stdlib scripts and registers them as commands.
func (e *EnhancedCommandService) loadStdlibScripts() error {
	logger.Debug("Loading stdlib scripts")

	scripts, err := e.stdlibSvc.LoadAllStdlibScripts()
	if err != nil {
		return fmt.Errorf("failed to load stdlib scripts: %w", err)
	}

	logger.Debug("Found stdlib scripts", "count", len(scripts))

	for scriptName, content := range scripts {
		// Create script command
		cmd := NewScriptCommand(scriptName, content, neurotypes.CommandTypeStdlib)

		// Register with enhanced registry
		err = e.registry.RegisterStdlibCommand(cmd)
		if err != nil {
			logger.Error("Failed to register stdlib command", "script", scriptName, "error", err)
			continue
		}

		logger.Debug("Registered stdlib command", "name", scriptName)
	}

	return nil
}

// RefreshStdlibCommands reloads all stdlib commands.
// This is useful for development or when stdlib scripts are updated.
func (e *EnhancedCommandService) RefreshStdlibCommands() error {
	if !e.initialized {
		return fmt.Errorf("enhanced command service not initialized")
	}

	return e.registry.RefreshStdlibCommands()
}

// ListCommands returns all available commands grouped by type.
func (e *EnhancedCommandService) ListCommands() map[string]neurotypes.CommandType {
	if !e.initialized {
		return make(map[string]neurotypes.CommandType)
	}

	return e.registry.ListCommands()
}

// ResolveCommand resolves a command using priority-based lookup.
func (e *EnhancedCommandService) ResolveCommand(name string) (*neurotypes.ResolvedCommand, error) {
	if !e.initialized {
		return nil, fmt.Errorf("enhanced command service not initialized")
	}

	return e.registry.ResolveCommand(name)
}

// Execute executes a command using the enhanced resolution system.
func (e *EnhancedCommandService) Execute(name string, args map[string]string, input string) error {
	if !e.initialized {
		return fmt.Errorf("enhanced command service not initialized")
	}

	return e.registry.Execute(name, args, input)
}

// GetParseMode returns the parse mode for a command.
func (e *EnhancedCommandService) GetParseMode(name string) neurotypes.ParseMode {
	if !e.initialized {
		return neurotypes.ParseModeKeyValue // Default fallback
	}

	return e.registry.GetParseMode(name)
}

// IsValidCommand checks if a command exists using enhanced resolution.
func (e *EnhancedCommandService) IsValidCommand(name string) bool {
	if !e.initialized {
		return false
	}

	return e.registry.IsValidCommand(name)
}

// Global enhanced command service instance
var globalEnhancedCommandService *EnhancedCommandService

// GetGlobalEnhancedCommandService returns the global enhanced command service instance.
func GetGlobalEnhancedCommandService() *EnhancedCommandService {
	return globalEnhancedCommandService
}

// SetGlobalEnhancedCommandService sets the global enhanced command service instance.
func SetGlobalEnhancedCommandService(service *EnhancedCommandService) {
	globalEnhancedCommandService = service
}
