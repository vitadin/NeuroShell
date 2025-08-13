// Package provider provides provider management commands for NeuroShell.
// It includes commands for managing and interacting with LLM provider configurations.
package provider

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// CatalogCommand implements the \provider-catalog command for listing available LLM providers.
// It provides access to the embedded provider catalog with filtering and search capabilities.
type CatalogCommand struct{}

// Name returns the command name "provider-catalog" for registration and lookup.
func (c *CatalogCommand) Name() string {
	return "provider-catalog"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *CatalogCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the provider-catalog command does.
func (c *CatalogCommand) Description() string {
	return "List available LLM providers from embedded catalog"
}

// Usage returns the syntax and usage examples for the provider-catalog command.
func (c *CatalogCommand) Usage() string {
	return `\provider-catalog[provider=openai|anthropic|gemini|all, sort=name|provider, search=query]

Examples:
  \provider-catalog                              %% List all available providers (default: sorted by provider)
  \provider-catalog[provider=openai]             %% List OpenAI providers only
  \provider-catalog[provider=anthropic]          %% List Anthropic providers only
  \provider-catalog[provider=gemini]             %% List Google Gemini providers only
  \provider-catalog[sort=name]                   %% Sort providers alphabetically by name
  \provider-catalog[search=chat]                 %% Search for providers containing "chat"
  \provider-catalog[search=GMC]                  %% Search by provider ID (case-insensitive)
  \provider-catalog[search=API]                  %% Search for providers with "API" in description
  \provider-catalog[provider=openai,sort=name]   %% OpenAI providers sorted by name
  \provider-catalog[search=completions,sort=name] %% Search for completion providers, sorted by name

Options:
  provider - Filter by provider: openai, anthropic, gemini, all (default: all)
  sort     - Sort order: name (alphabetical), provider (by provider then name)
  search   - Search query to filter providers by ID, name, display name, or description

Note: Options can be combined. Default sort is by provider.
      Provider catalog is stored in ${_output} variable.
      Shows provider ID, display name, provider type, and configuration details.
      Provider IDs are shown in format: [ID] Display Name (provider_type)`
}

// HelpInfo returns structured help information for the provider-catalog command.
func (c *CatalogCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\provider-catalog[provider=openai|anthropic|gemini|all, sort=name|provider, search=query]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "provider",
				Description: "Filter by provider: openai, anthropic, gemini, all",
				Required:    false,
				Type:        "string",
				Default:     "all",
			},
			{
				Name:        "sort",
				Description: "Sort order: name (alphabetical), provider (by provider then name)",
				Required:    false,
				Type:        "string",
				Default:     "provider",
			},
			{
				Name:        "search",
				Description: "Search query to filter providers by ID, name, display name, or description",
				Required:    false,
				Type:        "string",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\provider-catalog",
				Description: "List all available providers sorted by provider",
			},
			{
				Command:     "\\provider-catalog[provider=openai]",
				Description: "List OpenAI providers only",
			},
			{
				Command:     "\\provider-catalog[search=chat]",
				Description: "Search for providers containing 'chat'",
			},
			{
				Command:     "\\provider-catalog[search=OAC]",
				Description: "Search by provider ID (case-insensitive)",
			},
			{
				Command:     "\\provider-catalog[provider=anthropic,sort=name]",
				Description: "List Anthropic providers sorted alphabetically",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Formatted catalog listing of available providers",
				Type:        "command_output",
				Example:     "[OAC] OpenAI Chat API (openai)\\n  Provider: openai\\n  Base URL: https://api.openai.com/v1\\n  Endpoint: /chat/completions...",
			},
		},
		Notes: []string{
			"Options can be combined (e.g., provider=openai,sort=name)",
			"Default sort is by provider, then by name within each provider",
			"Shows provider ID, display name, provider type, configuration details",
			"Provider IDs are displayed in format: [ID] Display Name (provider_type)",
			"Embedded catalog includes popular providers: OpenAI, Anthropic, Gemini",
			"Search is case-insensitive and matches ID, name, display name, or description",
			"Provider IDs can be used with client factory for provider creation",
		},
	}
}

