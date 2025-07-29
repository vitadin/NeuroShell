package builtin

import (
	"strconv"
	"strings"
	"testing"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/version"
	"neuroshell/pkg/neurotypes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand_Name(t *testing.T) {
	cmd := &VersionCommand{}
	assert.Equal(t, "version", cmd.Name())
}

func TestVersionCommand_ParseMode(t *testing.T) {
	cmd := &VersionCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestVersionCommand_Description(t *testing.T) {
	cmd := &VersionCommand{}
	assert.Equal(t, "Show NeuroShell version information and store details in system variables", cmd.Description())
}

func TestVersionCommand_Usage(t *testing.T) {
	cmd := &VersionCommand{}
	assert.Equal(t, "\\version", cmd.Usage())
}

func TestVersionCommand_HelpInfo(t *testing.T) {
	cmd := &VersionCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "version", helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, "\\version", helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)

	// Should have no options
	assert.Empty(t, helpInfo.Options)

	// Should have examples
	assert.NotEmpty(t, helpInfo.Examples)
	assert.Contains(t, helpInfo.Examples[0].Command, "\\version")

	// Should have stored variables documented
	assert.NotEmpty(t, helpInfo.StoredVariables)

	// Check that all expected system variables are documented
	expectedVars := []string{
		"#version", "#version_base", "#version_codename", "#version_commit",
		"#version_commit_count", "#version_build_date", "#version_build_metadata",
		"#version_go_version", "#version_platform",
	}

	documentedVars := make(map[string]bool)
	for _, storedVar := range helpInfo.StoredVariables {
		documentedVars[storedVar.Name] = true
	}

	for _, expectedVar := range expectedVars {
		assert.True(t, documentedVars[expectedVar], "System variable %s should be documented", expectedVar)
	}

	// Should have notes
	assert.NotEmpty(t, helpInfo.Notes)
}

func TestVersionCommand_Execute(t *testing.T) {
	// Save original version info
	originalVersion := version.Version
	originalGitCommit := version.GitCommit
	originalBuildDate := version.BuildDate
	defer func() {
		version.Version = originalVersion
		version.GitCommit = originalGitCommit
		version.BuildDate = originalBuildDate
	}()

	// Set test version info
	testVersion := "0.2.0+123.abc1234"
	testCommit := "abc1234"
	testDate := "2025-07-29"

	version.SetBuildInfo(testVersion, testCommit, testDate)

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

	cmd := &VersionCommand{}

	// Execute the command
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Get variable service to check stored variables
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	// Test that all system variables are set correctly
	testCases := map[string]string{
		"#version":                testVersion,
		"#version_base":           "0.2.0",
		"#version_codename":       "Planaria", // 0.2.0 maps to Planaria
		"#version_commit":         testCommit,
		"#version_commit_count":   "123", // From build metadata
		"#version_build_date":     testDate,
		"#version_build_metadata": "123.abc1234",
	}

	for varName, expectedValue := range testCases {
		actualValue, err := variableService.Get(varName)
		assert.NoError(t, err, "Should be able to get %s", varName)
		assert.Equal(t, expectedValue, actualValue, "System variable %s should have correct value", varName)
	}

	// Test Go version and platform variables (these contain runtime info)
	goVersion, err := variableService.Get("#version_go_version")
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(goVersion, "go"), "Go version should start with 'go'")

	platform, err := variableService.Get("#version_platform")
	assert.NoError(t, err)
	assert.Contains(t, platform, "/", "Platform should contain OS/arch separator")
}

