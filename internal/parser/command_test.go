package parser

import (
	"strings"
	"testing"

	"neuroshell/pkg/neurotypes"

	"github.com/stretchr/testify/assert"
)

func TestParseInput_BackslashPrefix(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedName    string
		expectedMsg     string
		expectedOpts    map[string]string
		expectedBracket string
	}{
		{
			name:         "simple command with message",
			input:        "\\set hello world",
			expectedName: "set",
			expectedMsg:  "hello world",
			expectedOpts: map[string]string{},
		},
		{
			name:         "command with single word",
			input:        "\\help",
			expectedName: "help",
			expectedMsg:  "",
			expectedOpts: map[string]string{},
		},
		{
			name:            "command with bracket options",
			input:           "\\set[var=value] hello",
			expectedName:    "set",
			expectedMsg:     "hello",
			expectedBracket: "var=value",
			expectedOpts:    map[string]string{"var": "value"},
		},
		{
			name:            "command with multiple bracket options",
			input:           "\\bash[timeout=5, verbose] ls -la",
			expectedName:    "bash",
			expectedMsg:     "ls -la",
			expectedBracket: "timeout=5, verbose",
			expectedOpts:    map[string]string{"timeout": "5", "verbose": ""},
		},
		{
			name:            "command with quoted values",
			input:           "\\set[var=\"hello world\"] message",
			expectedName:    "set",
			expectedMsg:     "message",
			expectedBracket: "var=\"hello world\"",
			expectedOpts:    map[string]string{"var": "hello world"},
		},
		{
			name:            "command with empty brackets",
			input:           "\\command[] message",
			expectedName:    "command",
			expectedMsg:     "message",
			expectedBracket: "",
			expectedOpts:    map[string]string{},
		},
		{
			name:         "command with extra spaces",
			input:        "\\set    hello   world   ",
			expectedName: "set",
			expectedMsg:  "hello   world",
			expectedOpts: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseInput(tt.input)

			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.Equal(t, tt.expectedMsg, cmd.Message)
			assert.Equal(t, tt.expectedBracket, cmd.BracketContent)
			assert.Equal(t, tt.expectedOpts, cmd.Options)
		})
	}
}

func TestParseInput_NoBackslashPrefix(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedMsg  string
	}{
		{
			name:         "plain text becomes echo command",
			input:        "hello world",
			expectedName: "echo",
			expectedMsg:  "hello world",
		},
		{
			name:         "single word becomes echo",
			input:        "hello",
			expectedName: "echo",
			expectedMsg:  "hello",
		},
		{
			name:         "complex message becomes echo",
			input:        "analyze this data and create a report",
			expectedName: "echo",
			expectedMsg:  "analyze this data and create a report",
		},
		{
			name:         "message with special characters",
			input:        "hello! @user #hashtag $variable",
			expectedName: "echo",
			expectedMsg:  "hello! @user #hashtag $variable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseInput(tt.input)

			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.Equal(t, tt.expectedMsg, cmd.Message)
			assert.Equal(t, map[string]string{}, cmd.Options)
		})
	}
}

func TestParseInput_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedMsg  string
	}{
		{
			name:         "empty input",
			input:        "",
			expectedName: "echo",
			expectedMsg:  "",
		},
		{
			name:         "whitespace only",
			input:        "   ",
			expectedName: "echo",
			expectedMsg:  "",
		},
		{
			name:         "just backslash",
			input:        "\\",
			expectedName: "echo",
			expectedMsg:  "",
		},
		{
			name:         "backslash with spaces",
			input:        "\\   ",
			expectedName: "echo",
			expectedMsg:  "",
		},
		{
			name:         "malformed bracket syntax",
			input:        "\\command[unclosed",
			expectedName: "command[unclosed",
			expectedMsg:  "",
		},
		{
			name:         "multiple backslashes",
			input:        "\\\\command message",
			expectedName: "\\command",
			expectedMsg:  "message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseInput(tt.input)

			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.Equal(t, tt.expectedMsg, cmd.Message)
		})
	}
}

