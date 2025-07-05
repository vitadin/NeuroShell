package builtin

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

// setupHelpTestEnvironment sets up both command and service registries for testing help functionality
func setupHelpTestEnvironment(t *testing.T, testCommands []neurotypes.Command) neurotypes.Context {
	// Create separate registries for testing
	testCommandRegistry := commands.NewRegistry()
	testServiceRegistry := services.NewRegistry()

	// Register test commands
	for _, cmd := range testCommands {
		err := testCommandRegistry.Register(cmd)
		require.NoError(t, err)
	}

	// Temporarily replace global registries using thread-safe functions
	originalCommandRegistry := commands.GetGlobalRegistry()
	originalServiceRegistry := services.GetGlobalRegistry()
	commands.SetGlobalRegistry(testCommandRegistry)
	services.SetGlobalRegistry(testServiceRegistry)

	// Cleanup function
	t.Cleanup(func() {
		commands.SetGlobalRegistry(originalCommandRegistry)
		services.SetGlobalRegistry(originalServiceRegistry)
	})

	// Create and initialize help service
	helpService := services.NewHelpService()
	err := testServiceRegistry.RegisterService(helpService)
	require.NoError(t, err)

	// Create and initialize render service (required by new help command)
	renderService := services.NewRenderService()
	err = testServiceRegistry.RegisterService(renderService)
	require.NoError(t, err)

	// Create context and initialize services
	ctx := testutils.NewMockContext()
	err = helpService.Initialize(ctx)
	require.NoError(t, err)
	err = renderService.Initialize(ctx)
	require.NoError(t, err)

	return ctx
}

func TestHelpCommand_Name(t *testing.T) {
	cmd := &HelpCommand{}
	assert.Equal(t, "help", cmd.Name())
}

func TestHelpCommand_ParseMode(t *testing.T) {
	cmd := &HelpCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestHelpCommand_Description(t *testing.T) {
	cmd := &HelpCommand{}
	assert.Equal(t, "Show command help", cmd.Description())
}

func TestHelpCommand_Usage(t *testing.T) {
	cmd := &HelpCommand{}
	assert.Equal(t, "\\help[styled=true] [command]", cmd.Usage())
}

func TestHelpCommand_Execute(t *testing.T) {
	// Set up test environment with help service
	testCommands := []neurotypes.Command{
		&MockCommand{
			name:        "test1",
			description: "Test command 1",
			usage:       "\\test1",
		},
		&MockCommand{
			name:        "test2",
			description: "Test command 2",
			usage:       "\\test2 [arg]",
		},
		&MockCommand{
			name:        "aaa",
			description: "First alphabetically",
			usage:       "\\aaa",
		},
	}

	ctx := setupHelpTestEnvironment(t, testCommands)
	cmd := &HelpCommand{}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(map[string]string{}, "", ctx)

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)

	// Verify output contains expected elements
	assert.Contains(t, outputStr, "Neuro Shell Commands:")
	assert.Contains(t, outputStr, "Examples:")
	assert.Contains(t, outputStr, "Note: Text without \\ prefix is sent to LLM automatically")

	// Verify all test commands are listed
	assert.Contains(t, outputStr, "\\test1")
	assert.Contains(t, outputStr, "Test command 1")
	assert.Contains(t, outputStr, "\\test2 [arg]")
	assert.Contains(t, outputStr, "Test command 2")
	assert.Contains(t, outputStr, "\\aaa")
	assert.Contains(t, outputStr, "First alphabetically")

	// Verify example commands are shown
	assert.Contains(t, outputStr, "\\send Hello world")
	assert.Contains(t, outputStr, "\\set[name=\"John\"]")
	assert.Contains(t, outputStr, "\\get[name]")
	assert.Contains(t, outputStr, "\\bash[ls -la]")
}

