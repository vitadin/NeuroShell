package builtin

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// LLMCallCommand implements the \llm-call command for orchestrating LLM API calls.
// It provides pure service orchestration without message manipulation.
type LLMCallCommand struct{}

// Name returns the command name "llm-call" for registration and lookup.
func (c *LLMCallCommand) Name() string {
	return "llm-call"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *LLMCallCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the llm-call command does.
func (c *LLMCallCommand) Description() string {
	return "Orchestrate LLM API call using client, model, and session services"
}

// Usage returns the syntax and usage examples for the llm-call command.
func (c *LLMCallCommand) Usage() string {
	return `\llm-call[client_id=client_id, model_id=model_id, session_id=session_id, stream=false, dry_run=false]

Examples:
  \llm-call                                                  %% Use defaults (active model, active session, cached client)
  \llm-call[client_id=${_client_id}, model_id=my-gpt4]     %% Explicit client and model
  \llm-call[session_id=work-session, stream=true]          %% Use specific session with streaming
  \llm-call[dry_run=true]                                  %% Show what would be sent without API call
  \llm-call[client_id=openai:a1b2c3d4, model_id=creative-gpt4, session_id=creative-work]

Options:
  client_id  - LLM client ID (defaults to ${_client_id})
  model_id   - Model configuration ID (defaults to active model)
  session_id - Session ID (defaults to active session)
  stream     - Enable streaming mode (default: false)
  dry_run    - Show API payload without making call (default: false)

Notes:
  - This command does NOT accept input messages
  - Use \session-add-usermsg to add messages to sessions
  - Response stored in ${_output} and ${#llm_response} variables
  - Use \session-add-assistantmsg to add response to session`
}

// HelpInfo returns structured help information for the llm-call command.
func (c *LLMCallCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       `\llm-call[client_id=client_id, model_id=model_id, session_id=session_id, stream=false, dry_run=false]`,
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "client_id",
				Description: "LLM client ID (from \\llm-client-get)",
				Required:    false,
				Type:        "string",
				Default:     "${_client_id}",
			},
			{
				Name:        "model_id",
				Description: "Model configuration ID",
				Required:    false,
				Type:        "string",
				Default:     "active model",
			},
			{
				Name:        "session_id",
				Description: "Session ID for conversation context",
				Required:    false,
				Type:        "string",
				Default:     "active session",
			},
			{
				Name:        "stream",
				Description: "Enable streaming response mode",
				Required:    false,
				Type:        "boolean",
				Default:     "false",
			},
			{
				Name:        "dry_run",
				Description: "Show API payload without making actual call",
				Required:    false,
				Type:        "boolean",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     `\llm-call`,
				Description: "Use all defaults (cached client, active model, active session)",
			},
			{
				Command:     `\llm-call[model_id=my-gpt4, stream=true]`,
				Description: "Use specific model with streaming enabled",
			},
			{
				Command:     `\llm-call[dry_run=true]`,
				Description: "Preview API payload without making call",
			},
		},
		Notes: []string{
			"Pure service orchestration - does not modify sessions",
			"Input messages are ignored with warning - use \\session-add-usermsg",
			"Combines three independent components: client, model, session",
			"dry_run option shows complete API payload for debugging",
			"Response stored in ${_output} for use with \\session-add-assistantmsg",
			"All parameters support variable interpolation",
		},
	}
}

// Execute orchestrates an LLM API call using the three independent services.
func (c *LLMCallCommand) Execute(args map[string]string, input string) error {
	// IMPORTANT: Warn and discard any input message
	if input != "" {
		fmt.Printf("⚠️  Warning: \\llm-call does not accept input messages. Use \\session-add-usermsg first.\n")
		fmt.Printf("   Discarding input: %q\n", input)
	}

	// Get all required services
	clientFactory, err := services.GetGlobalClientFactoryService()
	if err != nil {
		return fmt.Errorf("client factory service not available: %w", err)
	}

	modelService, err := services.GetGlobalModelService()
	if err != nil {
		return fmt.Errorf("model service not available: %w", err)
	}

	sessionService, err := services.GetGlobalChatSessionService()
	if err != nil {
		return fmt.Errorf("session service not available: %w", err)
	}

	llmService, err := services.GetGlobalLLMService()
	if err != nil {
		return fmt.Errorf("llm service not available: %w", err)
	}

	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Resolve component IDs (with defaults)
	clientID := args["client_id"]
	if clientID == "" {
		if storedClientID, err := variableService.Get("_client_id"); err == nil {
			clientID = storedClientID
		}
	}
	if clientID == "" {
		return fmt.Errorf("client_id not specified and ${_client_id} not set. Use \\llm-client-get first")
	}

	var model *neurotypes.ModelConfig
	modelID := args["model_id"]
	if modelID == "" {
		// Try to get active model directly
		if activeModel, err := modelService.GetActiveModelConfigWithGlobalContext(); err == nil && activeModel != nil {
			model = activeModel
		}
	}
	if modelID == "" && model == nil {
		return fmt.Errorf("model_id not specified and no active model set. Use \\model-activate or specify model_id")
	}

	sessionID := args["session_id"]
	if sessionID == "" {
		// Try to get active session
		if activeSession, err := sessionService.GetActiveSession(); err == nil && activeSession != nil {
			sessionID = activeSession.ID
		}
	}
	if sessionID == "" {
		return fmt.Errorf("session_id not specified and no active session set. Use \\session-new or specify session_id")
	}

	// Retrieve the three independent components
	client, err := clientFactory.GetClientByID(clientID)
	if err != nil {
		return fmt.Errorf("failed to get client '%s': %w", clientID, err)
	}

	// Get model by name if not already obtained from active model
	if model == nil {
		var err error
		model, err = modelService.GetModelByNameWithGlobalContext(modelID)
		if err != nil {
			return fmt.Errorf("failed to get model '%s': %w", modelID, err)
		}
	}

	session, err := sessionService.GetSessionByNameOrID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session '%s': %w", sessionID, err)
	}

	// Handle dry_run option
	if args["dry_run"] == "true" {
		return c.handleDryRun(client, model, session, variableService)
	}

	// Make LLM call (pure service orchestration)
	stream := args["stream"] == "true"
	if stream {
		return c.handleStreamingCall(llmService, client, session, model, variableService)
	}
	return c.handleSyncCall(llmService, client, session, model, variableService)
}

