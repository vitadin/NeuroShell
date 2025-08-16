// Package model provides commands for managing LLM model configurations.
package model

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// OpenAIModelNewCommand implements the \openai-model-new command for creating OpenAI-specific models.
// It handles OpenAI reasoning models and their specialized parameters.
type OpenAIModelNewCommand struct{}

// Name returns the command name "openai-model-new" for registration and lookup.
func (c *OpenAIModelNewCommand) Name() string {
	return "openai-model-new"
}

// ParseMode returns ParseModeKeyValue for bracket parameter parsing.
func (c *OpenAIModelNewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of the openai-model-new command.
func (c *OpenAIModelNewCommand) Description() string {
	return "Create OpenAI model configurations with reasoning support"
}

// Usage returns the syntax and usage examples for the openai-model-new command.
func (c *OpenAIModelNewCommand) Usage() string {
	return `\openai-model-new[catalog_id=<ID>, reasoning_effort=medium, max_output_tokens=10000, ...] model_name

Examples:
  \openai-model-new[catalog_id=O3] my-o3                                    %% Create o3 model with defaults
  \openai-model-new[catalog_id=O3, reasoning_effort=high] thorough-o3       %% Create o3 with high reasoning effort
  \openai-model-new[catalog_id=O4M, reasoning_effort=low, max_output_tokens=5000] efficient-o4 %% Create efficient o4-mini
  \openai-model-new[catalog_id=G4O, reasoning_effort=medium, reasoning_summary=detailed] reasoning-gpt4o %% Create GPT-4o with reasoning summaries
  \openai-model-new[catalog_id=O1P, max_output_tokens=15000] powerful-o1    %% Create o1-pro with large token limit

Options:
  catalog_id - Short model ID from catalog (O3, O4M, O1, O1P, G4O, G41, etc.)
  reasoning_effort - Reasoning effort level (low, medium, high) - controls reasoning token usage
  max_output_tokens - Maximum total output tokens including reasoning tokens
  reasoning_summary - Enable reasoning summaries (auto only, recommended by OpenAI)
  temperature - Sampling temperature (0.0-2.0)
  max_tokens - Maximum completion tokens (for non-reasoning models)
  top_p - Nucleus sampling parameter (0.0-1.0)
  presence_penalty - Presence penalty (-2.0 to 2.0)
  frequency_penalty - Frequency penalty (-2.0 to 2.0)
  description - Human-readable description

Note: catalog_id is required.
      Reasoning models (o3, o4-mini, o1 series) automatically use /responses API.
      Non-reasoning models (GPT-4 series without reasoning_effort) use /chat/completions API.
      Use \model-catalog[provider=openai] to see available OpenAI models.`
}

// HelpInfo returns structured help information for the openai-model-new command.
func (c *OpenAIModelNewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\openai-model-new[catalog_id=<ID>, reasoning_effort=<level>, ...] model_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "catalog_id",
				Description: "Short model ID from catalog (O3, O4M, O1, O1P, G4O, etc.)",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "reasoning_effort",
				Description: "Reasoning effort level (low, medium, high)",
				Required:    false,
				Type:        "string",
				Default:     "medium",
			},
			{
				Name:        "max_output_tokens",
				Description: "Maximum total output tokens including reasoning tokens",
				Required:    false,
				Type:        "int",
			},
			{
				Name:        "reasoning_summary",
				Description: "Enable reasoning summaries (auto only, recommended by OpenAI)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "temperature",
				Description: "Sampling temperature (0.0-2.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "max_tokens",
				Description: "Maximum completion tokens (for non-reasoning models)",
				Required:    false,
				Type:        "int",
			},
			{
				Name:        "top_p",
				Description: "Nucleus sampling parameter (0.0-1.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "presence_penalty",
				Description: "Presence penalty (-2.0 to 2.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "frequency_penalty",
				Description: "Frequency penalty (-2.0 to 2.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "description",
				Description: "Human-readable description of the model configuration",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\openai-model-new[catalog_id=O3] my-o3",
				Description: "Create o3 model with default settings",
			},
			{
				Command:     "\\openai-model-new[catalog_id=O3, reasoning_effort=high] thorough-o3",
				Description: "Create o3 model with high reasoning effort",
			},
			{
				Command:     "\\openai-model-new[catalog_id=O4M, reasoning_effort=low, max_output_tokens=5000] efficient-o4",
				Description: "Create efficient o4-mini with token limit",
			},
			{
				Command:     "\\openai-model-new[catalog_id=G4O, reasoning_effort=medium, reasoning_summary=detailed] reasoning-gpt4o",
				Description: "Create GPT-4o with reasoning summaries",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_model_id",
				Description: "Contains the created model's unique ID",
				Type:        "system_output",
				Example:     "_model_id = \"a1b2c3d4-e5f6-7890-abcd-ef1234567890\"",
			},
			{
				Name:        "_model_name",
				Description: "Contains the created model's name",
				Type:        "system_output",
				Example:     "_model_name = \"my-o3\"",
			},
			{
				Name:        "#active_model_name",
				Description: "Contains the active model name (automatically set)",
				Type:        "system_metadata",
				Example:     "#active_model_name = \"my-o3\"",
			},
			{
				Name:        "#active_model_id",
				Description: "Contains the active model ID (automatically set)",
				Type:        "system_metadata",
				Example:     "#active_model_id = \"a1b2c3d4-e5f6-7890-abcd-ef1234567890\"",
			},
			{
				Name:        "#active_model_provider",
				Description: "Contains the active model provider (always 'openai')",
				Type:        "system_metadata",
				Example:     "#active_model_provider = \"openai\"",
			},
		},
		Notes: []string{
			"catalog_id is required",
			"Use \\model-catalog[provider=openai] to see available OpenAI models",
			"Reasoning models automatically use OpenAI Responses API (/responses)",
			"Non-reasoning models use Chat Completions API (/chat/completions)",
			"reasoning_effort: low=speed/economy, medium=balanced, high=thorough",
			"max_output_tokens includes reasoning tokens (reserve ~25k for complex reasoning)",
			"reasoning_summary='auto' shows internal reasoning process (default for reasoning models)",
			"Model is automatically activated after creation",
			"Variables in model name and parameters are interpolated",
		},
	}
}

