// Package model provides model management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with LLM model configurations.
package model

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ActivateCommand implements the \model-activate command for setting the active model.
// It provides model activation functionality with smart matching for model names and IDs.
type ActivateCommand struct{}

// Name returns the command name "model-activate" for registration and lookup.
func (c *ActivateCommand) Name() string {
	return "model-activate"
}

// ParseMode returns ParseModeKeyValue for argument parsing with options.
func (c *ActivateCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the model-activate command does.
func (c *ActivateCommand) Description() string {
	return "Set active model by name or ID with smart matching"
}

// Usage returns the syntax and usage examples for the model-activate command.
func (c *ActivateCommand) Usage() string {
	return `\model-activate[id=false] model_text
\model-activate[id=true] id_prefix
\model-activate

Examples:
  \model-activate my-gpt                    %% Activate by name (default) - matches any model name containing "my-gpt"
  \model-activate[id=true] 1234            %% Activate by ID prefix - matches any model ID starting with "1234"
  \model-activate my-claude-model          %% Activate by exact or partial name match
  \model-activate[id=true] abc123          %% Activate by ID prefix match
  \model-activate                          %% Show current active model, or activate latest if none active

Options:
  id - Search by model ID prefix instead of name (default: false)

Notes:
  - When called with no parameters: shows current active model or activates latest model
  - By default, searches model names for matches (partial matching supported)
  - With id=true, searches model ID prefixes
  - If multiple models match, shows list of matches and asks for more specific input
  - If no models match, shows helpful suggestions
  - Sets the matched model as the currently active model for all operations
  - Updates system variables with active model information`
}

// HelpInfo returns structured help information for the model-activate command.
func (c *ActivateCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\model-activate[id=false] [model_text]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "id",
				Description: "Search by model ID prefix instead of name",
				Required:    false,
				Type:        "boolean",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\model-activate my-gpt",
				Description: "Activate by name match (default behavior)",
			},
			{
				Command:     "\\model-activate[id=true] 1234",
				Description: "Activate by ID prefix match",
			},
			{
				Command:     "\\model-activate claude",
				Description: "Activate by partial name match",
			},
			{
				Command:     "\\model-activate",
				Description: "Show current active model or activate latest model",
			},
		},
		Notes: []string{
			"When called with no parameters: shows current active model or activates latest model",
			"By default searches model names (partial matching supported)",
			"Use id=true to search by model ID prefix instead",
			"If multiple models match, shows list and asks for more specific input",
			"If no models match, shows helpful suggestions",
			"Sets the matched model as currently active for all operations",
			"Variables in model text are interpolated before processing",
		},
	}
}

// Execute activates a model configuration using smart matching.
func (c *ActivateCommand) Execute(args map[string]string, input string) error {
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

	// Parse arguments
	idStr := args["id"]
	byID := idStr == "true"

	// Handle no-parameter case: show current active model or activate latest
	searchText := input
	if searchText == "" {
		return c.handleNoParameters(modelService, variableService)
	}

	// Get all models for searching
	models, err := modelService.ListModelsWithGlobalContext()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models found. Use \\model-new to create model configurations")
	}

	// Find matching models
	var matches []*neurotypes.ModelConfig
	if byID {
		matches = c.findModelsByIDPrefix(models, searchText)
	} else {
		matches = c.findModelsByName(models, searchText)
	}

	// Handle different match scenarios
	switch len(matches) {
	case 0:
		// No matches - provide helpful suggestions
		return c.handleNoMatches(models, searchText, byID)
	case 1:
		// Unique match - proceed with activation
		return c.activateModel(matches[0], modelService, variableService)
	default:
		// Multiple matches - ask for more specific input
		return c.handleMultipleMatches(matches, searchText, byID)
	}
}

// findModelsByName finds models whose names contain the search text (case-insensitive).
func (c *ActivateCommand) findModelsByName(models map[string]*neurotypes.ModelConfig, searchText string) []*neurotypes.ModelConfig {
	var matches []*neurotypes.ModelConfig
	searchLower := strings.ToLower(searchText)

	for _, model := range models {
		if strings.Contains(strings.ToLower(model.Name), searchLower) {
			matches = append(matches, model)
		}
	}

	return matches
}

// findModelsByIDPrefix finds models whose IDs start with the search text (case-insensitive).
func (c *ActivateCommand) findModelsByIDPrefix(models map[string]*neurotypes.ModelConfig, searchText string) []*neurotypes.ModelConfig {
	var matches []*neurotypes.ModelConfig
	searchLower := strings.ToLower(searchText)

	for _, model := range models {
		if strings.HasPrefix(strings.ToLower(model.ID), searchLower) {
			matches = append(matches, model)
		}
	}

	return matches
}

