// Package session provides session management commands for NeuroShell.
// It includes commands for creating, managing, and interacting with chat sessions.
package session

import (
	"fmt"
	"sort"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// ListCommand implements the \session-list command for listing chat sessions.
// It provides session listing functionality with support for filtering and sorting.
type ListCommand struct{}

// Name returns the command name "session-list" for registration and lookup.
func (c *ListCommand) Name() string {
	return "session-list"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *ListCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the session-list command does.
func (c *ListCommand) Description() string {
	return "List all existing chat sessions"
}

// Usage returns the syntax and usage examples for the session-list command.
func (c *ListCommand) Usage() string {
	return `\session-list[sort=name|created|updated, filter=active]

Examples:
  \session-list                           %% List all sessions (default: sorted by created, newest first)
  \session-list[sort=name]                %% List sessions sorted alphabetically by name
  \session-list[sort=updated]             %% List sessions sorted by last update (newest first)
  \session-list[filter=active]            %% Show only the active session
  \session-list[sort=name,filter=active]  %% Show active session only, sorted by name
  
Options:
  sort   - Sort order: name (alphabetical), created (newest first), updated (newest first)
  filter - Filter criteria: active (only active session), all (default)
  
Note: Options can be combined. Default sort is by creation time (newest first).
      Session list is stored in ${_output} variable.`
}

// HelpInfo returns structured help information for the session-list command.
func (c *ListCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       "\\session-list[sort=name|created|updated, filter=active]",
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "sort",
				Description: "Sort order: name (alphabetical), created (newest first), updated (newest first)",
				Required:    false,
				Type:        "string",
				Default:     "created",
			},
			{
				Name:        "filter",
				Description: "Filter criteria: active (only active session), all (default)",
				Required:    false,
				Type:        "string",
				Default:     "all",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\session-list",
				Description: "List all sessions sorted by creation time (newest first)",
			},
			{
				Command:     "\\session-list[sort=name]",
				Description: "List sessions sorted alphabetically by name",
			},
			{
				Command:     "\\session-list[filter=active]",
				Description: "Show only the currently active session",
			},
			{
				Command:     "\\session-list[sort=updated, filter=all]",
				Description: "List all sessions by last update time",
			},
		},
		Notes: []string{
			"Options can be combined (e.g., sort=name,filter=active)",
			"Default sort is by creation time with newest sessions first",
			"Session list output is stored in ${_output} variable",
			"Shows session name, ID (short), active status, message count, and creation date",
			"Active session is marked with 'active' status indicator",
		},
	}
}

// Execute lists chat sessions with optional filtering and sorting.
// Options:
//   - sort: name|created|updated (default: created)
//   - filter: active|all (default: all)
func (c *ListCommand) Execute(args map[string]string, _ string, ctx neurotypes.Context) error {
	// Get chat session service
	chatService, err := c.getChatSessionService()
	if err != nil {
		return fmt.Errorf("chat session service not available: %w", err)
	}

	// Get variable service for storing result variables
	variableService, err := c.getVariableService()
	if err != nil {
		return fmt.Errorf("variable service not available: %w", err)
	}

	// Parse arguments
	sortBy := args["sort"]
	if sortBy == "" {
		sortBy = "created" // default sort
	}
	filterBy := args["filter"]
	if filterBy == "" {
		filterBy = "all" // default filter
	}

	// Validate arguments
	if err := c.validateArguments(sortBy, filterBy); err != nil {
		return err
	}

	// Get all sessions from service
	allSessions := chatService.ListSessions()

	// Apply filtering
	sessions, err := c.filterSessions(allSessions, filterBy, chatService)
	if err != nil {
		return err
	}

	// Apply sorting
	c.sortSessions(sessions, sortBy)

	// Format output
	output := c.formatSessionList(sessions)

	// Store result in _output variable
	if err := variableService.SetSystemVariable("_output", output, ctx); err != nil {
		return fmt.Errorf("failed to store result: %w", err)
	}

	// Print the list
	fmt.Print(output)

	return nil
}

// validateArguments checks if the provided sort and filter options are valid.
func (c *ListCommand) validateArguments(sortBy, filterBy string) error {
	validSorts := map[string]bool{
		"name":    true,
		"created": true,
		"updated": true,
	}
	if !validSorts[sortBy] {
		return fmt.Errorf("invalid sort option '%s'. Valid options: name, created, updated", sortBy)
	}

	validFilters := map[string]bool{
		"all":    true,
		"active": true,
	}
	if !validFilters[filterBy] {
		return fmt.Errorf("invalid filter option '%s'. Valid options: all, active", filterBy)
	}

	return nil
}

// filterSessions applies the specified filter to the session list.
func (c *ListCommand) filterSessions(sessions []*neurotypes.ChatSession, filterBy string, chatService *services.ChatSessionService) ([]*neurotypes.ChatSession, error) {
	switch filterBy {
	case "all":
		return sessions, nil
	case "active":
		activeSession, err := chatService.GetActiveSession()
		if err != nil {
			if err.Error() == "no active session" {
				return []*neurotypes.ChatSession{}, nil
			}
			return nil, fmt.Errorf("failed to get active session: %w", err)
		}
		return []*neurotypes.ChatSession{activeSession}, nil
	default:
		return sessions, nil
	}
}

// sortSessions sorts the session list according to the specified criteria.
func (c *ListCommand) sortSessions(sessions []*neurotypes.ChatSession, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(sessions, func(i, j int) bool {
			return strings.ToLower(sessions[i].Name) < strings.ToLower(sessions[j].Name)
		})
	case "created":
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].CreatedAt.After(sessions[j].CreatedAt) // newest first
		})
	case "updated":
		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt) // newest first
		})
	}
}

// formatSessionList formats the session list for display.
func (c *ListCommand) formatSessionList(sessions []*neurotypes.ChatSession) string {
	if len(sessions) == 0 {
		return "No sessions found.\n"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Sessions (%d total):\n", len(sessions)))

	for _, session := range sessions {
		// Format: name (ID: shortid, active/inactive, X messages, created: date)
		shortID := session.ID
		if len(shortID) > 8 {
			shortID = shortID[:8]
		}

		status := ""
		if session.IsActive {
			status = ", active"
		}

		messageCount := len(session.Messages)
		messageText := "messages"
		if messageCount == 1 {
			messageText = "message"
		}

		createdDate := session.CreatedAt.Format("2006-01-02")

		result.WriteString(fmt.Sprintf("  %s    (ID: %s%s, %d %s, created: %s)\n",
			session.Name, shortID, status, messageCount, messageText, createdDate))
	}

	return result.String()
}

// getChatSessionService retrieves the chat session service from the global registry
func (c *ListCommand) getChatSessionService() (*services.ChatSessionService, error) {
	service, err := services.GetGlobalRegistry().GetService("chat_session")
	if err != nil {
		return nil, err
	}

	chatService, ok := service.(*services.ChatSessionService)
	if !ok {
		return nil, fmt.Errorf("chat session service has incorrect type")
	}

	return chatService, nil
}

// getVariableService retrieves the variable service from the global registry
func (c *ListCommand) getVariableService() (*services.VariableService, error) {
	service, err := services.GetGlobalRegistry().GetService("variable")
	if err != nil {
		return nil, err
	}

	variableService, ok := service.(*services.VariableService)
	if !ok {
		return nil, fmt.Errorf("variable service has incorrect type")
	}

	return variableService, nil
}

func init() {
	if err := commands.GlobalRegistry.Register(&ListCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register session-list command: %v", err))
	}
}