func TestParseInput_ComplexBracketSyntax(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedName    string
		expectedMsg     string
		expectedBracket string
		expectedOpts    map[string]string
	}{
		{
			name:            "nested quotes",
			input:           "\\set[key=\"value with 'quotes'\"] message",
			expectedName:    "set",
			expectedMsg:     "message",
			expectedBracket: "key=\"value with 'quotes'\"",
			expectedOpts:    map[string]string{"key": "value with 'quotes'"},
		},
		{
			name:            "comma in quoted value",
			input:           "\\bash[cmd=\"ls -la, pwd\"] execute",
			expectedName:    "bash",
			expectedMsg:     "execute",
			expectedBracket: "cmd=\"ls -la, pwd\"",
			expectedOpts:    map[string]string{"cmd": "ls -la, pwd"},
		},
		{
			name:            "multiple flags and values",
			input:           "\\run[verbose, timeout=30, force, file=\"test.txt\"] script",
			expectedName:    "run",
			expectedMsg:     "script",
			expectedBracket: "verbose, timeout=30, force, file=\"test.txt\"",
			expectedOpts:    map[string]string{"verbose": "", "timeout": "30", "force": "", "file": "test.txt"},
		},
		{
			name:            "single quotes",
			input:           "\\set[var='single quoted value'] msg",
			expectedName:    "set",
			expectedMsg:     "msg",
			expectedBracket: "var='single quoted value'",
			expectedOpts:    map[string]string{"var": "single quoted value"},
		},
		{
			name:            "equal sign in value",
			input:           "\\set[equation=\"x=y+z\"] formula",
			expectedName:    "set",
			expectedMsg:     "formula",
			expectedBracket: "equation=\"x=y+z\"",
			expectedOpts:    map[string]string{"equation": "x=y+z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseInput(tt.input)

			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.Equal(t, tt.expectedMsg, cmd.Message)
			assert.Equal(t, tt.expectedBracket, cmd.BracketContent)
			assert.Equal(t, tt.expectedOpts, cmd.Options)
		})
	}
}

func TestParseInput_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedName string
		expectedMsg  string
	}{
		{
			name:         "unicode characters",
			input:        "\\send こんにちは world 🌍",
			expectedName: "send",
			expectedMsg:  "こんにちは world 🌍",
		},
		{
			name:         "shell metacharacters",
			input:        "\\bash ls -la | grep test > output.txt",
			expectedName: "bash",
			expectedMsg:  "ls -la | grep test > output.txt",
		},
		{
			name:         "escape sequences",
			input:        "\\send hello\\nworld\\ttab",
			expectedName: "send",
			expectedMsg:  "hello\\nworld\\ttab",
		},
		{
			name:         "mixed quotes and brackets",
			input:        "analyze \"this [data]\" and 'that {info}'",
			expectedName: "echo",
			expectedMsg:  "analyze \"this [data]\" and 'that {info}'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseInput(tt.input)

			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.Equal(t, tt.expectedMsg, cmd.Message)
		})
	}
}

func TestParseInput_ParseMode(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedParseMode neurotypes.ParseMode
	}{
		{
			name:              "default command uses key-value mode",
			input:             "\\set[var=value] message",
			expectedParseMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:              "bash command uses key-value mode (default)",
			input:             "\\bash[cmd=ls] execute",
			expectedParseMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:              "echo command uses key-value mode",
			input:             "\\echo[urgent] message",
			expectedParseMode: neurotypes.ParseModeKeyValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseInput(tt.input)
			assert.Equal(t, tt.expectedParseMode, cmd.ParseMode)
		})
	}
}

func TestParseInput_LongInputs(t *testing.T) {
	// Test with very long inputs
	longMessage := strings.Repeat("very long message ", 1000)
	longKey := strings.Repeat("key", 100)
	longValue := strings.Repeat("value ", 500)

	tests := []struct {
		name         string
		input        string
		expectedName string
	}{
		{
			name:         "very long message",
			input:        "\\send " + longMessage,
			expectedName: "send",
		},
		{
			name:         "long bracket content",
			input:        "\\set[" + longKey + "=\"" + longValue + "\"] message",
			expectedName: "set",
		},
		{
			name:         "long plain text",
			input:        longMessage,
			expectedName: "echo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ParseInput(tt.input)
			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.NotNil(t, cmd.Options)
		})
	}
}

