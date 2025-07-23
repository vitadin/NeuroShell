// Package model provides model management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with LLM model configurations.
package model

import (
	"fmt"
	"strconv"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// NewCommand implements the \model-new command for creating new model configurations.
// It provides model creation functionality with configurable parameters for different LLM providers.
type NewCommand struct{}

// Name returns the command name "model-new" for registration and lookup.
func (c *NewCommand) Name() string {
	return "model-new"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *NewCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the model-new command does.
func (c *NewCommand) Description() string {
	return "Create new LLM model configuration"
}

// Usage returns the syntax and usage examples for the model-new command.
func (c *NewCommand) Usage() string {
	return `\model-new[provider=provider_name, base_model=model_name, temperature=0.7, max_tokens=1000, ...] model_name
\model-new[catalog_id=<ID>, temperature=0.7, max_tokens=1000, ...] model_name

Examples:
  \model-new[catalog_id=CS4] my-claude                                   %% Create from catalog (Claude Sonnet 4)
  \model-new[catalog_id=O3, temperature=0.8] creative-model              %% Create from catalog with custom parameters
  \model-new[provider=openai, base_model=gpt-4] my-gpt4                  %% Create OpenAI GPT-4 model (manual)
  \model-new[provider=anthropic, base_model=claude-3-sonnet] claude-work %% Create Anthropic Claude model (manual)
  \model-new[catalog_id=CO4, max_tokens=4000] analysis-opus              %% Create Claude Opus 4 with custom max tokens

Required Options (choose one):
  Option A: catalog_id - Short model ID from catalog (e.g., CS4, O3, CO37) - auto-populates provider and base_model
  Option B: provider + base_model - Manual specification
    provider - LLM provider name (e.g., openai, anthropic, local)
    base_model - Provider's model identifier (e.g., gpt-4, claude-3-sonnet, llama-2)

Optional Parameters:
  temperature - Controls randomness (0.0-1.0, default varies by provider)
  max_tokens - Maximum tokens to generate (positive integer)
  top_p - Nucleus sampling parameter (0.0-1.0)
  top_k - Top-k sampling parameter (positive integer)
  presence_penalty - Presence penalty (-2.0 to 2.0)
  frequency_penalty - Frequency penalty (-2.0 to 2.0)
  description - Human-readable description of the model configuration

Note: Model name is required and taken from the input parameter.
      Model names must be unique and cannot contain spaces.
      Use \model-catalog to see available catalog IDs.
      When using catalog_id, provider and base_model are auto-populated from catalog.
      Additional provider-specific parameters can be passed and will be stored.`
}

// HelpInfo returns structured help information for the model-new command.
func (c *NewCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\model-new[catalog_id=<ID>] model_name OR \\model-new[provider=provider_name, base_model=model_name, ...] model_name",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "catalog_id",
				Description: "Short model ID from catalog (e.g., CS4, O3, CO37) - auto-populates provider and base_model",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "provider",
				Description: "LLM provider name (e.g., openai, anthropic, local) - required if catalog_id not used",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "base_model",
				Description: "Provider's model identifier (e.g., gpt-4, claude-3-sonnet) - required if catalog_id not used",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "temperature",
				Description: "Controls randomness in model output (0.0-1.0)",
				Required:    false,
				Type:        "float",
			},
			{
				Name:        "max_tokens",
				Description: "Maximum number of tokens to generate",
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
				Name:        "top_k",
				Description: "Top-k sampling parameter",
				Required:    false,
				Type:        "int",
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
				Command:     "\\model-new[catalog_id=CS4] my-claude",
				Description: "Create model from catalog (Claude Sonnet 4)",
			},
			{
				Command:     "\\model-new[catalog_id=O3, temperature=0.8] creative-model",
				Description: "Create from catalog with custom parameters",
			},
			{
				Command:     "\\model-new[provider=openai, base_model=gpt-4] my-gpt4",
				Description: "Create OpenAI GPT-4 model configuration (manual)",
			},
			{
				Command:     "\\model-new[provider=anthropic, base_model=claude-3-sonnet] claude-work",
				Description: "Create Anthropic Claude model configuration (manual)",
			},
			{
				Command:     "\\model-new[catalog_id=CO4, max_tokens=4000] analysis-opus",
				Description: "Create Claude Opus 4 with custom max tokens",
			},
			{
				Command:     "\\model-new[provider=local, base_model=llama-2, max_tokens=2048] local-llama",
				Description: "Create local model with custom token limit",
			},
		},
		Notes: []string{
			"Model name is required and taken from the input parameter",
			"Model names must be unique and cannot contain spaces",
			"Either catalog_id OR both provider and base_model are required",
			"Use \\model-catalog to see available catalog IDs",
			"When using catalog_id, provider and base_model are auto-populated from catalog",
			"Variables in model name and parameters are interpolated",
			"Additional provider-specific parameters can be included",
			"Created model ID and metadata are stored in system variables",
		},
	}
}

