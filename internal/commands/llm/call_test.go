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

func TestCallCommand_Name(t *testing.T) {
	cmd := &CallCommand{}
	assert.Equal(t, "llm-call", cmd.Name())
}

func TestCallCommand_ParseMode(t *testing.T) {
	cmd := &CallCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestCallCommand_Description(t *testing.T) {
	cmd := &CallCommand{}
	assert.Contains(t, cmd.Description(), "Orchestrate LLM API call")
}

func TestCallCommand_Usage(t *testing.T) {
	cmd := &CallCommand{}
	usage := cmd.Usage()
	assert.Contains(t, usage, "\\llm-call")
	assert.Contains(t, usage, "dry_run")
	assert.Contains(t, usage, "stream")
}

func TestCallCommand_HelpInfo(t *testing.T) {
	cmd := &CallCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, cmd.Name(), helpInfo.Command)
	assert.Equal(t, cmd.Description(), helpInfo.Description)
	assert.Equal(t, cmd.ParseMode(), helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.NotEmpty(t, helpInfo.Options)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Verify specific options
	optionNames := make(map[string]bool)
	for _, option := range helpInfo.Options {
		optionNames[option.Name] = true
	}
	assert.True(t, optionNames["client_id"])
	assert.True(t, optionNames["model_id"])
	assert.True(t, optionNames["session_id"])
	// Note: stream option removed as streaming is no longer supported
	assert.True(t, optionNames["dry_run"])
}

func TestCallCommand_Execute_InputWarning(t *testing.T) {
	// Create test context and services
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService()) // Use mock for testing
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Create command and test with input (should show warning)
	cmd := &CallCommand{}

	// This should fail due to missing required components, but we're testing the warning
	err := cmd.Execute(map[string]string{}, "Hello, this should show a warning")

	// We expect this to fail due to missing client_id, but the warning should have been shown
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client_id not specified")
}

