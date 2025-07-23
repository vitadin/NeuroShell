// Package model provides model management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with LLM model configurations.
package model

import (
	"fmt"
	"sort"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// StatusCommand implements the \model-status command for displaying model configurations.
// It provides listing, filtering, and detailed information about created models.
type StatusCommand struct{}

// Name returns the command name "model-status" for registration and lookup.
func (c *StatusCommand) Name() string {
	return "model-status"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *StatusCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the model-status command does.
func (c *StatusCommand) Description() string {
	return "Display status and details of model configurations"
}

// Usage returns the syntax and usage examples for the model-status command.
func (c *StatusCommand) Usage() string {
	return `\model-status[name=model_name, provider=provider_name, sort=name|created|provider]

Examples:
  \model-status                                    %% Show all model configurations
  \model-status[name=my-gpt4]                     %% Show specific model by name
  \model-status[provider=openai]                  %% Show models for specific provider
  \model-status[sort=name]                        %% Sort models by name
  \model-status[sort=created]                     %% Sort models by creation date
  \model-status[sort=provider]                    %% Sort models by provider
  \model-status[provider=anthropic, sort=name]    %% Filter by provider and sort

Options:
  name - Filter by specific model name
  provider - Filter by LLM provider (e.g., openai, anthropic, local)
  sort - Sort models by field (name, created, provider, default: name)

Note: Displays comprehensive information including ID, provider, base model,
      parameters, creation date, and description for each model configuration.
      Results are stored in system variables for further processing.`
}

// HelpInfo returns structured help information for the model-status command.
func (c *StatusCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\model-status[name=model_name, provider=provider_name, sort=name|created|provider]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "name",
				Description: "Filter by specific model name",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "provider",
				Description: "Filter by LLM provider (e.g., openai, anthropic, local)",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "sort",
				Description: "Sort models by field (name, created, provider)",
				Required:    false,
				Type:        "string",
				Default:     "name",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\model-status",
				Description: "Show all model configurations",
			},
			{
				Command:     "\\model-status[name=my-gpt4]",
				Description: "Show specific model by name",
			},
			{
				Command:     "\\model-status[provider=openai]",
				Description: "Show models for specific provider",
			},
			{
				Command:     "\\model-status[sort=created]",
				Description: "Sort models by creation date",
			},
			{
				Command:     "\\model-status[provider=anthropic, sort=name]",
				Description: "Filter by provider and sort by name",
			},
		},
		Notes: []string{
			"Shows comprehensive model information including parameters",
			"Variables in filter criteria are interpolated",
			"Results are stored in system variables for further processing",
			"Empty result set returns appropriate message",
			"Model count and filtered count stored in metadata variables",
		},
	}
}

// Execute displays model configurations based on the provided filters and sorting.
// Optional filters: name, provider
// Optional sorting: name (default), created, provider
func (c *StatusCommand) Execute(args map[string]string, _ string) error {
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

	// Get all models
	models, err := modelService.ListModelsWithGlobalContext()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	// Parse filter criteria (variable interpolation handled by state machine)
	nameFilter := args["name"]
	providerFilter := args["provider"]
	sortBy := args["sort"]

	if sortBy == "" {
		sortBy = "name" // Default sort
	}

	// Validate sort option
	if err := c.validateSortOption(sortBy); err != nil {
		return err
	}

	// Filter models
	filteredModels := c.filterModels(models, nameFilter, providerFilter)

	// Sort models
	sortedModels := c.sortModels(filteredModels, sortBy)

	// Format output
	output := c.formatModelStatus(sortedModels, nameFilter, providerFilter)

	// Update status variables
	if err := c.updateStatusVariables(models, filteredModels, variableService); err != nil {
		return fmt.Errorf("failed to update status variables: %w", err)
	}

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", output); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print output
	fmt.Print(output)

	return nil
}

// validateSortOption validates the sort parameter.
func (c *StatusCommand) validateSortOption(sortBy string) error {
	validSorts := map[string]bool{
		"name":     true,
		"created":  true,
		"provider": true,
	}

	if !validSorts[sortBy] {
		return fmt.Errorf("invalid sort option '%s'. Valid options: name, created, provider", sortBy)
	}

	return nil
}

// filterModels applies name and provider filters to the model list.
func (c *StatusCommand) filterModels(models map[string]*neurotypes.ModelConfig, nameFilter, providerFilter string) []*neurotypes.ModelConfig {
	var filtered []*neurotypes.ModelConfig

	for _, model := range models {
		// Apply name filter
		if nameFilter != "" && !strings.Contains(strings.ToLower(model.Name), strings.ToLower(nameFilter)) {
			continue
		}

		// Apply provider filter
		if providerFilter != "" && !strings.EqualFold(model.Provider, providerFilter) {
			continue
		}

		filtered = append(filtered, model)
	}

	return filtered
}

