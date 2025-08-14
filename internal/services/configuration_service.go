package services

import (
	"fmt"
	"strings"

	neuroshellcontext "neuroshell/internal/context"
)

// Note: Environment variable prefixes are now managed centrally in the context layer

// APIKeySource represents an API key found from a specific source with provider attribution
type APIKeySource struct {
	Source       string // "os", "config", "local"
	OriginalName string // "A_OPENAI_KEY", "OPENAI_API_KEY"
	Value        string // actual key
	Provider     string // "openai" (detected)
}

// ConfigPaths represents configuration file paths and their loading status
type ConfigPaths struct {
	ConfigDir       string // Configuration directory path
	ConfigDirExists bool   // Whether configuration directory exists
	ConfigEnvPath   string // Config .env file path (if exists)
	ConfigEnvLoaded bool   // Whether config .env was loaded
	LocalEnvPath    string // Local .env file path (if exists)
	LocalEnvLoaded  bool   // Whether local .env was loaded
	NeuroRCPath     string // Executed .neurorc file path
	NeuroRCExecuted bool   // Whether .neurorc was executed
}

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

	if err := ctx.LoadEnvironmentVariables(ctx.GetProviderEnvPrefixes()); err != nil {
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

	// Check for common API key patterns using context's provider list
	contextProviders := ctx.GetSupportedProviders()
	providers := make([]string, len(contextProviders))
	for i, provider := range contextProviders {
		providers[i] = strings.ToUpper(provider)
	}
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

// GetAllAPIKeys scans multiple sources and collects all configuration variables.
// Sources scanned: (a) OS environment variables, (b) config folder .env, (c) local .env
// Returns all variables with source attribution for transparent user control.
// Filtering is delegated to the calling command for better separation of concerns.
func (c *ConfigurationService) GetAllAPIKeys() ([]APIKeySource, error) {
	if !c.initialized {
		return nil, fmt.Errorf("configuration service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()

	// Load all sources with prefixes into context configuration map
	if err := ctx.LoadEnvironmentVariablesWithPrefix("os."); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}
	if err := ctx.LoadConfigDotEnvWithPrefix("config."); err != nil {
		return nil, fmt.Errorf("failed to load config .env: %w", err)
	}
	if err := ctx.LoadLocalDotEnvWithPrefix("local."); err != nil {
		return nil, fmt.Errorf("failed to load local .env: %w", err)
	}

	// Get all config values and collect all variables
	configMap, err := c.GetAllConfigValues()
	if err != nil {
		return nil, fmt.Errorf("failed to get config values: %w", err)
	}

	var keys []APIKeySource

	// Collect all prefixed configuration values
	for configKey, configValue := range configMap {
		// Skip empty values or very short values (less than 10 chars)
		if strings.TrimSpace(configValue) == "" || len(strings.TrimSpace(configValue)) < 10 {
			continue
		}

		// Check if this is a prefixed key (os., config., local.)
		var source, originalName string
		switch {
		case strings.HasPrefix(configKey, "os."):
			source = "os"
			originalName = configKey[3:] // Remove "os." prefix
		case strings.HasPrefix(configKey, "config."):
			source = "config"
			originalName = configKey[7:] // Remove "config." prefix
		case strings.HasPrefix(configKey, "local."):
			source = "local"
			originalName = configKey[6:] // Remove "local." prefix
		default:
			continue // Skip non-prefixed keys
		}

		// Collect all variables without provider filtering
		// Provider detection and API-related filtering will be done by the calling command
		keys = append(keys, APIKeySource{
			Source:       source,
			OriginalName: originalName,
			Value:        configValue,
			Provider:     "", // Will be determined by the calling command
		})
	}

	return keys, nil
}

// GetSupportedProviders returns the list of supported provider names from the context.
// This provides a service-layer interface to the centralized provider registry.
func (c *ConfigurationService) GetSupportedProviders() []string {
	if !c.initialized {
		// Return empty slice if not initialized
		return []string{}
	}

	ctx := neuroshellcontext.GetGlobalContext()
	return ctx.GetSupportedProviders()
}

// GetConfigurationPaths returns configuration file paths and their loading status.
// This information is determined by checking paths and file existence directly via context.
func (c *ConfigurationService) GetConfigurationPaths() (*ConfigPaths, error) {
	if !c.initialized {
		return nil, fmt.Errorf("configuration service not initialized")
	}

	ctx := neuroshellcontext.GetGlobalContext()

	// Get configuration directory
	configDir, err := ctx.GetUserConfigDir()
	if err != nil {
		configDir = "" // Default to empty if cannot get
	}

	// Check if config directory exists
	configDirExists := false
	if configDir != "" {
		configDirExists = ctx.FileExists(configDir)
	}

	// Check config .env file
	configEnvPath := ""
	configEnvLoaded := false
	if configDir != "" {
		configEnvPath = configDir + "/.env"
		configEnvLoaded = ctx.FileExists(configEnvPath)
	}

	// Check local .env file
	localEnvPath := ""
	localEnvLoaded := false
	workDir, err := ctx.GetWorkingDir()
	if err == nil {
		localEnvPath = workDir + "/.env"
		localEnvLoaded = ctx.FileExists(localEnvPath)
	}

	return &ConfigPaths{
		ConfigDir:       configDir,
		ConfigDirExists: configDirExists,
		ConfigEnvPath:   configEnvPath,
		ConfigEnvLoaded: configEnvLoaded,
		LocalEnvPath:    localEnvPath,
		LocalEnvLoaded:  localEnvLoaded,
		NeuroRCPath:     "",    // Will be set by command from system variables
		NeuroRCExecuted: false, // Will be set by command from system variables
	}, nil
}
