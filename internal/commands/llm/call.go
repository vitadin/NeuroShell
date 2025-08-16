// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"
	"strings"
	"time"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// CallCommand implements the \llm-call command for orchestrating LLM API calls.
// It provides pure service orchestration without message manipulation.
type CallCommand struct{}

// Name returns the command name "llm-call" for registration and lookup.
func (c *CallCommand) Name() string {
	return "llm-call"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *CallCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the llm-call command does.
func (c *CallCommand) Description() string {
	return "Orchestrate LLM API call using client, model, and session services"
}

// Usage returns the syntax and usage examples for the llm-call command.
func (c *CallCommand) Usage() string {
	return `\llm-call[client_id=client_id, model_id=model_id, session_id=session_id, stream=false, dry_run=false]

Examples:
  \llm-call                                                %% Use defaults (active model, active session, cached client)
  \llm-call[client_id=${_client_id}, model_id=my-gpt4]     %% Explicit client and model
  \llm-call[session_id=work-session, stream=true]          %% Use specific session with streaming
  \llm-call[dry_run=true]                                  %% Show what would be sent without API call
  \llm-call[client_id=OAR:a1b2c3d4, model_id=creative-gpt4, session_id=creative-work]

Options:
  client_id     - LLM client ID (defaults to ${_client_id})
  model_id      - Model configuration ID (defaults to active model)
  session_id    - Session ID (defaults to active session)
  stream        - Enable streaming mode (default: false)
  dry_run       - Show API payload without making call (default: false)

Notes:
  - This command does NOT accept input messages
  - Use \session-add-usermsg to add messages to sessions
  - Response stored in ${_output} and ${#llm_response} variables
  - Network debug data always available in ${_debug_network}
  - Use \session-add-assistantmsg to add response to session`
}

// HelpInfo returns structured help information for the llm-call command.
func (c *CallCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       `\llm-call[client_id=client_id, model_id=model_id, session_id=session_id, stream=false, dry_run=false]`,
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "client_id",
				Description: "LLM client ID (from \\*-client-new + \\llm-client-activate)",
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
			"Network debug data always available in ${_debug_network}",
			"All parameters support variable interpolation",
		},
	}
}

// Execute orchestrates an LLM API call using the three independent services.
func (c *CallCommand) Execute(args map[string]string, input string) error {
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
		return fmt.Errorf("client_id not specified and ${_client_id} not set. Use \\*-client-new and \\llm-client-activate first")
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

	// Validate session has messages before making LLM call
	if len(session.Messages) == 0 {
		// For dry run, show warnings but continue to display debug info
		if args["dry_run"] == "true" {
			return c.handleDryRun(client, model, session, variableService)
		}
		// For actual calls, error out immediately
		return fmt.Errorf("session '%s' contains no messages. Use \\session-add-usermsg to add messages before calling LLM", session.Name)
	}

	// Handle dry_run option
	if args["dry_run"] == "true" {
		return c.handleDryRun(client, model, session, variableService)
	}

	// Make LLM call (pure service orchestration)
	// Check streaming mode from both args and _stream variable
	stream := args["stream"] == "true"
	if !stream {
		// Check if _stream variable is set to true (case-insensitive)
		if streamVar, err := variableService.Get("_stream"); err == nil {
			stream = strings.ToLower(strings.TrimSpace(streamVar)) == "true"
		}
	}

	if stream {
		return c.handleStreamingCall(llmService, client, session, model, variableService)
	}
	return c.handleSyncCall(llmService, client, session, model, variableService)
}

