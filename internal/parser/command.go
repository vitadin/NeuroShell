// Package parser provides command parsing functionality for NeuroShell input.
// It handles the parsing of user input into command structures with support for different syntax modes.
package parser

import (
	"fmt"
	"regexp"
	"strings"

	"neuroshell/pkg/types"
)

// Command represents a parsed NeuroShell command with its arguments and metadata.
type Command struct {
	Name           string
	BracketContent string
	Options        map[string]string
	Message        string
	ParseMode      ParseMode
}

// ParseMode defines how command arguments should be parsed from user input.
type ParseMode = types.ParseMode

const (
	// ParseModeKeyValue parses arguments as key=value pairs within brackets
	ParseModeKeyValue = types.ParseModeKeyValue
	// ParseModeRaw treats the entire input as raw text without parsing
	ParseModeRaw = types.ParseModeRaw
)

// ParseInput parses user input into a Command structure with name, options, and message.
func ParseInput(input string) *Command {
	input = strings.TrimSpace(input)

	// If doesn't start with \, treat as \send message
	if !strings.HasPrefix(input, "\\") {
		return &Command{
			Name:    "send",
			Message: input,
			Options: make(map[string]string),
		}
	}

	// Remove the leading backslash
	input = input[1:]

	cmd := &Command{
		Options: make(map[string]string),
	}

	// Try to parse command with brackets: command[content] message
	bracketRe := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\[([^\]]*)\](.*)$`)
	if matches := bracketRe.FindStringSubmatch(input); matches != nil {
		cmd.Name = matches[1]
		cmd.BracketContent = matches[2]
		cmd.Message = strings.TrimSpace(matches[3])
	} else {
		// Simple command without brackets: command message
		parts := strings.SplitN(input, " ", 2)
		cmd.Name = parts[0]
		if len(parts) > 1 {
			cmd.Message = strings.TrimSpace(parts[1])
		}
	}

	// If no command name (malformed), treat as \send
	if cmd.Name == "" {
		cmd.Name = "send"
		cmd.Message = input
	}

	// Determine parse mode based on command
	cmd.ParseMode = getParseMode(cmd.Name)

	// Parse bracket content based on mode
	if cmd.BracketContent != "" {
		if cmd.ParseMode != ParseModeRaw {
			// Parse as key=value pairs
			parseKeyValueOptions(cmd.BracketContent, cmd.Options)
		}
	}

	return cmd
}

func getParseMode(_ string) ParseMode {
	// Default to key-value parsing for all commands
	// Commands that need raw parsing will handle it internally
	return ParseModeKeyValue
}

func parseKeyValueOptions(content string, options map[string]string) {
	if content == "" {
		return
	}

	// Split by comma, handling quoted values
	parts := splitByComma(content)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "=") {
			// key=value format
			kv := strings.SplitN(part, "=", 2)
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			// Remove quotes if present
			value = unquote(value)
			options[key] = value
		} else {
			// Just a flag
			options[part] = ""
		}
	}
}

func splitByComma(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]

		if !inQuotes && (c == '"' || c == '\'') {
			inQuotes = true
			quoteChar = c
			current.WriteByte(c)
		} else if inQuotes && c == quoteChar {
			inQuotes = false
			quoteChar = 0
			current.WriteByte(c)
		} else if !inQuotes && c == ',' {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func (c *Command) String() string {
	result := fmt.Sprintf("\\%s", c.Name)
	if len(c.Options) > 0 {
		result += "["
		first := true
		for k, v := range c.Options {
			if !first {
				result += ", "
			}
			if v != "" {
				result += fmt.Sprintf("%s=%q", k, v)
			} else {
				result += k
			}
			first = false
		}
		result += "]"
	}
	if c.Message != "" {
		result += " " + c.Message
	}
	return result
}
