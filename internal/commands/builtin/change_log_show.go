package builtin

import (
	"fmt"
	"strings"
	"unicode"

	"neuroshell/internal/commands"
	"neuroshell/internal/commands/printing"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ChangeLogShowCommand implements the \change-log-show command for displaying NeuroShell development history.
// It provides access to the embedded change log with search functionality.
type ChangeLogShowCommand struct{}

// Name returns the command name "change-log-show" for registration and lookup.
func (c *ChangeLogShowCommand) Name() string {
	return "change-log-show"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ChangeLogShowCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the change-log-show command does.
func (c *ChangeLogShowCommand) Description() string {
	return "Show NeuroShell development change log with search capabilities"
}

// Usage returns the syntax and usage examples for the change-log-show command.
func (c *ChangeLogShowCommand) Usage() string {
	return `\change-log-show[search=query,order=asc|desc]

Examples:
  \change-log-show                         %% Show all change log entries (newest at bottom)
  \change-log-show[order=desc]             %% Show entries with newest at top
  \change-log-show[search=bug]             %% Search for entries containing "bug"
  \change-log-show[search=streaming]       %% Search for streaming-related changes
  \change-log-show[search=feature]         %% Search for feature-type entries
  \change-log-show[search=CL003]           %% Search by change log ID
  \change-log-show[search=temporal,order=desc]  %% Search with newest first

Options:
  search - Search query to filter entries by ID, version, type, title, description, or impact
  order  - Sort order: 'asc' (oldest first, newest at bottom) or 'desc' (newest first). Default: asc

Note: Search is case-insensitive and matches across all entry fields.
      Change log is stored in ${_output} variable.
      Default order shows newest entries at bottom for better visibility.`
}

// HelpInfo returns structured help information for the change-log-show command.
func (c *ChangeLogShowCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\change-log-show[search=query,order=asc|desc]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "search",
				Description: "Search query to filter entries by ID, version, type, title, description, or impact",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "order",
				Description: "Sort order: 'asc' (oldest first, newest at bottom) or 'desc' (newest first). Default: asc",
				Required:    false,
				Type:        "string",
				Default:     "asc",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\change-log-show",
				Description: "Show all change log entries (newest at bottom)",
			},
			{
				Command:     "\\change-log-show[order=desc]",
				Description: "Show entries with newest at top",
			},
			{
				Command:     "\\change-log-show[search=bug]",
				Description: "Search for entries containing 'bug'",
			},
			{
				Command:     "\\change-log-show[search=streaming]",
				Description: "Search for streaming-related changes",
			},
			{
				Command:     "\\change-log-show[search=CL003]",
				Description: "Search by change log ID",
			},
			{
				Command:     "\\change-log-show[search=feature,order=desc]",
				Description: "Search for feature-type entries with newest first",
			},
		},
		StoredVariables: []neurotypes.HelpStoredVariable{
			{
				Name:        "_output",
				Description: "Formatted change log listing",
				Type:        "command_output",
				Example:     "[CL003] Added streaming mode support via _stream variable\\n  Version: 0.2.0+development\\n  Date: 2025-01-30\\n  Type: feature...",
			},
		},
		Notes: []string{
			"Search is case-insensitive and matches across all entry fields",
			"Default order is ascending (oldest first, newest at bottom) for better visibility",
			"Use order=desc to show newest entries at the top (traditional reverse chronological)",
			"Change log entries include ID, version, date, type, title, description, and impact",
			"Entry types include: bugfix, feature, enhancement, testing, refactor, docs, chore",
			"Files changed information shows which source files were modified",
		},
	}
}

// Execute displays change log entries with optional search filtering and sort order.
// Options:
//   - search: query string for filtering (optional)
//   - order: sort order "asc" (default) or "desc" (optional)
func (c *ChangeLogShowCommand) Execute(args map[string]string, _ string) error {
	// Get change log service
	changeLogService, err := services.GetGlobalChangeLogService()
	if err != nil {
		return fmt.Errorf("change log service not available: %w", err)
	}

	// Get variable service for storing result variables
	variableService, err := services.GetGlobalVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Get theme object for styling
	themeObj := c.getThemeObject()

	// Parse arguments
	searchQuery := args["search"]
	sortOrder := args["order"]
	if sortOrder == "" {
		sortOrder = "asc" // Default to ascending (oldest first, newest at bottom)
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		return fmt.Errorf("invalid sort order '%s': must be 'asc' or 'desc'", sortOrder)
	}

	// Get change log entries based on search
	var entries []neurotypes.ChangeLogEntry
	if searchQuery != "" {
		entries, err = changeLogService.SearchChangeLogWithOrder(searchQuery, sortOrder)
	} else {
		entries, err = changeLogService.GetChangeLogWithOrder(sortOrder)
	}
	if err != nil {
		return fmt.Errorf("failed to get change log: %w", err)
	}

	// Format output with theme styling
	output := c.formatChangeLog(entries, searchQuery, sortOrder, themeObj)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", output); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print the change log using semantic output
	printer := printing.NewDefaultPrinter()
	printer.Print(output)

	return nil
}

