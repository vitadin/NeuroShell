package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
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
	context.SetGlobalContext(ctx)

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
MOONSHOT_BASE_URL=config-moonshot-url
RANDOM_KEY=config-random
`
	err = os.WriteFile(filepath.Join(tempConfigDir, ".env"), []byte(configEnvContent), 0644)
	require.NoError(t, err)

	// Create local .env file (higher priority) with all types of keys
	localEnvContent := `NEURO_OPENAI_API_KEY=local-key
ANTHROPIC_API_KEY=local-anthropic-key
NEURO_LOG_LEVEL=debug
NEURO_LOCAL_ONLY=local-value
MOONSHOT_API_KEY=local-moonshot-key
ANOTHER_RANDOM_KEY=local-random
`
	err = os.WriteFile(filepath.Join(tempWorkDir, ".env"), []byte(localEnvContent), 0644)
	require.NoError(t, err)

	// Set test working directory to point to our temporary directory
	neuroCtx := ctx.(*context.NeuroContext)
	configSubctx := context.NewConfigurationSubcontextFromContext(neuroCtx)
	configSubctx.SetTestWorkingDir(tempWorkDir)

	// Set environment variable (highest priority) - only prefixed ones will be loaded
	ctx.SetTestEnvOverride("NEURO_OPENAI_API_KEY", "env-key")
	ctx.SetTestEnvOverride("OPENAI_API_KEY", "env-legacy-key")
	ctx.SetTestEnvOverride("MOONSHOT_BASE_URL", "env-moonshot-url")
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

	// Test MOONSHOT prefix priority: env var > local .env > config .env
	moonshotBaseURL, err := service.GetConfigValue("MOONSHOT_BASE_URL")
	require.NoError(t, err)
	assert.Equal(t, "env-moonshot-url", moonshotBaseURL, "Environment variable should override .env files")

	moonshotAPIKey, err := service.GetConfigValue("MOONSHOT_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "local-moonshot-key", moonshotAPIKey, "Local .env should override config .env when no env var")
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
			context.SetGlobalContext(ctx)
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
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Set up test environment variable
	ctx.SetTestEnvOverride("NEURO_LOG_LEVEL", "info")
	defer ctx.ClearAllTestEnvOverrides()

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
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Set up test environment variable
	ctx.SetTestEnvOverride("NEURO_LOG_LEVEL", "info")
	defer ctx.ClearAllTestEnvOverrides()

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
	context.SetGlobalContext(ctx)
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
	context.SetGlobalContext(ctx)

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
	context.SetGlobalContext(ctx)

	// Set up environment variables - only prefixed ones should be loaded
	ctx.SetTestEnvOverride("NEURO_SHOULD_LOAD", "loaded")
	ctx.SetTestEnvOverride("OPENAI_SHOULD_LOAD", "loaded")
	ctx.SetTestEnvOverride("ANTHROPIC_SHOULD_LOAD", "loaded")
	ctx.SetTestEnvOverride("MOONSHOT_SHOULD_LOAD", "loaded")
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

	moonshotValue, err := service.GetConfigValue("MOONSHOT_SHOULD_LOAD")
	require.NoError(t, err)
	assert.Equal(t, "loaded", moonshotValue)

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

	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Set up test environment variable
	ctx.SetTestEnvOverride("NEURO_TIMEOUT", "30s")
	defer ctx.ClearAllTestEnvOverrides()

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

	// Clear environment variable override to test file loading
	ctx.ClearTestEnvOverride("NEURO_TIMEOUT")

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

func TestConfigurationService_EnvironmentVariablePrefixMatching(t *testing.T) {
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Test all supported prefixes with various environment variables
	testCases := []struct {
		name        string
		envKey      string
		envValue    string
		shouldLoad  bool
		description string
	}{
		// NEURO_ prefix
		{"NEURO_API_KEY", "NEURO_API_KEY", "neuro-key", true, "NEURO_ prefix should be loaded"},
		{"NEURO_BASE_URL", "NEURO_BASE_URL", "https://neuro.example.com", true, "NEURO_BASE_URL should be loaded"},
		{"NEURO_TIMEOUT", "NEURO_TIMEOUT", "60s", true, "NEURO_TIMEOUT should be loaded"},

		// OPENAI_ prefix
		{"OPENAI_API_KEY", "OPENAI_API_KEY", "sk-openai-key", true, "OPENAI_ prefix should be loaded"},
		{"OPENAI_BASE_URL", "OPENAI_BASE_URL", "https://api.openai.com", true, "OPENAI_BASE_URL should be loaded"},
		{"OPENAI_ORG_ID", "OPENAI_ORG_ID", "org-123", true, "OPENAI_ORG_ID should be loaded"},

		// ANTHROPIC_ prefix
		{"ANTHROPIC_API_KEY", "ANTHROPIC_API_KEY", "ant-anthropic-key", true, "ANTHROPIC_ prefix should be loaded"},
		{"ANTHROPIC_BASE_URL", "ANTHROPIC_BASE_URL", "https://api.anthropic.com", true, "ANTHROPIC_BASE_URL should be loaded"},

		// MOONSHOT_ prefix
		{"MOONSHOT_API_KEY", "MOONSHOT_API_KEY", "mk-moonshot-key", true, "MOONSHOT_ prefix should be loaded"},
		{"MOONSHOT_BASE_URL", "MOONSHOT_BASE_URL", "https://api.moonshot.cn", true, "MOONSHOT_BASE_URL should be loaded"},
		{"MOONSHOT_MODEL", "MOONSHOT_MODEL", "moonshot-v1-8k", true, "MOONSHOT_MODEL should be loaded"},

		// Non-prefixed variables (should NOT be loaded from env)
		{"RANDOM_VAR", "RANDOM_VAR", "should-not-load", false, "Non-prefixed env vars should NOT be loaded"},
		{"PATH", "PATH", "/usr/bin", false, "PATH env var should NOT be loaded"},
		{"HOME", "HOME", "/home/user", false, "HOME env var should NOT be loaded"},
		{"USER", "USER", "testuser", false, "USER env var should NOT be loaded"},

		// Partial prefix matches (should NOT be loaded)
		{"NEUR_API_KEY", "NEUR_API_KEY", "should-not-load", false, "Partial prefix match should NOT be loaded"},
		{"OPENAI", "OPENAI", "should-not-load", false, "Exact prefix without underscore should NOT be loaded"},
		{"OPENAI_KEY_TEST", "OPENAIKEY_TEST", "should-not-load", false, "Wrong prefix format should NOT be loaded"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear any previous test env overrides
			ctx.ClearAllTestEnvOverrides()

			// Set the test environment variable
			ctx.SetTestEnvOverride(tc.envKey, tc.envValue)

			service := NewConfigurationService()
			err := service.Initialize()
			require.NoError(t, err)

			value, err := service.GetConfigValue(tc.envKey)
			require.NoError(t, err)

			if tc.shouldLoad {
				assert.Equal(t, tc.envValue, value, tc.description)
			} else {
				assert.Equal(t, "", value, tc.description)
			}

			// Clean up
			ctx.ClearTestEnvOverride(tc.envKey)
		})
	}
}

func TestConfigurationService_AllSupportedPrefixes(t *testing.T) {
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Test that all expected prefixes are supported
	expectedPrefixes := []string{"NEURO_", "OPENAI_", "ANTHROPIC_", "MOONSHOT_", "GOOGLE_"}

	// Set one environment variable for each prefix
	testVars := map[string]string{
		"NEURO_TEST":     "neuro-value",
		"OPENAI_TEST":    "openai-value",
		"ANTHROPIC_TEST": "anthropic-value",
		"MOONSHOT_TEST":  "moonshot-value",
		"GOOGLE_TEST":    "google-value",
	}

	// Set all test environment variables
	for key, value := range testVars {
		ctx.SetTestEnvOverride(key, value)
	}

	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Verify all prefixed variables were loaded
	for key, expectedValue := range testVars {
		value, err := service.GetConfigValue(key)
		require.NoError(t, err)
		assert.Equal(t, expectedValue, value, "Variable %s should be loaded with prefix matching", key)
	}

	// Clean up
	ctx.ClearAllTestEnvOverrides()

	// Verify the expected prefixes list matches what's used internally through context
	globalCtx := context.GetGlobalContext()
	actualPrefixes := globalCtx.GetProviderEnvPrefixes()
	assert.ElementsMatch(t, expectedPrefixes, actualPrefixes, "context provider prefixes should match expected supported prefixes")
}

func TestConfigurationService_MoonshotBaseURLExample(t *testing.T) {
	// This test specifically verifies the example mentioned by the user:
	// MOONSHOT_BASE_URL should be stored in the config map when present as OS env var
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Set the specific example environment variable
	ctx.SetTestEnvOverride("MOONSHOT_BASE_URL", "https://api.moonshot.cn/v1")
	defer ctx.ClearAllTestEnvOverrides()

	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Verify that MOONSHOT_BASE_URL was loaded and stored in config map
	baseURL, err := service.GetConfigValue("MOONSHOT_BASE_URL")
	require.NoError(t, err)
	assert.Equal(t, "https://api.moonshot.cn/v1", baseURL, "MOONSHOT_BASE_URL should be loaded from environment variables")

	// Also verify it appears in the full config map
	allConfig, err := service.GetAllConfigValues()
	require.NoError(t, err)
	assert.Contains(t, allConfig, "MOONSHOT_BASE_URL", "MOONSHOT_BASE_URL should be present in full config map")
	assert.Equal(t, "https://api.moonshot.cn/v1", allConfig["MOONSHOT_BASE_URL"], "MOONSHOT_BASE_URL value should match in full config map")
}

func TestConfigurationService_GetAllAPIKeys(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(t *testing.T, ctx neurotypes.Context)
		expectKeys      int
		expectSources   []string
		expectProviders []string
	}{
		{
			name: "collect variables from OS environment variables",
			setup: func(_ *testing.T, ctx neurotypes.Context) {
				ctx.SetTestEnvOverride("A_OPENAI_KEY", "sk-1234567890abcdef")
				ctx.SetTestEnvOverride("MY_ANTHROPIC_API_KEY", "ant-1234567890abcdef")
				ctx.SetTestEnvOverride("OPENROUTER_SECRET", "or-1234567890abcdef")
			},
			expectKeys:      3,
			expectSources:   []string{"os", "os", "os"},
			expectProviders: []string{"", "", ""}, // Provider detection moved to command layer
		},
		{
			name: "skip short variables",
			setup: func(_ *testing.T, ctx neurotypes.Context) {
				ctx.SetTestEnvOverride("OPENAI_API_KEY", "short")                 // Too short, should be skipped
				ctx.SetTestEnvOverride("VALID_OPENAI_KEY", "sk-1234567890abcdef") // Valid length
			},
			expectKeys:      1,
			expectSources:   []string{"os"},
			expectProviders: []string{""}, // Provider detection moved to command layer
		},
		{
			name: "skip empty variables",
			setup: func(_ *testing.T, ctx neurotypes.Context) {
				ctx.SetTestEnvOverride("OPENAI_API_KEY", "")                      // Empty, should be skipped
				ctx.SetTestEnvOverride("ANTHROPIC_KEY", "   ")                    // Whitespace only, should be skipped
				ctx.SetTestEnvOverride("MOONSHOT_API_KEY", "mk-1234567890abcdef") // Valid
			},
			expectKeys:      1,
			expectSources:   []string{"os"},
			expectProviders: []string{""}, // Provider detection moved to command layer
		},
		{
			name: "case insensitive variable collection",
			setup: func(_ *testing.T, ctx neurotypes.Context) {
				ctx.SetTestEnvOverride("UPPER_OPENAI_KEY", "sk-1234567890abcdef")
				ctx.SetTestEnvOverride("lower_anthropic_key", "ant-1234567890abcdef")
				ctx.SetTestEnvOverride("MiXeD_OpenRouter_Key", "or-1234567890abcdef")
			},
			expectKeys:      3,
			expectSources:   []string{"os", "os", "os"},
			expectProviders: []string{"", "", ""}, // Provider detection moved to command layer
		},
		{
			name: "collect all variables regardless of content",
			setup: func(_ *testing.T, ctx neurotypes.Context) {
				ctx.SetTestEnvOverride("RANDOM_KEY", "some-random-value-1234567890")
				ctx.SetTestEnvOverride("ANOTHER_VAR", "another-value-1234567890")
			},
			expectKeys:      2, // Now collects all variables with sufficient length
			expectSources:   []string{"os", "os"},
			expectProviders: []string{"", ""}, // Provider detection moved to command layer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.NewTestContext()
			context.SetGlobalContext(ctx)
			ctx.ClearAllTestEnvOverrides()

			service := NewConfigurationService()
			err := service.Initialize()
			require.NoError(t, err)

			tt.setup(t, ctx)

			keys, err := service.GetAllAPIKeys()
			require.NoError(t, err)

			assert.Equal(t, tt.expectKeys, len(keys), "Number of keys should match")

			if tt.expectKeys > 0 {
				// Verify sources and providers
				actualSources := make([]string, len(keys))
				actualProviders := make([]string, len(keys))

				for i, key := range keys {
					actualSources[i] = key.Source
					actualProviders[i] = key.Provider

					// Verify key properties
					assert.NotEmpty(t, key.OriginalName, "OriginalName should not be empty")
					assert.NotEmpty(t, key.Value, "Value should not be empty")
					assert.GreaterOrEqual(t, len(key.Value), 10, "Value should be at least 10 characters")
				}

				assert.ElementsMatch(t, tt.expectSources, actualSources, "Sources should match")
				assert.ElementsMatch(t, tt.expectProviders, actualProviders, "Providers should match")
			}

			// Clean up
			ctx.ClearAllTestEnvOverrides()
		})
	}
}

func TestConfigurationService_GetAllAPIKeys_MultipleProviderMatches(t *testing.T) {
	// Test that configuration service collects variables regardless of provider names
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)
	ctx.ClearAllTestEnvOverrides()

	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Set an env var that contains multiple provider names
	ctx.SetTestEnvOverride("OPENAI_ANTHROPIC_COMBINED_KEY", "sk-1234567890abcdef")

	keys, err := service.GetAllAPIKeys()
	require.NoError(t, err)

	assert.Equal(t, 1, len(keys), "Should find exactly one key")

	if len(keys) > 0 {
		key := keys[0]
		assert.Equal(t, "os", key.Source)
		assert.Equal(t, "OPENAI_ANTHROPIC_COMBINED_KEY", key.OriginalName)
		assert.Equal(t, "sk-1234567890abcdef", key.Value)
		// Provider detection moved to command layer
		assert.Equal(t, "", key.Provider, "Provider should be empty string")
	}

	ctx.ClearAllTestEnvOverrides()
}

func TestConfigurationService_GetAllAPIKeys_ConfigAndLocalEnvFiles(t *testing.T) {
	// This test would require creating actual .env files in test directories
	// For now, we'll test the logic assuming the files exist and are readable
	_ = context.NewTestContext()

	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test the method doesn't crash when files don't exist
	keys, err := service.GetAllAPIKeys()
	require.NoError(t, err)
	// Should return empty slice when no files exist and no env vars are set
	assert.IsType(t, []APIKeySource{}, keys)
}

func TestConfigurationService_GetAllAPIKeys_NotInitialized(t *testing.T) {
	service := NewConfigurationService()
	// Don't initialize the service

	keys, err := service.GetAllAPIKeys()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
	assert.Nil(t, keys)
}

func TestConfigurationService_ReadOnlyOverrides(t *testing.T) {
	_ = context.NewTestContext()
	service := NewConfigurationService()
	err := service.Initialize()
	require.NoError(t, err)

	// Initially should have no overrides
	overrides, err := service.GetReadOnlyOverrides()
	require.NoError(t, err)
	assert.Empty(t, overrides)

	// Set a read-only override
	err = service.SetReadOnlyOverride("test-command", true)
	require.NoError(t, err)

	// Should now appear in overrides
	overrides, err = service.GetReadOnlyOverrides()
	require.NoError(t, err)
	assert.Contains(t, overrides, "test-command")
	assert.True(t, overrides["test-command"])

	// Set another override with different value
	err = service.SetReadOnlyOverride("another-command", false)
	require.NoError(t, err)

	// Should now have both overrides
	overrides, err = service.GetReadOnlyOverrides()
	require.NoError(t, err)
	assert.Len(t, overrides, 2)
	assert.True(t, overrides["test-command"])
	assert.False(t, overrides["another-command"])

	// Remove first override
	err = service.RemoveReadOnlyOverride("test-command")
	require.NoError(t, err)

	// Should now only have the second override
	overrides, err = service.GetReadOnlyOverrides()
	require.NoError(t, err)
	assert.Len(t, overrides, 1)
	assert.Contains(t, overrides, "another-command")
	assert.NotContains(t, overrides, "test-command")
}

func TestConfigurationService_ReadOnlyOverrides_NotInitialized(t *testing.T) {
	service := NewConfigurationService()
	// Don't initialize the service

	// All methods should return errors when not initialized
	_, err := service.GetReadOnlyOverrides()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	err = service.SetReadOnlyOverride("command", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	err = service.RemoveReadOnlyOverride("command")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	err = service.LoadReadOnlyOverrides()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestConfigurationService_LoadReadOnlyOverrides(t *testing.T) {
	tests := []struct {
		name              string
		configValue       string
		expectedOverrides map[string]bool
	}{
		{
			name:        "load single override",
			configValue: "get:true",
			expectedOverrides: map[string]bool{
				"get": true,
			},
		},
		{
			name:        "load multiple overrides",
			configValue: "get:true,set:false,vars:true",
			expectedOverrides: map[string]bool{
				"get":  true,
				"set":  false,
				"vars": true,
			},
		},
		{
			name:        "load with spaces",
			configValue: "get : true , set : false",
			expectedOverrides: map[string]bool{
				"get": true,
				"set": false,
			},
		},
		{
			name:              "empty configuration",
			configValue:       "",
			expectedOverrides: map[string]bool{},
		},
		{
			name:        "malformed entries ignored",
			configValue: "get:true,invalid,set:false,another:invalid,vars:true",
			expectedOverrides: map[string]bool{
				"get":  true,
				"set":  false,
				"vars": true,
			},
		},
		{
			name:              "only command names",
			configValue:       "get,set,vars",
			expectedOverrides: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = context.NewTestContext()
			service := NewConfigurationService()
			err := service.Initialize()
			require.NoError(t, err)

			// Set the NEURO_READONLY_COMMANDS configuration
			err = service.SetConfigValue("NEURO_READONLY_COMMANDS", tt.configValue)
			require.NoError(t, err)

			// Load read-only overrides from configuration
			err = service.LoadReadOnlyOverrides()
			require.NoError(t, err)

			// Check that the overrides match expectations
			overrides, err := service.GetReadOnlyOverrides()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOverrides, overrides)
		})
	}
}
