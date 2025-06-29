package parser

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Command struct {
	Name    string            `parser:"'\\' @Ident"`
	Options map[string]string `parser:"'[' @@? ']'?"`
	Message string            `parser:"@@?"`
}

type Option struct {
	Key   string `parser:"@Ident"`
	Value string `parser:"('=' @String)?"`
}

var commandLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: "String", Pattern: `"([^"\\]|\\.)*"`},
	{Name: "Punct", Pattern: `[\\[\]=,]`},
	{Name: "Whitespace", Pattern: `\s+`},
})

var parser = participle.MustBuild[Command](
	participle.Lexer(commandLexer),
	participle.Unquote("String"),
)

func ParseCommand(input string) (*Command, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "\\") {
		return nil, fmt.Errorf("command must start with '\\'")
	}
	
	return parser.ParseString("", input)
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