func TestHelpCommand_Execute_AlphabeticalOrder(t *testing.T) {
	// Register commands in non-alphabetical order
	testCommands := []neurotypes.Command{
		&MockCommand{name: "zebra", description: "Last", usage: "\\zebra"},
		&MockCommand{name: "apple", description: "First", usage: "\\apple"},
		&MockCommand{name: "banana", description: "Middle", usage: "\\banana"},
	}

	ctx := setupHelpTestEnvironment(t, testCommands)
	cmd := &HelpCommand{}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(map[string]string{}, "", ctx)

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)

	// Find positions of commands in output
	applePos := strings.Index(outputStr, "\\apple")
	bananaPos := strings.Index(outputStr, "\\banana")
	zebraPos := strings.Index(outputStr, "\\zebra")

	// Verify they appear in alphabetical order
	assert.True(t, applePos < bananaPos, "apple should appear before banana")
	assert.True(t, bananaPos < zebraPos, "banana should appear before zebra")
}

func TestHelpCommand_Execute_EmptyRegistry(t *testing.T) {
	// Use empty command list
	testCommands := []neurotypes.Command{}

	ctx := setupHelpTestEnvironment(t, testCommands)
	cmd := &HelpCommand{}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(map[string]string{}, "", ctx)

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)

	// Should still show header and examples even with no commands
	assert.Contains(t, outputStr, "Neuro Shell Commands:")
	assert.Contains(t, outputStr, "Examples:")
	assert.Contains(t, outputStr, "Note: Text without \\ prefix is sent to LLM automatically")
}

func TestHelpCommand_Execute_WithArgs(t *testing.T) {
	// Test help command with specific command requested
	testCommands := []neurotypes.Command{
		&MockCommand{name: "test", description: "Test", usage: "\\test"},
	}

	ctx := setupHelpTestEnvironment(t, testCommands)
	cmd := &HelpCommand{}

	// Test with args - request help for specific command
	args := map[string]string{"test": ""}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(args, "", ctx)

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)
	// Should show specific command help
	assert.Contains(t, outputStr, "Command: test")
	assert.Contains(t, outputStr, "Description: Test")
	assert.Contains(t, outputStr, "Usage: \\test")
}

func TestHelpCommand_Execute_WithInput(t *testing.T) {
	// Test that help command uses input to show specific command help
	testCommands := []neurotypes.Command{
		&MockCommand{name: "test", description: "Test", usage: "\\test"},
	}

	ctx := setupHelpTestEnvironment(t, testCommands)
	cmd := &HelpCommand{}

	// Test with valid command name in input
	input := "test"

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(map[string]string{}, input, ctx)

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)
	// Should show specific command help
	assert.Contains(t, outputStr, "Command: test")
	assert.Contains(t, outputStr, "Description: Test")
}

func TestHelpCommand_Execute_FormatConsistency(t *testing.T) {
	// Test output formatting consistency
	// Register commands with various length names and descriptions
	testCommands := []neurotypes.Command{
		&MockCommand{
			name:        "short",
			description: "Short description",
			usage:       "\\short",
		},
		&MockCommand{
			name:        "verylongcommandname",
			description: "This is a very long description that tests formatting",
			usage:       "\\verylongcommandname [arg1] [arg2]",
		},
		&MockCommand{
			name:        "mid",
			description: "Medium length description",
			usage:       "\\mid [optional]",
		},
	}

	ctx := setupHelpTestEnvironment(t, testCommands)
	cmd := &HelpCommand{}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(map[string]string{}, "", ctx)

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)

	// Verify all commands are present
	assert.Contains(t, outputStr, "\\short")
	assert.Contains(t, outputStr, "Short description")
	assert.Contains(t, outputStr, "\\verylongcommandname [arg1] [arg2]")
	assert.Contains(t, outputStr, "This is a very long description that tests formatting")
	assert.Contains(t, outputStr, "\\mid [optional]")
	assert.Contains(t, outputStr, "Medium length description")
}