func TestCallCommand_Execute_DryRun(t *testing.T) {
	// Create test context and services
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services for setup
	clientFactory, _ := services.GetGlobalClientFactoryService()
	modelService, _ := services.GetGlobalModelService()
	sessionService, _ := services.GetGlobalChatSessionService()
	variableService, _ := services.GetGlobalVariableService()

	// Create test client (using mock API key)
	_, clientID, err := clientFactory.GetClientWithID("OAR", "test-api-key")
	require.NoError(t, err)

	// Store client ID in variable service for dry run display
	_ = variableService.SetSystemVariable("_client_id", clientID)

	// Create test model
	model, err := modelService.CreateModelWithGlobalContext("test-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7}, "Test model", "")
	require.NoError(t, err)

	// Create test session
	session, err := sessionService.CreateSession("test-session", "You are helpful", "")
	require.NoError(t, err)

	// Add a message to the session
	err = sessionService.AddMessage(session.ID, "user", "Hello, how are you?")
	require.NoError(t, err)

	// Test dry run
	cmd := &CallCommand{}
	args := map[string]string{
		"client_id":  clientID,
		"model_id":   model.Name,
		"session_id": session.ID,
		"dry_run":    "true",
	}

	err = cmd.Execute(args, "")
	require.NoError(t, err)

	// Verify dry run variables were set
	output, err := variableService.Get("_output")
	require.NoError(t, err)
	assert.Equal(t, "DRY RUN - No API call made", output)

	dryRunMode, err := variableService.Get("#dry_run_mode")
	require.NoError(t, err)
	assert.Equal(t, "true", dryRunMode)

	dryRunClient, err := variableService.Get("#dry_run_client")
	require.NoError(t, err)
	assert.Equal(t, clientID, dryRunClient)
}

func TestCallCommand_Execute_SyncCall(t *testing.T) {
	// Create test context and services
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services for setup
	clientFactory, _ := services.GetGlobalClientFactoryService()
	modelService, _ := services.GetGlobalModelService()
	sessionService, _ := services.GetGlobalChatSessionService()
	variableService, _ := services.GetGlobalVariableService()

	// Create test client
	_, clientID, err := clientFactory.GetClientWithID("OAR", "test-api-key")
	require.NoError(t, err)

	// Create test model
	model, err := modelService.CreateModelWithGlobalContext("test-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7}, "Test model", "")
	require.NoError(t, err)

	// Create test session with a message
	session, err := sessionService.CreateSession("test-session", "You are helpful", "")
	require.NoError(t, err)

	err = sessionService.AddMessage(session.ID, "user", "Hello, how are you?")
	require.NoError(t, err)

	// Test sync call
	cmd := &CallCommand{}
	args := map[string]string{
		"client_id":  clientID,
		"model_id":   model.Name,
		"session_id": session.ID,
		"stream":     "false",
	}

	err = cmd.Execute(args, "")
	require.NoError(t, err)

	// Verify response variables were set
	output, err := variableService.Get("_output")
	require.NoError(t, err)
	assert.Contains(t, output, "mocking reply")

	llmResponse, err := variableService.Get("#llm_response")
	require.NoError(t, err)
	assert.Equal(t, output, llmResponse)

	callSuccess, err := variableService.Get("#llm_call_success")
	require.NoError(t, err)
	assert.Equal(t, "true", callSuccess)

	callMode, err := variableService.Get("#llm_call_mode")
	require.NoError(t, err)
	assert.Equal(t, "http", callMode)
}

func TestCallCommand_Execute_StreamingIgnored(t *testing.T) {
	// Create test context and services
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services for setup
	clientFactory, _ := services.GetGlobalClientFactoryService()
	modelService, _ := services.GetGlobalModelService()
	sessionService, _ := services.GetGlobalChatSessionService()
	variableService, _ := services.GetGlobalVariableService()

	// Create test client
	_, clientID, err := clientFactory.GetClientWithID("OAR", "test-api-key")
	require.NoError(t, err)

	// Create test model
	model, err := modelService.CreateModelWithGlobalContext("test-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7}, "Test model", "")
	require.NoError(t, err)

	// Create test session with a message
	session, err := sessionService.CreateSession("test-session", "You are helpful", "")
	require.NoError(t, err)

	err = sessionService.AddMessage(session.ID, "user", "Write a story")
	require.NoError(t, err)

	// Test streaming call
	cmd := &CallCommand{}
	args := map[string]string{
		"client_id":  clientID,
		"model_id":   model.Name,
		"session_id": session.ID,
		"stream":     "true",
	}

	err = cmd.Execute(args, "")
	require.NoError(t, err)

	// Verify response variables were set
	output, err := variableService.Get("_output")
	require.NoError(t, err)
	assert.Contains(t, output, "mocking reply")

	callMode, err := variableService.Get("#llm_call_mode")
	require.NoError(t, err)
	assert.Equal(t, "http", callMode) // Streaming is ignored, should use HTTP
}

func TestCallCommand_Execute_DefaultResolution(t *testing.T) {
	// Create test context and services
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services for setup
	clientFactory, _ := services.GetGlobalClientFactoryService()
	modelService, _ := services.GetGlobalModelService()
	sessionService, _ := services.GetGlobalChatSessionService()
	variableService, _ := services.GetGlobalVariableService()

	// Create test client and store in variable
	_, clientID, err := clientFactory.GetClientWithID("OAR", "test-api-key")
	require.NoError(t, err)
	_ = variableService.SetSystemVariable("_client_id", clientID)

	// Create and activate test model
	model, err := modelService.CreateModelWithGlobalContext("test-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7}, "Test model", "")
	require.NoError(t, err)
	err = modelService.SetActiveModelWithGlobalContext(model.ID)
	require.NoError(t, err)

	// Create and set active session
	session, err := sessionService.CreateSession("test-session", "You are helpful", "")
	require.NoError(t, err)
	err = sessionService.SetActiveSession(session.ID)
	require.NoError(t, err)

	// Add message to session
	err = sessionService.AddMessage(session.ID, "user", "Test message")
	require.NoError(t, err)

	// Test with no explicit parameters (should use defaults)
	cmd := &CallCommand{}
	err = cmd.Execute(map[string]string{}, "")
	require.NoError(t, err)

	// Verify call succeeded using defaults
	callSuccess, err := variableService.Get("#llm_call_success")
	require.NoError(t, err)
	assert.Equal(t, "true", callSuccess)
}

func TestCallCommand_Execute_MissingComponents(t *testing.T) {
	// Create fresh test context and services for isolation
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services with fresh registry
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context with isolation
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	cmd := &CallCommand{}

	// Test missing client_id (most basic error)
	err := cmd.Execute(map[string]string{}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client_id not specified")

	// Test missing session_id (after providing valid client and using default model)
	variableService, _ := services.GetGlobalVariableService()
	clientFactory, _ := services.GetGlobalClientFactoryService()
	_, clientID, _ := clientFactory.GetClientWithID("OAR", "test-api-key")
	_ = variableService.SetSystemVariable("_client_id", clientID)

	err = cmd.Execute(map[string]string{}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session_id not specified")

	// Test invalid model_id (with valid client and session)
	sessionService, _ := services.GetGlobalChatSessionService()
	session, _ := sessionService.CreateSession("test-session", "You are helpful", "")

	err = cmd.Execute(map[string]string{
		"model_id":   "nonexistent-model",
		"session_id": session.ID,
	}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get model")
}

func TestCallCommand_Execute_EmptySession(t *testing.T) {
	// Create test context and services
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services for setup
	clientFactory, _ := services.GetGlobalClientFactoryService()
	modelService, _ := services.GetGlobalModelService()
	sessionService, _ := services.GetGlobalChatSessionService()

	// Create test client
	_, clientID, err := clientFactory.GetClientWithID("OAR", "test-api-key")
	require.NoError(t, err)

	// Create test model
	model, err := modelService.CreateModelWithGlobalContext("test-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7}, "Test model", "")
	require.NoError(t, err)

	// Create test session WITHOUT messages
	session, err := sessionService.CreateSession("empty-session", "You are helpful", "")
	require.NoError(t, err)

	// Test that normal llm-call errors out with empty session
	cmd := &CallCommand{}
	args := map[string]string{
		"client_id":  clientID,
		"model_id":   model.Name,
		"session_id": session.ID,
		"stream":     "false",
	}

	err = cmd.Execute(args, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contains no messages")
	assert.Contains(t, err.Error(), "Use \\session-add-usermsg")
}

func TestCallCommand_Execute_EmptySession_DryRun(t *testing.T) {
	// Create test context and services
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	// Get services for setup
	clientFactory, _ := services.GetGlobalClientFactoryService()
	modelService, _ := services.GetGlobalModelService()
	sessionService, _ := services.GetGlobalChatSessionService()
	variableService, _ := services.GetGlobalVariableService()

	// Create test client
	_, clientID, err := clientFactory.GetClientWithID("OAR", "test-api-key")
	require.NoError(t, err)
	_ = variableService.SetSystemVariable("_client_id", clientID)

	// Create test model
	model, err := modelService.CreateModelWithGlobalContext("test-model", "openai", "gpt-4",
		map[string]any{"temperature": 0.7}, "Test model", "")
	require.NoError(t, err)

	// Create test session WITHOUT messages
	session, err := sessionService.CreateSession("empty-session", "You are helpful", "")
	require.NoError(t, err)

	// Test that dry run works with empty session but shows warnings
	cmd := &CallCommand{}
	args := map[string]string{
		"client_id":  clientID,
		"model_id":   model.Name,
		"session_id": session.ID,
		"dry_run":    "true",
	}

	err = cmd.Execute(args, "")
	require.NoError(t, err) // Dry run should succeed

	// Verify dry run variables were set
	output, err := variableService.Get("_output")
	require.NoError(t, err)
	assert.Equal(t, "DRY RUN - No API call made", output)

	dryRunMode, err := variableService.Get("#dry_run_mode")
	require.NoError(t, err)
	assert.Equal(t, "true", dryRunMode)
}

func TestCallCommand_Execute_InvalidComponents(t *testing.T) {
	// Create test context and services
	ctx := context.New()
	ctx.SetTestMode(true)

	// Initialize services
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewModelService())
	_ = registry.RegisterService(services.NewChatSessionService())
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set global registry and context
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	oldCtx := context.GetGlobalContext()
	context.SetGlobalContext(ctx)
	defer context.SetGlobalContext(oldCtx)

	cmd := &CallCommand{}

	// Test invalid client_id
	args := map[string]string{
		"client_id":  "nonexistent-client",
		"model_id":   "test-model",
		"session_id": "test-session",
	}

	err := cmd.Execute(args, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get client")

	// Create valid client but test invalid model
	clientFactory, _ := services.GetGlobalClientFactoryService()
	_, clientID, _ := clientFactory.GetClientWithID("OAR", "test-api-key")

	args["client_id"] = clientID
	args["model_id"] = "nonexistent-model"

	err = cmd.Execute(args, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get model")

	// Create valid model but test invalid session
	modelService, _ := services.GetGlobalModelService()
	model, _ := modelService.CreateModelWithGlobalContext("test-model", "openai", "gpt-4", nil, "Test model", "")

	args["model_id"] = model.Name
	args["session_id"] = "nonexistent-session"

	err = cmd.Execute(args, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get session")
}

func TestCallCommand_handleDryRun(t *testing.T) {
	cmd := &CallCommand{}

	// Create mock client
	clientFactory := services.NewClientFactoryService()
	_ = clientFactory.Initialize()
	client, clientID, _ := clientFactory.GetClientWithID("OAR", "test-key")
	_ = clientID // Use clientID to avoid unused variable error

	// Create test model
	model := &neurotypes.ModelConfig{
		ID:         "test-model-id",
		Name:       "test-model",
		Provider:   "openai",
		BaseModel:  "gpt-4",
		Parameters: map[string]any{"temperature": 0.7, "max_tokens": 1000},
	}

	// Create test session
	session := &neurotypes.ChatSession{
		ID:           "test-session-id",
		Name:         "test-session",
		SystemPrompt: "You are a helpful assistant",
		Messages: []neurotypes.Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
			{Role: "user", Content: "How are you?"},
		},
	}

	// Create variable service
	variableService := services.NewVariableService()
	_ = variableService.Initialize()

	// Test dry run
	err := cmd.handleDryRun(client, model, session, variableService)
	require.NoError(t, err)

	// Verify variables were set
	output, err := variableService.Get("_output")
	require.NoError(t, err)
	assert.Equal(t, "DRY RUN - No API call made", output)

	dryRunMode, err := variableService.Get("#dry_run_mode")
	require.NoError(t, err)
	assert.Equal(t, "true", dryRunMode)

	messageCount, err := variableService.Get("#dry_run_message_count")
	require.NoError(t, err)
	assert.Equal(t, "3", messageCount)
}

func TestCallCommand_handleSyncCall(t *testing.T) {
	cmd := &CallCommand{}

	// Create mock services with registry setup
	registry := services.NewRegistry()
	_ = registry.RegisterService(services.NewMockLLMService())
	_ = registry.RegisterService(services.NewClientFactoryService())
	_ = registry.RegisterService(services.NewVariableService())
	_ = registry.RegisterService(services.NewDebugTransportService())
	_ = registry.InitializeAll()

	// Set up global registry for the test
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(registry)
	defer services.SetGlobalRegistry(oldRegistry)

	// Get services from registry
	llmService, _ := services.GetGlobalLLMService()
	clientFactory, _ := services.GetGlobalClientFactoryService()
	variableService, _ := services.GetGlobalVariableService()

	client, clientID, _ := clientFactory.GetClientWithID("OAR", "test-key")
	_ = clientID // Use clientID to avoid unused variable error

	model := &neurotypes.ModelConfig{
		ID:        "test-model-id",
		Name:      "test-model",
		Provider:  "openai",
		BaseModel: "gpt-4",
	}

	session := &neurotypes.ChatSession{
		ID:   "test-session-id",
		Name: "test-session",
		Messages: []neurotypes.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	// Test sync call
	err := cmd.handleSyncCall(llmService, client, session, model, variableService)
	require.NoError(t, err)

	// Verify variables were set
	output, err := variableService.Get("_output")
	require.NoError(t, err)
	assert.Contains(t, output, "mocking reply")

	callMode, err := variableService.Get("#llm_call_mode")
	require.NoError(t, err)
	assert.Equal(t, "http", callMode)
}
