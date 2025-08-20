package builtin

import (
	"os"
	"strings"
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPathCommand_Name(t *testing.T) {
	cmd := &ConfigPathCommand{}
	assert.Equal(t, "config-path", cmd.Name())
}

func TestConfigPathCommand_Description(t *testing.T) {
	cmd := &ConfigPathCommand{}
	assert.Contains(t, cmd.Description(), "configuration file paths")
}

func TestConfigPathCommand_Usage(t *testing.T) {
	cmd := &ConfigPathCommand{}
	assert.Equal(t, "\\config-path", cmd.Usage())
}

func TestConfigPathCommand_HelpInfo(t *testing.T) {
	cmd := &ConfigPathCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "config-path", helpInfo.Command)
	assert.Contains(t, helpInfo.Description, "configuration file paths")
	assert.Equal(t, "\\config-path", helpInfo.Usage)
	assert.Len(t, helpInfo.Examples, 1)
	assert.Len(t, helpInfo.StoredVariables, 8) // 8 system variables

	// Verify all expected system variables are documented
	expectedVars := []string{"#config_dir", "#config_dir_exists", "#config_env_path", "#config_env_loaded", "#local_env_path", "#local_env_loaded", "#neurorc_path", "#neurorc_executed"}
	for _, expectedVar := range expectedVars {
		found := false
		for _, storedVar := range helpInfo.StoredVariables {
			if storedVar.Name == expectedVar {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected system variable %s not found in help info", expectedVar)
	}
}

func TestConfigPathCommand_Execute_Success(t *testing.T) {
	// Setup test context and services
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Setup service registry
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	// Setup temp directories for test
	tempConfigDir := "/tmp/neuroshell-test-config"
	tempWorkDir := "/tmp/neuroshell-test-workdir"

	// Clean up and create directories
	_ = os.RemoveAll(tempConfigDir)
	_ = os.RemoveAll(tempWorkDir)
	err := os.MkdirAll(tempConfigDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(tempWorkDir, 0755)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
		_ = os.RemoveAll(tempWorkDir)
	}()

	// Create config .env file
	configEnvContent := "NEURO_TEST_KEY=config-value\n"
	err = os.WriteFile(tempConfigDir+"/.env", []byte(configEnvContent), 0644)
	require.NoError(t, err)

	// Create local .env file
	localEnvContent := "LOCAL_TEST_KEY=local-value\n"
	err = os.WriteFile(tempWorkDir+"/.env", []byte(localEnvContent), 0644)
	require.NoError(t, err)

	// Register required services
	err = registry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = registry.RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	// Initialize services
	err = registry.InitializeAll()
	require.NoError(t, err)

	// Set up .neurorc execution tracking
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)
	err = variableService.SetSystemVariable("#neurorc_path", "/home/test/.neurorc")
	require.NoError(t, err)
	err = variableService.SetSystemVariable("#neurorc_executed", "true")
	require.NoError(t, err)

	// Execute the command
	cmd := &ConfigPathCommand{}

	// Capture output
	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(map[string]string{}, "")
		require.NoError(t, err)
	})

	// Verify output contains expected paths
	assert.Contains(t, output, "Config Directory: /tmp/neuroshell-test-config (exists)")
	assert.Contains(t, output, "Config .env: /tmp/neuroshell-test-config/.env (loaded)")
	assert.Contains(t, output, "Local .env: /tmp/neuroshell-test-workdir/.env (loaded)")
	assert.Contains(t, output, ".neurorc: /home/test/.neurorc (executed)")

	// Verify system variables are set
	configDir, err := variableService.Get("#config_dir")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/neuroshell-test-config", configDir)

	configDirExists, err := variableService.Get("#config_dir_exists")
	require.NoError(t, err)
	assert.Equal(t, "true", configDirExists)

	configEnvPath, err := variableService.Get("#config_env_path")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/neuroshell-test-config/.env", configEnvPath)

	configEnvLoaded, err := variableService.Get("#config_env_loaded")
	require.NoError(t, err)
	assert.Equal(t, "true", configEnvLoaded)

	localEnvPath, err := variableService.Get("#local_env_path")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/neuroshell-test-workdir/.env", localEnvPath)

	localEnvLoaded, err := variableService.Get("#local_env_loaded")
	require.NoError(t, err)
	assert.Equal(t, "true", localEnvLoaded)

	neuroRCPath, err := variableService.Get("#neurorc_path")
	require.NoError(t, err)
	assert.Equal(t, "/home/test/.neurorc", neuroRCPath)

	neuroRCExecuted, err := variableService.Get("#neurorc_executed")
	require.NoError(t, err)
	assert.Equal(t, "true", neuroRCExecuted)
}