// handleDryRun shows the complete API payload that would be sent without making the call.
func (c *CallCommand) handleDryRun(client neurotypes.LLMClient, model *neurotypes.ModelConfig, session *neurotypes.ChatSession, variableService *services.VariableService) error {
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
		fmt.Println("\n⚠️  WARNING: This session contains no messages!")
		fmt.Println("    Actual \\llm-call would fail with: session contains no messages")
		fmt.Println("    Use \\session-add-usermsg to add messages before calling LLM")
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

// handleSyncCall performs a synchronous LLM API call with always-on debug transport.
func (c *CallCommand) handleSyncCall(llmService neurotypes.LLMService, client neurotypes.LLMClient, session *neurotypes.ChatSession, model *neurotypes.ModelConfig, variableService *services.VariableService) error {
	// Start temporal display for thinking indicator
	displayID := "llm-call-sync"
	displayStarted := c.startLLMThinkingDisplay(displayID, "Thinking...")

	// Ensure display is stopped on function exit (success or error)
	if displayStarted {
		defer c.stopLLMDisplay(displayID)
	}

	// Get debug transport service and inject debug transport into client
	debugTransportService, err := services.GetGlobalDebugTransportService()
	if err != nil {
		return fmt.Errorf("debug transport service not available: %w", err)
	}

	// Create and inject debug transport into the client
	debugTransport := debugTransportService.CreateTransport()
	client.SetDebugTransport(debugTransport)

	// Make structured LLM call (debug capture happens automatically via transport)
	structuredResponse, err := llmService.SendStructuredCompletion(client, session, model)
	if err != nil {
		return fmt.Errorf("LLM call failed: %w", err)
	}

	// Get captured debug data from the debug transport service
	debugData := debugTransportService.GetCapturedData()

	// Render thinking blocks if present and store them separately
	var renderedThinking string
	if len(structuredResponse.ThinkingBlocks) > 0 {
		// Get thinking renderer service
		thinkingRenderer, err := services.GetGlobalThinkingRendererService()
		if err != nil {
			// No thinking renderer available - store empty string
			renderedThinking = ""
		} else {
			// Create render configuration with theme integration
			renderConfig, err := c.createRenderConfig(variableService)
			if err != nil {
				// Create a default render configuration if theme service fails
				renderConfig = c.createDefaultRenderConfig()
			}
			// Calculate the message index for the assistant's response (next message in session)
			messageIndex := len(session.Messages) + 1
			// Render thinking blocks with XML format and proper message indexing
			renderedThinking = thinkingRenderer.RenderThinkingBlocksWithMessageIndex(structuredResponse.ThinkingBlocks, renderConfig, messageIndex)
		}
	} else {
		// No thinking blocks
		renderedThinking = ""
	}

	// For backward compatibility, create full response (text + thinking)
	// but also store components separately for scripts to use
	var fullResponse string
	if renderedThinking != "" {
		fullResponse = renderedThinking + structuredResponse.TextContent
	} else {
		fullResponse = structuredResponse.TextContent
	}

	// Store response and debug data in variables
	_ = variableService.SetSystemVariable("_output", fullResponse)                             // Backward compatibility
	_ = variableService.SetSystemVariable("#llm_response", fullResponse)                       // Backward compatibility
	_ = variableService.SetSystemVariable("#llm_text_content", structuredResponse.TextContent) // Clean text only
	_ = variableService.SetSystemVariable("#llm_thinking_blocks_rendered", renderedThinking)   // Rendered thinking blocks
	_ = variableService.SetSystemVariable("#llm_thinking_blocks_count", fmt.Sprintf("%d", len(structuredResponse.ThinkingBlocks)))
	_ = variableService.SetSystemVariable("#llm_call_success", "true")
	_ = variableService.SetSystemVariable("#llm_call_mode", "sync")
	_ = variableService.SetSystemVariable("_debug_network", debugData)

	// Store error information if present
	if structuredResponse.Error != nil {
		_ = variableService.SetSystemVariable("#llm_error_code", structuredResponse.Error.Code)
		_ = variableService.SetSystemVariable("#llm_error_message", structuredResponse.Error.Message)
		_ = variableService.SetSystemVariable("#llm_error_type", structuredResponse.Error.Type)
	} else {
		_ = variableService.SetSystemVariable("#llm_error_code", "")
		_ = variableService.SetSystemVariable("#llm_error_message", "")
		_ = variableService.SetSystemVariable("#llm_error_type", "")
	}

	// Clear debug data for next call
	debugTransportService.ClearCapturedData()

	// Don't output response here - let calling script handle formatting
	return nil
}

// handleStreamingCall performs a streaming LLM API call.
func (c *CallCommand) handleStreamingCall(llmService neurotypes.LLMService, client neurotypes.LLMClient, session *neurotypes.ChatSession, model *neurotypes.ModelConfig, variableService *services.VariableService) error {
	// Start temporal display for connection indicator
	displayID := "llm-call-stream"
	displayStarted := c.startLLMThinkingDisplay(displayID, "Connecting...")

	// Ensure display is stopped if we exit early due to error
	if displayStarted {
		defer func() {
			// Only stop if still active (in case we stopped it manually below)
			if temporalService := c.getTemporalDisplayService(); temporalService != nil {
				if temporalService.IsActive(displayID) {
					c.stopLLMDisplay(displayID)
				}
			}
		}()
	}

	// Get debug transport service and inject debug transport into client
	debugTransportService, err := services.GetGlobalDebugTransportService()
	if err != nil {
		return fmt.Errorf("debug transport service not available: %w", err)
	}

	// Create and inject debug transport into the client
	debugTransport := debugTransportService.CreateTransport()
	client.SetDebugTransport(debugTransport)

	// Pure service orchestration for streaming (debug capture happens automatically via transport)
	stream, err := llmService.StreamCompletion(client, session, model)
	if err != nil {
		return fmt.Errorf("streaming LLM call failed: %w", err)
	}

	// Switch to streaming content display and accumulate response
	var fullResponse strings.Builder
	var streamingContent strings.Builder

	// Start streaming content display
	streamingStarted := false
	for chunk := range stream {
		if chunk.Error != nil {
			return fmt.Errorf("streaming error: %w", chunk.Error)
		}

		// Switch from "Connecting..." to streaming content display on first chunk
		if !streamingStarted {
			if displayStarted {
				c.stopLLMDisplay(displayID)
			}
			displayStarted = c.startStreamingContentDisplay(displayID, &streamingContent)
			streamingStarted = true
		}

		// Accumulate response and update streaming display content
		fullResponse.WriteString(chunk.Content)
		streamingContent.WriteString(chunk.Content)
	}

	// Get captured debug data from the debug transport service
	debugData := debugTransportService.GetCapturedData()

	// Store complete response and debug data
	response := fullResponse.String()
	_ = variableService.SetSystemVariable("_output", response)
	_ = variableService.SetSystemVariable("#llm_response", response)
	_ = variableService.SetSystemVariable("#llm_call_success", "true")
	_ = variableService.SetSystemVariable("#llm_call_mode", "stream")
	_ = variableService.SetSystemVariable("_debug_network", debugData)

	// Clear debug data for next call
	debugTransportService.ClearCapturedData()

	// Don't output response here - let _send.neuro handle final markdown rendering
	return nil
}

// getTemporalDisplayService attempts to get the temporal display service.
// Returns nil if service is not available (graceful degradation).
func (c *CallCommand) getTemporalDisplayService() *services.TemporalDisplayService {
	serviceInterface, err := services.GetGlobalRegistry().GetService("temporal-display")
	if err != nil {
		return nil // Graceful degradation
	}

	temporalService, ok := serviceInterface.(*services.TemporalDisplayService)
	if !ok {
		return nil // Graceful degradation
	}

	return temporalService
}

// startLLMThinkingDisplay starts a temporal display showing "Thinking..." with elapsed time.
// Returns true if display was started successfully, false otherwise.
func (c *CallCommand) startLLMThinkingDisplay(id string, message string) bool {
	temporalService := c.getTemporalDisplayService()
	if temporalService == nil {
		return false // Graceful degradation
	}

	// Create a custom renderer for LLM thinking display
	renderer := func(elapsed time.Duration) string {
		seconds := int(elapsed.Seconds())
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
		return style.Render(fmt.Sprintf("%s %ds", message, seconds))
	}

	// Create a condition that never stops automatically (we'll stop manually)
	condition := func(_ time.Duration) bool {
		return false // Never auto-stop, we'll stop manually
	}

	err := temporalService.StartCustomDisplay(id, condition, renderer)
	return err == nil
}

// startStreamingContentDisplay starts a temporal display showing actual streaming content.
// Returns true if display was started successfully, false otherwise.
func (c *CallCommand) startStreamingContentDisplay(id string, content *strings.Builder) bool {
	temporalService := c.getTemporalDisplayService()
	if temporalService == nil {
		return false // Graceful degradation
	}

	// Create a custom renderer for streaming content display
	renderer := func(elapsed time.Duration) string {
		seconds := int(elapsed.Seconds())
		currentContent := content.String()

		// Use single-line display for reliability (no multi-line stacking issues)
		// Show the last meaningful chunk of content
		displayContent := currentContent

		// Get last 200 characters for preview
		if len(displayContent) > 80 {
			displayContent = "..." + displayContent[len(displayContent)-77:]
		}

		// Replace newlines and tabs with spaces for single-line display
		displayContent = strings.ReplaceAll(displayContent, "\n", " ")
		displayContent = strings.ReplaceAll(displayContent, "\t", " ")

		// Collapse multiple spaces
		for strings.Contains(displayContent, "  ") {
			displayContent = strings.ReplaceAll(displayContent, "  ", " ")
		}

		// Trim and ensure reasonable length
		displayContent = strings.TrimSpace(displayContent)
		if len(displayContent) > 80 {
			displayContent = displayContent[:77] + "..."
		}

		// Show character count and preview
		charCount := len(currentContent)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow for streaming
		return style.Render(fmt.Sprintf("Streaming (%ds): %d chars | %s", seconds, charCount, displayContent))
	}

	// Create a condition that never stops automatically (we'll stop manually)
	condition := func(_ time.Duration) bool {
		return false // Never auto-stop, we'll stop manually
	}

	err := temporalService.StartCustomDisplay(id, condition, renderer)
	return err == nil
}

// stopLLMDisplay stops a temporal display and handles any errors gracefully.
func (c *CallCommand) stopLLMDisplay(id string) {
	temporalService := c.getTemporalDisplayService()
	if temporalService == nil {
		return // Nothing to stop
	}

	// Stop the display, ignore errors (graceful degradation)
	_ = temporalService.Stop(id)
}

// createRenderConfig creates a RenderConfig that integrates with the theme service.
func (c *CallCommand) createRenderConfig(variableService *services.VariableService) (neurotypes.RenderConfig, error) {
	// Get theme service for styling
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		return nil, fmt.Errorf("theme service not available: %w", err)
	}

	// Get current theme (try from variable, fallback to default)
	themeName := "default"
	if themeVar, err := variableService.Get("_theme"); err == nil && themeVar != "" {
		themeName = themeVar
	}

	// Get theme configuration
	theme := themeService.GetThemeByName(themeName)

	// Create the command render config
	config := &CommandRenderConfig{
		theme:         theme,
		themeName:     themeName,
		showThinking:  true, // Default to showing thinking blocks
		thinkingStyle: "full",
		compactMode:   false,
		maxWidth:      120, // Default terminal width
	}

	// Check for thinking display preferences in variables
	if thinkingVar, err := variableService.Get("_thinking_display"); err == nil {
		switch strings.ToLower(strings.TrimSpace(thinkingVar)) {
		case "hidden", "false", "off":
			config.showThinking = false
		case "summary", "compact":
			config.thinkingStyle = "summary"
		case "full", "true", "on":
			config.thinkingStyle = "full"
		}
	}

	// Check for compact mode preference
	if compactVar, err := variableService.Get("_compact_mode"); err == nil {
		config.compactMode = strings.ToLower(strings.TrimSpace(compactVar)) == "true"
	}

	return config, nil
}