// handleNoMatches provides helpful error message when no models match.
func (c *ActivateCommand) handleNoMatches(models map[string]*neurotypes.ModelConfig, searchText string, byID bool) error {
	searchType := "name"
	if byID {
		searchType = "ID prefix"
	}

	errorMsg := fmt.Sprintf("No models found matching %s '%s'.\n\nAvailable models:", searchType, searchText)

	// Show available models
	for _, model := range models {
		if byID {
			errorMsg += fmt.Sprintf("\n  ID: %s (name: %s)", model.ID[:8], model.Name)
		} else {
			errorMsg += fmt.Sprintf("\n  %s (ID: %s)", model.Name, model.ID[:8])
		}
	}

	return fmt.Errorf("%s", errorMsg)
}

// handleMultipleMatches provides helpful error message when multiple models match.
func (c *ActivateCommand) handleMultipleMatches(matches []*neurotypes.ModelConfig, searchText string, byID bool) error {
	searchType := "name"
	if byID {
		searchType = "ID prefix"
	}

	errorMsg := fmt.Sprintf("Multiple models match %s '%s'. Please be more specific:\n", searchType, searchText)

	for _, model := range matches {
		if byID {
			errorMsg += fmt.Sprintf("  ID: %s (name: %s, provider: %s)\n", model.ID[:8], model.Name, model.Provider)
		} else {
			errorMsg += fmt.Sprintf("  %s (ID: %s, provider: %s)\n", model.Name, model.ID[:8], model.Provider)
		}
	}

	errorMsg += "\nTip: Use the full name or a longer ID prefix to uniquely identify the model."

	return fmt.Errorf("%s", errorMsg)
}

// activateModel performs the actual model activation.
func (c *ActivateCommand) activateModel(model *neurotypes.ModelConfig, modelService *services.ModelService, variableService *services.VariableService) error {
	// Set the model as active
	err := modelService.SetActiveModelWithGlobalContext(model.ID)
	if err != nil {
		return fmt.Errorf("failed to activate model: %w", err)
	}

	// Auto-push client creation command to stack service for seamless UX
	if stackService, err := services.GetGlobalStackService(); err == nil {
		clientCommand := fmt.Sprintf("\\llm-client-get[provider=%s]", model.Provider)
		stackService.PushCommand(clientCommand)
	}

	// Prepare success message
	outputMsg := fmt.Sprintf("Activated model '%s' (ID: %s, Provider: %s, Base: %s)",
		model.Name, model.ID[:8], model.Provider, model.BaseModel)

	// Update activation-related variables
	if err := c.updateActivationVariables(model, variableService); err != nil {
		return fmt.Errorf("failed to update activation variables: %w", err)
	}

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	fmt.Println(outputMsg)

	return nil
}

// updateActivationVariables sets model activation-related system variables.
func (c *ActivateCommand) updateActivationVariables(model *neurotypes.ModelConfig, variableService *services.VariableService) error {
	// Set activation result variables
	variables := map[string]string{
		"#active_model_id":       model.ID,
		"#active_model_name":     model.Name,
		"#active_model_provider": model.Provider,
		"#active_model_base":     model.BaseModel,
	}

	// Add parameter count
	variables["#active_model_param_count"] = fmt.Sprintf("%d", len(model.Parameters))

	// Add description if present
	if model.Description != "" {
		variables["#active_model_description"] = model.Description
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

// handleNoParameters handles the case when model-activate is called without parameters.
// It shows current active model if exists, otherwise activates the latest created model.
func (c *ActivateCommand) handleNoParameters(modelService *services.ModelService, variableService *services.VariableService) error {
	// Check if there's already an active model
	activeModelName, err := variableService.Get("#active_model_name")
	if err == nil && activeModelName != "" {
		// There's already an active model, show current status
		activeModelID, _ := variableService.Get("#active_model_id")
		activeModelProvider, _ := variableService.Get("#active_model_provider")
		activeModelBase, _ := variableService.Get("#active_model_base")

		outputMsg := fmt.Sprintf("Current active model: '%s' (ID: %s, Provider: %s, Base: %s)",
			activeModelName, activeModelID[:8], activeModelProvider, activeModelBase)

		// Store result in _output variable
		if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
			return fmt.Errorf("failed to store result: %w", err)
		}

		fmt.Println(outputMsg)
		return nil
	}

	// No active model exists, try to activate the latest created model
	models, err := modelService.ListModelsWithGlobalContext()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		// No models exist, show friendly message
		outputMsg := "No models available. Use \\model-new to create one."

		if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
			return fmt.Errorf("failed to store result: %w", err)
		}

		fmt.Println(outputMsg)
		return nil
	}

	// Find the latest created model
	latestModel := c.findLatestModel(models)
	if latestModel == nil {
		return fmt.Errorf("failed to find latest model")
	}

	// Activate the latest model
	return c.activateModel(latestModel, modelService, variableService)
}

// findLatestModel finds the most recently created model from the models map.
func (c *ActivateCommand) findLatestModel(models map[string]*neurotypes.ModelConfig) *neurotypes.ModelConfig {
	var latestModel *neurotypes.ModelConfig

	for _, model := range models {
		if latestModel == nil || model.CreatedAt.After(latestModel.CreatedAt) {
			latestModel = model
		}
	}

	return latestModel
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&ActivateCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-activate command: %v", err))
	}
}
