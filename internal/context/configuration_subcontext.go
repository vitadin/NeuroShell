package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

// ConfigurationSubcontext defines the interface for configuration management functionality.
// This manages configuration maps, environment variables, and file operations.
type ConfigurationSubcontext interface {

	// Configuration map operations
	GetConfigMap() map[string]string
	SetConfigMap(configMap map[string]string)
	GetConfigValue(key string) (string, bool)
	SetConfigValue(key, value string)

	// Configuration loading operations
	LoadDefaults() error
	LoadConfigDotEnv() error
	LoadLocalDotEnv() error
	LoadEnvironmentVariables(prefixes []string) error
	LoadEnvironmentVariablesWithPrefix(sourcePrefix string) error
	LoadConfigDotEnvWithPrefix(sourcePrefix string) error
	LoadLocalDotEnvWithPrefix(sourcePrefix string) error

	// Environment variable operations
	GetEnv(key string) string
	SetEnvVariable(key, value string) error
	GetEnvVariable(key string) string
	SetTestEnvOverride(key, value string)
	ClearTestEnvOverride(key string)
	ClearAllTestEnvOverrides()
	GetTestEnvOverrides() map[string]string

	// Test working directory operations
	SetTestWorkingDir(path string)

	// Parent context operations
	SetParentContext(parent TestModeProvider)

	// File system operations
	GetUserConfigDir() (string, error)
	GetWorkingDir() (string, error)
	FileExists(path string) bool
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error

	// Allowed global variables management
	IsAllowedGlobalVariable(name string) bool
	GetAllowedGlobalVariables() []string

	// Default command configuration
	GetDefaultCommand() string
	SetDefaultCommand(command string)

	// Read-only command overrides
	SetCommandReadOnly(commandName string, readOnly bool)
	RemoveCommandReadOnlyOverride(commandName string)
	IsCommandReadOnlyOverride(commandName string) (bool, bool)
	GetReadOnlyOverrides() map[string]bool
	ClearAllReadOnlyOverrides()
}

// TestModeProvider interface allows subcontexts to check test mode from parent context
type TestModeProvider interface {
	IsTestMode() bool
}

// configurationSubcontext implements the ConfigurationSubcontext interface.
type configurationSubcontext struct {
	// Configuration management
	configMap   map[string]string // Configuration key-value store
	configMutex sync.RWMutex      // Protects configMap

	// Test environment overrides
	parentContext    TestModeProvider  // Reference to parent context for test mode
	testEnvOverrides map[string]string // Test-specific environment variable overrides
	testWorkingDir   string            // Test working directory override
	testMutex        sync.RWMutex      // Protects test environment overrides

	// Allowed global variables configuration
	allowedGlobalVariables []string // Defines which global variables (starting with _) can be set by users

	// Default command configuration
	defaultCommand string // Command to use when input doesn't start with \\

	// Read-only command overrides
	readOnlyOverrides map[string]bool // Dynamic overrides: true=readonly, false=writable
	readOnlyMutex     sync.RWMutex    // Protects readOnlyOverrides
}

// NewConfigurationSubcontext creates a new ConfigurationSubcontext instance.
func NewConfigurationSubcontext() ConfigurationSubcontext {
	return &configurationSubcontext{
		configMap:        make(map[string]string),
		parentContext:    nil, // Will be set when attached to a parent context
		testEnvOverrides: make(map[string]string),
		allowedGlobalVariables: []string{
			"_style",
			"_reply_way",
			"_echo_command",
			"_render_markdown",
			"_default_command",
			"_stream",
			"_editor",
			"_session_autosave",
			"_completion_mode",
			// Shell prompt configuration variables
			"_prompt_lines_count",
			"_prompt_line1",
			"_prompt_line2",
			"_prompt_line3",
			"_prompt_line4",
			"_prompt_line5",
		},
		defaultCommand:    "echo", // Default to echo for development convenience
		readOnlyOverrides: make(map[string]bool),
	}
}

// NewConfigurationSubcontextFromContext creates a ConfigurationSubcontext from an existing NeuroContext.
// This is used by services to get a reference to the context's configuration subcontext.
func NewConfigurationSubcontextFromContext(ctx *NeuroContext) ConfigurationSubcontext {
	return ctx.configurationCtx
}

// IsTestMode returns whether the configuration subcontext is in test mode by checking the parent context.
func (c *configurationSubcontext) IsTestMode() bool {
	if c.parentContext != nil {
		return c.parentContext.IsTestMode()
	}
	return false // Default to false if no parent context
}

// GetConfigMap returns a copy of the configuration map.
func (c *configurationSubcontext) GetConfigMap() map[string]string {
	c.configMutex.RLock()
	defer c.configMutex.RUnlock()

	result := make(map[string]string)
	for key, value := range c.configMap {
		result[key] = value
	}
	return result
}

