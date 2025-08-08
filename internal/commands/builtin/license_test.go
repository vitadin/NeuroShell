package builtin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseCommand_Name(t *testing.T) {
	cmd := &LicenseCommand{}
	assert.Equal(t, "license", cmd.Name())
}

func TestLicenseCommand_Description(t *testing.T) {
	cmd := &LicenseCommand{}
	description := cmd.Description()
	assert.Contains(t, description, "license")
	assert.Contains(t, description, "information")
}

func TestLicenseCommand_Usage(t *testing.T) {
	cmd := &LicenseCommand{}
	usage := cmd.Usage()
	assert.Equal(t, "\\license", usage)
}

func TestLicenseCommand_ParseMode(t *testing.T) {
	cmd := &LicenseCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestLicenseCommand_HelpInfo(t *testing.T) {
	cmd := &LicenseCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "license", helpInfo.Command)
	assert.Contains(t, helpInfo.Description, "license")
	assert.Equal(t, "\\license", helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Check examples
	assert.NotEmpty(t, helpInfo.Examples)
	assert.True(t, len(helpInfo.Examples) >= 2)

	// Check stored variables
	assert.NotEmpty(t, helpInfo.StoredVariables)
	variableNames := make([]string, len(helpInfo.StoredVariables))
	for i, v := range helpInfo.StoredVariables {
		variableNames[i] = v.Name
	}
	assert.Contains(t, variableNames, "#license_name")
	assert.Contains(t, variableNames, "#license_short_name")
	assert.Contains(t, variableNames, "#license_url")
	assert.Contains(t, variableNames, "#license_file_path")
}

func TestLicenseCommand_Execute_BasicFunctionality(t *testing.T) {
	// Initialize test services
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)
	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()

	cmd := &LicenseCommand{}

	// Execute command
	err = cmd.Execute(map[string]string{}, "")
	require.NoError(t, err)
}

func TestLicenseCommand_Execute_SystemVariableStorage(t *testing.T) {
	// Initialize test services
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	err := services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)
	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()

	// Get variable service
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	cmd := &LicenseCommand{}

	// Execute command
	err = cmd.Execute(map[string]string{}, "")
	require.NoError(t, err)

	// Verify system variables were set
	expectedVars := map[string]string{
		"#license_name":       "GNU Lesser General Public License v3.0",
		"#license_short_name": "LGPL-3.0",
		"#license_url":        "https://www.gnu.org/licenses/lgpl-3.0.html",
	}

	for varName, expectedValue := range expectedVars {
		actualValue, err := variableService.Get(varName)
		require.NoError(t, err, "Failed to get variable %s", varName)
		assert.Equal(t, expectedValue, actualValue, "Variable %s has wrong value", varName)
	}

	// Check that license_file_path was set (value depends on file system)
	licensePath, err := variableService.Get("#license_file_path")
	require.NoError(t, err)
	// Should be empty string if no LICENSE file found, or a valid path
	assert.True(t, licensePath == "" || strings.Contains(licensePath, "LICENSE"))
}

func TestLicenseCommand_Execute_WithLicenseFile(t *testing.T) {
	// Create a temporary LICENSE file
	tempDir := t.TempDir()
	licensePath := filepath.Join(tempDir, "LICENSE")
	licenseContent := `GNU LESSER GENERAL PUBLIC LICENSE
Version 3, 29 June 2007

Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>
Everyone is permitted to copy and distribute verbatim copies
of this license document, but changing it is not allowed.

This is a test license file for NeuroShell.`

	err := os.WriteFile(licensePath, []byte(licenseContent), 0644)
	require.NoError(t, err)

	// Change to temp directory so the command can find the LICENSE file
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalDir)
	}()
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Initialize test services
	ctx := context.NewTestContext()
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	context.SetGlobalContext(ctx)
	_ = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)
	defer func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	}()

	cmd := &LicenseCommand{}

	// Execute command
	err = cmd.Execute(map[string]string{}, "")
	require.NoError(t, err)

	// Verify license_file_path variable was set correctly
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	storedPath, err := variableService.Get("#license_file_path")
	require.NoError(t, err)
	assert.Contains(t, storedPath, "LICENSE")
}

func TestLicenseCommand_Execute_ServiceNotAvailable(t *testing.T) {
	// Create an empty registry without services to simulate error condition
	originalRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	t.Cleanup(func() {
		services.SetGlobalRegistry(originalRegistry)
		context.ResetGlobalContext()
	})

	cmd := &LicenseCommand{}

	// Execute should fail when variable service is not available
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}