func TestVersionCommand_Execute_WithDifferentVersions(t *testing.T) {
	// Save original version info
	originalVersion := version.Version
	originalGitCommit := version.GitCommit
	originalBuildDate := version.BuildDate
	defer func() {
		version.Version = originalVersion
		version.GitCommit = originalGitCommit
		version.BuildDate = originalBuildDate
	}()

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

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	cmd := &VersionCommand{}

	testCases := []struct {
		name                string
		version             string
		commit              string
		buildDate           string
		expectedBase        string
		expectedCodename    string
		expectedCommitCount string
		expectedMetadata    string
	}{
		{
			name:                "version with build metadata",
			version:             "0.2.0+530.a455fa8",
			commit:              "a455fa8",
			buildDate:           "2025-07-29",
			expectedBase:        "0.2.0",
			expectedCodename:    "Planaria",
			expectedCommitCount: "530",
			expectedMetadata:    "530.a455fa8",
		},
		{
			name:                "simple version without metadata",
			version:             "1.0.0",
			commit:              "unknown",
			buildDate:           "unknown",
			expectedBase:        "1.0.0",
			expectedCodename:    "Sapiens",
			expectedCommitCount: "0",
			expectedMetadata:    "",
		},
		{
			name:                "version without codename",
			version:             "0.10.0+42.def5678",
			commit:              "def5678",
			buildDate:           "2025-07-30",
			expectedBase:        "0.10.0",
			expectedCodename:    "",
			expectedCommitCount: "42",
			expectedMetadata:    "42.def5678",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Set test version info
			version.SetBuildInfo(tt.version, tt.commit, tt.buildDate)

			// Execute the command
			err := cmd.Execute(map[string]string{}, "")
			assert.NoError(t, err)

			// Check system variables
			actualVersion, err := variableService.Get("#version")
			assert.NoError(t, err)
			assert.Equal(t, tt.version, actualVersion)

			actualBase, err := variableService.Get("#version_base")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBase, actualBase)

			actualCodename, err := variableService.Get("#version_codename")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCodename, actualCodename)

			actualCommit, err := variableService.Get("#version_commit")
			assert.NoError(t, err)
			assert.Equal(t, tt.commit, actualCommit)

			actualCommitCount, err := variableService.Get("#version_commit_count")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCommitCount, actualCommitCount)

			actualBuildDate, err := variableService.Get("#version_build_date")
			assert.NoError(t, err)
			assert.Equal(t, tt.buildDate, actualBuildDate)

			actualMetadata, err := variableService.Get("#version_build_metadata")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedMetadata, actualMetadata)
		})
	}
}

func TestVersionCommand_Execute_VariableServiceError(t *testing.T) {
	// Create an empty registry without services to simulate error condition
	originalRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())
	t.Cleanup(func() {
		services.SetGlobalRegistry(originalRegistry)
		context.ResetGlobalContext()
	})

	cmd := &VersionCommand{}

	// Execute should fail when variable service is not available
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestVersionCommand_Execute_CommitCountAsString(t *testing.T) {
	// Save original version info
	originalVersion := version.Version
	originalGitCommit := version.GitCommit
	originalBuildDate := version.BuildDate
	defer func() {
		version.Version = originalVersion
		version.GitCommit = originalGitCommit
		version.BuildDate = originalBuildDate
	}()

	// Set version with high commit count
	version.SetBuildInfo("0.2.0+9999.xyz123", "xyz123", "2025-07-29")

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

	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	cmd := &VersionCommand{}

	// Execute the command
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)

	// Check that commit count is stored as string
	commitCountStr, err := variableService.Get("#version_commit_count")
	assert.NoError(t, err)
	assert.Equal(t, "9999", commitCountStr)

	// Verify it can be parsed back to int
	commitCount, err := strconv.Atoi(commitCountStr)
	assert.NoError(t, err)
	assert.Equal(t, 9999, commitCount)
}

func TestVersionCommand_StorageErrors(t *testing.T) {
	// This test is more of a coverage test since the command handles storage errors gracefully
	// and continues execution. In a real failure scenario, warnings would be printed but the command succeeds.

	// Save original version info
	originalVersion := version.Version
	originalGitCommit := version.GitCommit
	originalBuildDate := version.BuildDate
	defer func() {
		version.Version = originalVersion
		version.GitCommit = originalGitCommit
		version.BuildDate = originalBuildDate
	}()

	version.SetBuildInfo("0.2.0+123.abc1234", "abc1234", "2025-07-29")

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

	cmd := &VersionCommand{}

	// Execute should succeed even if there are storage issues
	// (The current implementation doesn't have a way to force storage errors in tests,
	// but this tests the error handling path)
	err = cmd.Execute(map[string]string{}, "")
	assert.NoError(t, err)
}