// SetConfigMap replaces the entire configuration map.
func (c *configurationSubcontext) SetConfigMap(configMap map[string]string) {
	c.configMutex.Lock()
	defer c.configMutex.Unlock()

	c.configMap = make(map[string]string)
	for key, value := range configMap {
		c.configMap[key] = value
	}
}

// GetConfigValue retrieves a configuration value by key.
func (c *configurationSubcontext) GetConfigValue(key string) (string, bool) {
	c.configMutex.RLock()
	defer c.configMutex.RUnlock()

	value, exists := c.configMap[key]
	return value, exists
}

// SetConfigValue sets a configuration value.
func (c *configurationSubcontext) SetConfigValue(key, value string) {
	c.configMutex.Lock()
	defer c.configMutex.Unlock()

	c.configMap[key] = value
}

// LoadDefaults sets up default configuration values.
func (c *configurationSubcontext) LoadDefaults() error {
	defaults := map[string]string{
		"NEURO_API_KEY":    "",
		"NEURO_MODEL":      "claude-3-5-sonnet-20241022",
		"NEURO_MAX_TOKENS": "8192",
	}

	for key, value := range defaults {
		c.SetConfigValue(key, value)
	}
	return nil
}

// LoadConfigDotEnv loads .env file from the user's config directory (~/.config/neuroshell/.env).
func (c *configurationSubcontext) LoadConfigDotEnv() error {
	configDir, err := c.GetUserConfigDir()
	if err != nil {
		// Config directory access failure is not fatal
		return nil
	}

	envPath := filepath.Join(configDir, ".env")
	if !c.FileExists(envPath) {
		// Missing config .env file is not an error
		return nil
	}

	return c.loadDotEnvFile(envPath)
}

// LoadLocalDotEnv loads .env file from the current working directory.
func (c *configurationSubcontext) LoadLocalDotEnv() error {
	workDir, err := c.GetWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	envPath := filepath.Join(workDir, ".env")
	if !c.FileExists(envPath) {
		// Missing local .env file is not an error
		return nil
	}

	return c.loadDotEnvFile(envPath)
}

// LoadEnvironmentVariables loads specific prefixed environment variables into context configuration map.
// This has the highest priority and will override all file-based configuration.
func (c *configurationSubcontext) LoadEnvironmentVariables(prefixes []string) error {
	// In test mode, only check test environment overrides for clean testing
	if c.IsTestMode() {
		c.testMutex.RLock()
		for key, value := range c.testEnvOverrides {
			for _, prefix := range prefixes {
				if strings.HasPrefix(key, prefix) {
					c.SetConfigValue(key, value)
					break
				}
			}
		}
		c.testMutex.RUnlock()
		return nil // Don't load OS environment variables in test mode
	}

	// In production mode, check actual OS environment variables
	environ := os.Environ()
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			for _, prefix := range prefixes {
				if strings.HasPrefix(key, prefix) {
					// Store in configuration map (highest priority)
					c.SetConfigValue(key, value)
				}
			}
		}
	}

	return nil
}

// LoadEnvironmentVariablesWithPrefix loads OS environment variables with a source prefix.
// Used by Configuration Service for multi-source API key collection.
func (c *configurationSubcontext) LoadEnvironmentVariablesWithPrefix(sourcePrefix string) error {
	// In test mode, only load test environment overrides for clean testing
	if c.IsTestMode() {
		c.testMutex.RLock()
		for key, value := range c.testEnvOverrides {
			prefixedKey := sourcePrefix + key
			c.SetConfigValue(prefixedKey, value)
		}
		c.testMutex.RUnlock()
		return nil // Don't load OS environment variables in test mode
	}

	// Load actual OS environment variables with prefix (production mode only)
	environ := os.Environ()
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			prefixedKey := sourcePrefix + key
			c.SetConfigValue(prefixedKey, value)
		}
	}

	return nil
}

// LoadConfigDotEnvWithPrefix loads config .env file with a source prefix.
// Used by Configuration Service for multi-source API key collection.
func (c *configurationSubcontext) LoadConfigDotEnvWithPrefix(sourcePrefix string) error {
	// In test mode, don't load actual config .env files for clean testing
	if c.IsTestMode() {
		return nil // Don't load config .env files in test mode
	}

	configDir, err := c.GetUserConfigDir()
	if err != nil {
		return nil // Config directory access failure is not fatal
	}

	envPath := filepath.Join(configDir, ".env")
	if !c.FileExists(envPath) {
		return nil // Missing config .env file is not an error
	}

	return c.loadDotEnvFileWithPrefix(envPath, sourcePrefix)
}