// Execute lists available LLM providers with optional filtering, sorting, and searching.
// Options:
//   - provider: openai|anthropic|gemini|all (default: all)
//   - sort: name|provider (default: provider)
//   - search: query string for filtering (optional)
func (c *CatalogCommand) Execute(args map[string]string, _ string) error {
	// Get provider catalog service
	providerCatalogService, err := services.GetGlobalProviderCatalogService()
	if err != nil {
		return fmt.Errorf("provider catalog service not available: %w", err)
	}

	// Get variable service for storing result variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get theme object for styling
	themeObj := c.getThemeObject()

	// Parse arguments
	provider := args["provider"]
	if provider == "" {
		provider = "all" // default provider filter
	}
	sortBy := args["sort"]
	if sortBy == "" {
		sortBy = "provider" // default sort
	}
	searchQuery := args["search"]

	// Validate arguments
	if err := c.validateArguments(provider, sortBy); err != nil {
		return err
	}

	// Get providers based on provider filter
	var providers []neurotypes.ProviderCatalogEntry
	if provider == "all" {
		providers, err = providerCatalogService.GetProviderCatalog()
	} else {
		providers, err = providerCatalogService.GetProvidersByProvider(provider)
	}
	if err != nil {
		return fmt.Errorf("failed to get provider catalog: %w", err)
	}

	// Apply search filter if provided
	if searchQuery != "" {
		providers, err = c.filterProvidersBySearch(providers, searchQuery)
		if err != nil {
			return fmt.Errorf("failed to search providers: %w", err)
		}
	}

	// Apply sorting
	c.sortProviders(providers, sortBy, provider)

	// Format output with theme styling
	output := c.formatProviderCatalog(providers, provider, sortBy, searchQuery, themeObj)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", output); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print the catalog
	fmt.Print(output)

	return nil
}

// validateArguments checks if the provided provider and sort options are valid.
func (c *CatalogCommand) validateArguments(provider, sortBy string) error {
	validProviders := map[string]bool{
		"all":       true,
		"openai":    true,
		"anthropic": true,
		"gemini":    true,
	}
	if !validProviders[provider] {
		return fmt.Errorf("invalid provider option '%s'. Valid options: all, openai, anthropic, gemini", provider)
	}

	validSorts := map[string]bool{
		"name":     true,
		"provider": true,
	}
	if !validSorts[sortBy] {
		return fmt.Errorf("invalid sort option '%s'. Valid options: name, provider", sortBy)
	}

	return nil
}

// filterProvidersBySearch filters providers based on the search query.
func (c *CatalogCommand) filterProvidersBySearch(providers []neurotypes.ProviderCatalogEntry, query string) ([]neurotypes.ProviderCatalogEntry, error) {
	var matches []neurotypes.ProviderCatalogEntry
	queryLower := strings.ToLower(query)

	for _, provider := range providers {
		// Search in provider ID, provider name, display name, and description
		if strings.Contains(strings.ToLower(provider.ID), queryLower) ||
			strings.Contains(strings.ToLower(provider.Provider), queryLower) ||
			strings.Contains(strings.ToLower(provider.DisplayName), queryLower) ||
			strings.Contains(strings.ToLower(provider.Description), queryLower) {
			matches = append(matches, provider)
		}
	}

	return matches, nil
}

// sortProviders sorts the provider list according to the specified criteria.
func (c *CatalogCommand) sortProviders(providers []neurotypes.ProviderCatalogEntry, sortBy, provider string) {
	switch sortBy {
	case "name":
		sort.Slice(providers, func(i, j int) bool {
			return strings.ToLower(providers[i].DisplayName) < strings.ToLower(providers[j].DisplayName)
		})
	case "provider":
		sort.Slice(providers, func(i, j int) bool {
			// First sort by provider (if showing all providers)
			if provider == "all" {
				if providers[i].Provider != providers[j].Provider {
					return providers[i].Provider < providers[j].Provider
				}
			}
			// Then sort by display name within provider
			return strings.ToLower(providers[i].DisplayName) < strings.ToLower(providers[j].DisplayName)
		})
	}
}

