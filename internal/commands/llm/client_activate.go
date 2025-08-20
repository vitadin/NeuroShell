// Package llm contains LLM-related commands for the NeuroShell CLI.
package llm

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ClientActivateCommand implements the \llm-client-activate command for activating LLM clients.
// It provides clean separation between client creation and activation, supporting both
// provider catalog ID mode and specific client ID mode.
type ClientActivateCommand struct{}

// Name returns the command name "llm-client-activate" for registration and lookup.
func (c *ClientActivateCommand) Name() string {
	return "llm-client-activate"
}

// ParseMode returns ParseModeRaw for direct input parameter parsing.
func (c *ClientActivateCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeRaw
}

// Description returns a brief description of what the llm-client-activate command does.
func (c *ClientActivateCommand) Description() string {
	return "Activate LLM client by provider catalog ID or specific client ID"
}

// Usage returns the syntax and usage examples for the llm-client-activate command.
func (c *ClientActivateCommand) Usage() string {
	return `\llm-client-activate provider_catalog_id_or_client_id

Examples:
  \llm-client-activate OAR                    %% Activate by provider catalog ID (OpenAI Reasoning)
  \llm-client-activate OAC                    %% Activate by provider catalog ID (OpenAI Chat)  
  \llm-client-activate ANC                    %% Activate by provider catalog ID (Anthropic)
  \llm-client-activate GMC                    %% Activate by provider catalog ID (Gemini)
  \llm-client-activate OAR:a3f2cae8           %% Activate by specific client ID
  \llm-client-activate ANC:b4c5d7e9           %% Activate by specific Anthropic client ID

Notes:
  - Provider catalog IDs: OAR (OpenAI Reasoning), OAC (OpenAI Chat), ANC (Anthropic), GMC (Gemini)
  - Client IDs have format: ProviderCatalogID:hash (e.g., OAR:a3f2cae8)
  - Sets ${#active_client_id} to the activated client
  - If input contains ':', treated as specific client ID
  - If input has no ':', treated as provider catalog ID and finds any matching client`
}

// HelpInfo returns structured help information for the llm-client-activate command.
func (c *ClientActivateCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\llm-client-activate provider_catalog_id_or_client_id",
		ParseMode:   c.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\llm-client-activate OAR",
				Description: "Activate any OpenAI reasoning client",
			},
			{
				Command:     "\\llm-client-activate OAC",
				Description: "Activate any OpenAI chat client",
			},
			{
				Command:     "\\llm-client-activate ANC",
				Description: "Activate any Anthropic client",
			},
			{
				Command:     "\\llm-client-activate OAR:a3f2cae8",
				Description: "Activate specific OpenAI reasoning client by ID",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "#active_client_id",
				Description: "ID of the currently active client",
				Type:        "system_metadata",
				Example:     "OAR:a3f2cae8",
			},
			{
				Name:        "_output",
				Description: "Command result message",
				Type:        "command_output",
				Example:     "Activated client OAR:a3f2cae8 (type: OpenAI Reasoning)",
			},
		},
		Notes: []string{
			"Supports two input modes: provider catalog ID or specific client ID",
			"Provider catalog ID mode finds any existing client of that type",
			"Client ID mode activates the exact client specified",
			"Sets #active_client_id system variable for use by \\llm-call",
			"Use \\*-client-new commands to create clients before activation",
		},
	}
}

// Execute activates an LLM client by provider catalog ID or specific client ID.
func (c *ClientActivateCommand) Execute(_ map[string]string, input string) error {
	// Validate input
	identifier := strings.TrimSpace(input)
	if identifier == "" {
		return fmt.Errorf("provider catalog ID or client ID is required\n\nUsage: %s", c.Usage())
	}

	// Get required services
	clientFactory, err := services.GetGlobalClientFactoryService()
	if err != nil {
		return fmt.Errorf("client factory service not available: %w", err)
	}

	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	var clientID string

	// Determine mode based on input format
	if strings.Contains(identifier, ":") {
		// Mode 2: Specific client ID (e.g., "OAR:a3f2cae8")
		_, err = clientFactory.GetClientByID(identifier)
		if err != nil {
			return fmt.Errorf("client with ID '%s' not found: %w", identifier, err)
		}
		clientID = identifier
	} else {
		// Mode 1: Provider catalog ID (e.g., "OAR", "OAC", "ANC")
		_, clientID, err = clientFactory.FindClientByProviderCatalogID(identifier)
		if err != nil {
			return fmt.Errorf("no client found for provider catalog ID '%s': %w", identifier, err)
		}
	}

	// Set the active client
	if err := variableService.SetSystemVariable("#active_client_id", clientID); err != nil {
		return fmt.Errorf("failed to set active client ID: %w", err)
	}

	// Prepare success message with client type information
	clientTypeInfo := c.getClientTypeDescription(clientID)
	outputMsg := fmt.Sprintf("Activated client %s (%s)", clientID, clientTypeInfo)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// getClientTypeDescription returns a human-readable description of the client type.
func (c *ClientActivateCommand) getClientTypeDescription(clientID string) string {
	switch {
	case strings.HasPrefix(clientID, "OAR:"):
		return "type: OpenAI Reasoning/Dual-mode"
	case strings.HasPrefix(clientID, "OAC:"):
		return "type: OpenAI Chat"
	case strings.HasPrefix(clientID, "ANC:"):
		return "type: Anthropic"
	case strings.HasPrefix(clientID, "GMC:"):
		return "type: Gemini"
	default:
		return "type: Unknown"
	}
}

// IsReadOnly returns false as the llm command modifies system state.
func (c *ClientActivateCommand) IsReadOnly() bool {
	return false
}
func init() {
	if err := commands.GetGlobalRegistry().Register(&ClientActivateCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register llm-client-activate command: %v", err))
	}
}
