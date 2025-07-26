package builtin

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestGetEnvCommand_Name(t *testing.T) {
	cmd := &GetEnvCommand{}
	assert.Equal(t, "get-env", cmd.Name())
}

func TestGetEnvCommand_ParseMode(t *testing.T) {
	cmd := &GetEnvCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestGetEnvCommand_Description(t *testing.T) {
	cmd := &GetEnvCommand{}
	assert.Equal(t, "Get an environment variable and create #os.VAR neuro variable", cmd.Description())
}

func TestGetEnvCommand_Usage(t *testing.T) {
	cmd := &GetEnvCommand{}
	assert.Equal(t, "\\get-env[VAR] or \\get-env VAR", cmd.Usage())
}

func TestGetEnvCommand_HelpInfo(t *testing.T) {
	cmd := &GetEnvCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "get-env", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Options)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Check that VAR option exists
	found := false
	for _, option := range helpInfo.Options {
		if option.Name == "VAR" {
			found = true
			assert.True(t, option.Required)
			assert.Equal(t, "string", option.Type)
			break
		}
	}
	assert.True(t, found, "VAR option should be present in help info")
}

func TestGetEnvCommand_Execute_BracketSyntax_TestMode(t *testing.T) {
	// Setup test environment
	neuroCtx, variableService := setupGetEnvCommandTestRegistry(t)
	cmd := &GetEnvCommand{}

	// Test case: Set test env override and retrieve it
	testVarName := "TEST_VAR_BRACKET"
	testVarValue := "test_value_bracket"

	// Set test environment override
	neuroCtx.SetTestEnvOverride(testVarName, testVarValue)

	// Execute command with bracket syntax
	args := map[string]string{testVarName: ""}
	err := cmd.Execute(args, "")
	require.NoError(t, err)

	// Verify #os.TEST_VAR_BRACKET neuro variable was created
	neuroVarName := "#os." + testVarName
	neuroVarValue, err := variableService.Get(neuroVarName)
	require.NoError(t, err)
	assert.Equal(t, testVarValue, neuroVarValue)
}

func TestGetEnvCommand_Execute_SpaceSyntax_TestMode(t *testing.T) {
	// Setup test environment
	neuroCtx, variableService := setupGetEnvCommandTestRegistry(t)
	cmd := &GetEnvCommand{}

	// Test case: Set test env override and retrieve it
	testVarName := "TEST_VAR_SPACE"
	testVarValue := "test_value_space"

	// Set test environment override
	neuroCtx.SetTestEnvOverride(testVarName, testVarValue)

	// Execute command with space syntax
	err := cmd.Execute(map[string]string{}, testVarName)
	require.NoError(t, err)

	// Verify #os.TEST_VAR_SPACE neuro variable was created
	neuroVarName := "#os." + testVarName
	neuroVarValue, err := variableService.Get(neuroVarName)
	require.NoError(t, err)
	assert.Equal(t, testVarValue, neuroVarValue)
}

func TestGetEnvCommand_Execute_ProductionMode(t *testing.T) {
	// Setup production environment
	variableService := setupGetEnvCommandProductionRegistry(t)
	cmd := &GetEnvCommand{}

	// Set actual OS environment variable
	testVarName := "TEST_GETENV_PROD"
	testVarValue := "production_value"

	// Set OS environment variable
	err := os.Setenv(testVarName, testVarValue)
	require.NoError(t, err)
	defer func() {
		_ = os.Unsetenv(testVarName)
	}()

	// Execute command
	args := map[string]string{testVarName: ""}
	err = cmd.Execute(args, "")
	require.NoError(t, err)

	// Verify #os.TEST_GETENV_PROD neuro variable was created
	neuroVarName := "#os." + testVarName
	neuroVarValue, err := variableService.Get(neuroVarName)
	require.NoError(t, err)
	assert.Equal(t, testVarValue, neuroVarValue)
}

func TestGetEnvCommand_Execute_NoVariableSpecified(t *testing.T) {
	// Setup test environment
	_, _ = setupGetEnvCommandTestRegistry(t)
	cmd := &GetEnvCommand{}

	// Test case: No variable specified (should return usage error)
	err := cmd.Execute(map[string]string{}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Usage:")
}

func TestGetEnvCommand_Execute_NeuroVariableCreationFlow(t *testing.T) {
	// Setup test environment
	neuroCtx, variableService := setupGetEnvCommandTestRegistry(t)
	cmd := &GetEnvCommand{}

	// Test the complete flow: get env var -> create neuro var -> verify both
	testVarName := "COMPLETE_FLOW_TEST"
	testVarValue := "complete_flow_value"

	// Set test environment override
	neuroCtx.SetTestEnvOverride(testVarName, testVarValue)

	// Execute command
	args := map[string]string{testVarName: ""}
	err := cmd.Execute(args, "")
	require.NoError(t, err)

	// Verify the neuro variable was created correctly
	neuroVarName := "#os." + testVarName
	neuroVarValue, err := variableService.Get(neuroVarName)
	require.NoError(t, err)
	assert.Equal(t, testVarValue, neuroVarValue, "Neuro variable should have the same value as environment variable")

	// Verify we can retrieve the neuro variable independently
	allVars, err := variableService.GetAllVariables()
	require.NoError(t, err)

	found := false
	for varName, varValue := range allVars {
		if varName == neuroVarName {
			found = true
			assert.Equal(t, testVarValue, varValue, "Neuro variable in full listing should match")
			break
		}
	}
	assert.True(t, found, "Neuro variable should be present in full variable listing")
}

// setupGetEnvCommandTestRegistry sets up test environment and services for GetEnvCommand tests
func setupGetEnvCommandTestRegistry(t *testing.T) (*context.NeuroContext, *services.VariableService) {
	ctx := context.NewTestContext()
	context.SetGlobalContext(ctx)

	// Create a new registry for testing
	services.SetGlobalRegistry(services.NewRegistry())
	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	// Initialize the variable service
	err = variableService.Initialize()
	require.NoError(t, err)

	neuroCtx := ctx.(*context.NeuroContext)
	return neuroCtx, variableService
}

// setupGetEnvCommandProductionRegistry sets up production environment and services for GetEnvCommand tests
func setupGetEnvCommandProductionRegistry(t *testing.T) *services.VariableService {
	ctx := context.New()
	context.SetGlobalContext(ctx)

	// Create a new registry for testing
	services.SetGlobalRegistry(services.NewRegistry())
	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	// Initialize the variable service
	err = variableService.Initialize()
	require.NoError(t, err)

	return variableService
}