// formatChangeLog formats the change log entries for display with theme styling.
func (c *ChangeLogShowCommand) formatChangeLog(entries []neurotypes.ChangeLogEntry, searchQuery string, sortOrder string, themeObj *services.Theme) string {
	if len(entries) == 0 {
		searchText := ""
		if searchQuery != "" {
			searchText = fmt.Sprintf(" matching %s", themeObj.Variable.Render("'"+searchQuery+"'"))
		}
		noEntriesMsg := fmt.Sprintf("No change log entries found%s.", searchText)
		return themeObj.Warning.Render(noEntriesMsg) + "\n"
	}

	var result strings.Builder

	// Header with professional styling
	headerParts := []string{themeObj.Success.Render("NeuroShell Change Log")}
	if searchQuery != "" {
		searchPart := fmt.Sprintf("- Search: %s", themeObj.Variable.Render("'"+searchQuery+"'"))
		headerParts = append(headerParts, searchPart)
	}

	// Add sort order information
	var orderText string
	if sortOrder == "asc" {
		orderText = "oldest→newest"
	} else {
		orderText = "newest→oldest"
	}
	orderPart := fmt.Sprintf("- Order: %s", themeObj.Keyword.Render(orderText))
	headerParts = append(headerParts, orderPart)

	countPart := fmt.Sprintf("(%s)", themeObj.Info.Render(fmt.Sprintf("%d entries", len(entries))))
	headerParts = append(headerParts, countPart)

	result.WriteString(fmt.Sprintf("%s:\n\n", strings.Join(headerParts, " ")))

	// Format each entry
	for i, entry := range entries {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(c.formatChangeLogEntry(entry, themeObj))
	}

	return result.String()
}

// formatChangeLogEntry formats a single change log entry for display with professional theme styling.
func (c *ChangeLogShowCommand) formatChangeLogEntry(entry neurotypes.ChangeLogEntry, themeObj *services.Theme) string {
	var result strings.Builder

	// Entry header: [ID] Title with prominent ID
	entryID := themeObj.Highlight.Render(fmt.Sprintf("[%s]", entry.ID))
	title := themeObj.Command.Render(entry.Title)
	entryHeader := fmt.Sprintf("%s %s\n", entryID, title)
	result.WriteString(entryHeader)

	// Version and date
	versionLine := fmt.Sprintf("  %s %s\n",
		themeObj.Info.Render("Version:"),
		themeObj.Variable.Render(entry.Version))
	result.WriteString(versionLine)

	dateLine := fmt.Sprintf("  %s %s\n",
		themeObj.Info.Render("Date:"),
		themeObj.Variable.Render(entry.Date))
	result.WriteString(dateLine)

	// Type with color coding
	var typeStyled string
	switch entry.Type {
	case "bugfix":
		typeStyled = themeObj.Error.Render(c.toTitle(entry.Type))
	case "feature":
		typeStyled = themeObj.Success.Render(c.toTitle(entry.Type))
	case "enhancement":
		typeStyled = themeObj.Info.Render(c.toTitle(entry.Type))
	case "testing":
		typeStyled = themeObj.Variable.Render(c.toTitle(entry.Type))
	default:
		typeStyled = themeObj.Keyword.Render(c.toTitle(entry.Type))
	}
	typeLine := fmt.Sprintf("  %s %s\n",
		themeObj.Info.Render("Type:"),
		typeStyled)
	result.WriteString(typeLine)

	// Description
	if entry.Description != "" {
		descriptionLine := fmt.Sprintf("  %s %s\n",
			themeObj.Info.Render("Description:"),
			themeObj.Info.Render(entry.Description))
		result.WriteString(descriptionLine)
	}

	// Impact
	if entry.Impact != "" {
		impactLine := fmt.Sprintf("  %s %s\n",
			themeObj.Info.Render("Impact:"),
			themeObj.Success.Render(entry.Impact))
		result.WriteString(impactLine)
	}

	// Files changed
	if len(entry.FilesChanged) > 0 {
		filesList := make([]string, len(entry.FilesChanged))
		for i, file := range entry.FilesChanged {
			filesList[i] = themeObj.Variable.Render(file)
		}
		filesLine := fmt.Sprintf("  %s %s\n",
			themeObj.Info.Render("Files:"),
			strings.Join(filesList, ", "))
		result.WriteString(filesLine)
	}

	return result.String()
}

// toTitle converts the first character of a string to uppercase.
func (c *ChangeLogShowCommand) toTitle(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// getThemeObject retrieves the theme object based on the _style variable
func (c *ChangeLogShowCommand) getThemeObject() *services.Theme {
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

// IsReadOnly returns true as the change-log-show command doesn't modify system state.
func (c *ChangeLogShowCommand) IsReadOnly() bool {
	return true
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&ChangeLogShowCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register change-log-show command: %v", err))
	}
}