// Execute creates a new OpenAI model configuration.
func (c *OpenAIModelNewCommand) Execute(args map[string]string, input string) error {
	// Validate model name (input parameter)
	modelName := strings.TrimSpace(input)
	if modelName == "" {
		return fmt.Errorf("model name is required. Usage: %s", c.Usage())
	}

	// Get required services
	modelService, err := services.GetGlobalModelService()
	if err != nil {
		return fmt.Errorf("model service not available: %w", err)
	}

	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	catalogService, err := services.GetGlobalModelCatalogService()
	if err != nil {
		return fmt.Errorf("model catalog service not available: %w", err)
	}

	// Determine base model and provider
	provider := "openai"
	var baseModel string
	var catalogModel *neurotypes.ModelCatalogEntry

	catalogID := args["catalog_id"]

	if catalogID == "" {
		return fmt.Errorf("catalog_id is required")
	}

	// Look up model in catalog
	entry, err := catalogService.GetModelByID(catalogID)
	if err != nil {
		return fmt.Errorf("failed to find model with catalog_id '%s': %w", catalogID, err)
	}

	if entry.Provider != "openai" {
		return fmt.Errorf("catalog_id '%s' is not an OpenAI model (provider: %s). Use \\openai-model-new only for OpenAI models", catalogID, entry.Provider)
	}

	catalogModel = &entry
	baseModel = entry.Name

	// Get parameter validator service
	paramValidatorService, err := services.GetGlobalParameterValidatorService()
	if err != nil {
		return fmt.Errorf("parameter validator service not available: %w", err)
	}

	// Validate parameters using the model's parameter definitions
	parameters, err := paramValidatorService.ValidateParameters(args, catalogModel.Parameters)
	if err != nil {
		return fmt.Errorf("failed to validate parameters: %w", err)
	}

	// Get description
	description := args["description"]
	if description == "" && catalogModel != nil {
		description = fmt.Sprintf("OpenAI %s model", catalogModel.DisplayName)
	}

	// Validate parameters
	if err := modelService.ValidateModelParameters(parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create model using the proper service method
	createdModel, err := modelService.CreateModelWithGlobalContext(modelName, provider, baseModel, parameters, description, catalogID)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Set result variables
	_ = variableService.SetSystemVariable("_model_id", createdModel.ID)
	_ = variableService.SetSystemVariable("_model_name", createdModel.Name)
	_ = variableService.SetSystemVariable("_output", fmt.Sprintf("Created model '%s' (ID: %s, Provider: %s, Base: %s)", createdModel.Name, createdModel.ID[:8], createdModel.Provider, createdModel.BaseModel))

	// Set metadata variables (consistent with regular model-new command)
	_ = variableService.SetSystemVariable("#model_id", createdModel.ID)
	_ = variableService.SetSystemVariable("#model_name", createdModel.Name)
	_ = variableService.SetSystemVariable("#model_provider", createdModel.Provider)
	_ = variableService.SetSystemVariable("#model_base", createdModel.BaseModel)
	_ = variableService.SetSystemVariable("#model_created", createdModel.CreatedAt.Format("2006-01-02 15:04:05"))
	_ = variableService.SetSystemVariable("#model_param_count", fmt.Sprintf("%d", len(createdModel.Parameters)))

	// Activate the model automatically
	_ = variableService.SetSystemVariable("#active_model_name", createdModel.Name)
	_ = variableService.SetSystemVariable("#active_model_id", createdModel.ID)
	_ = variableService.SetSystemVariable("#active_model_provider", createdModel.Provider)
	_ = variableService.SetSystemVariable("#active_model_base", createdModel.BaseModel)
	if createdModel.Description != "" {
		_ = variableService.SetSystemVariable("#model_description", createdModel.Description)
	}

	// Auto-push client creation and model activation commands for seamless UX
	// Note: Stack is LIFO, so we push in reverse order of execution
	if stackService, err := services.GetGlobalStackService(); err == nil {
		// Use the model's provider catalog ID as the client type (now 1:1 mapping)
		preferredClientType := catalogModel.ProviderCatalogID

		if preferredClientType != "" {
			// Push commands in reverse order (LIFO stack)
			// 1. Push model activation (executed last)
			modelActivateCommand := fmt.Sprintf("\\silent \\model-activate[id=true] %s", createdModel.ID)
			stackService.PushCommand(modelActivateCommand)

			// 2. Push client activation (executed second)
			activateCommand := fmt.Sprintf("\\silent \\llm-client-activate %s", preferredClientType)
			stackService.PushCommand(activateCommand)

			// 3. Push client creation (executed first)
			createCommand := c.generateClientNewCommand(preferredClientType)
			if createCommand != "" {
				stackService.PushCommand(createCommand)
			}
		}
	}

	// Output success message
	outputMsg := fmt.Sprintf("Created model '%s' (ID: %s, Provider: %s, Base: %s)", createdModel.Name, createdModel.ID[:8], createdModel.Provider, createdModel.BaseModel)
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

	return nil
}

// generateClientNewCommand generates the appropriate client creation command for a client type.
func (c *OpenAIModelNewCommand) generateClientNewCommand(clientType string) string {
	switch clientType {
	case "OAC":
		return "\\silent \\openai-client-new[client_type=OAC]"
	case "OAR":
		return "\\silent \\openai-client-new[client_type=OAR]"
	default:
		return ""
	}
}

func init() {
	if err := commands.GlobalRegistry.Register(&OpenAIModelNewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register openai-model-new command: %v", err))
	}
}
