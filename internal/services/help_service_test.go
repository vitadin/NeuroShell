package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/commands"
	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

func TestHelpService_Name(t *testing.T) {
	service := NewHelpService()
	assert.Equal(t, "help", service.Name())
}

func TestHelpService_Initialize(t *testing.T) {
	// Create test commands
	testCommands := []neurotypes.Command{
		&MockHelpCommand{
			name:        "test1",
			description: "Test command 1",
			usage:       "\\test1",
			parseMode:   neurotypes.ParseModeKeyValue,
		},
		&MockHelpCommand{
			name:        "test2",
			description: "Test command 2",
			usage:       "\\test2 [arg]",
			parseMode:   neurotypes.ParseModeRaw,
		},
	}

	// Set up test environment
	ctx := setupHelpServiceTest(t, testCommands)

	// Set the global context for the service
	oldCtx := neuroshellcontext.GetGlobalContext()
	neuroshellcontext.SetGlobalContext(ctx)
	defer neuroshellcontext.SetGlobalContext(oldCtx)

	service := NewHelpService()
	assert.False(t, service.initialized)

	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)

	// Verify commands were stored
	assert.Len(t, service.commands, 2)
	assert.Contains(t, service.commands, "test1")
	assert.Contains(t, service.commands, "test2")
}

func TestHelpService_GetAllCommands(t *testing.T) {
	testCommands := []neurotypes.Command{
		&MockHelpCommand{name: "zebra", description: "Last", usage: "\\zebra"},
		&MockHelpCommand{name: "apple", description: "First", usage: "\\apple"},
		&MockHelpCommand{name: "banana", description: "Middle", usage: "\\banana"},
	}

	service, _ := setupInitializedHelpService(t, testCommands)

	allCommands, err := service.GetAllCommands()
	assert.NoError(t, err)
	assert.Len(t, allCommands, 3)

	// Verify they are sorted alphabetically
	assert.Equal(t, "apple", allCommands[0].Name)
	assert.Equal(t, "banana", allCommands[1].Name)
	assert.Equal(t, "zebra", allCommands[2].Name)
}

func TestHelpService_GetCommand(t *testing.T) {
	testCommands := []neurotypes.Command{
		&MockHelpCommand{
			name:        "test",
			description: "Test command",
			usage:       "\\test [arg]",
			parseMode:   neurotypes.ParseModeRaw,
		},
	}

	service, _ := setupInitializedHelpService(t, testCommands)

	// Test existing command
	cmdInfo, err := service.GetCommand("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", cmdInfo.Name)
	assert.Equal(t, "Test command", cmdInfo.Description)
	assert.Equal(t, "\\test [arg]", cmdInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeRaw, cmdInfo.ParseMode)

	// Test non-existent command
	_, err = service.GetCommand("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command 'nonexistent' not found")
}

func TestHelpService_GetCommandNames(t *testing.T) {
	testCommands := []neurotypes.Command{
		&MockHelpCommand{name: "zebra", description: "Last", usage: "\\zebra"},
		&MockHelpCommand{name: "apple", description: "First", usage: "\\apple"},
	}

	service, _ := setupInitializedHelpService(t, testCommands)

	names, err := service.GetCommandNames()
	assert.NoError(t, err)
	assert.Len(t, names, 2)
	// Should be sorted
	assert.Equal(t, []string{"apple", "zebra"}, names)
}

func TestHelpService_IsValidCommand(t *testing.T) {
	testCommands := []neurotypes.Command{
		&MockHelpCommand{name: "valid", description: "Valid command", usage: "\\valid"},
	}

	service, _ := setupInitializedHelpService(t, testCommands)

	assert.True(t, service.IsValidCommand("valid"))
	assert.False(t, service.IsValidCommand("invalid"))
}

func TestHelpService_GetCommandCount(t *testing.T) {
	testCommands := []neurotypes.Command{
		&MockHelpCommand{name: "cmd1", description: "Command 1", usage: "\\cmd1"},
		&MockHelpCommand{name: "cmd2", description: "Command 2", usage: "\\cmd2"},
		&MockHelpCommand{name: "cmd3", description: "Command 3", usage: "\\cmd3"},
	}

	service, _ := setupInitializedHelpService(t, testCommands)

	assert.Equal(t, 3, service.GetCommandCount())
}