// formatProviderCatalog formats the provider catalog for display with theme styling.
func (c *CatalogCommand) formatProviderCatalog(providers []neurotypes.ProviderCatalogEntry, provider, sortBy, searchQuery string, themeObj *services.Theme) string {
	if len(providers) == 0 {
		searchText := ""
		if searchQuery != "" {
			searchText = fmt.Sprintf(" matching %s", themeObj.Variable.Render("'"+searchQuery+"'"))
		}
		providerText := ""
		if provider != "all" {
			providerText = fmt.Sprintf(" from %s", themeObj.Keyword.Render(provider))
		}
		noProvidersMsg := fmt.Sprintf("No providers found%s%s.", providerText, searchText)
		return themeObj.Warning.Render(noProvidersMsg) + "\n"
	}

	var result strings.Builder

	// Header with professional styling
	headerParts := []string{themeObj.Success.Render("Provider Catalog")}
	if provider != "all" {
		providerPart := fmt.Sprintf("(%s)", themeObj.Keyword.Render(c.toTitle(provider)))
		headerParts = append(headerParts, providerPart)
	}
	if searchQuery != "" {
		searchPart := fmt.Sprintf("- Search: %s", themeObj.Variable.Render("'"+searchQuery+"'"))
		headerParts = append(headerParts, searchPart)
	}
	countPart := fmt.Sprintf("(%s)", themeObj.Info.Render(fmt.Sprintf("%d providers", len(providers))))
	headerParts = append(headerParts, countPart)

	result.WriteString(fmt.Sprintf("%s:\n", strings.Join(headerParts, " ")))

	// Group by provider if showing all providers
	if provider == "all" && sortBy == "provider" {
		currentProvider := ""
		for _, providerEntry := range providers {
			if providerEntry.Provider != currentProvider {
				if currentProvider != "" {
					result.WriteString("\n")
				}
				providerHeader := fmt.Sprintf("%s Providers:", c.toTitle(providerEntry.Provider))
				result.WriteString(themeObj.Success.Render(providerHeader) + "\n")
				currentProvider = providerEntry.Provider
			}
			result.WriteString(c.formatProviderEntry(providerEntry, true, themeObj))
		}
	} else {
		// Simple list format
		for _, providerEntry := range providers {
			result.WriteString(c.formatProviderEntry(providerEntry, provider == "all", themeObj))
		}
	}

	return result.String()
}

// formatProviderEntry formats a single provider entry for display with professional theme styling.
func (c *CatalogCommand) formatProviderEntry(provider neurotypes.ProviderCatalogEntry, showProvider bool, themeObj *services.Theme) string {
	var result strings.Builder

	// Provider header: [ID] Display Name (client_type) with prominent catalog ID
	catalogID := themeObj.Highlight.Render(fmt.Sprintf("[%s]", provider.ID))
	displayName := themeObj.Command.Render(provider.DisplayName)
	clientType := themeObj.Variable.Render(fmt.Sprintf("(%s)", provider.ClientType))
	providerHeader := fmt.Sprintf("  %s %s %s\n", catalogID, displayName, clientType)
	result.WriteString(providerHeader)

	// Provider type (if showing all providers)
	if showProvider {
		providerLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Provider:"),
			themeObj.Keyword.Render(provider.Provider))
		result.WriteString(providerLine)
	}

	// Base URL
	if provider.BaseURL != "" {
		baseURLLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Base URL:"),
			themeObj.Variable.Render(provider.BaseURL))
		result.WriteString(baseURLLine)
	}

	// Endpoint
	if provider.Endpoint != "" {
		endpointLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Endpoint:"),
			themeObj.Variable.Render(provider.Endpoint))
		result.WriteString(endpointLine)
	}

	// Client type
	clientTypeLine := fmt.Sprintf("    %s %s\n",
		themeObj.Info.Render("Client Type:"),
		themeObj.Keyword.Render(provider.ClientType))
	result.WriteString(clientTypeLine)

	// Headers
	if len(provider.Headers) > 0 {
		headersList := make([]string, 0, len(provider.Headers))
		for key, value := range provider.Headers {
			headersList = append(headersList, fmt.Sprintf("%s: %s",
				themeObj.Variable.Render(key),
				themeObj.Variable.Render(value)))
		}
		sort.Strings(headersList) // Sort for consistent output
		headersLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Headers:"),
			strings.Join(headersList, ", "))
		result.WriteString(headersLine)
	}

	// Description
	if len(provider.Description) > 0 {
		descriptionLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Description:"),
			themeObj.Info.Render(provider.Description))
		result.WriteString(descriptionLine)
	}

	// Implementation notes
	if len(provider.ImplementationNotes) > 0 {
		implementationLine := fmt.Sprintf("    %s %s\n",
			themeObj.Info.Render("Implementation:"),
			themeObj.Variable.Render(provider.ImplementationNotes))
		result.WriteString(implementationLine)
	}

	return result.String()
}

// toTitle converts the first character of a string to uppercase.
func (c *CatalogCommand) toTitle(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// getThemeObject retrieves the theme object based on the _style variable
func (c *CatalogCommand) getThemeObject() *services.Theme {
	// Get _style variable for theme selection
	styleValue := ""
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		if value, err := variableService.Get("_style"); err == nil {
			styleValue = value
		}
	}

	// Get theme service and theme object (always returns valid theme)
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		// This should rarely happen, but we need to return something
		panic(fmt.Sprintf("theme service not available: %v", err))
	}

	return themeService.GetThemeByName(styleValue)
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&CatalogCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register provider-catalog command: %v", err))
	}
}
