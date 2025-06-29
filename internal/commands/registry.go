package commands

import (
	"neuroshell/internal/parser"
)

type CommandInfo struct {
	Name        string
	ParseMode   parser.ParseMode
	Description string
	Usage       string
}

var CommandRegistry = map[string]CommandInfo{
	"send": {
		Name:        "send",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "Send message to LLM agent",
		Usage:       "\\send message",
	},
	"set": {
		Name:        "set",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "Set a variable",
		Usage:       "\\set[var=value] or \\set var value",
	},
	"get": {
		Name:        "get",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "Get a variable",
		Usage:       "\\get[var] or \\get var",
	},
	"bash": {
		Name:        "bash",
		ParseMode:   parser.ParseModeRaw,
		Description: "Execute system command",
		Usage:       "\\bash[command] or \\bash command",
	},
	"help": {
		Name:        "help",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "Show command help",
		Usage:       "\\help [command]",
	},
	"new": {
		Name:        "new",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "Start new session",
		Usage:       "\\new [name]",
	},
	"save": {
		Name:        "save",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "Save current session",
		Usage:       "\\save[name=\"session_name\"]",
	},
	"load": {
		Name:        "load",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "Load saved session",
		Usage:       "\\load[name=\"session_name\"]",
	},
	"clear": {
		Name:        "clear",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "Clear current session",
		Usage:       "\\clear",
	},
	"list": {
		Name:        "list",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "List all variables",
		Usage:       "\\list [pattern]",
	},
	"history": {
		Name:        "history",
		ParseMode:   parser.ParseModeKeyValue,
		Description: "View recent exchanges",
		Usage:       "\\history[n=5]",
	},
}

func GetCommandInfo(name string) (CommandInfo, bool) {
	info, exists := CommandRegistry[name]
	return info, exists
}

func GetParseMode(commandName string) parser.ParseMode {
	if info, exists := CommandRegistry[commandName]; exists {
		return info.ParseMode
	}
	return parser.ParseModeKeyValue // Default
}

func IsValidCommand(name string) bool {
	_, exists := CommandRegistry[name]
	return exists
}

func GetAllCommands() []CommandInfo {
	commands := make([]CommandInfo, 0, len(CommandRegistry))
	for _, info := range CommandRegistry {
		commands = append(commands, info)
	}
	return commands
}