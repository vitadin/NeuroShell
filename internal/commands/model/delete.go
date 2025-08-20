// Package model provides model management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with LLM model configurations.
package model

import (
	"fmt"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// DeleteCommand implements the \model-delete command for removing model configurations.
// It provides model deletion functionality with smart matching for model names and IDs.
type DeleteCommand struct{}

// Name returns the command name "model-delete" for registration and lookup.
func (c *DeleteCommand) Name() string {
	return "model-delete"
}

// ParseMode returns ParseModeKeyValue for argument parsing with options.
func (c *DeleteCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the model-delete command does.
func (c *DeleteCommand) Description() string {
	return "Delete model configuration by name or ID with smart matching"
}

// Usage returns the syntax and usage examples for the model-delete command.
func (c *DeleteCommand) Usage() string {
	return `\model-delete[id=false] model_text
\model-delete[id=true] id_prefix

Examples:
  \model-delete my-gpt                      %% Delete by name (default) - matches any model name containing "my-gpt"
  \model-delete[id=true] 1234              %% Delete by ID prefix - matches any model ID starting with "1234"
  \model-delete my-claude-model            %% Delete by exact or partial name match
  \model-delete[id=true] abc123            %% Delete by ID prefix match

Options:
  id - Search by model ID prefix instead of name (default: false)

Notes:
  - By default, searches model names for matches (partial matching supported)
  - With id=true, searches model ID prefixes
  - If multiple models match, shows list of matches and asks for more specific input
  - If no models match, shows helpful suggestions
  - After deletion, automatically re-evaluates active model (shows current or activates latest)
  - Deletion is permanent and cannot be undone`
}

// HelpInfo returns structured help information for the model-delete command.
func (c *DeleteCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\model-delete[id=false] model_text",
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
				Command:     "\\model-delete my-gpt",
				Description: "Delete by name match (default behavior)",
			},
			{
				Command:     "\\model-delete[id=true] 1234",
				Description: "Delete by ID prefix match",
			},
			{
				Command:     "\\model-delete claude",
				Description: "Delete by partial name match",
			},
		},
		Notes: []string{
			"By default searches model names (partial matching supported)",
			"Use id=true to search by model ID prefix instead",
			"If multiple models match, shows list and asks for more specific input",
			"If no models match, shows helpful suggestions",
			"After deletion, automatically re-evaluates active model state",
			"Variables in model text are interpolated before processing",
		},
	}
}

// Execute deletes a model configuration using smart matching.
func (c *DeleteCommand) Execute(args map[string]string, input string) error {
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

	// Validate input
	searchText := input
	if searchText == "" {
		return fmt.Errorf("model name or ID prefix is required\n\nUsage: %s", c.Usage())
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
		// Unique match - proceed with deletion
		return c.deleteModel(matches[0], modelService, variableService)
	default:
		// Multiple matches - ask for more specific input
		return c.handleMultipleMatches(matches, searchText, byID)
	}
}

// findModelsByName finds models whose names contain the search text (case-insensitive).
func (c *DeleteCommand) findModelsByName(models map[string]*neurotypes.ModelConfig, searchText string) []*neurotypes.ModelConfig {
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
func (c *DeleteCommand) findModelsByIDPrefix(models map[string]*neurotypes.ModelConfig, searchText string) []*neurotypes.ModelConfig {
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
func (c *DeleteCommand) handleNoMatches(models map[string]*neurotypes.ModelConfig, searchText string, byID bool) error {
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
func (c *DeleteCommand) handleMultipleMatches(matches []*neurotypes.ModelConfig, searchText string, byID bool) error {
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

// deleteModel performs the actual model deletion.
func (c *DeleteCommand) deleteModel(model *neurotypes.ModelConfig, modelService *services.ModelService, variableService *services.VariableService) error {
	// Store model info for result message before deletion
	modelName := model.Name
	modelID := model.ID
	modelProvider := model.Provider
	modelBase := model.BaseModel

	// Delete the model by ID (more reliable than by name)
	err := modelService.DeleteModelWithGlobalContext(modelID)
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	// Prepare success message
	outputMsg := fmt.Sprintf("Deleted model '%s' (ID: %s, Provider: %s, Base: %s)",
		modelName, modelID[:8], modelProvider, modelBase)

	// Auto-push model activation command to stack service to handle active model state
	// This will either show current active model, activate latest model, or show "no models" message
	if stackService, err := services.GetGlobalStackService(); err == nil {
		activateCommand := "\\silent \\model-activate"
		stackService.PushCommand(activateCommand)
	}

	// Update deletion-related variables
	if err := c.updateDeletionVariables(modelName, modelID, modelProvider, modelBase, variableService); err != nil {
		return fmt.Errorf("failed to update deletion variables: %w", err)
	}

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", outputMsg); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print confirmation
	printer := printing.NewDefaultPrinter()
	printer.Success(outputMsg)

	return nil
}

// updateDeletionVariables sets model deletion-related system variables.
func (c *DeleteCommand) updateDeletionVariables(modelName, modelID, provider, baseModel string, variableService *services.VariableService) error {
	// Set deletion result variables
	variables := map[string]string{
		"#deleted_model_id":       modelID,
		"#deleted_model_name":     modelName,
		"#deleted_model_provider": provider,
		"#deleted_model_base":     baseModel,
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

// IsReadOnly returns false as the model command modifies system state.
func (c *DeleteCommand) IsReadOnly() bool {
	return false
}
func init() {
	if err := commands.GetGlobalRegistry().Register(&DeleteCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-delete command: %v", err))
	}
}