// LoadLocalDotEnvWithPrefix loads local .env file with a source prefix.
// Used by Configuration Service for multi-source API key collection.
func (c *configurationSubcontext) LoadLocalDotEnvWithPrefix(sourcePrefix string) error {
	// In test mode, don't load actual local .env files for clean testing
	if c.IsTestMode() {
		return nil // Don't load local .env files in test mode
	}

	workDir, err := c.GetWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	envPath := filepath.Join(workDir, ".env")
	if !c.FileExists(envPath) {
		return nil // Missing local .env file is not an error
	}

	return c.loadDotEnvFileWithPrefix(envPath, sourcePrefix)
}

// loadDotEnvFile loads a specific .env file and stores all values in context configuration map.
// This is a private helper method used by LoadConfigDotEnv and LoadLocalDotEnv.
func (c *configurationSubcontext) loadDotEnvFile(envPath string) error {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file %s: %w", envPath, err)
	}

	// Parse .env file
	envMap, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse .env file %s: %w", envPath, err)
	}

	// Store all values in context configuration map
	for key, value := range envMap {
		c.SetConfigValue(key, value)
	}

	return nil
}

// loadDotEnvFileWithPrefix loads a .env file and stores values with a source prefix.
func (c *configurationSubcontext) loadDotEnvFileWithPrefix(envPath, sourcePrefix string) error {
	data, err := os.ReadFile(envPath)
	if err != nil {
		return fmt.Errorf("failed to read .env file %s: %w", envPath, err)
	}

	// Parse .env file
	envMap, err := godotenv.Unmarshal(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse .env file %s: %w", envPath, err)
	}

	// Store all values with source prefix in context configuration map
	for key, value := range envMap {
		prefixedKey := sourcePrefix + key
		c.SetConfigValue(prefixedKey, value)
	}

	return nil
}

// GetEnv retrieves environment variables, providing test mode appropriate values.
// In test mode, returns predefined test values. In normal mode, returns os.Getenv().
func (c *configurationSubcontext) GetEnv(key string) string {
	if c.IsTestMode() {
		return c.getTestEnvValue(key)
	}
	return os.Getenv(key)
}

// getTestEnvValue returns test mode appropriate values for environment variables.
func (c *configurationSubcontext) getTestEnvValue(key string) string {
	c.testMutex.RLock()
	defer c.testMutex.RUnlock()

	// Check test overrides first
	if value, exists := c.testEnvOverrides[key]; exists {
		return value
	}

	// Default test values
	switch key {
	case "OPENAI_API_KEY":
		return "test-openai-key"
	case "ANTHROPIC_API_KEY":
		return "test-anthropic-key"
	case "GOOGLE_API_KEY":
		return "test-google-key"
	case "EDITOR":
		return "test-editor"
	case "GOOS":
		return "test-os"
	case "GOARCH":
		return "test-arch"
	default:
		return ""
	}
}

// SetTestEnvOverride sets a test-specific environment variable override.
// This allows tests to control what GetEnv returns for specific keys without affecting the OS environment.
func (c *configurationSubcontext) SetTestEnvOverride(key, value string) {
	c.testMutex.Lock()
	defer c.testMutex.Unlock()
	c.testEnvOverrides[key] = value
}

// SetEnvVariable sets an environment variable, respecting test mode.
// In test mode, this sets a test environment override.
// In production mode, this sets an actual OS environment variable.
func (c *configurationSubcontext) SetEnvVariable(key, value string) error {
	if c.IsTestMode() {
		// In test mode, set test environment override
		c.SetTestEnvOverride(key, value)
		return nil
	}

	// In production mode, set actual OS environment variable
	return os.Setenv(key, value)
}

// GetEnvVariable retrieves an environment variable value, respecting test mode.
// This is a pure function that only gets the environment variable without side effects.
func (c *configurationSubcontext) GetEnvVariable(key string) string {
	return c.GetEnv(key)
}

// ClearTestEnvOverride removes a test-specific environment variable override.
func (c *configurationSubcontext) ClearTestEnvOverride(key string) {
	c.testMutex.Lock()
	defer c.testMutex.Unlock()
	delete(c.testEnvOverrides, key)
}

// ClearAllTestEnvOverrides removes all test-specific environment variable overrides.
func (c *configurationSubcontext) ClearAllTestEnvOverrides() {
	c.testMutex.Lock()
	defer c.testMutex.Unlock()
	c.testEnvOverrides = make(map[string]string)
}

// GetTestEnvOverrides returns a copy of all test environment variable overrides.
func (c *configurationSubcontext) GetTestEnvOverrides() map[string]string {
	c.testMutex.RLock()
	defer c.testMutex.RUnlock()

	overrides := make(map[string]string)
	for key, value := range c.testEnvOverrides {
		overrides[key] = value
	}
	return overrides
}