func TestHelpCommand_Execute_StaticContent(t *testing.T) {
	// Test that static content (examples, notes) is always present
	// Use empty registry
	testCommands := []neurotypes.Command{}

	ctx := setupHelpTestEnvironment(t, testCommands)
	cmd := &HelpCommand{}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Execute(map[string]string{}, "", ctx)

	// Restore stdout
	_ = w.Close()
	os.Stdout = originalStdout

	// Read captured output
	output, _ := io.ReadAll(r)
	outputStr := string(output)

	assert.NoError(t, err)

	// Verify static content is present
	expectedStaticContent := []string{
		"Neuro Shell Commands:",
		"Examples:",
		"\\send Hello world",
		"\\set[name=\"John\"]",
		"\\get[name]",
		"\\bash[ls -la]",
		"Note: Text without \\ prefix is sent to LLM automatically",
	}

	for _, content := range expectedStaticContent {
		assert.Contains(t, outputStr, content, "Missing static content: %s", content)
	}
}

func TestHelpCommand_Execute_SpecificCommand(t *testing.T) {
	// Test help for a specific command using \help[command] syntax
	testCommands := []neurotypes.Command{
		&MockCommand{
			name:        "bash",
			description: "Execute system commands via bash",
			usage:       "\\bash command_to_execute",
			parseMode:   neurotypes.ParseModeRaw,
		},
		&MockCommand{
			name:        "set",
			description: "Set a variable",
			usage:       "\\set[var=value] or \\set var value",
			parseMode:   neurotypes.ParseModeKeyValue,
		},
	}

	ctx := setupHelpTestEnvironment(t, testCommands)
	cmd := &HelpCommand{}

	tests := []struct {
		name         string
		args         map[string]string
		shouldError  bool
		expectedText []string
	}{
		{
			name: "help for bash command",
			args: map[string]string{"bash": ""},
			expectedText: []string{
				"Command: bash",
				"Description: Execute system commands via bash",
				"Usage: \\bash command_to_execute",
				"Parse Mode: Raw",
				"Examples:",
				"\\bash command_to_execute",
				"%% Basic usage example",
			},
		},
		{
			name: "help for set command",
			args: map[string]string{"set": ""},
			expectedText: []string{
				"Command: set",
				"Description: Set a variable",
				"Usage: \\set[var=value] or \\set var value",
				"Parse Mode: Key-Value",
				"Examples:",
				"\\set[var=value] or \\set var value",
				"%% Basic usage example",
			},
		},
		{
			name:        "help for nonexistent command",
			args:        map[string]string{"nonexistent": ""},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmd.Execute(tt.args, "", ctx)

			// Restore stdout
			_ = w.Close()
			os.Stdout = originalStdout

			// Read captured output
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				assert.NoError(t, err)
				for _, expectedText := range tt.expectedText {
					assert.Contains(t, outputStr, expectedText, "Missing expected text: %s", expectedText)
				}
			}
		})
	}
}

func TestHelpCommand_Execute_ServiceUnavailable(t *testing.T) {
	// Test when help service is not available
	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

	// Don't set up help service - this will cause service not found error
	err := cmd.Execute(map[string]string{}, "", ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "help service not available")
}

