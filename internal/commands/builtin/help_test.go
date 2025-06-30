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
	"neuroshell/internal/testutils"
	"neuroshell/pkg/neurotypes"
)

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
	assert.Equal(t, "\\help [command]", cmd.Usage())
}

func TestHelpCommand_Execute(t *testing.T) {
	// Create a separate registry for testing to avoid polluting global state
	testRegistry := commands.NewRegistry()

	// Register some test commands
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

	for _, cmd := range testCommands {
		err := testRegistry.Register(cmd)
		require.NoError(t, err)
	}

	// Temporarily replace global registry
	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

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
	// Create a separate registry for testing
	testRegistry := commands.NewRegistry()

	// Register commands in non-alphabetical order
	testCommands := []neurotypes.Command{
		&MockCommand{name: "zebra", description: "Last", usage: "\\zebra"},
		&MockCommand{name: "apple", description: "First", usage: "\\apple"},
		&MockCommand{name: "banana", description: "Middle", usage: "\\banana"},
	}

	for _, cmd := range testCommands {
		err := testRegistry.Register(cmd)
		require.NoError(t, err)
	}

	// Temporarily replace global registry
	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

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
	// Create an empty registry for testing
	testRegistry := commands.NewRegistry()

	// Temporarily replace global registry
	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

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
	// Test that help command ignores arguments (current implementation doesn't use them)
	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

	// Create minimal registry
	testRegistry := commands.NewRegistry()
	testRegistry.Register(&MockCommand{name: "test", description: "Test", usage: "\\test"})

	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

	// Test with args
	args := map[string]string{"command": "test"}

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
	// Should still show all commands (current implementation doesn't filter by specific command)
	assert.Contains(t, outputStr, "Neuro Shell Commands:")
}

func TestHelpCommand_Execute_WithInput(t *testing.T) {
	// Test that help command ignores input (current implementation doesn't use it)
	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

	// Create minimal registry
	testRegistry := commands.NewRegistry()
	testRegistry.Register(&MockCommand{name: "test", description: "Test", usage: "\\test"})

	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

	// Test with input
	input := "some input text"

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
	// Should still show all commands (current implementation doesn't use input)
	assert.Contains(t, outputStr, "Neuro Shell Commands:")
}

func TestHelpCommand_Execute_FormatConsistency(t *testing.T) {
	// Test output formatting consistency
	testRegistry := commands.NewRegistry()

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

	for _, cmd := range testCommands {
		err := testRegistry.Register(cmd)
		require.NoError(t, err)
	}

	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

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
	cmd := &HelpCommand{}
	ctx := testutils.NewMockContext()

	// Use empty registry
	testRegistry := commands.NewRegistry()
	originalRegistry := commands.GlobalRegistry
	commands.GlobalRegistry = testRegistry
	defer func() {
		commands.GlobalRegistry = originalRegistry
	}()

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
		testRegistry.Register(cmd)
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
		testRegistry.Register(cmd)
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
