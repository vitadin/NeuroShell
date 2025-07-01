package services

import (
	"fmt"
	"sort"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/pkg/neurotypes"
)

// CommandInfo holds metadata about a command for help display
type CommandInfo struct {
	Name        string
	Description string
	Usage       string
	ParseMode   neurotypes.ParseMode
}

// HelpService provides command help information by storing command metadata in context
// and providing access methods. It follows the architecture rule that commands access
// services only, not registries directly.
type HelpService struct {
	initialized bool
	commands    map[string]CommandInfo
}

// NewHelpService creates a new HelpService instance
func NewHelpService() *HelpService {
	return &HelpService{
		initialized: false,
		commands:    make(map[string]CommandInfo),
	}
}

// Name returns the service name "help" for registration
func (h *HelpService) Name() string {
	return "help"
}

// Initialize collects command metadata from the command registry and stores it in context
// as system variables, following the architecture pattern
func (h *HelpService) Initialize(ctx neurotypes.Context) error {
	// Collect all commands from the global registry using thread-safe accessor
	allCommands := commands.GetGlobalRegistry().GetAll()

	// Store command metadata
	h.commands = make(map[string]CommandInfo)
	commandNames := make([]string, 0, len(allCommands))

	for _, cmd := range allCommands {
		info := CommandInfo{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Usage:       cmd.Usage(),
			ParseMode:   cmd.ParseMode(),
		}
		h.commands[cmd.Name()] = info
		commandNames = append(commandNames, cmd.Name())

		// Store individual command metadata in context as system variables
		if err := h.setCommandMetadata(ctx, info); err != nil {
			return fmt.Errorf("failed to store metadata for command %s: %w", cmd.Name(), err)
		}
	}

	// Sort command names for consistent ordering
	sort.Strings(commandNames)

	// Store list of all command names in context
	if err := h.setSystemVariable(ctx, "#cmd_list", strings.Join(commandNames, ",")); err != nil {
		return fmt.Errorf("failed to store command list: %w", err)
	}

	// Store count of commands
	if err := h.setSystemVariable(ctx, "#cmd_count", fmt.Sprintf("%d", len(commandNames))); err != nil {
		return fmt.Errorf("failed to store command count: %w", err)
	}

	h.initialized = true
	return nil
}

// GetAllCommands returns metadata for all registered commands
func (h *HelpService) GetAllCommands() ([]CommandInfo, error) {
	if !h.initialized {
		return nil, fmt.Errorf("help service not initialized")
	}

	result := make([]CommandInfo, 0, len(h.commands))
	for _, info := range h.commands {
		result = append(result, info)
	}

	// Sort by name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// GetCommand returns metadata for a specific command by name
func (h *HelpService) GetCommand(name string) (CommandInfo, error) {
	if !h.initialized {
		return CommandInfo{}, fmt.Errorf("help service not initialized")
	}

	info, exists := h.commands[name]
	if !exists {
		return CommandInfo{}, fmt.Errorf("command '%s' not found", name)
	}

	return info, nil
}

// GetCommandNames returns a sorted list of all command names
func (h *HelpService) GetCommandNames() ([]string, error) {
	if !h.initialized {
		return nil, fmt.Errorf("help service not initialized")
	}

	names := make([]string, 0, len(h.commands))
	for name := range h.commands {
		names = append(names, name)
	}

	sort.Strings(names)
	return names, nil
}

// IsValidCommand checks if a command exists
func (h *HelpService) IsValidCommand(name string) bool {
	if !h.initialized {
		return false
	}

	_, exists := h.commands[name]
	return exists
}

// GetCommandCount returns the total number of registered commands
func (h *HelpService) GetCommandCount() int {
	return len(h.commands)
}

// setCommandMetadata stores individual command metadata as system variables
func (h *HelpService) setCommandMetadata(ctx neurotypes.Context, info CommandInfo) error {
	prefix := fmt.Sprintf("#cmd_%s_", info.Name)

	// Store command description
	if err := h.setSystemVariable(ctx, prefix+"desc", info.Description); err != nil {
		return err
	}

	// Store command usage
	if err := h.setSystemVariable(ctx, prefix+"usage", info.Usage); err != nil {
		return err
	}

	// Store command parse mode
	parseModeStr := h.parseModeToString(info.ParseMode)
	if err := h.setSystemVariable(ctx, prefix+"parsemode", parseModeStr); err != nil {
		return err
	}

	return nil
}

// setSystemVariable safely sets a system variable in context
func (h *HelpService) setSystemVariable(ctx neurotypes.Context, name, value string) error {
	// Try to use SetSystemVariable if available (from concrete context implementation)
	if setter, ok := ctx.(interface {
		SetSystemVariable(string, string) error
	}); ok {
		return setter.SetSystemVariable(name, value)
	}

	// Fallback: this shouldn't happen in normal operation but provides safety
	return fmt.Errorf("context does not support setting system variables")
}

// parseModeToString converts parse mode enum to readable string
func (h *HelpService) parseModeToString(mode neurotypes.ParseMode) string {
	switch mode {
	case neurotypes.ParseModeKeyValue:
		return "KeyValue"
	case neurotypes.ParseModeRaw:
		return "Raw"
	case neurotypes.ParseModeWithOptions:
		return "WithOptions"
	default:
		return "Unknown"
	}
}
