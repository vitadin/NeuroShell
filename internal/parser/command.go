package parser

import (
	"fmt"
	"regexp"
	"strings"
)

type Command struct {
	Name    string
	Options map[string]string
	Message string
}

func ParseCommand(input string) (*Command, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "\\") {
		return nil, fmt.Errorf("command must start with '\\'")
	}
	
	// Remove the leading backslash
	input = input[1:]
	
	// Parse command name
	parts := strings.SplitN(input, " ", 2)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	
	nameAndOptions := parts[0]
	var message string
	if len(parts) > 1 {
		message = strings.TrimSpace(parts[1])
	}
	
	// Check if command has options in brackets
	cmd := &Command{
		Options: make(map[string]string),
		Message: message,
	}
	
	if strings.Contains(nameAndOptions, "[") {
		// Parse command with options: command[opt1=val1,opt2=val2]
		re := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\[([^\]]*)\]$`)
		matches := re.FindStringSubmatch(nameAndOptions)
		if matches == nil {
			return nil, fmt.Errorf("invalid command format")
		}
		
		cmd.Name = matches[1]
		optionsStr := matches[2]
		
		if optionsStr != "" {
			// Parse options
			if err := parseOptions(optionsStr, cmd.Options); err != nil {
				return nil, fmt.Errorf("invalid options: %v", err)
			}
		}
	} else {
		// Simple command without options
		cmd.Name = nameAndOptions
	}
	
	return cmd, nil
}

func parseOptions(optionsStr string, options map[string]string) error {
	// Split by comma, but handle quoted values
	parts := splitOptions(optionsStr)
	
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
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
			
			options[key] = value
		} else {
			// Just a flag
			options[part] = ""
		}
	}
	
	return nil
}

func splitOptions(s string) []string {
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