// SetTestWorkingDir sets the test working directory override for test mode.
func (c *configurationSubcontext) SetTestWorkingDir(path string) {
	c.testMutex.Lock()
	defer c.testMutex.Unlock()
	c.testWorkingDir = path
}

// SetParentContext sets the parent context reference for accessing test mode.
func (c *configurationSubcontext) SetParentContext(parent TestModeProvider) {
	c.testMutex.Lock()
	defer c.testMutex.Unlock()
	c.parentContext = parent
}

// GetUserConfigDir returns the user's configuration directory.
// In test mode, returns a temporary directory to avoid polluting the user's system.
func (c *configurationSubcontext) GetUserConfigDir() (string, error) {
	if c.IsTestMode() {
		// In test mode, return a predictable test path
		return "/tmp/neuroshell-test-config", nil
	}

	// Get XDG config home or fall back to ~/.config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configHome, "neuroshell"), nil
}

// GetWorkingDir returns the current working directory.
// In test mode, returns the test working directory if set, otherwise returns default test path.
func (c *configurationSubcontext) GetWorkingDir() (string, error) {
	if c.IsTestMode() {
		c.testMutex.RLock()
		testWorkDir := c.testWorkingDir
		c.testMutex.RUnlock()

		if testWorkDir != "" {
			return testWorkDir, nil
		}
		// Return default test working directory if test mode is on but no specific path is set
		return "/tmp/neuroshell-test-workdir", nil
	}

	return os.Getwd()
}

// FileExists checks if a file or directory exists at the given path.
func (c *configurationSubcontext) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads the contents of a file, supporting test mode isolation.
func (c *configurationSubcontext) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file with the specified permissions, supporting test mode isolation.
func (c *configurationSubcontext) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// MkdirAll creates a directory path with the specified permissions, including any necessary parents.
func (c *configurationSubcontext) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Allowed global variables management

// IsAllowedGlobalVariable checks if a variable name is in the allowed global variables list.
func (c *configurationSubcontext) IsAllowedGlobalVariable(name string) bool {
	for _, allowedVar := range c.allowedGlobalVariables {
		if name == allowedVar {
			return true
		}
	}
	return false
}

// GetAllowedGlobalVariables returns a copy of the allowed global variables list.
func (c *configurationSubcontext) GetAllowedGlobalVariables() []string {
	result := make([]string, len(c.allowedGlobalVariables))
	copy(result, c.allowedGlobalVariables)
	return result
}

// Default command configuration

// GetDefaultCommand returns the default command to use when input doesn't start with \\
func (c *configurationSubcontext) GetDefaultCommand() string {
	return c.defaultCommand
}

// SetDefaultCommand sets the default command to use when input doesn't start with \\
func (c *configurationSubcontext) SetDefaultCommand(command string) {
	c.defaultCommand = command
}

// Read-only command overrides

// SetCommandReadOnly sets or removes a read-only override for a specific command.
// This allows dynamic configuration of read-only status at runtime.
func (c *configurationSubcontext) SetCommandReadOnly(commandName string, readOnly bool) {
	c.readOnlyMutex.Lock()
	defer c.readOnlyMutex.Unlock()

	c.readOnlyOverrides[commandName] = readOnly
}

// RemoveCommandReadOnlyOverride removes any read-only override for a command,
// reverting to the command's self-declared IsReadOnly() status.
func (c *configurationSubcontext) RemoveCommandReadOnlyOverride(commandName string) {
	c.readOnlyMutex.Lock()
	defer c.readOnlyMutex.Unlock()

	delete(c.readOnlyOverrides, commandName)
}

// IsCommandReadOnlyOverride checks if there is a read-only override for a command.
// Returns (override_value, has_override).
func (c *configurationSubcontext) IsCommandReadOnlyOverride(commandName string) (bool, bool) {
	c.readOnlyMutex.RLock()
	defer c.readOnlyMutex.RUnlock()

	override, exists := c.readOnlyOverrides[commandName]
	return override, exists
}

// GetReadOnlyOverrides returns a copy of all current read-only overrides.
// This is useful for configuration services and debugging.
func (c *configurationSubcontext) GetReadOnlyOverrides() map[string]bool {
	c.readOnlyMutex.RLock()
	defer c.readOnlyMutex.RUnlock()

	// Return a copy to prevent external modification
	overrides := make(map[string]bool)
	for name, readOnly := range c.readOnlyOverrides {
		overrides[name] = readOnly
	}
	return overrides
}

// ClearAllReadOnlyOverrides removes all read-only overrides.
// This is useful for testing purposes.
func (c *configurationSubcontext) ClearAllReadOnlyOverrides() {
	c.readOnlyMutex.Lock()
	defer c.readOnlyMutex.Unlock()

	c.readOnlyOverrides = make(map[string]bool)
}