func TestParseKeyValueOptions(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedOptions map[string]string
	}{
		{
			name:            "empty content",
			content:         "",
			expectedOptions: map[string]string{},
		},
		{
			name:            "single key=value",
			content:         "key=value",
			expectedOptions: map[string]string{"key": "value"},
		},
		{
			name:            "multiple key=value pairs",
			content:         "key1=value1, key2=value2",
			expectedOptions: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:            "mixed flags and key=value",
			content:         "flag1, key=value, flag2",
			expectedOptions: map[string]string{"flag1": "", "key": "value", "flag2": ""},
		},
		{
			name:            "quoted values",
			content:         "key1=\"quoted value\", key2='single quoted'",
			expectedOptions: map[string]string{"key1": "quoted value", "key2": "single quoted"},
		},
		{
			name:            "comma in quoted value",
			content:         "cmd=\"ls -la, pwd\", verbose",
			expectedOptions: map[string]string{"cmd": "ls -la, pwd", "verbose": ""},
		},
		{
			name:            "equal sign in quoted value",
			content:         "equation=\"x=y+z\", formula=\"a=b\"",
			expectedOptions: map[string]string{"equation": "x=y+z", "formula": "a=b"},
		},
		{
			name:            "nested quotes",
			content:         "key=\"value with 'nested' quotes\"",
			expectedOptions: map[string]string{"key": "value with 'nested' quotes"},
		},
		{
			name:            "whitespace handling",
			content:         " key1 = value1 ,  key2  =  value2  ",
			expectedOptions: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:            "empty values",
			content:         "key1=, key2=\"\"",
			expectedOptions: map[string]string{"key1": "", "key2": ""},
		},
		{
			name:            "special characters in keys",
			content:         "key_1=value, key-2=value, key.3=value",
			expectedOptions: map[string]string{"key_1": "value", "key-2": "value", "key.3": "value"},
		},
		{
			name:            "special characters in values",
			content:         "url=\"https://example.com?q=test&p=1\", path=\"/usr/local/bin\"",
			expectedOptions: map[string]string{"url": "https://example.com?q=test&p=1", "path": "/usr/local/bin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make(map[string]string)
			parseKeyValueOptions(tt.content, options)
			assert.Equal(t, tt.expectedOptions, options)
		})
	}
}

func TestParseKeyValueOptions_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedOptions map[string]string
	}{
		{
			name:            "only commas",
			content:         ",,,",
			expectedOptions: map[string]string{},
		},
		{
			name:            "trailing comma",
			content:         "key=value,",
			expectedOptions: map[string]string{"key": "value"},
		},
		{
			name:            "leading comma",
			content:         ",key=value",
			expectedOptions: map[string]string{"key": "value"},
		},
		{
			name:            "multiple consecutive commas",
			content:         "key1=value1,,,key2=value2",
			expectedOptions: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:            "malformed key=value (multiple equals)",
			content:         "key=value=extra",
			expectedOptions: map[string]string{"key": "value=extra"},
		},
		{
			name:            "unmatched quotes",
			content:         "key=\"unmatched quote",
			expectedOptions: map[string]string{"key": "\"unmatched quote"},
		},
		{
			name:            "mixed quote neurotypes",
			content:         "key1=\"double\", key2='single', key3=unquoted",
			expectedOptions: map[string]string{"key1": "double", "key2": "single", "key3": "unquoted"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := make(map[string]string)
			parseKeyValueOptions(tt.content, options)
			assert.Equal(t, tt.expectedOptions, options)
		})
	}
}

