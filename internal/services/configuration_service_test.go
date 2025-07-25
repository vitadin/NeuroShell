package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
)

func TestConfigurationService_Name(t *testing.T) {
	service := NewConfigurationService()
	assert.Equal(t, "configuration", service.Name())
}

func TestConfigurationService_Initialize(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful initialization",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = context.NewTestContext()
			service := NewConfigurationService()
			err := service.Initialize()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, service.initialized)
			}
		})
	}
}

func TestConfigurationService_ConfigurationPriority(t *testing.T) {
	// Test the priority system: env vars > local .env > config .env > defaults
	ctx := context.NewTestContext()

	// Create temporary directories for test isolation
	tempConfigDir := "/tmp/neuroshell-test-config"
	tempWorkDir := "/tmp/neuroshell-test-workdir"

	// Ensure clean test directories
	_ = os.RemoveAll(tempConfigDir)
	_ = os.RemoveAll(tempWorkDir)
	err := os.MkdirAll(tempConfigDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(tempWorkDir, 0755)
	require.NoError(t, err)

	// Clean up after test
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
		_ = os.RemoveAll(tempWorkDir)
	}()

	// Create config .env file (lower priority) with all types of keys
	configEnvContent := `NEURO_OPENAI_API_KEY=config-key
OPENAI_API_KEY=config-legacy-key
NEURO_TIMEOUT=30s
NEURO_CONFIG_ONLY=config-value
RANDOM_KEY=config-random
`
	err = os.WriteFile(filepath.Join(tempConfigDir, ".env"), []byte(configEnvContent), 0644)
	require.NoError(t, err)

	// Create local .env file (higher priority) with all types of keys
	localEnvContent := `NEURO_OPENAI_API_KEY=local-key
ANTHROPIC_API_KEY=local-anthropic-key
NEURO_LOG_LEVEL=debug
NEURO_LOCAL_ONLY=local-value
ANOTHER_RANDOM_KEY=local-random
`
	err = os.WriteFile(filepath.Join(tempWorkDir, ".env"), []byte(localEnvContent), 0644)
	require.NoError(t, err)

	// Set environment variable (highest priority) - only prefixed ones will be loaded
	ctx.SetTestEnvOverride("NEURO_OPENAI_API_KEY", "env-key")
	ctx.SetTestEnvOverride("OPENAI_API_KEY", "env-legacy-key")
	defer ctx.ClearAllTestEnvOverrides()

	service := NewConfigurationService()
	err = service.Initialize()
	require.NoError(t, err)

	// Test priority: env var should override local .env, which should override config .env
	apiKey, err := service.GetAPIKey("openai")
	require.NoError(t, err)
	assert.Equal(t, "env-key", apiKey, "Environment variable should have highest priority")

	// Test legacy key priority
	legacyKey, err := service.GetConfigValue("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "env-legacy-key", legacyKey, "Environment variable should override config .env")

	// Test config-only value (should be present)
	configOnly, err := service.GetConfigValue("NEURO_CONFIG_ONLY")
	require.NoError(t, err)
	assert.Equal(t, "config-value", configOnly)

	// Test local-only value (should be present)
	localOnly, err := service.GetConfigValue("NEURO_LOCAL_ONLY")
	require.NoError(t, err)
	assert.Equal(t, "local-value", localOnly)

	// Test local Anthropic key (should be present)
	anthropicKey, err := service.GetConfigValue("ANTHROPIC_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "local-anthropic-key", anthropicKey)

	// Test that all .env file keys are loaded (even non-prefixed ones)
	randomKey, err := service.GetConfigValue("RANDOM_KEY")
	require.NoError(t, err)
	assert.Equal(t, "config-random", randomKey)

	anotherRandomKey, err := service.GetConfigValue("ANOTHER_RANDOM_KEY")
	require.NoError(t, err)
	assert.Equal(t, "local-random", anotherRandomKey)
}

func TestConfigurationService_GetAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, service *ConfigurationService)
		provider  string
		expectKey string
		expectErr bool
	}{
		{
			name: "get provider-specific API key with NEURO_ prefix",
			setup: func(t *testing.T, service *ConfigurationService) {
				err := service.SetConfigValue("NEURO_OPENAI_API_KEY", "sk-test-openai-key")
				require.NoError(t, err)
			},
			provider:  "openai",
			expectKey: "sk-test-openai-key",
			expectErr: false,
		},
		{
			name: "get provider-specific API key with legacy format",
			setup: func(t *testing.T, service *ConfigurationService) {
				err := service.SetConfigValue("ANTHROPIC_API_KEY", "anthropic-legacy-key")
				require.NoError(t, err)
			},
			provider:  "anthropic",
			expectKey: "anthropic-legacy-key",
			expectErr: false,
		},
		{
			name: "get generic API key when provider-specific not available",
			setup: func(t *testing.T, service *ConfigurationService) {
				err := service.SetConfigValue("NEURO_API_KEY", "generic-api-key")
				require.NoError(t, err)
			},
			provider:  "anthropic",
			expectKey: "generic-api-key",
			expectErr: false,
		},
		{
			name: "NEURO_ prefixed key takes precedence over legacy",
			setup: func(t *testing.T, service *ConfigurationService) {
				err := service.SetConfigValue("OPENAI_API_KEY", "legacy-key")
				require.NoError(t, err)
				err = service.SetConfigValue("NEURO_OPENAI_API_KEY", "neuro-specific-key")
				require.NoError(t, err)
			},
			provider:  "openai",
			expectKey: "neuro-specific-key",
			expectErr: false,
		},
		{
			name: "error when no API key configured",
			setup: func(_ *testing.T, _ *ConfigurationService) {
				// No API key setup - ensure no test env vars either
			},
			provider:  "moonshot",
			expectKey: "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.NewTestContext()
			// Clear any default test environment variables
			ctx.ClearAllTestEnvOverrides()

			service := NewConfigurationService()
			err := service.Initialize()
			require.NoError(t, err)

			tt.setup(t, service)

			key, err := service.GetAPIKey(tt.provider)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, "", key)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectKey, key)
			}
		})
	}
}

