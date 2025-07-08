package services

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"neuroshell/internal/context"
)

func TestMarkdownService_Name(t *testing.T) {
	service := NewMarkdownService()
	assert.Equal(t, "markdown", service.Name())
}

func TestMarkdownService_Initialize(t *testing.T) {
	service := NewMarkdownService()
	assert.False(t, service.initialized)

	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
	assert.NotNil(t, service.renderer)
}

func TestMarkdownService_Render(t *testing.T) {
	service := NewMarkdownService()

	// Test uninitialized service
	_, err := service.Render("# Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize service
	err = service.Initialize()
	require.NoError(t, err)

	// Test empty markdown
	_, err = service.Render("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	_, err = service.Render("   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Test valid markdown
	result, err := service.Render("# Hello World")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// The result should contain ANSI escape sequences for formatting
	assert.True(t, containsText(result, "Hello World"), "Result should contain 'Hello World' text")
}

func TestMarkdownService_RenderWithStyle(t *testing.T) {
	service := NewMarkdownService()

	// Test uninitialized service
	_, err := service.RenderWithStyle("# Test", "dark")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize service
	err = service.Initialize()
	require.NoError(t, err)

	// Test empty markdown
	_, err = service.RenderWithStyle("", "dark")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Test valid markdown with different styles
	testCases := []struct {
		name     string
		markdown string
		style    string
	}{
		{"dark style", "# Hello World", "dark"},
		{"light style", "# Hello World", "light"},
		{"auto style", "# Hello World", "auto"},
		{"notty style", "# Hello World", "notty"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.RenderWithStyle(tc.markdown, tc.style)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.True(t, containsText(result, "Hello World"), "Result should contain 'Hello World' text")
		})
	}

	// Test invalid style (should fall back to default)
	result, err := service.RenderWithStyle("# Hello World", "invalid-style")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.True(t, containsText(result, "Hello World"), "Result should contain 'Hello World' text")
}

func TestMarkdownService_RenderWithTheme(t *testing.T) {
	service := NewMarkdownService()

	// Test uninitialized service
	_, err := service.RenderWithTheme("# Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize service
	err = service.Initialize()
	require.NoError(t, err)

	// Set up test environment
	setupMarkdownTestRegistry(t)

	// Test with theme integration
	result, err := service.RenderWithTheme("# Hello World")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.True(t, containsText(result, "Hello World"), "Result should contain 'Hello World' text")
}

func TestMarkdownService_SetWordWrap(t *testing.T) {
	service := NewMarkdownService()

	// Test uninitialized service
	err := service.SetWordWrap(80)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	// Initialize service
	err = service.Initialize()
	require.NoError(t, err)

	// Test invalid word wrap
	err = service.SetWordWrap(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")

	err = service.SetWordWrap(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be positive")

	// Test valid word wrap
	err = service.SetWordWrap(100)
	assert.NoError(t, err)

	// Test that rendering still works after changing word wrap
	result, err := service.Render("# Hello World")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.True(t, containsText(result, "Hello World"), "Result should contain 'Hello World' text")
}

func TestMarkdownService_MapThemeToGlamourStyle(t *testing.T) {
	service := NewMarkdownService()

	testCases := []struct {
		themeName     string
		expectedStyle string
	}{
		{"dark", "dark"},
		{"dark1", "dark"},
		{"light", "light"},
		{"plain", "notty"},
		{"default", "auto"},
		{"unknown", "auto"},
		{"", "auto"},
	}

	for _, tc := range testCases {
		t.Run(tc.themeName, func(t *testing.T) {
			result := service.mapThemeToGlamourStyle(tc.themeName)
			assert.Equal(t, tc.expectedStyle, result)
		})
	}
}

func TestMarkdownService_GetAvailableStyles(t *testing.T) {
	service := NewMarkdownService()
	styles := service.GetAvailableStyles()

	assert.NotEmpty(t, styles)
	assert.Contains(t, styles, "auto")
	assert.Contains(t, styles, "dark")
	assert.Contains(t, styles, "light")
	assert.Contains(t, styles, "notty")
	assert.Contains(t, styles, "ascii")
}

func TestMarkdownService_GetServiceInfo(t *testing.T) {
	service := NewMarkdownService()

	info := service.GetServiceInfo()

	assert.Equal(t, "markdown", info["name"])
	assert.Equal(t, false, info["initialized"])
	assert.NotNil(t, info["styles"])
	assert.NotNil(t, info["description"])

	// Test after initialization
	err := service.Initialize()
	require.NoError(t, err)

	info = service.GetServiceInfo()
	assert.Equal(t, true, info["initialized"])
}

func TestMarkdownService_GetCurrentTheme(t *testing.T) {
	service := NewMarkdownService()

	// Test with no variable service
	theme := service.getCurrentTheme()
	assert.Equal(t, "default", theme)

	// Set up test environment
	setupMarkdownTestRegistry(t)

	// Test with variable service but no _style variable
	theme = service.getCurrentTheme()
	assert.Equal(t, "default", theme)

	// Set a theme variable
	ctx := context.New()
	context.SetGlobalContext(ctx)

	err := ctx.SetSystemVariable("_style", "dark")
	require.NoError(t, err)

	theme = service.getCurrentTheme()
	assert.Equal(t, "dark", theme)
}

func TestMarkdownService_CleanShellMarkers(t *testing.T) {
	service := NewMarkdownService()
	
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple continuation", "line1\\n... line2", "line1\\nline2"},
		{"marker only line", "line1\\n...\\nline2", "line1\\nline2"},
		{"multiple markers", "line1\\n... line2\\n... line3", "line1\\nline2\\nline3"},
		{"no markers", "line1\\nline2", "line1\\nline2"},
		{"mixed content", "# Title\\n... This is content\\n...\\n## Section", "# Title\\nThis is content\\n## Section"},
		{"actual newlines with markers", "line1\n... line2\nline3", "line1\nline2\nline3"},
		{"marker with spaces", "line1\\n...    line2   ", "line1\\nline2"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.cleanShellMarkers(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMarkdownService_ProcessEscapeSequences(t *testing.T) {
	service := NewMarkdownService()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"newline escape", "Hello\\nWorld", "Hello\nWorld"},
		{"tab escape", "Hello\\tWorld", "Hello\tWorld"},
		{"carriage return escape", "Hello\\rWorld", "Hello\rWorld"},
		{"escaped backslash", "Hello\\\\World", "Hello\\World"},
		{"multiple escapes", "Line1\\nLine2\\tTabbed", "Line1\nLine2\tTabbed"},
		{"no escapes", "Hello World", "Hello World"},
		{"mixed content", "# Title\\n\\nThis is **bold** text", "# Title\n\nThis is **bold** text"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := service.processEscapeSequences(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMarkdownService_RenderWithEscapeSequences(t *testing.T) {
	service := NewMarkdownService()
	err := service.Initialize()
	require.NoError(t, err)

	// Test that escape sequences are processed before rendering
	markdown := "# Hello World\\n\\nThis is **bold** text"
	result, err := service.Render(markdown)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// The result should contain both "Hello World" and "bold" on separate rendered lines
	assert.True(t, containsText(result, "Hello World"), "Result should contain 'Hello World' text")
	assert.True(t, containsText(result, "bold"), "Result should contain 'bold' text")
}

func TestMarkdownService_RenderWithShellMarkersAndEscapeSequences(t *testing.T) {
	service := NewMarkdownService()
	err := service.Initialize()
	require.NoError(t, err)
	
	// Test input similar to what user reported: shell continuation markers + escape sequences
	markdown := "# dwdwdwd \\n dwdwdwd `dddd` \\n\\n...\\n... ## this is th"
	result, err := service.Render(markdown)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	
	// The result should contain the content without shell markers
	assert.True(t, containsText(result, "dwdwdwd"), "Result should contain 'dwdwdwd' text")
	assert.True(t, containsText(result, "dddd"), "Result should contain 'dddd' text")
	assert.True(t, containsText(result, "## this is th"), "Result should contain '## this is th' text")
	
	// The result should NOT contain shell continuation markers
	assert.False(t, containsText(result, "..."), "Result should not contain '...' continuation markers")
}

func TestMarkdownService_ComplexMarkdown(t *testing.T) {
	service := NewMarkdownService()
	err := service.Initialize()
	require.NoError(t, err)

	complexMarkdown := `# Main Title

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
> with multiple lines

| Column 1 | Column 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |
`

	result, err := service.Render(complexMarkdown)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// Check that various markdown elements are present
	assert.True(t, containsText(result, "Main Title"), "Result should contain 'Main Title' text")
	assert.True(t, containsText(result, "Features"), "Result should contain 'Features' text")
	assert.True(t, containsText(result, "example.com"), "Result should contain 'example.com' text")
	assert.True(t, containsText(result, "fmt.Println"), "Result should contain 'fmt.Println' text")
	assert.True(t, containsText(result, "blockquote"), "Result should contain 'blockquote' text")
	assert.True(t, containsText(result, "Column 1"), "Result should contain 'Column 1' text")
}

// containsText checks if the given text is present in the result, stripping ANSI escape sequences
func containsText(result, text string) bool {
	// Regular expression to match ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanResult := ansiRegex.ReplaceAllString(result, "")
	return strings.Contains(cleanResult, text)
}

// setupMarkdownTestRegistry creates a clean test registry for markdown tests
func setupMarkdownTestRegistry(t *testing.T) {
	// Create a new service registry for testing
	oldServiceRegistry := GetGlobalRegistry()
	SetGlobalRegistry(NewRegistry())

	// Create a test context
	ctx := context.New()
	context.SetGlobalContext(ctx)

	// Register variable service
	err := GetGlobalRegistry().RegisterService(NewVariableService())
	require.NoError(t, err)

	// Initialize services
	err = GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		SetGlobalRegistry(oldServiceRegistry)
		context.ResetGlobalContext()
	})
}