func TestSplitByComma(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil, // splitByComma returns nil for empty string
		},
		{
			name:     "single item",
			input:    "item",
			expected: []string{"item"},
		},
		{
			name:     "multiple items",
			input:    "item1,item2,item3",
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name:     "items with spaces",
			input:    "item1, item2, item3",
			expected: []string{"item1", " item2", " item3"},
		},
		{
			name:     "quoted item with comma",
			input:    "\"item1, with comma\",item2",
			expected: []string{"\"item1, with comma\"", "item2"},
		},
		{
			name:     "single quoted item with comma",
			input:    "'item1, with comma',item2",
			expected: []string{"'item1, with comma'", "item2"},
		},
		{
			name:     "mixed quotes",
			input:    "\"double quoted\", 'single quoted', unquoted",
			expected: []string{"\"double quoted\"", " 'single quoted'", " unquoted"},
		},
		{
			name:     "nested quotes",
			input:    "\"outer 'inner' quotes\", other",
			expected: []string{"\"outer 'inner' quotes\"", " other"},
		},
		{
			name:     "complex nested quotes",
			input:    "'outer \"inner\" quotes', \"outer 'inner' quotes\"",
			expected: []string{"'outer \"inner\" quotes'", " \"outer 'inner' quotes\""},
		},
		{
			name:     "consecutive commas",
			input:    "item1,,item2,,,item3",
			expected: []string{"item1", "", "item2", "", "", "item3"},
		},
		{
			name:     "trailing comma",
			input:    "item1,item2,",
			expected: []string{"item1", "item2"}, // Implementation doesn't include empty trailing element
		},
		{
			name:     "leading comma",
			input:    ",item1,item2",
			expected: []string{"", "item1", "item2"},
		},
		{
			name:     "only commas",
			input:    ",,,",
			expected: []string{"", "", ""}, // Implementation returns n-1 elements for n commas
		},
		{
			name:     "unmatched quote at end",
			input:    "item1,\"unmatched",
			expected: []string{"item1", "\"unmatched"},
		},
		{
			name:     "unmatched quote at start",
			input:    "\"unmatched,item2",
			expected: []string{"\"unmatched,item2"},
		},
		{
			name:     "empty quoted strings",
			input:    "\"\",''",
			expected: []string{"\"\"", "''"},
		},
		{
			name:     "quotes with equals",
			input:    "key=\"value, with comma\",other=simple",
			expected: []string{"key=\"value, with comma\"", "other=simple"},
		},
		{
			name:     "complex real-world example",
			input:    "verbose, timeout=30, cmd=\"ls -la, pwd\", force, name='test file'",
			expected: []string{"verbose", " timeout=30", " cmd=\"ls -la, pwd\"", " force", " name='test file'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByComma(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitByComma_ComplexQuoting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "alternating quotes",
			input:    "\"item1\",'item2',\"item3\",'item4'",
			expected: []string{"\"item1\"", "'item2'", "\"item3\"", "'item4'"},
		},
		{
			name:     "quotes with special characters",
			input:    "\"item with \\\"escaped\\\" quotes\", 'item with \\'escaped\\' quotes'",
			expected: []string{"\"item with \\\"escaped\\\" quotes\"", " 'item with \\'escaped\\' quotes'"},
		},
		{
			name:     "urls with commas in query parameters",
			input:    "url=\"https://example.com?a=1,b=2\", other=value",
			expected: []string{"url=\"https://example.com?a=1,b=2\"", " other=value"},
		},
		{
			name:     "shell command with pipes and redirects",
			input:    "cmd=\"ls -la | grep test, cat file.txt > output.txt\"",
			expected: []string{"cmd=\"ls -la | grep test, cat file.txt > output.txt\""},
		},
		{
			name:     "json-like structure",
			input:    "data='{\"key\": \"value, with comma\", \"other\": 123}'",
			expected: []string{"data='{\"key\": \"value, with comma\", \"other\": 123}'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByComma(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double quoted string",
			input:    "\"hello world\"",
			expected: "hello world",
		},
		{
			name:     "single quoted string",
			input:    "'hello world'",
			expected: "hello world",
		},
		{
			name:     "unquoted string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "empty double quotes",
			input:    "\"\"",
			expected: "",
		},
		{
			name:     "empty single quotes",
			input:    "''",
			expected: "",
		},
		{
			name:     "single character quoted",
			input:    "\"a\"",
			expected: "a",
		},
		{
			name:     "single character unquoted",
			input:    "a",
			expected: "a",
		},
		{
			name:     "partial quote at start only",
			input:    "\"hello world",
			expected: "\"hello world",
		},
		{
			name:     "partial quote at end only",
			input:    "hello world\"",
			expected: "hello world\"",
		},
		{
			name:     "mismatched quotes",
			input:    "\"hello world'",
			expected: "\"hello world'",
		},
		{
			name:     "quotes in middle",
			input:    "hello\"world",
			expected: "hello\"world",
		},
		{
			name:     "nested quotes (double outer)",
			input:    "\"hello 'world'\"",
			expected: "hello 'world'",
		},
		{
			name:     "nested quotes (single outer)",
			input:    "'hello \"world\"'",
			expected: "hello \"world\"",
		},
		{
			name:     "escaped quotes inside",
			input:    "\"hello \\\"world\\\"\"",
			expected: "hello \"world\"",
		},
		{
			name:     "just quotes",
			input:    "\"",
			expected: "\"",
		},
		{
			name:     "three quotes",
			input:    "\"\"\"",
			expected: "\"", // unquote("\"\"\"") removes outer quotes, leaving middle quote
		},
		{
			name:     "whitespace with quotes",
			input:    "\"  hello world  \"",
			expected: "  hello world  ",
		},
		{
			name:     "special characters",
			input:    "\"hello@#$%^&*()world\"",
			expected: "hello@#$%^&*()world",
		},
		{
			name:     "unicode characters",
			input:    "\"こんにちは世界\"",
			expected: "こんにちは世界",
		},
		{
			name:     "very long quoted string",
			input:    "\"" + strings.Repeat("very long string ", 100) + "\"",
			expected: strings.Repeat("very long string ", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unquote(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommand_String(t *testing.T) {
	tests := []struct {
		name     string
		command  *Command
		expected string
	}{
		{
			name: "simple command with message",
			command: &Command{
				Name:    "send",
				Message: "hello world",
				Options: map[string]string{},
			},
			expected: "\\send hello world",
		},
		{
			name: "command without message",
			command: &Command{
				Name:    "help",
				Options: map[string]string{},
			},
			expected: "\\help",
		},
		{
			name: "command with single option",
			command: &Command{
				Name:    "set",
				Message: "hello",
				Options: map[string]string{"var": "value"},
			},
			expected: "\\set[var=\"value\"] hello",
		},
		{
			name: "command with flag option",
			command: &Command{
				Name:    "bash",
				Message: "ls -la",
				Options: map[string]string{"verbose": ""},
			},
			expected: "\\bash[verbose] ls -la",
		},
		{
			name: "command with multiple options",
			command: &Command{
				Name:    "run",
				Message: "script.sh",
				Options: map[string]string{
					"timeout": "30",
					"verbose": "",
					"file":    "test.txt",
				},
			},
			// Note: map iteration order is not guaranteed, so we'll check this differently
		},
		{
			name: "command with empty message but options",
			command: &Command{
				Name:    "set",
				Options: map[string]string{"var": "value"},
			},
			expected: "\\set[var=\"value\"]",
		},
		{
			name: "command with quoted value containing spaces",
			command: &Command{
				Name:    "set",
				Message: "test",
				Options: map[string]string{"var": "hello world"},
			},
			expected: "\\set[var=\"hello world\"] test",
		},
		{
			name: "command with special characters in value",
			command: &Command{
				Name:    "bash",
				Message: "execute",
				Options: map[string]string{"cmd": "ls -la | grep test"},
			},
			expected: "\\bash[cmd=\"ls -la | grep test\"] execute",
		},
		{
			name: "command with unicode",
			command: &Command{
				Name:    "send",
				Message: "こんにちは世界",
				Options: map[string]string{"lang": "ja"},
			},
			expected: "\\send[lang=\"ja\"] こんにちは世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.String()

			// For commands with multiple options, we need to check components due to map iteration order
			if len(tt.command.Options) > 1 {
				assert.Contains(t, result, "\\"+tt.command.Name)
				if tt.command.Message != "" {
					assert.Contains(t, result, tt.command.Message)
				}
				for key, value := range tt.command.Options {
					if value != "" {
						assert.Contains(t, result, key+"=\""+value+"\"")
					} else {
						assert.Contains(t, result, key)
					}
				}
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCommand_String_Deterministic(t *testing.T) {
	// Test that String() is deterministic for the same command
	cmd := &Command{
		Name:    "test",
		Message: "message",
		Options: map[string]string{"key": "value"},
	}

	result1 := cmd.String()
	result2 := cmd.String()

	assert.Equal(t, result1, result2, "String() should be deterministic")
}

func TestCommand_String_EmptyCommand(t *testing.T) {
	// Test edge case with completely empty command
	cmd := &Command{
		Options: map[string]string{},
	}

	result := cmd.String()
	assert.Equal(t, "\\", result)
}

func TestCommand_String_LongContent(t *testing.T) {
	// Test with very long content
	longMessage := strings.Repeat("very long message ", 100)
	longValue := strings.Repeat("long value ", 50)

	cmd := &Command{
		Name:    "send",
		Message: longMessage,
		Options: map[string]string{"data": longValue},
	}

	result := cmd.String()
	assert.Contains(t, result, "\\send")
	assert.Contains(t, result, longMessage)
	assert.Contains(t, result, "data=\""+longValue+"\"")
}

func TestGetParseMode(t *testing.T) {
	tests := []struct {
		name         string
		commandName  string
		expectedMode neurotypes.ParseMode
	}{
		{
			name:         "set command",
			commandName:  "set",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:         "get command",
			commandName:  "get",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:         "bash command",
			commandName:  "bash",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:         "echo command",
			commandName:  "echo",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:         "help command",
			commandName:  "help",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:         "unknown command",
			commandName:  "unknown",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:         "empty command name",
			commandName:  "",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
		{
			name:         "special characters in command name",
			commandName:  "test-command_123",
			expectedMode: neurotypes.ParseModeKeyValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getParseMode(tt.commandName)
			assert.Equal(t, tt.expectedMode, result)
		})
	}
}

func TestParseInput_ErrorHandling(t *testing.T) {
	// These tests verify that the parser handles malformed input gracefully
	// without panicking and produces reasonable output
	tests := []struct {
		name           string
		input          string
		expectedName   string
		shouldNotPanic bool
	}{
		{
			name:           "extremely long input",
			input:          strings.Repeat("very long input ", 10000),
			expectedName:   "echo",
			shouldNotPanic: true,
		},
		{
			name:           "input with null bytes",
			input:          "\\command\x00test\x00message",
			expectedName:   "command\x00test\x00message", // Parser preserves null bytes in command name
			shouldNotPanic: true,
		},
		{
			name:           "input with control characters",
			input:          "\\test\r\n\t message",
			expectedName:   "test\r\n\t", // Parser preserves control characters in command name
			shouldNotPanic: true,
		},
		{
			name:           "deeply nested brackets",
			input:          "\\cmd[key=\"[[[[nested]]]]\" value]",
			expectedName:   "cmd",
			shouldNotPanic: true,
		},
		{
			name:           "malformed regex patterns",
			input:          "\\test[.*+?{}()[]^$|\\] message",
			expectedName:   "test",
			shouldNotPanic: true,
		},
		{
			name:           "extremely large options",
			input:          "\\set[data=\"" + strings.Repeat("x", 100000) + "\"] msg",
			expectedName:   "set",
			shouldNotPanic: true,
		},
		{
			name:           "binary data",
			input:          string([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}),
			expectedName:   "echo",
			shouldNotPanic: true,
		},
		{
			name:           "unicode normalization edge cases",
			input:          "\\test é vs é message", // Different unicode representations
			expectedName:   "test",
			shouldNotPanic: true,
		},
		{
			name:           "extremely nested quotes",
			input:          "\\cmd[val=\"'\"'\"'\"'\"'\"'\"'\"] msg",
			expectedName:   "cmd",
			shouldNotPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && tt.shouldNotPanic {
					t.Errorf("ParseInput panicked on input %q: %v", tt.input, r)
				}
			}()

			cmd := ParseInput(tt.input)
			assert.NotNil(t, cmd)
			assert.Equal(t, tt.expectedName, cmd.Name)
			assert.NotNil(t, cmd.Options)
		})
	}
}

func TestParseInput_SecurityConsiderations(t *testing.T) {
	// Test potential security edge cases
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "script injection attempt",
			input: "\\bash[cmd=\"rm -rf /; echo malicious\"] execute",
		},
		{
			name:  "path traversal attempt",
			input: "\\run[file=\"../../../etc/passwd\"] read",
		},
		{
			name:  "environment variable injection",
			input: "\\set[var=\"$HOME; rm -rf /\"] value",
		},
		{
			name:  "sql injection style",
			input: "\\query[sql=\"'; DROP TABLE users; --\"] execute",
		},
		{
			name:  "command substitution attempt",
			input: "\\bash[cmd=\"$(rm -rf /)\"] execute",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parser should not execute or interpret these, just parse them
			cmd := ParseInput(tt.input)
			assert.NotNil(t, cmd)
			assert.NotNil(t, cmd.Options)
			// The important thing is that parsing doesn't execute anything
		})
	}
}

// Benchmark tests for performance analysis
func BenchmarkParseInput_Simple(b *testing.B) {
	input := "\\set[var=value] hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseInput(input)
	}
}

func BenchmarkParseInput_Complex(b *testing.B) {
	input := "\\bash[timeout=30, verbose, cmd=\"ls -la | grep test\", force] execute complex command"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseInput(input)
	}
}

func BenchmarkParseInput_NoBackslash(b *testing.B) {
	input := "this is a simple message without backslash"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseInput(input)
	}
}

func BenchmarkParseInput_LongInput(b *testing.B) {
	longInput := "\\send " + strings.Repeat("very long message ", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseInput(longInput)
	}
}

func BenchmarkParseKeyValueOptions_Simple(b *testing.B) {
	content := "key1=value1, key2=value2, flag"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		options := make(map[string]string) // Create fresh map for each iteration
		parseKeyValueOptions(content, options)
	}
}