// handleDryRun shows the complete API payload that would be sent without making the call.
func (c *LLMCallCommand) handleDryRun(client neurotypes.LLMClient, model *neurotypes.ModelConfig, session *neurotypes.ChatSession, variableService *services.VariableService) error {
	fmt.Println("=== LLM CALL DRY RUN ===")
	// Get client ID from variable service for debugging display
	clientID, _ := variableService.Get("_client_id")
	fmt.Printf("Client: %s (%s)\n", clientID, client.GetProviderName())
	fmt.Printf("Model: %s (Base: %s, Provider: %s)\n", model.Name, model.BaseModel, model.Provider)
	fmt.Printf("Session: %s (%d messages)\n", session.Name, len(session.Messages))

	fmt.Println("\n=== MODEL CONFIGURATION ===")
	if len(model.Parameters) == 0 {
		fmt.Println("No parameters set")
	} else {
		for k, v := range model.Parameters {
			fmt.Printf("%s: %v\n", k, v)
		}
	}

	fmt.Println("\n=== SESSION PAYLOAD (EXACT API FORMAT) ===")
	if session.SystemPrompt != "" {
		fmt.Printf("System: %s\n", session.SystemPrompt)
	} else {
		fmt.Println("System: (no system prompt)")
	}

	fmt.Println("Messages:")
	if len(session.Messages) == 0 {
		fmt.Println("  (no messages in session)")
	} else {
		for i, msg := range session.Messages {
			fmt.Printf("  [%d] %s: %s\n", i+1, msg.Role, msg.Content)
		}
	}

	fmt.Printf("\nTotal Messages: %d\n", len(session.Messages))

	// Store dry run results
	_ = variableService.SetSystemVariable("_output", "DRY RUN - No API call made")
	_ = variableService.SetSystemVariable("#dry_run_mode", "true")
	_ = variableService.SetSystemVariable("#dry_run_client", clientID)
	_ = variableService.SetSystemVariable("#dry_run_model", model.Name)
	_ = variableService.SetSystemVariable("#dry_run_session", session.Name)
	_ = variableService.SetSystemVariable("#dry_run_message_count", fmt.Sprintf("%d", len(session.Messages)))

	return nil
}

// handleSyncCall performs a synchronous LLM API call.
func (c *LLMCallCommand) handleSyncCall(llmService neurotypes.LLMService, client neurotypes.LLMClient, session *neurotypes.ChatSession, model *neurotypes.ModelConfig, variableService *services.VariableService) error {
	// Pure service orchestration - no message manipulation
	response, err := llmService.SendCompletion(client, session, model)
	if err != nil {
		return fmt.Errorf("LLM call failed: %w", err)
	}

	// Store response in variables
	_ = variableService.SetSystemVariable("_output", response)
	_ = variableService.SetSystemVariable("#llm_response", response)
	_ = variableService.SetSystemVariable("#llm_call_success", "true")
	_ = variableService.SetSystemVariable("#llm_call_mode", "sync")

	// Output response (read-only display)
	fmt.Println(response)
	return nil
}

// handleStreamingCall performs a streaming LLM API call.
func (c *LLMCallCommand) handleStreamingCall(llmService neurotypes.LLMService, client neurotypes.LLMClient, session *neurotypes.ChatSession, model *neurotypes.ModelConfig, variableService *services.VariableService) error {
	// Pure service orchestration for streaming
	stream, err := llmService.StreamCompletion(client, session, model)
	if err != nil {
		return fmt.Errorf("streaming LLM call failed: %w", err)
	}

	var fullResponse strings.Builder
	for chunk := range stream {
		if chunk.Error != nil {
			return fmt.Errorf("streaming error: %w", chunk.Error)
		}
		fmt.Print(chunk.Content)
		fullResponse.WriteString(chunk.Content)
	}
	fmt.Println() // Final newline

	// Store complete response
	response := fullResponse.String()
	_ = variableService.SetSystemVariable("_output", response)
	_ = variableService.SetSystemVariable("#llm_response", response)
	_ = variableService.SetSystemVariable("#llm_call_success", "true")
	_ = variableService.SetSystemVariable("#llm_call_mode", "stream")

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&LLMCallCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register llm-call command: %v", err))
	}
}