// sortModels sorts the model list by the specified field.
func (c *StatusCommand) sortModels(models []*neurotypes.ModelConfig, sortBy string) []*neurotypes.ModelConfig {
	sort.Slice(models, func(i, j int) bool {
		switch sortBy {
		case "name":
			return strings.ToLower(models[i].Name) < strings.ToLower(models[j].Name)
		case "created":
			return models[i].CreatedAt.Before(models[j].CreatedAt)
		case "provider":
			if models[i].Provider == models[j].Provider {
				return strings.ToLower(models[i].Name) < strings.ToLower(models[j].Name)
			}
			return strings.ToLower(models[i].Provider) < strings.ToLower(models[j].Provider)
		default:
			return strings.ToLower(models[i].Name) < strings.ToLower(models[j].Name)
		}
	})

	return models
}

// formatModelStatus formats the model list for display.
func (c *StatusCommand) formatModelStatus(models []*neurotypes.ModelConfig, nameFilter, providerFilter string) string {
	// Get active model ID from model service
	modelService, err := services.GetGlobalModelService()
	var activeModelID string
	if err == nil {
		if activeModel, err := modelService.GetActiveModelConfigWithGlobalContext(); err == nil {
			activeModelID = activeModel.ID
		}
	}
	if len(models) == 0 {
		if nameFilter != "" || providerFilter != "" {
			var filters []string
			if nameFilter != "" {
				filters = append(filters, fmt.Sprintf("name='%s'", nameFilter))
			}
			if providerFilter != "" {
				filters = append(filters, fmt.Sprintf("provider='%s'", providerFilter))
			}
			return fmt.Sprintf("No model configurations found matching filter: %s\n", strings.Join(filters, ", "))
		}
		return "No model configurations found. Use \\model-new to create model configurations.\n"
	}

	var output strings.Builder

	// Header
	if nameFilter != "" || providerFilter != "" {
		var filters []string
		if nameFilter != "" {
			filters = append(filters, fmt.Sprintf("name='%s'", nameFilter))
		}
		if providerFilter != "" {
			filters = append(filters, fmt.Sprintf("provider='%s'", providerFilter))
		}
		output.WriteString(fmt.Sprintf("Model Configurations - Filter: %s (%d models):\n", strings.Join(filters, ", "), len(models)))
	} else {
		output.WriteString(fmt.Sprintf("Model Configurations (%d models):\n", len(models)))
	}

	// Model details
	for _, model := range models {
		output.WriteString(c.formatModelDetails(model, activeModelID))
		output.WriteString("\n")
	}

	return output.String()
}

// formatModelDetails formats detailed information for a single model.
func (c *StatusCommand) formatModelDetails(model *neurotypes.ModelConfig, activeModelID string) string {
	var details strings.Builder

	// Basic info
	details.WriteString(fmt.Sprintf("  %s (%s)\n", model.Name, model.ID[:8]))
	details.WriteString(fmt.Sprintf("    Provider: %s\n", model.Provider))
	details.WriteString(fmt.Sprintf("    Base Model: %s\n", model.BaseModel))

	// Parameters
	if len(model.Parameters) > 0 {
		details.WriteString(fmt.Sprintf("    Parameters: %d (", len(model.Parameters)))
		var params []string
		for key, value := range model.Parameters {
			params = append(params, fmt.Sprintf("%s=%v", key, value))
		}
		sort.Strings(params) // Sort for consistent output
		details.WriteString(strings.Join(params, ", "))
		details.WriteString(")\n")
	} else {
		details.WriteString("    Parameters: none\n")
	}

	// Timestamps
	details.WriteString(fmt.Sprintf("    Created: %s\n", model.CreatedAt.Format("2006-01-02 15:04:05")))
	if !model.UpdatedAt.Equal(model.CreatedAt) {
		details.WriteString(fmt.Sprintf("    Updated: %s\n", model.UpdatedAt.Format("2006-01-02 15:04:05")))
	}

	// Description
	if model.Description != "" {
		details.WriteString(fmt.Sprintf("    Description: %s\n", model.Description))
	}

	// Active flag - check if this model is the currently active one
	if model.ID == activeModelID {
		details.WriteString("    Active: true\n")
	}

	return details.String()
}

// updateStatusVariables sets model status-related system variables.
func (c *StatusCommand) updateStatusVariables(allModels map[string]*neurotypes.ModelConfig, filteredModels []*neurotypes.ModelConfig, variableService *services.VariableService) error {
	// Set status variables
	variables := map[string]string{
		"#model_count":          fmt.Sprintf("%d", len(allModels)),
		"#model_filtered_count": fmt.Sprintf("%d", len(filteredModels)),
	}

	// If showing single model, set additional variables
	if len(filteredModels) == 1 {
		model := filteredModels[0]
		variables["#model_current_id"] = model.ID
		variables["#model_current_name"] = model.Name
		variables["#model_current_provider"] = model.Provider
		variables["#model_current_base"] = model.BaseModel
		variables["#model_current_created"] = model.CreatedAt.Format("2006-01-02 15:04:05")
		variables["#model_current_param_count"] = fmt.Sprintf("%d", len(model.Parameters))

		if model.Description != "" {
			variables["#model_current_description"] = model.Description
		}
	}

	for name, value := range variables {
		if err := variableService.SetSystemVariable(name, value); err != nil {
			return fmt.Errorf("failed to set variable %s: %w", name, err)
		}
	}

	return nil
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&StatusCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register model-status command: %v", err))
	}
}
