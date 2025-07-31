// Package llm contains tests for LLM-related commands.
package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestClientActivateCommand_Name(t *testing.T) {
	cmd := &ClientActivateCommand{}
	assert.Equal(t, "llm-client-activate", cmd.Name())
}

func TestClientActivateCommand_ParseMode(t *testing.T) {
	cmd := &ClientActivateCommand{}
	assert.Equal(t, neurotypes.ParseModeRaw, cmd.ParseMode())
}

func TestClientActivateCommand_Description(t *testing.T) {
	cmd := &ClientActivateCommand{}
	assert.Equal(t, "Activate LLM client by provider catalog ID or specific client ID", cmd.Description())
}

func TestClientActivateCommand_Usage(t *testing.T) {
	cmd := &ClientActivateCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "llm-client-activate")
	assert.Contains(t, usage, "provider_catalog_id")
	assert.Contains(t, usage, "client_id")
	assert.Contains(t, usage, "OAR")
	assert.Contains(t, usage, "OAC")
}

func TestClientActivateCommand_HelpInfo(t *testing.T) {
	cmd := &ClientActivateCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "llm-client-activate", helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, neurotypes.ParseModeRaw, helpInfo.ParseMode)

	// Check examples
	assert.Len(t, helpInfo.Examples, 4)
	assert.Contains(t, helpInfo.Examples[0].Command, "OAR")
	assert.Contains(t, helpInfo.Examples[1].Command, "OAC")
	assert.Contains(t, helpInfo.Examples[2].Command, "ANC")
	assert.Contains(t, helpInfo.Examples[3].Command, "OAR:a3f2cae8")

	// Check stored variables
	assert.Len(t, helpInfo.StoredVariables, 2)

	variableNames := make(map[string]bool)
	for _, variable := range helpInfo.StoredVariables {
		variableNames[variable.Name] = true
	}

	assert.True(t, variableNames["#active_client_id"])
	assert.True(t, variableNames["_output"])

	// Check notes
	assert.Len(t, helpInfo.Notes, 5)
	assert.Contains(t, helpInfo.Notes[0], "Supports two input modes")
	assert.Contains(t, helpInfo.Notes[1], "Provider catalog ID")
	assert.Contains(t, helpInfo.Notes[2], "Client ID mode")
	assert.Contains(t, helpInfo.Notes[3], "Sets #active_client_id")
	assert.Contains(t, helpInfo.Notes[4], "Use \\*-client-new commands")
}

func TestClientActivateCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &ClientActivateCommand{}
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider catalog ID or client ID is required")
}

func TestClientActivateCommand_Execute_WhitespaceOnlyInput(t *testing.T) {
	cmd := &ClientActivateCommand{}
	err := cmd.Execute(map[string]string{}, "   \t\n  ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider catalog ID or client ID is required")
}

func TestClientActivateCommand_Execute_VariableServiceError(t *testing.T) {
	// Create empty registry without services
	originalRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	t.Cleanup(func() {
		services.SetGlobalRegistry(originalRegistry)
		context.ResetGlobalContext()
	})

	cmd := &ClientActivateCommand{}
	err := cmd.Execute(map[string]string{}, "OAR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client factory service not available")
}

func TestClientActivateCommand_Execute_ClientFactoryServiceError(t *testing.T) {
	// Initialize test environment with only variable service
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	// Register only variable service (no client factory)
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	cmd := &ClientActivateCommand{}
	err = cmd.Execute(map[string]string{}, "OAR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client factory service not available")
}

func TestClientActivateCommand_Execute_ProviderCatalogID_ClientNotFound(t *testing.T) {
	// Initialize test environment
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	// Register required services
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Don't create any clients
	cmd := &ClientActivateCommand{}
	err = cmd.Execute(map[string]string{}, "OAR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no client found with provider catalog ID 'OAR'")
}

func TestClientActivateCommand_Execute_ClientID_ClientNotFound(t *testing.T) {
	// Initialize test environment
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)

	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})

	// Register required services
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	_ = services.GetGlobalRegistry().RegisterService(services.NewClientFactoryService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Don't create any clients
	cmd := &ClientActivateCommand{}
	err = cmd.Execute(map[string]string{}, "OAR:nonexistent123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client with ID 'OAR:nonexistent123' not found")
}
