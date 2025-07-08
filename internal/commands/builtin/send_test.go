package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/commands"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestSendCommand_Name(t *testing.T) {
	cmd := &SendCommand{}
	assert.Equal(t, "send", cmd.Name())
}

func TestSendCommand_ParseMode(t *testing.T) {
	cmd := &SendCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestSendCommand_Description(t *testing.T) {
	cmd := &SendCommand{}
	assert.Equal(t, "Send message to LLM agent", cmd.Description())
}

func TestSendCommand_Usage(t *testing.T) {
	cmd := &SendCommand{}
	assert.Equal(t, "\\send message", cmd.Usage())
}

func TestSendCommand_HelpInfo(t *testing.T) {
	cmd := &SendCommand{}
	help := cmd.HelpInfo()

	assert.Equal(t, "send", help.Command)
	assert.Equal(t, "Send message to LLM agent", help.Description)
	assert.Equal(t, neurotypes.ParseModeKeyValue, help.ParseMode)
	assert.Contains(t, help.Notes, "Routes to send-stream or send-sync based on _reply_way variable")
}

func TestSendCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &SendCommand{}

	err := cmd.Execute(map[string]string{}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Usage:")
}

func TestSendCommand_Execute_RouterLogic(t *testing.T) {
	// Setup test environment
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Create a new registry for testing
	registry := commands.NewRegistry()

	// Register mock send-sync command
	mockSendSync := &MockSendSyncCommand{}
	err := registry.Register(mockSendSync)
	require.NoError(t, err)

	// Temporarily replace global registry
	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = registry
	defer func() { commands.GlobalRegistry = originalRegistry }()

	// Setup variable service
	varService := services.NewVariableService()
	err = varService.Initialize()
	require.NoError(t, err)

	// Create test registry for services
	serviceRegistry := services.NewRegistry()
	err = serviceRegistry.RegisterService(varService)
	require.NoError(t, err)

	// Temporarily replace global service registry
	originalServiceRegistry := services.GlobalRegistry
	services.GlobalRegistry = serviceRegistry
	defer func() { services.GlobalRegistry = originalServiceRegistry }()

	cmd := &SendCommand{}

	// Test default routing (should route to send-sync)
	err = cmd.Execute(map[string]string{}, "hello")
	assert.NoError(t, err) // Router never fails
	assert.True(t, mockSendSync.executed)
	assert.Equal(t, "hello", mockSendSync.lastInput)

	// Reset mock
	mockSendSync.executed = false
	mockSendSync.lastInput = ""

	// Test with _reply_way=sync
	err = varService.Set("_reply_way", "sync")
	require.NoError(t, err)

	err = cmd.Execute(map[string]string{}, "test sync")
	assert.NoError(t, err)
	assert.True(t, mockSendSync.executed)
	assert.Equal(t, "test sync", mockSendSync.lastInput)
}

// MockSendSyncCommand provides a mock implementation for testing
type MockSendSyncCommand struct {
	executed  bool
	lastInput string
	lastArgs  map[string]string
}

func (m *MockSendSyncCommand) Name() string {
	return "send-sync"
}

func (m *MockSendSyncCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

func (m *MockSendSyncCommand) Description() string {
	return "Mock send-sync command"
}

func (m *MockSendSyncCommand) Usage() string {
	return "\\send-sync message"
}

func (m *MockSendSyncCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     "send-sync",
		Description: "Mock send-sync command",
		Usage:       "\\send-sync message",
		ParseMode:   neurotypes.ParseModeKeyValue,
	}
}

func (m *MockSendSyncCommand) Execute(args map[string]string, input string) error {
	m.executed = true
	m.lastInput = input
	m.lastArgs = args
	return nil
}
