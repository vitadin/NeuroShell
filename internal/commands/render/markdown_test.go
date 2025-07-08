package render

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

func TestMarkdownCommand_Name(t *testing.T) {
	cmd := &MarkdownCommand{}
	assert.Equal(t, "render-markdown", cmd.Name())
}

func TestMarkdownCommand_ParseMode(t *testing.T) {
	cmd := &MarkdownCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestMarkdownCommand_Description(t *testing.T) {
	cmd := &MarkdownCommand{}
	desc := cmd.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "markdown")
	assert.Contains(t, desc, "ANSI")
}

func TestMarkdownCommand_Usage(t *testing.T) {
	cmd := &MarkdownCommand{}
	usage := cmd.Usage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "\\render-markdown")
}

func TestMarkdownCommand_HelpInfo(t *testing.T) {
	cmd := &MarkdownCommand{}
	helpInfo := cmd.HelpInfo()

	assert.Equal(t, "render-markdown", helpInfo.Command)
	assert.NotEmpty(t, helpInfo.Description)
	assert.NotEmpty(t, helpInfo.Usage)
	assert.Equal(t, neurotypes.ParseModeKeyValue, helpInfo.ParseMode)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)

	// Check that examples contain valid markdown
	assert.True(t, len(helpInfo.Examples) > 0)
	for _, example := range helpInfo.Examples {
		assert.Contains(t, example.Command, "\\render-markdown")
		assert.NotEmpty(t, example.Description)
	}
}

func TestMarkdownCommand_Execute_EmptyInput(t *testing.T) {
	cmd := &MarkdownCommand{}

	// Test empty input
	err := cmd.Execute(map[string]string{}, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Usage:")
}

func TestMarkdownCommand_Execute_Success(t *testing.T) {
	cmd := &MarkdownCommand{}

	// Set up test environment
	setupMarkdownCommandTestRegistry(t)

	// Test with simple markdown
	args := map[string]string{}
	markdown := "# Hello World\n\nThis is **bold** text."

	err := cmd.Execute(args, markdown)
	assert.NoError(t, err)

	// Check that output was stored in _output variable
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.True(t, containsText(output, "Hello World"), "Output should contain 'Hello World' text")
}

func TestMarkdownCommand_Execute_ComplexMarkdown(t *testing.T) {
	cmd := &MarkdownCommand{}

	// Set up test environment
	setupMarkdownCommandTestRegistry(t)

	// Test with complex markdown
	args := map[string]string{}
	markdown := `# Main Title

## Features

- **Bold text** and *italic text*
- [Links](https://example.com)
- ` + "`inline code`" + `

### Code Block

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

> This is a blockquote

| Column 1 | Column 2 |
|----------|----------|
| Cell 1   | Cell 2   |
`

	err := cmd.Execute(args, markdown)
	assert.NoError(t, err)

	// Check that output was stored in _output variable
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that various markdown elements are present
	assert.True(t, containsText(output, "Main Title"), "Output should contain 'Main Title' text")
	assert.True(t, containsText(output, "Features"), "Output should contain 'Features' text")
	assert.True(t, containsText(output, "example.com"), "Output should contain 'example.com' text")
	assert.True(t, containsText(output, "fmt.Println"), "Output should contain 'fmt.Println' text")
	assert.True(t, containsText(output, "blockquote"), "Output should contain 'blockquote' text")
	assert.True(t, containsText(output, "Column 1"), "Output should contain 'Column 1' text")
}

func TestMarkdownCommand_Execute_WithTheme(t *testing.T) {
	cmd := &MarkdownCommand{}

	// Set up test environment
	setupMarkdownCommandTestRegistry(t)

	// Set a theme
	ctx := context.GetGlobalContext()
	neuroCtx := ctx.(*context.NeuroContext)
	err := neuroCtx.SetSystemVariable("_style", "dark")
	require.NoError(t, err)

	// Test with markdown
	args := map[string]string{}
	markdown := "# Hello World\n\nThis is **bold** text."

	err = cmd.Execute(args, markdown)
	assert.NoError(t, err)

	// Check that output was stored
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.True(t, containsText(output, "Hello World"), "Output should contain 'Hello World' text")
}

func TestMarkdownCommand_Execute_MarkdownServiceError(t *testing.T) {
	cmd := &MarkdownCommand{}

	// Set up test environment without markdown service
	setupMarkdownCommandTestRegistryWithoutMarkdown(t)

	// Test execution without markdown service
	args := map[string]string{}
	markdown := "# Hello World"

	err := cmd.Execute(args, markdown)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "markdown service not available")
}

func TestMarkdownCommand_Execute_VariableServiceError(t *testing.T) {
	cmd := &MarkdownCommand{}

	// Set up test environment without variable service
	setupMarkdownCommandTestRegistryWithoutVariable(t)

	// Test execution without variable service
	args := map[string]string{}
	markdown := "# Hello World"

	err := cmd.Execute(args, markdown)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable service not available")
}

func TestMarkdownCommand_Execute_WithEscapeSequences(t *testing.T) {
	cmd := &MarkdownCommand{}

	// Set up test environment
	setupMarkdownCommandTestRegistry(t)

	// Test with markdown containing escape sequences
	args := map[string]string{}
	markdown := "# Hello World\\n\\nThis is **bold** text\\nand some more content"

	err := cmd.Execute(args, markdown)
	assert.NoError(t, err)

	// Check that output was stored and escape sequences were processed
	variableService, err := services.GetGlobalVariableService()
	require.NoError(t, err)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Verify that the content was properly rendered with escape sequences processed
	assert.True(t, containsText(output, "Hello World"), "Output should contain 'Hello World' text")
	assert.True(t, containsText(output, "bold"), "Output should contain 'bold' text")
	assert.True(t, containsText(output, "some more content"), "Output should contain 'some more content' text")
}

func TestMarkdownCommand_Execute_InvalidMarkdown(t *testing.T) {
	cmd := &MarkdownCommand{}

	// Set up test environment
	setupMarkdownCommandTestRegistry(t)

	// Test with whitespace-only input
	args := map[string]string{}
	markdown := "   \n\t  "

	err := cmd.Execute(args, markdown)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to render markdown")
}

// containsText checks if the given text is present in the result, stripping ANSI escape sequences
func containsText(result, text string) bool {
	// Regular expression to match ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanResult := ansiRegex.ReplaceAllString(result, "")
	return strings.Contains(cleanResult, text)
}

// setupMarkdownCommandTestRegistry creates a clean test registry for markdown command tests
func setupMarkdownCommandTestRegistry(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Create a test context
	ctx := context.New()
	context.SetGlobalContext(ctx)

	// Register required services
	err := services.GetGlobalRegistry().RegisterService(services.NewMarkdownService())
	require.NoError(t, err)

	err = services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}

// setupMarkdownCommandTestRegistryWithoutMarkdown creates a test registry without markdown service
func setupMarkdownCommandTestRegistryWithoutMarkdown(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Create a test context
	ctx := context.New()
	context.SetGlobalContext(ctx)

	// Register only variable service (no markdown service)
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}

// setupMarkdownCommandTestRegistryWithoutVariable creates a test registry without variable service
func setupMarkdownCommandTestRegistryWithoutVariable(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Create a test context
	ctx := context.New()
	context.SetGlobalContext(ctx)

	// Register only markdown service (no variable service)
	err := services.GetGlobalRegistry().RegisterService(services.NewMarkdownService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}