func TestConfigPathCommand_Execute_NoConfigFiles(t *testing.T) {
	// Setup test context and services
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Setup service registry
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	// Clean up test directories first
	_ = os.RemoveAll("/tmp/neuroshell-test-config")
	_ = os.RemoveAll("/tmp/neuroshell-test-workdir")

	// Register required services
	err := registry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = registry.RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	// Initialize services
	err = registry.InitializeAll()
	require.NoError(t, err)

	// Execute the command without any config files
	cmd := &ConfigPathCommand{}

	// Capture output
	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(map[string]string{}, "")
		require.NoError(t, err)
	})

	// Should still show config directory (not found since we didn't create it)
	assert.Contains(t, output, "Config Directory: /tmp/neuroshell-test-config (not found)")
	// Should show .env files as not found
	assert.Contains(t, output, "Config .env: /tmp/neuroshell-test-config/.env (not found)")
	assert.Contains(t, output, "Local .env: /tmp/neuroshell-test-workdir/.env (not found)")
	// No .neurorc output since no system variables set
}

func TestConfigPathCommand_Execute_PartialConfigFiles(t *testing.T) {
	// Setup test context and services
	ctx := context.NewTestContext()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Setup service registry
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	// Setup temp directories
	tempConfigDir := "/tmp/neuroshell-test-config"
	_ = os.RemoveAll(tempConfigDir)
	_ = os.RemoveAll("/tmp/neuroshell-test-workdir")
	err := os.MkdirAll(tempConfigDir, 0755)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempConfigDir)
	}()

	// Create only config .env file (no local .env)
	configEnvContent := "NEURO_TEST_KEY=config-value\n"
	err = os.WriteFile(tempConfigDir+"/.env", []byte(configEnvContent), 0644)
	require.NoError(t, err)

	// Register required services
	err = registry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = registry.RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	// Initialize services
	err = registry.InitializeAll()
	require.NoError(t, err)

	// Set up only .neurorc path (not executed)
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)
	err = variableService.SetSystemVariable("#neurorc_path", "/home/test/.neurorc")
	require.NoError(t, err)
	err = variableService.SetSystemVariable("#neurorc_executed", "false")
	require.NoError(t, err)

	// Execute the command
	cmd := &ConfigPathCommand{}

	// Capture output
	output := stringprocessing.CaptureOutput(func() {
		err := cmd.Execute(map[string]string{}, "")
		require.NoError(t, err)
	})

	// Verify mixed status
	assert.Contains(t, output, "Config Directory: /tmp/neuroshell-test-config (exists)")
	assert.Contains(t, output, "Config .env: /tmp/neuroshell-test-config/.env (loaded)")
	assert.Contains(t, output, "Local .env: /tmp/neuroshell-test-workdir/.env (not found)")
	assert.Contains(t, output, ".neurorc: /home/test/.neurorc (not executed)")

	// Verify system variables reflect correct status
	configEnvLoaded, err := variableService.Get("#config_env_loaded")
	require.NoError(t, err)
	assert.Equal(t, "true", configEnvLoaded)

	localEnvLoaded, err := variableService.Get("#local_env_loaded")
	require.NoError(t, err)
	assert.Equal(t, "false", localEnvLoaded)
}

func TestConfigPathCommand_Execute_ServiceNotAvailable(t *testing.T) {
	// Setup empty registry (no services)
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	cmd := &ConfigPathCommand{}

	// Should return error when configuration service is not available
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration service not available")
}

func TestConfigPathCommand_Execute_VariableServiceNotAvailable(t *testing.T) {
	// Setup registry with only configuration service
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	err := registry.RegisterService(services.NewConfigurationService())
	require.NoError(t, err)
	err = registry.InitializeAll()
	require.NoError(t, err)

	cmd := &ConfigPathCommand{}

	// Should return error when variable service is not available
	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestConfigPathCommand_Execute_ConfigurationServiceNotInitialized(t *testing.T) {
	// Setup registry with uninitialized configuration service
	registry := services.NewRegistry()
	services.SetGlobalRegistry(registry)

	err := registry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = registry.RegisterService(services.NewConfigurationService())
	require.NoError(t, err)

	// Initialize only variable service, not configuration service
	variableService, err := registry.GetService("variable")
	require.NoError(t, err)
	err = variableService.Initialize()
	require.NoError(t, err)

	cmd := &ConfigPathCommand{}

	// Should return error when configuration service is not initialized
	err = cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "not initialized")
}
