// Package parser provides command parsing functionality for NeuroShell input.
// It handles the parsing of user input into command structures with support for different syntax modes.
package parser

import (
	"fmt"
	"regexp"
	"strings"

	"neuroshell/pkg/neurotypes"
)

// Command represents a parsed NeuroShell command with its arguments and metadata.
type Command struct {
	Name           string
	BracketContent string
	Options        map[string]string
	Message        string
	ParseMode      neurotypes.ParseMode
	OriginalText   string // Store original command text for deterministic output
}

// ParseInput parses user input into a Command structure with name, options, and message.
// For backward compatibility, it uses "echo" as the default command.
func ParseInput(input string) *Command {
	return ParseInputWithContext(input, nil)
}

// ParseInputWithContext parses user input using the provided context for default command configuration.
func ParseInputWithContext(input string, ctx neurotypes.Context) *Command {
	originalInput := input // Store original before any processing
	input = strings.TrimSpace(input)

	// Determine default command from context
	defaultCmd := "echo" // fallback default
	if ctx != nil {
		defaultCmd = ctx.GetDefaultCommand()
	}

	// If doesn't start with \, treat as default command message
	if !strings.HasPrefix(input, "\\") {
		return &Command{
			Name:         defaultCmd,
			Message:      input,
			Options:      make(map[string]string),
			OriginalText: originalInput,
		}
	}

	// Remove the leading backslash
	input = input[1:]

	cmd := &Command{
		Options:      make(map[string]string),
		OriginalText: originalInput,
	}

	// Try to parse command with brackets: command[content] message
	if parsed := parseCommandWithBrackets(input); parsed != nil {
		cmd.Name = parsed.Name
		cmd.BracketContent = parsed.BracketContent
		cmd.Message = parsed.Message
	} else {
		// Simple command without brackets: command message
		parts := strings.SplitN(input, " ", 2)
		cmd.Name = parts[0]
		if len(parts) > 1 {
			cmd.Message = strings.TrimSpace(parts[1])
		}
	}

	// If no command name (malformed), treat as default command
	if cmd.Name == "" {
		cmd.Name = defaultCmd
		cmd.Message = input
	}

	// Determine parse mode based on command
	cmd.ParseMode = getParseMode(cmd.Name)

	// Parse bracket content based on mode
	if cmd.BracketContent != "" {
		if cmd.ParseMode != neurotypes.ParseModeRaw {
			// Parse as key=value pairs
			parseKeyValueOptions(cmd.BracketContent, cmd.Options)
		}
	}

	return cmd
}

func getParseMode(_ string) neurotypes.ParseMode {
	// Default to key-value parsing for all commands
	// Commands that need raw parsing will handle it internally
	return neurotypes.ParseModeKeyValue
}

// ParsedCommand represents the result of parsing a command with brackets
type ParsedCommand struct {
	Name           string
	BracketContent string
	Message        string
}

// parseCommandWithBrackets parses a command with bracket-aware logic to handle nested brackets
func parseCommandWithBrackets(input string) *ParsedCommand {
	// Find command name (first word)
	spaceIdx := strings.Index(input, " ")
	bracketIdx := strings.Index(input, "[")

	if bracketIdx == -1 {
		return nil // No brackets found
	}

	var commandEnd int
	if spaceIdx != -1 && spaceIdx < bracketIdx {
		return nil // Space before bracket, not a bracket command
	}

	// Find the end of command name (where bracket starts)
	commandEnd = bracketIdx

	if commandEnd == 0 {
		return nil // No command name
	}

	commandName := input[:commandEnd]
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`).MatchString(commandName) {
		return nil // Invalid command name
	}

	// Parse brackets with proper nesting
	bracketDepth := 0
	contentStart := bracketIdx + 1
	contentEnd := -1

	for i := bracketIdx; i < len(input); i++ {
		switch input[i] {
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
			if bracketDepth == 0 {
				contentEnd = i
				goto bracketParsingDone
			}
		}
	}

bracketParsingDone:

	if contentEnd == -1 {
		return nil // Unclosed brackets
	}

	bracketContent := input[contentStart:contentEnd]
	var message string
	if contentEnd+1 < len(input) {
		message = strings.TrimSpace(input[contentEnd+1:])
	}

	return &ParsedCommand{
		Name:           commandName,
		BracketContent: bracketContent,
		Message:        message,
	}
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
	inBrackets := false
	quoteChar := byte(0)
	bracketDepth := 0

	for i := 0; i < len(s); i++ {
		c := s[i]

		switch {
		case !inQuotes && (c == '"' || c == '\''):
			inQuotes = true
			quoteChar = c
			current.WriteByte(c)
		case inQuotes && c == quoteChar:
			inQuotes = false
			quoteChar = 0
			current.WriteByte(c)
		case !inQuotes && c == '[':
			inBrackets = true
			bracketDepth++
			current.WriteByte(c)
		case !inQuotes && c == ']':
			bracketDepth--
			if bracketDepth <= 0 {
				inBrackets = false
				bracketDepth = 0
			}
			current.WriteByte(c)
		case !inQuotes && !inBrackets && c == ',':
			parts = append(parts, current.String())
			current.Reset()
		default:
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

// ParseArrayValue parses a string that may contain an array notation like "[item1, item2, item3]"
// and returns the items as a slice. If the input is not an array, returns a single-item slice.
func ParseArrayValue(value string) []string {
	value = strings.TrimSpace(value)

	// Check if it's an array notation [...]
	if len(value) >= 2 && value[0] == '[' && value[len(value)-1] == ']' {
		// Extract content inside brackets
		content := value[1 : len(value)-1]
		content = strings.TrimSpace(content)

		if content == "" {
			return []string{}
		}

		// Split by comma and clean up each item
		items := strings.Split(content, ",")
		var result []string
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item != "" {
				// Remove quotes if present
				item = unquote(item)
				result = append(result, item)
			}
		}
		return result
	}

	// Not an array, return single item
	return []string{unquote(value)}
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