func BenchmarkParseKeyValueOptions_Complex(b *testing.B) {
	content := "verbose, timeout=30, cmd=\"ls -la, pwd\", force, file='test file', url=\"https://example.com?a=1,b=2\""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		options := make(map[string]string) // Create fresh map for each iteration
		parseKeyValueOptions(content, options)
	}
}

func BenchmarkSplitByComma_Simple(b *testing.B) {
	input := "item1,item2,item3,item4,item5"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitByComma(input)
	}
}

func BenchmarkSplitByComma_WithQuotes(b *testing.B) {
	input := "\"item1, with comma\", 'item2, also with comma', item3, \"item4\""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitByComma(input)
	}
}

func BenchmarkSplitByComma_LongInput(b *testing.B) {
	items := make([]string, 1000)
	for i := range items {
		items[i] = "item" + string(rune(i))
	}
	input := strings.Join(items, ",")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitByComma(input)
	}
}

func BenchmarkUnquote_Double(b *testing.B) {
	input := "\"hello world with quotes\""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unquote(input)
	}
}

func BenchmarkUnquote_Single(b *testing.B) {
	input := "'hello world with quotes'"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unquote(input)
	}
}

func BenchmarkUnquote_NoQuotes(b *testing.B) {
	input := "hello world without quotes"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = unquote(input)
	}
}

func BenchmarkCommand_String_Simple(b *testing.B) {
	cmd := &Command{
		Name:    "send",
		Message: "hello world",
		Options: map[string]string{"flag": ""},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.String()
	}
}

func BenchmarkCommand_String_Complex(b *testing.B) {
	cmd := &Command{
		Name:    "bash",
		Message: "execute complex command",
		Options: map[string]string{
			"timeout": "30",
			"verbose": "",
			"cmd":     "ls -la | grep test",
			"force":   "",
			"file":    "test.txt",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.String()
	}
}

// Memory allocation benchmarks
func BenchmarkParseInput_MemAlloc(b *testing.B) {
	input := "\\set[var=value, flag] hello world"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseInput(input)
	}
}

func BenchmarkSplitByComma_MemAlloc(b *testing.B) {
	input := "item1,item2,item3,item4,item5"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitByComma(input)
	}
}