func TestHelpService_NotInitialized(t *testing.T) {
	service := NewHelpService()

	// All methods should return errors when not initialized
	_, err := service.GetAllCommands()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = service.GetCommand("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	_, err = service.GetCommandNames()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	assert.False(t, service.IsValidCommand("test"))
}

func TestHelpService_ParseModeToString(t *testing.T) {
	service := NewHelpService()

	tests := []struct {
		mode     neurotypes.ParseMode
		expected string
	}{
		{neurotypes.ParseModeKeyValue, "KeyValue"},
		{neurotypes.ParseModeRaw, "Raw"},
		{neurotypes.ParseModeWithOptions, "WithOptions"},
		{neurotypes.ParseMode(999), "Unknown"}, // Invalid mode
	}

	for _, tt := range tests {
		result := service.parseModeToString(tt.mode)
		assert.Equal(t, tt.expected, result)
	}
}

func TestHelpService_EmptyCommandRegistry(t *testing.T) {
	// Test with no commands
	testCommands := []neurotypes.Command{}
	service, _ := setupInitializedHelpService(t, testCommands)

	allCommands, err := service.GetAllCommands()
	assert.NoError(t, err)
	assert.Len(t, allCommands, 0)

	names, err := service.GetCommandNames()
	assert.NoError(t, err)
	assert.Len(t, names, 0)

	assert.Equal(t, 0, service.GetCommandCount())
}

func TestHelpService_SystemVariableStorage(t *testing.T) {
	testCommands := []neurotypes.Command{
		&MockHelpCommand{
			name:        "testcmd",
			description: "Test description",
			usage:       "\\testcmd [options]",
			parseMode:   neurotypes.ParseModeKeyValue,
		},
	}

	ctx := setupHelpServiceTest(t, testCommands)

	// Set the global context for the service
	oldCtx := neuroshellcontext.GetGlobalContext()
	neuroshellcontext.SetGlobalContext(ctx)
	defer neuroshellcontext.SetGlobalContext(oldCtx)

	service := NewHelpService()

	err := service.Initialize()
	require.NoError(t, err)

	// Verify system variables were set
	allVars := ctx.GetAllVariables()

	// Check command list
	assert.Contains(t, allVars, "#cmd_list")
	assert.Equal(t, "testcmd", allVars["#cmd_list"])

	// Check command count
	assert.Contains(t, allVars, "#cmd_count")
	assert.Equal(t, "1", allVars["#cmd_count"])

	// Check individual command metadata
	assert.Contains(t, allVars, "#cmd_testcmd_desc")
	assert.Equal(t, "Test description", allVars["#cmd_testcmd_desc"])

	assert.Contains(t, allVars, "#cmd_testcmd_usage")
	assert.Equal(t, "\\testcmd [options]", allVars["#cmd_testcmd_usage"])

	assert.Contains(t, allVars, "#cmd_testcmd_parsemode")
	assert.Equal(t, "KeyValue", allVars["#cmd_testcmd_parsemode"])
}

// Helper functions

func setupHelpServiceTest(t *testing.T, testCommands []neurotypes.Command) neurotypes.Context {
	// Create test command registry
	testCommandRegistry := commands.NewRegistry()

	// Register test commands
	for _, cmd := range testCommands {
		err := testCommandRegistry.Register(cmd)
		require.NoError(t, err)
	}

	// Temporarily replace global command registry using thread-safe functions
	originalCommandRegistry := commands.GetGlobalRegistry()
	commands.SetGlobalRegistry(testCommandRegistry)

	t.Cleanup(func() {
		commands.SetGlobalRegistry(originalCommandRegistry)
	})

	neuroshellcontext.ResetGlobalContext()
	ctx := neuroshellcontext.GetGlobalContext()
	ctx.SetTestMode(true)
	return ctx
}

func setupInitializedHelpService(t *testing.T, testCommands []neurotypes.Command) (*HelpService, neurotypes.Context) {
	ctx := setupHelpServiceTest(t, testCommands)

	// Set the global context for the service
	oldCtx := neuroshellcontext.GetGlobalContext()
	neuroshellcontext.SetGlobalContext(ctx)
	t.Cleanup(func() {
		neuroshellcontext.SetGlobalContext(oldCtx)
	})

	service := NewHelpService()

	err := service.Initialize()
	require.NoError(t, err)

	return service, ctx
}

// MockHelpCommand for testing HelpService
type MockHelpCommand struct {
	name        string
	description string
	usage       string
	parseMode   neurotypes.ParseMode
}

func (m *MockHelpCommand) Name() string {
	return m.name
}

func (m *MockHelpCommand) Description() string {
	return m.description
}

func (m *MockHelpCommand) Usage() string {
	return m.usage
}

func (m *MockHelpCommand) ParseMode() neurotypes.ParseMode {
	if m.parseMode == 0 {
		return neurotypes.ParseModeKeyValue
	}
	return m.parseMode
}

func (m *MockHelpCommand) Execute(_ map[string]string, _ string) error {
	return nil
}

func (m *MockHelpCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     m.Name(),
		Description: m.Description(),
		Usage:       m.Usage(),
		ParseMode:   m.ParseMode(),
		Examples:    []neurotypes.HelpExample{},
	}
}
