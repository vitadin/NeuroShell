package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestSendSyncCommand_Name(t *testing.T) {
	cmd := &SendSyncCommand{}
	assert.Equal(t, "send-sync", cmd.Name())
}

func TestSendSyncCommand_ParseMode(t *testing.T) {
	cmd := &SendSyncCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestSendSyncCommand_Description(t *testing.T) {
	cmd := &SendSyncCommand{}
	assert.Equal(t, "Send message to LLM agent with synchronous response", cmd.Description())
}

func TestSendSyncCommand_Usage(t *testing.T) {
	cmd := &SendSyncCommand{}
	assert.Equal(t, "\\send-sync message", cmd.Usage())
}

func TestSendSyncCommand_HelpInfo(t *testing.T) {
	cmd := &SendSyncCommand{}
	help := cmd.HelpInfo()

	assert.Equal(t, "send-sync", help.Command)
	assert.Equal(t, "Send message to LLM agent with synchronous response", help.Description)
	assert.Equal(t, neurotypes.ParseModeKeyValue, help.ParseMode)
	assert.Contains(t, help.Notes, "Messages are sent to the active chat session")
}

func TestSendSyncCommand_Execute_EmptyInput(t *testing.T) {
	// Setup minimal service for variable service requirement
	varService := services.NewVariableService()
	err := varService.Initialize()
	require.NoError(t, err)

	serviceRegistry := services.NewRegistry()
	err = serviceRegistry.RegisterService(varService)
	require.NoError(t, err)

	originalServiceRegistry := services.GlobalRegistry
	services.GlobalRegistry = serviceRegistry
	defer func() { services.GlobalRegistry = originalServiceRegistry }()

	cmd := &SendSyncCommand{}

	err = cmd.Execute(map[string]string{}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Usage:")
}

func TestSendSyncCommand_Execute_WithMockServices(t *testing.T) {
	// Setup test environment
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Create and setup services
	serviceRegistry := services.NewRegistry()

	// Setup chat session service
	chatService := services.NewChatSessionService()
	err := chatService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(chatService)
	require.NoError(t, err)

	// Setup model service
	modelService := services.NewModelService()
	err = modelService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(modelService)
	require.NoError(t, err)

	// Setup variable service
	varService := services.NewVariableService()
	err = varService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(varService)
	require.NoError(t, err)

	// Setup mock LLM service
	mockLLMService := services.NewMockLLMService()
	err = mockLLMService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(mockLLMService)
	require.NoError(t, err)

	// Setup client factory service
	clientFactory := services.NewClientFactoryService()
	err = clientFactory.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(clientFactory)
	require.NoError(t, err)

	// Temporarily replace global service registry
	originalServiceRegistry := services.GlobalRegistry
	services.GlobalRegistry = serviceRegistry
	defer func() { services.GlobalRegistry = originalServiceRegistry }()

	cmd := &SendSyncCommand{}

	// Test successful execution
	err = cmd.Execute(map[string]string{}, "Hello, test!")
	require.NoError(t, err)

	// Verify session was created and message was added
	activeSession, err := chatService.GetActiveSession()
	require.NoError(t, err)
	assert.Equal(t, "auto", activeSession.Name)
	assert.Len(t, activeSession.Messages, 2) // user message + assistant response

	// Check messages
	assert.Equal(t, "user", activeSession.Messages[0].Role)
	assert.Equal(t, "Hello, test!", activeSession.Messages[0].Content)
	assert.Equal(t, "assistant", activeSession.Messages[1].Role)
	assert.Equal(t, "This is a mocking reply message for the sending message: Hello, test!", activeSession.Messages[1].Content)

	// Check that variables were updated
	var1, err := varService.Get("1")
	require.NoError(t, err)
	assert.Equal(t, "This is a mocking reply message for the sending message: Hello, test!", var1) // Latest assistant response

	var2, err := varService.Get("2")
	require.NoError(t, err)
	assert.Equal(t, "Hello, test!", var2) // Latest user message
}

func TestSendSyncCommand_Execute_WithExistingSession(t *testing.T) {
	// Setup test environment
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Create and setup services
	serviceRegistry := services.NewRegistry()

	// Setup chat session service
	chatService := services.NewChatSessionService()
	err := chatService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(chatService)
	require.NoError(t, err)

	// Setup model service
	modelService := services.NewModelService()
	err = modelService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(modelService)
	require.NoError(t, err)

	// Setup variable service
	varService := services.NewVariableService()
	err = varService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(varService)
	require.NoError(t, err)

	// Setup mock LLM service
	mockLLMService := services.NewMockLLMService()
	err = mockLLMService.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(mockLLMService)
	require.NoError(t, err)

	// Setup client factory service
	clientFactory := services.NewClientFactoryService()
	err = clientFactory.Initialize()
	require.NoError(t, err)
	err = serviceRegistry.RegisterService(clientFactory)
	require.NoError(t, err)

	// Temporarily replace global service registry
	originalServiceRegistry := services.GlobalRegistry
	services.GlobalRegistry = serviceRegistry
	defer func() { services.GlobalRegistry = originalServiceRegistry }()

	// Create existing session
	existingSession, err := chatService.CreateSession("test-session", "Custom system prompt", "")
	require.NoError(t, err)
	err = chatService.SetActiveSession(existingSession.ID)
	require.NoError(t, err)

	cmd := &SendSyncCommand{}

	// Test execution with existing session
	err = cmd.Execute(map[string]string{}, "Second message")
	require.NoError(t, err)

	// Verify messages were added to existing session
	activeSession, err := chatService.GetActiveSession()
	require.NoError(t, err)
	assert.Equal(t, "test-session", activeSession.Name)
	assert.Len(t, activeSession.Messages, 2) // user message + assistant response

	// Check messages
	assert.Equal(t, "user", activeSession.Messages[0].Role)
	assert.Equal(t, "Second message", activeSession.Messages[0].Content)
	assert.Equal(t, "assistant", activeSession.Messages[1].Role)
	assert.Equal(t, "This is a mocking reply message for the sending message: Second message", activeSession.Messages[1].Content)
}

func TestSendSyncCommand_Execute_ServiceErrors(t *testing.T) {
	// Setup test environment with minimal services
	ctx := context.New()
	ctx.SetTestMode(true)
	context.SetGlobalContext(ctx)

	// Create empty service registry (missing services)
	serviceRegistry := services.NewRegistry()

	// Temporarily replace global service registry
	originalServiceRegistry := services.GlobalRegistry
	services.GlobalRegistry = serviceRegistry
	defer func() { services.GlobalRegistry = originalServiceRegistry }()

	cmd := &SendSyncCommand{}

	// Test with missing variable service (first service checked)
	err := cmd.Execute(map[string]string{}, "Hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get variable service")
}
