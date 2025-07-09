package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestSendCommand_Execute_PipelineOrchestration(t *testing.T) {
	// Setup test environment
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Create and setup services
	serviceRegistry := services.NewRegistry()

	// Variable service
	varService := services.NewVariableService()
	err := varService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(varService)
	require.NoError(t, err)

	// Chat session service
	chatSessionService := services.NewChatSessionService()
	err = chatSessionService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(chatSessionService)
	require.NoError(t, err)

	// Model service
	modelService := services.NewModelService()
	err = modelService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(modelService)
	require.NoError(t, err)

	// Client factory service
	clientFactory := services.NewClientFactoryService()
	err = clientFactory.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(clientFactory)
	require.NoError(t, err)

	// LLM service (mock)
	llmService := services.NewMockLLMService()
	err = llmService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(llmService)
	require.NoError(t, err)

	// Temporarily replace global service registry
	originalServiceRegistry := services.GlobalRegistry
	services.GlobalRegistry = serviceRegistry
	defer func() { services.GlobalRegistry = originalServiceRegistry }()

	cmd := &SendCommand{}

	// Test with mock services (should succeed)
	err = cmd.Execute(map[string]string{}, "hello")
	require.NoError(t, err)

	// Test with sync mode (default)
	_ = varService.Set("_reply_way", "sync")
	err = cmd.Execute(map[string]string{}, "hello sync")
	require.NoError(t, err)

	// Test with stream mode
	_ = varService.Set("_reply_way", "stream")
	err = cmd.Execute(map[string]string{}, "hello stream")
	require.NoError(t, err)
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
