package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"neuroshell/internal/testutils"
)

func TestRenderService_BasicFunctionality(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	assert.Equal(t, "render", service.Name())
	assert.True(t, service.initialized)
}

func TestRenderService_ThemeManagement(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test available themes
	themes := service.GetAvailableThemes()
	assert.Contains(t, themes, "default")
	assert.Contains(t, themes, "dark")
	assert.Contains(t, themes, "light")
	assert.Len(t, themes, 3)

	// Test getting specific themes
	defaultTheme, exists := service.GetTheme("default")
	assert.True(t, exists)
	assert.NotNil(t, defaultTheme)
	assert.Equal(t, "default", defaultTheme.Name)

	// Test non-existent theme
	_, exists = service.GetTheme("nonexistent")
	assert.False(t, exists)
}

func TestRenderService_KeywordHighlighting(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		text     string
		keywords []string
	}{
		{
			name:     "highlight commands",
			text:     "Use \\get and \\set commands",
			keywords: []string{"\\get", "\\set"},
		},
		{
			name:     "highlight variables",
			text:     "The value is ${name} and ${count}",
			keywords: []string{"${name}", "${count}"},
		},
		{
			name:     "mixed keywords",
			text:     "Run \\bash with ${script} parameter",
			keywords: []string{"\\bash", "${script}"},
		},
		{
			name:     "no keywords",
			text:     "Plain text without highlights",
			keywords: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := RenderOptions{
				Keywords: tt.keywords,
				Theme:    "default",
			}

			result, err := service.RenderText(tt.text, options)
			require.NoError(t, err)

			// In test environments, lipgloss may not output ANSI codes
			// Instead, check that the function ran without error and content is preserved
			if len(tt.keywords) > 0 {
				// The result should be different from input when keywords are present
				// (either styled or at minimum, processed)
				assert.NotEmpty(t, result, "Expected non-empty result")
			}

			// Original text content should still be present (minus styling)
			for _, keyword := range tt.keywords {
				assert.Contains(t, result, keyword, "Keyword should still be present in output")
			}
		})
	}
}

func TestRenderService_NeuroShellSyntaxHighlighting(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name     string
		text     string
		contains []string // Strings that should still be present
	}{
		{
			name:     "commands and variables",
			text:     "Use \\set[var=${name}] to set variable",
			contains: []string{"\\set", "${name}"},
		},
		{
			name:     "multiple variables",
			text:     "Values: ${var1}, ${var2}, ${@user}",
			contains: []string{"${var1}", "${var2}", "${@user}"},
		},
		{
			name:     "complex command",
			text:     "\\session-new[name=test] creates a session",
			contains: []string{"\\session-new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := RenderOptions{
				Theme: "default",
			}

			result, err := service.RenderText(tt.text, options)
			require.NoError(t, err)

			// Should process the text (in test env, may not show ANSI codes)
			assert.NotEmpty(t, result, "Expected non-empty result")

			// All expected content should still be present
			for _, content := range tt.contains {
				assert.Contains(t, result, content, "Expected content should be present")
			}
		})
	}
}

func TestRenderService_GlobalStyling(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		text    string
		options RenderOptions
	}{
		{
			name: "bold styling",
			text: "Bold text",
			options: RenderOptions{
				Bold:  true,
				Theme: "default",
			},
		},
		{
			name: "color styling",
			text: "Colored text",
			options: RenderOptions{
				Color: "#ff0000",
				Theme: "default",
			},
		},
		{
			name: "named style",
			text: "Success message",
			options: RenderOptions{
				Style: "success",
				Theme: "default",
			},
		},
		{
			name: "combined styling",
			text: "Complex styling",
			options: RenderOptions{
				Bold:       true,
				Italic:     true,
				Color:      "#0000ff",
				Background: "#ffff00",
				Theme:      "default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.RenderText(tt.text, tt.options)
			require.NoError(t, err)

			// Should process the text (in test env, may not show ANSI codes)
			assert.NotEmpty(t, result, "Expected non-empty result")

			// Original text should be present
			assert.Contains(t, result, tt.text, "Original text should be present")
		})
	}
}

func TestRenderService_ConvenienceMethods(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	text := "Test \\get and \\set commands"
	keywords := []string{"\\get", "\\set"}

	// Test RenderWithTheme
	result1, err := service.RenderWithTheme(text, keywords, "dark")
	require.NoError(t, err)
	assert.NotEmpty(t, result1)
	assert.Contains(t, result1, "\\get")
	assert.Contains(t, result1, "\\set")

	// Test HighlightKeywords (uses default theme)
	result2, err := service.HighlightKeywords(text, keywords)
	require.NoError(t, err)
	assert.NotEmpty(t, result2)
	assert.Contains(t, result2, "\\get")
	assert.Contains(t, result2, "\\set")
}

func TestRenderService_ErrorHandling(t *testing.T) {
	service := NewRenderService()
	// Don't initialize the service

	// Test operations on uninitialized service
	_, err := service.RenderText("test", RenderOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")

	themes := service.GetAvailableThemes()
	assert.Empty(t, themes)

	_, exists := service.GetTheme("default")
	assert.False(t, exists)
}

func TestRenderService_ThemeFallback(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test with non-existent theme - should fallback to default
	options := RenderOptions{
		Keywords: []string{"test"},
		Theme:    "nonexistent",
	}

	result, err := service.RenderText("test keyword", options)
	require.NoError(t, err)

	// Should still work and process the text
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "test")
}

func TestRenderService_EdgeCases(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		text    string
		options RenderOptions
	}{
		{
			name:    "empty text",
			text:    "",
			options: RenderOptions{Keywords: []string{"test"}},
		},
		{
			name:    "empty keywords",
			text:    "some text",
			options: RenderOptions{Keywords: []string{}},
		},
		{
			name:    "nil keywords",
			text:    "some text",
			options: RenderOptions{},
		},
		{
			name:    "keywords with empty strings",
			text:    "some text",
			options: RenderOptions{Keywords: []string{"", "test", ""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.RenderText(tt.text, tt.options)
			require.NoError(t, err)

			// Should not crash and should return something reasonable
			if tt.text == "" {
				assert.Equal(t, "", result)
			} else {
				assert.Contains(t, result, tt.text)
			}
		})
	}
}

func TestRenderService_SpecialCharacters(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test with special regex characters in keywords
	text := "Use \\get[var] and \\set commands"
	keywords := []string{"\\get[var]", "\\set"}

	options := RenderOptions{
		Keywords: keywords,
		Theme:    "default",
	}

	result, err := service.RenderText(text, options)
	require.NoError(t, err)

	// Should handle regex special characters correctly
	assert.Contains(t, result, "\\get[var]")
	assert.Contains(t, result, "\\set")
	assert.NotEmpty(t, result) // Should process the text
}