// NOTE: TestHelpCommand_ShowCommandExamples was removed because showCommandExamples method
// was deprecated in favor of the new HelpInfo-based approach with RenderService integration
/*
func TestHelpCommand_ShowCommandExamples(t *testing.T) {
	// Test the showCommandExamples function with different command types
	cmd := &HelpCommand{}

	tests := []struct {
		name         string
		cmdInfo      services.CommandInfo
		expectedText []string
	}{
		{
			name: "KeyValue parse mode command",
			cmdInfo: services.CommandInfo{
				Name:        "set",
				Usage:       "\\set[var=value] or \\set var value",
				ParseMode:   neurotypes.ParseModeKeyValue,
				Description: "Set a variable",
			},
			expectedText: []string{
				"Examples:",
				"\\set[var=value] or \\set var value",
				"\\set[option=value]",
			},
		},
		{
			name: "Raw parse mode command",
			cmdInfo: services.CommandInfo{
				Name:        "bash",
				Usage:       "\\bash command_to_execute",
				ParseMode:   neurotypes.ParseModeRaw,
				Description: "Execute system commands via bash",
			},
			expectedText: []string{
				"Examples:",
				"\\bash command_to_execute",
			},
		},
		{
			name: "WithOptions parse mode command",
			cmdInfo: services.CommandInfo{
				Name:        "test",
				Usage:       "\\test [options] message",
				ParseMode:   neurotypes.ParseModeWithOptions,
				Description: "Test command",
			},
			expectedText: []string{
				"Examples:",
				"\\test [options] message",
				"\\test[option=value]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmd.showCommandExamples(tt.cmdInfo)

			// Restore stdout
			_ = w.Close()
			os.Stdout = originalStdout

			// Read captured output
			output, _ := io.ReadAll(r)
			outputStr := string(output)

			// Verify all expected text is present
			for _, expectedText := range tt.expectedText {
				assert.Contains(t, outputStr, expectedText, "Missing expected text: %s", expectedText)
			}

			// Verify the primary usage is always shown
			assert.Contains(t, outputStr, tt.cmdInfo.Usage)

			// Verify KeyValue and WithOptions modes get generic parameter example
			if tt.cmdInfo.ParseMode == neurotypes.ParseModeKeyValue || tt.cmdInfo.ParseMode == neurotypes.ParseModeWithOptions {
				expectedGeneric := fmt.Sprintf("\\%s[option=value]", tt.cmdInfo.Name)
				assert.Contains(t, outputStr, expectedGeneric)
			}
		})
	}
}
*/

// MockCommand for testing (reuse from registry_test.go structure)
type MockCommand struct {
	name        string
	parseMode   neurotypes.ParseMode
	description string
	usage       string
	executeFunc func(args map[string]string, input string, ctx neurotypes.Context) error
}

func (m *MockCommand) Name() string {
	return m.name
}

func (m *MockCommand) ParseMode() neurotypes.ParseMode {
	if m.parseMode == 0 {
		return neurotypes.ParseModeKeyValue
	}
	return m.parseMode
}

func (m *MockCommand) Description() string {
	return m.description
}

func (m *MockCommand) Usage() string {
	return m.usage
}

func (m *MockCommand) Execute(args map[string]string, input string, ctx neurotypes.Context) error {
	if m.executeFunc != nil {
		return m.executeFunc(args, input, ctx)
	}
	return nil
}

func (m *MockCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     m.Name(),
		Description: m.Description(),
		Usage:       m.Usage(),
		ParseMode:   m.ParseMode(),
		Examples: []neurotypes.HelpExample{
			{
				Command:     m.Usage(),
				Description: "Basic usage example",
			},
		},
	}
}

// Benchmark tests
func BenchmarkHelpCommand_Execute_SmallRegistry(b *testing.B) {
	// Create test registry with few commands
	testRegistry := commands.NewRegistry()
	for i := 0; i < 5; i++ {
		cmd := &MockCommand{
			name:        fmt.Sprintf("cmd%d", i),
			description: fmt.Sprintf("Command %d", i),
			usage:       fmt.Sprintf("\\cmd%d", i),
		}
		if err := testRegistry.Register(cmd); err != nil {
			b.Fatal(err)
		}
	}

	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = originalStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(map[string]string{}, "", ctx)
	}
}

func BenchmarkHelpCommand_Execute_LargeRegistry(b *testing.B) {
	// Create test registry with many commands
	testRegistry := commands.NewRegistry()
	for i := 0; i < 100; i++ {
		cmd := &MockCommand{
			name:        fmt.Sprintf("command%d", i),
			description: fmt.Sprintf("Description for command %d with some details", i),
			usage:       fmt.Sprintf("\\command%d [arg1] [arg2]", i),
		}
		if err := testRegistry.Register(cmd); err != nil {
			b.Fatal(err)
		}
	}

	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

	// Redirect stdout to avoid benchmark noise
	originalStdout := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = originalStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.Execute(map[string]string{}, "", ctx)
	}
}