// Execute creates a new model configuration with the specified parameters.
// The input parameter is used as the model name (required).
// Required options: provider, base_model
// Optional parameters: temperature, max_tokens, top_p, top_k, presence_penalty, frequency_penalty, description
func (c *NewCommand) Execute(args map[string]string, input string) error {
	// Get model service
	modelService, err := services.GetGlobalModelService()
	if err != nil {
		return fmt.Errorf("model service not available: %w", err)
	}

	// Get variable service for storing result variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Note: Variable interpolation is now handled by the state machine before commands execute

	// Get model catalog service for catalog_id support
	modelCatalogService, err := services.GetGlobalModelCatalogService()
	if err != nil {
		return fmt.Errorf("model catalog service not available: %w", err)
	}

	// Parse required arguments
	modelName := input
	provider := args["provider"]
	baseModel := args["base_model"]
	catalogID := args["catalog_id"]

	// Handle catalog_id parameter - auto-populate provider and base_model from catalog
	if catalogID != "" {
		catalogModel, err := modelCatalogService.GetModelByID(catalogID)
		if err != nil {
			return fmt.Errorf("failed to find model with catalog_id '%s': %w", catalogID, err)
		}

		// Auto-populate provider and base_model from catalog (catalog_id overrides manual values)
		provider = catalogModel.Provider
		baseModel = catalogModel.Name
	}

	// Validate required parameters
	if modelName == "" {
		return fmt.Errorf("model name is required\\n\\nUsage: %s", c.Usage())
	}

	// Either catalog_id OR both provider and base_model are required
	if catalogID == "" {
		if provider == "" {
			return fmt.Errorf("provider is required (or use catalog_id)\\n\\nUsage: %s", c.Usage())
		}
		if baseModel == "" {
			return fmt.Errorf("base_model is required (or use catalog_id)\\n\\nUsage: %s", c.Usage())
		}
	} else if provider == "" || baseModel == "" {
		// If catalog_id is provided, ensure provider and base_model are populated
		return fmt.Errorf("failed to auto-populate provider/base_model from catalog_id '%s'", catalogID)
	}

	// Note: Variable interpolation for model name, provider, and base_model is handled by state machine

	// Parse optional parameters
	parameters := make(map[string]any)
	description := ""

	// Handle description separately
	if desc, exists := args["description"]; exists {
		description = desc
		// Note: Variable interpolation for description is handled by state machine
	}

	// Parse numeric parameters
	if err := c.parseParameters(args, parameters); err != nil {
		return fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Validate parameters
	if err := modelService.ValidateModelParameters(parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Create model configuration
	model, err := modelService.CreateModelWithGlobalContext(modelName, provider, baseModel, parameters, description, catalogID)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	// Update model-related variables
	if err := c.updateModelVariables(model, variableService); err != nil {
		return fmt.Errorf("failed to update model variables: %w", err)
	}

	// Prepare output message
	outputMsg := fmt.Sprintf("Created model '%s' (ID: %s, Provider: %s, Base: %s)",
		model.Name, model.ID[:8], model.Provider, model.BaseModel)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// parseParameters parses numeric and other typed parameters from string arguments.
func (c *NewCommand) parseParameters(args map[string]string, parameters map[string]any) error {
	// Parse temperature
	if temp, exists := args["temperature"]; exists {
		tempFloat, err := strconv.ParseFloat(temp, 64)
		if err != nil {
			return fmt.Errorf("invalid temperature value: %s", temp)
		}
		parameters["temperature"] = tempFloat
	}

	// Parse max_tokens
	if maxTokens, exists := args["max_tokens"]; exists {
		maxTokensInt, err := strconv.Atoi(maxTokens)
		if err != nil {
			return fmt.Errorf("invalid max_tokens value: %s", maxTokens)
		}
		parameters["max_tokens"] = maxTokensInt
	}

	// Parse top_p
	if topP, exists := args["top_p"]; exists {
		topPFloat, err := strconv.ParseFloat(topP, 64)
		if err != nil {
			return fmt.Errorf("invalid top_p value: %s", topP)
		}
		parameters["top_p"] = topPFloat
	}

	// Parse top_k
	if topK, exists := args["top_k"]; exists {
		topKInt, err := strconv.Atoi(topK)
		if err != nil {
			return fmt.Errorf("invalid top_k value: %s", topK)
		}
		parameters["top_k"] = topKInt
	}

	// Parse presence_penalty
	if presencePenalty, exists := args["presence_penalty"]; exists {
		presencePenaltyFloat, err := strconv.ParseFloat(presencePenalty, 64)
		if err != nil {
			return fmt.Errorf("invalid presence_penalty value: %s", presencePenalty)
		}
		parameters["presence_penalty"] = presencePenaltyFloat
	}

	// Parse frequency_penalty
	if frequencyPenalty, exists := args["frequency_penalty"]; exists {
		frequencyPenaltyFloat, err := strconv.ParseFloat(frequencyPenalty, 64)
		if err != nil {
			return fmt.Errorf("invalid frequency_penalty value: %s", frequencyPenalty)
		}
		parameters["frequency_penalty"] = frequencyPenaltyFloat
	}

	// Add any other string parameters that aren't specially handled
	excludedParams := map[string]bool{
		"provider": true, "base_model": true, "description": true, "catalog_id": true,
		"temperature": true, "max_tokens": true, "top_p": true, "top_k": true,
		"presence_penalty": true, "frequency_penalty": true,
	}

	for key, value := range args {
		if !excludedParams[key] {
			parameters[key] = value
		}
	}

	return nil
}

// updateModelVariables sets model-related system variables.
func (c *NewCommand) updateModelVariables(model *neurotypes.ModelConfig, variableService *services.VariableService) error {
	// Set model variables
	variables := map[string]string{
		"#model_id":       model.ID,
		"#model_name":     model.Name,
		"#model_provider": model.Provider,
		"#model_base":     model.BaseModel,
		"#model_created":  model.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// Add parameter count
	variables["#model_param_count"] = fmt.Sprintf("%d", len(model.Parameters))

	// Add description if present
	if model.Description != "" {
		variables["#model_description"] = model.Description
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&NewCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-new command: %v", err))
	}
}
