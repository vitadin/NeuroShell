package parser

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Command struct {
	Name          string
	BracketContent string
	Options       map[string]string
	Message       string
	ParseMode     ParseMode
}

type ParseMode int

const (
	ParseModeKeyValue ParseMode = iota // Default: parse as key=value pairs
	ParseModeRaw                       // Raw content for commands like \bash
)

type RawCommand struct {
	Name           string `parser:"'\\' @Ident"`
	BracketContent string `parser:"( '[' @BracketContent ']' )?"`
	Message        string `parser:"@Rest?"`
}

var commandLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: "BracketContent", Pattern: `[^\]]*`},
	{Name: "Rest", Pattern: `.*`},
	{Name: "Punct", Pattern: `[\\[\]]`},
	{Name: "Whitespace", Pattern: `\s+`},
})

var parser = participle.MustBuild[RawCommand](
	participle.Lexer(commandLexer),
)

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
	
	// Parse with participle
	rawCmd, err := parser.ParseString("", input)
	if err != nil {
		// If parsing fails, treat as \send message (removing the \)
		return &Command{
			Name:    "send",
			Message: strings.TrimPrefix(input, "\\"),
			Options: make(map[string]string),
		}
	}
	
	cmd := &Command{
		Name:          rawCmd.Name,
		BracketContent: rawCmd.BracketContent,
		Message:       strings.TrimSpace(rawCmd.Message),
		Options:       make(map[string]string),
	}
	
	// Determine parse mode based on command
	cmd.ParseMode = getParseMode(cmd.Name)
	
	// Parse bracket content based on mode
	if cmd.BracketContent != "" {
		if cmd.ParseMode == ParseModeRaw {
			// Keep as raw string - don't parse
		} else {
			// Parse as key=value pairs
			parseKeyValueOptions(cmd.BracketContent, cmd.Options)
		}
	}
	
	return cmd
}

func getParseMode(commandName string) ParseMode {
	// This will be replaced by command registry lookup
	switch commandName {
	case "bash":
		return ParseModeRaw
	default:
		return ParseModeKeyValue
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