// createDefaultRenderConfig creates a basic render configuration when theme service is not available.
func (c *CallCommand) createDefaultRenderConfig() neurotypes.RenderConfig {
	return &services.DefaultRenderConfig{
		ShowThinkingEnabled: true,
		ThinkingStyleValue:  "full",
		CompactModeEnabled:  false,
		MaxWidthValue:       120,
		ThemeValue:          "default",
	}
}

// CommandRenderConfig implements neurotypes.RenderConfig using the theme service.
type CommandRenderConfig struct {
	theme         *services.Theme
	themeName     string
	showThinking  bool
	thinkingStyle string
	compactMode   bool
	maxWidth      int
}

// GetStyle returns theme-based styles for different elements.
func (c *CommandRenderConfig) GetStyle(element string) lipgloss.Style {
	if c.theme == nil {
		return lipgloss.NewStyle()
	}

	// Get the base style from theme
	var baseStyle lipgloss.Style
	switch element {
	case "info":
		baseStyle = c.theme.Info
	case "italic":
		baseStyle = c.theme.Italic
	case "background":
		baseStyle = c.theme.Background
	case "warning":
		baseStyle = c.theme.Warning
	case "highlight":
		baseStyle = c.theme.Highlight
	case "bold":
		baseStyle = c.theme.Bold
	case "underline":
		baseStyle = c.theme.Underline
	default:
		baseStyle = c.theme.Info // Default to info style
	}

	// Check if colors should be disabled (NO_COLOR environment variable or --no-color flag)
	// This respects the global color profile setting
	if lipgloss.ColorProfile() == termenv.Ascii {
		// Strip all colors from the style, keeping only formatting
		return lipgloss.NewStyle().
			Bold(baseStyle.GetBold()).
			Italic(baseStyle.GetItalic()).
			Underline(baseStyle.GetUnderline()).
			Strikethrough(baseStyle.GetStrikethrough()).
			Reverse(baseStyle.GetReverse()).
			Blink(baseStyle.GetBlink()).
			Faint(baseStyle.GetFaint())
	}

	return baseStyle
}

// GetTheme returns the current theme name.
func (c *CommandRenderConfig) GetTheme() string {
	return c.themeName
}

// IsCompactMode returns whether compact rendering is enabled.
func (c *CommandRenderConfig) IsCompactMode() bool {
	return c.compactMode
}

// GetMaxWidth returns the maximum width for content rendering.
func (c *CommandRenderConfig) GetMaxWidth() int {
	return c.maxWidth
}

// ShowThinking returns whether thinking blocks should be displayed.
func (c *CommandRenderConfig) ShowThinking() bool {
	return c.showThinking
}

// GetThinkingStyle returns the thinking display style preference.
func (c *CommandRenderConfig) GetThinkingStyle() string {
	return c.thinkingStyle
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&CallCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register llm-call command: %v", err))
	}
}
