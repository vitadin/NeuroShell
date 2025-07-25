package services

import (
	"fmt"
	"strings"

	neuroshellcontext "neuroshell/internal/context"
)

// Environment variable prefixes that should be loaded from OS
var envPrefixes = []string{"NEURO_", "OPENAI_", "ANTHROPIC_"}

// ConfigurationService provides configuration management for NeuroShell.
// It follows the three-layer architecture by being stateless and interacting only with Context.
// All configuration values are loaded and stored in the context's configuration map.
// This service contains only business logic and orchestrates loading via context methods.
type ConfigurationService struct {
	initialized bool
}

// NewConfigurationService creates a new ConfigurationService instance.
func NewConfigurationService() *ConfigurationService {
	return &ConfigurationService{
		initialized: false,
	}
}

// Name returns the service name "configuration" for registration.
func (c *ConfigurationService) Name() string {
	return "configuration"
}

// Initialize orchestrates configuration loading from multiple sources with proper priority.
// Priority (highest to lowest): Environment variables > Local .env > Config .env > Defaults
// All actual loading is delegated to the Context layer.
func (c *ConfigurationService) Initialize() error {
	if c.initialized {
		return nil
	}

	ctx := neuroshellcontext.GetGlobalContext()

	// Initialize empty configuration map
	ctx.SetConfigMap(make(map[string]string))

	// Orchestrate configuration loading in priority order (lowest to highest)
	// Each method is implemented in the Context layer
	if err := ctx.LoadDefaults(); err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	if err := ctx.LoadConfigDotEnv(); err != nil {
		return fmt.Errorf("failed to load config .env: %w", err)
	}

	if err := ctx.LoadLocalDotEnv(); err != nil {
		return fmt.Errorf("failed to load local .env: %w", err)
	}

	if err := ctx.LoadEnvironmentVariables(envPrefixes); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	c.initialized = true
	return nil
}

// GetAPIKey retrieves an API key for a specific provider from the configuration map.
// Returns error if the API key is not configured.
func (c *ConfigurationService) GetAPIKey(provider string) (string, error) {
	if !c.initialized {
		return "", fmt.Errorf("configuration service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()

	// Try provider-specific key first
	providerKey := fmt.Sprintf("NEURO_%s_API_KEY", strings.ToUpper(provider))
	if apiKey, exists := ctx.GetConfigValue(providerKey); exists && apiKey != "" {
		return apiKey, nil
	}

	// Try legacy provider-specific key (without NEURO_ prefix)
	legacyKey := fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider))
	if apiKey, exists := ctx.GetConfigValue(legacyKey); exists && apiKey != "" {
		return apiKey, nil
	}

	// Try generic key for backward compatibility
	if apiKey, exists := ctx.GetConfigValue("NEURO_API_KEY"); exists && apiKey != "" {
		return apiKey, nil
	}

	return "", fmt.Errorf("API key not configured for provider %s (expected %s or %s)", provider, providerKey, legacyKey)
}

// GetConfigValue retrieves a configuration value by key from the configuration map.
// Returns empty string if the configuration value doesn't exist (no error).
func (c *ConfigurationService) GetConfigValue(key string) (string, error) {
	if !c.initialized {
		return "", fmt.Errorf("configuration service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	value, _ := ctx.GetConfigValue(key)
	return value, nil
}

// SetConfigValue sets a configuration value in the configuration map.
// This is primarily for testing purposes.
func (c *ConfigurationService) SetConfigValue(key, value string) error {
	if !c.initialized {
		return fmt.Errorf("configuration service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	ctx.SetConfigValue(key, value)
	return nil
}

// LoadConfiguration reloads all configuration sources.
// This is useful for testing or when configuration files change.
func (c *ConfigurationService) LoadConfiguration() error {
	if !c.initialized {
		return fmt.Errorf("configuration service not initialized")
	}

	// Re-run the initialization process
	c.initialized = false
	return c.Initialize()
}

// ValidateConfiguration checks that required configuration values are present.
// Currently focuses on API key validation but can be extended.
func (c *ConfigurationService) ValidateConfiguration() error {
	if !c.initialized {
		return fmt.Errorf("configuration service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	configMap := ctx.GetConfigMap()

	// Check for common API key patterns
	providers := []string{"OPENAI", "ANTHROPIC", "OPENROUTER", "MOONSHOT"}
	hasAnyAPIKey := false

	for _, provider := range providers {
		// Check NEURO_ prefixed keys
		neuroKey := fmt.Sprintf("NEURO_%s_API_KEY", provider)
		if value, exists := configMap[neuroKey]; exists && value != "" {
			hasAnyAPIKey = true
			// Basic format validation
			if len(strings.TrimSpace(value)) < 10 {
				return fmt.Errorf("API key for %s appears to be too short", provider)
			}
		}

		// Check legacy keys (without NEURO_ prefix)
		legacyKey := fmt.Sprintf("%s_API_KEY", provider)
		if value, exists := configMap[legacyKey]; exists && value != "" {
			hasAnyAPIKey = true
			// Basic format validation
			if len(strings.TrimSpace(value)) < 10 {
				return fmt.Errorf("API key for %s appears to be too short", provider)
			}
		}
	}

	// Check for generic API key
	if value, exists := configMap["NEURO_API_KEY"]; exists && value != "" {
		hasAnyAPIKey = true
	}

	// Having at least one API key is recommended but not required
	if !hasAnyAPIKey {
		// This is a warning, not an error - the shell can still function
		return nil
	}

	return nil
}

// GetAllConfigValues returns all configuration values from the configuration map.
// This is primarily for debugging and listing purposes.
func (c *ConfigurationService) GetAllConfigValues() (map[string]string, error) {
	if !c.initialized {
		return nil, fmt.Errorf("configuration service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()
	return ctx.GetConfigMap(), nil
}
