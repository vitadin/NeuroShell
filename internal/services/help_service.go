package services

import (
	"fmt"
	"sort"

	neuroshellcontext "neuroshell/internal/context"
	"neuroshell/pkg/neurotypes"
)

// HelpService provides command help information by accessing metadata stored in context.
// It follows the architecture rule that services access state through context.
type HelpService struct {
	initialized bool
}

// NewHelpService creates a new HelpService instance
func NewHelpService() *HelpService {
	return &HelpService{
		initialized: false,
	}
}

// Name returns the service name "help" for registration
func (h *HelpService) Name() string {
	return "help"
}

// Initialize collects command metadata from the command registry and stores it in context
// as system variables, following the architecture pattern
func (h *HelpService) Initialize() error {
	h.initialized = true
	return nil
}

// GetAllCommands returns metadata for all registered commands
func (h *HelpService) GetAllCommands() ([]*neurotypes.CommandInfo, error) {
	if !h.initialized {
		return nil, fmt.Errorf("help service not initialized")
	}

	// Get all command info from global context
	globalCtx := neuroshellcontext.GetGlobalContext()
	if globalCtx == nil {
		return nil, fmt.Errorf("global context not available")
	}

	neuroCtx, ok := globalCtx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	commandInfoMap := neuroCtx.GetAllCommandInfo()

	result := make([]*neurotypes.CommandInfo, 0, len(commandInfoMap))
	for _, info := range commandInfoMap {
		result = append(result, info)
	}

	// Sort by name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// GetCommand returns metadata for a specific command by name
func (h *HelpService) GetCommand(name string) (*neurotypes.CommandInfo, error) {
	if !h.initialized {
		return nil, fmt.Errorf("help service not initialized")
	}

	// Get command info from global context
	globalCtx := neuroshellcontext.GetGlobalContext()
	if globalCtx == nil {
		return nil, fmt.Errorf("global context not available")
	}

	neuroCtx, ok := globalCtx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	info, exists := neuroCtx.GetCommandInfo(name)
	if !exists {
		return nil, fmt.Errorf("command '%s' not found", name)
	}

	return info, nil
}

// GetCommandNames returns a sorted list of all command names
func (h *HelpService) GetCommandNames() ([]string, error) {
	if !h.initialized {
		return nil, fmt.Errorf("help service not initialized")
	}

	// Get command names from global context
	globalCtx := neuroshellcontext.GetGlobalContext()
	if globalCtx == nil {
		return nil, fmt.Errorf("global context not available")
	}

	neuroCtx, ok := globalCtx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return nil, fmt.Errorf("invalid context type")
	}

	names := neuroCtx.GetRegisteredCommands()
	sort.Strings(names)
	return names, nil
}

// IsValidCommand checks if a command exists
func (h *HelpService) IsValidCommand(name string) bool {
	if !h.initialized {
		return false
	}

	// Get command existence from global context
	globalCtx := neuroshellcontext.GetGlobalContext()
	if globalCtx == nil {
		return false
	}

	neuroCtx, ok := globalCtx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return false
	}

	return neuroCtx.IsCommandRegistered(name)
}

// GetCommandCount returns the total number of registered commands
func (h *HelpService) GetCommandCount() int {
	if !h.initialized {
		return 0
	}

	// Get command count from global context
	globalCtx := neuroshellcontext.GetGlobalContext()
	if globalCtx == nil {
		return 0
	}

	neuroCtx, ok := globalCtx.(*neuroshellcontext.NeuroContext)
	if !ok {
		return 0
	}

	return len(neuroCtx.GetRegisteredCommands())
}