func TestConfigurationService_GetConfigValue(t *testing.T) {
	_ = context.NewTestContext()
	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test getting a default value
	logLevel, err := service.GetConfigValue("NEURO_LOG_LEVEL")
	require.NoError(t, err)
	assert.Equal(t, "info", logLevel)

	// Test getting a non-existent value
	nonExistent, err := service.GetConfigValue("NEURO_NON_EXISTENT")
	require.NoError(t, err)
	assert.Equal(t, "", nonExistent)

	// Test setting and getting a custom value
	err = service.SetConfigValue("NEURO_CUSTOM", "custom-value")
	require.NoError(t, err)

	customValue, err := service.GetConfigValue("NEURO_CUSTOM")
	require.NoError(t, err)
	assert.Equal(t, "custom-value", customValue)
}

func TestConfigurationService_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, service *ConfigurationService)
		wantErr bool
	}{
		{
			name: "valid configuration with NEURO_ prefixed API key",
			setup: func(t *testing.T, service *ConfigurationService) {
				err := service.SetConfigValue("NEURO_OPENAI_API_KEY", "sk-valid-openai-key-1234567890")
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "valid configuration with legacy API key",
			setup: func(t *testing.T, service *ConfigurationService) {
				err := service.SetConfigValue("ANTHROPIC_API_KEY", "ant-valid-anthropic-key-1234567890")
				require.NoError(t, err)
			},
			wantErr: false,
		},
		{
			name: "valid configuration with no API keys",
			setup: func(_ *testing.T, _ *ConfigurationService) {
				// No API keys - should still be valid
			},
			wantErr: false,
		},
		{
			name: "invalid configuration with too short NEURO_ API key",
			setup: func(t *testing.T, service *ConfigurationService) {
				err := service.SetConfigValue("NEURO_OPENAI_API_KEY", "short")
				require.NoError(t, err)
			},
			wantErr: true,
		},
		{
			name: "invalid configuration with too short legacy API key",
			setup: func(t *testing.T, service *ConfigurationService) {
				err := service.SetConfigValue("ANTHROPIC_API_KEY", "short")
				require.NoError(t, err)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = context.NewTestContext()
			service := NewConfigurationService()
			err := service.Initialize()
			require.NoError(t, err)

			tt.setup(t, service)

			err = service.ValidateConfiguration()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigurationService_GetAllConfigValues(t *testing.T) {
	_ = context.NewTestContext()
	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Set some test values (mix of NEURO_ and non-NEURO_)
	err = service.SetConfigValue("NEURO_TEST_KEY", "test-value")
	require.NoError(t, err)
	err = service.SetConfigValue("NEURO_ANOTHER_KEY", "another-value")
	require.NoError(t, err)
	err = service.SetConfigValue("NON_NEURO_KEY", "should-appear-too")
	require.NoError(t, err)

	configValues, err := service.GetAllConfigValues()
	require.NoError(t, err)

	// Should contain NEURO_ prefixed values (including defaults)
	assert.Contains(t, configValues, "NEURO_TEST_KEY")
	assert.Equal(t, "test-value", configValues["NEURO_TEST_KEY"])

	assert.Contains(t, configValues, "NEURO_ANOTHER_KEY")
	assert.Equal(t, "another-value", configValues["NEURO_ANOTHER_KEY"])

	assert.Contains(t, configValues, "NEURO_LOG_LEVEL")
	assert.Equal(t, "info", configValues["NEURO_LOG_LEVEL"])

	// Should also contain non-NEURO_ prefixed values since we load complete .env files
	assert.Contains(t, configValues, "NON_NEURO_KEY")
	assert.Equal(t, "should-appear-too", configValues["NON_NEURO_KEY"])
}

func TestConfigurationService_TestModeIsolation(t *testing.T) {
	// This test verifies that test mode doesn't pollute the user's actual config
	ctx := context.NewTestContext()
	assert.True(t, ctx.IsTestMode(), "Context should be in test mode")

	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Verify that the config directory paths are test-specific
	configDir, err := ctx.GetUserConfigDir()
	require.NoError(t, err)
	assert.Contains(t, configDir, "test", "Config directory should contain 'test' in test mode")

	workDir, err := ctx.GetWorkingDir()
	require.NoError(t, err)
	assert.Contains(t, workDir, "test", "Working directory should contain 'test' in test mode")
}

func TestConfigurationService_EnvironmentVariableOverride(t *testing.T) {
	ctx := context.NewTestContext()

	// Set up test environment variables with proper prefixes
	ctx.SetTestEnvOverride("NEURO_OVERRIDE_TEST", "env-override-value")
	ctx.SetTestEnvOverride("OPENAI_API_KEY", "env-openai-key")
	ctx.SetTestEnvOverride("ANTHROPIC_API_KEY", "env-anthropic-key")
	defer ctx.ClearAllTestEnvOverrides()

	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Should get the environment variable values for prefixed vars
	value, err := service.GetConfigValue("NEURO_OVERRIDE_TEST")
	require.NoError(t, err)
	assert.Equal(t, "env-override-value", value)

	openaiKey, err := service.GetConfigValue("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "env-openai-key", openaiKey)

	anthropicKey, err := service.GetConfigValue("ANTHROPIC_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "env-anthropic-key", anthropicKey)
}

func TestConfigurationService_EnvironmentVariablePrefixFiltering(t *testing.T) {
	ctx := context.NewTestContext()

	// Set up environment variables - only prefixed ones should be loaded
	ctx.SetTestEnvOverride("NEURO_SHOULD_LOAD", "loaded")
	ctx.SetTestEnvOverride("OPENAI_SHOULD_LOAD", "loaded")
	ctx.SetTestEnvOverride("ANTHROPIC_SHOULD_LOAD", "loaded")
	ctx.SetTestEnvOverride("RANDOM_SHOULD_NOT_LOAD", "not-loaded")
	defer ctx.ClearAllTestEnvOverrides()

	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Should load prefixed environment variables
	neuroValue, err := service.GetConfigValue("NEURO_SHOULD_LOAD")
	require.NoError(t, err)
	assert.Equal(t, "loaded", neuroValue)

	openaiValue, err := service.GetConfigValue("OPENAI_SHOULD_LOAD")
	require.NoError(t, err)
	assert.Equal(t, "loaded", openaiValue)

	anthropicValue, err := service.GetConfigValue("ANTHROPIC_SHOULD_LOAD")
	require.NoError(t, err)
	assert.Equal(t, "loaded", anthropicValue)

	// Should NOT load non-prefixed environment variables
	randomValue, err := service.GetConfigValue("RANDOM_SHOULD_NOT_LOAD")
	require.NoError(t, err)
	assert.Equal(t, "", randomValue)
}

func TestConfigurationService_LoadConfiguration(t *testing.T) {
	tempConfigDir := "/tmp/neuroshell-test-config"

	// Clean up
	_ = os.RemoveAll(tempConfigDir)
	err := os.MkdirAll(tempConfigDir, 0755)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
	}()

	_ = context.NewTestContext()
	service := NewConfigurationService()
	err = service.Initialize()
	require.NoError(t, err)

	// Initially should have default timeout
	timeout, err := service.GetConfigValue("NEURO_TIMEOUT")
	require.NoError(t, err)
	assert.Equal(t, "30s", timeout)

	// Create a new config file
	envContent := "NEURO_TIMEOUT=120s\nNEURO_NEW_KEY=new-value\n"
	err = os.WriteFile(filepath.Join(tempConfigDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	// Reload configuration
	err = service.LoadConfiguration()
	require.NoError(t, err)

	// Should now have updated values
	timeout, err = service.GetConfigValue("NEURO_TIMEOUT")
	require.NoError(t, err)
	assert.Equal(t, "120s", timeout)

	newKey, err := service.GetConfigValue("NEURO_NEW_KEY")
	require.NoError(t, err)
	assert.Equal(t, "new-value", newKey)
}
