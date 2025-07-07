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

	// Test available themes (should include plain theme now)
	themes := service.GetAvailableThemes()
	assert.Contains(t, themes, "default")
	assert.Contains(t, themes, "dark")
	assert.Contains(t, themes, "light")
	assert.Contains(t, themes, "plain")
	assert.Len(t, themes, 4)

	// Test getting specific themes
	defaultTheme, exists := service.GetTheme("default")
	assert.True(t, exists)
	assert.NotNil(t, defaultTheme)
	assert.Equal(t, "default", defaultTheme.Name)

	// Test plain theme
	plainTheme, exists := service.GetTheme("plain")
	assert.True(t, exists)
	assert.NotNil(t, plainTheme)
	assert.Equal(t, "plain", plainTheme.Name)

	// Test non-existent theme
	_, exists = service.GetTheme("nonexistent")
	assert.False(t, exists)
}

func TestRenderService_GetThemeByName(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	tests := []struct {
		name         string
		input        string
		expectedName string
		description  string
	}{
		{
			name:         "empty string returns plain theme",
			input:        "",
			expectedName: "plain",
			description:  "Empty theme name should return plain theme",
		},
		{
			name:         "plain returns plain theme",
			input:        "plain",
			expectedName: "plain",
			description:  "Explicit plain theme should return plain theme",
		},
		{
			name:         "dark1 returns dark theme (alias)",
			input:        "dark1",
			expectedName: "dark",
			description:  "dark1 alias should return dark theme",
		},
		{
			name:         "DARK1 returns dark theme (case insensitive)",
			input:        "DARK1",
			expectedName: "dark",
			description:  "DARK1 should be case insensitive",
		},
		{
			name:         "dark returns dark theme",
			input:        "dark",
			expectedName: "dark",
			description:  "dark theme should return dark theme",
		},
		{
			name:         "default returns default theme",
			input:        "default",
			expectedName: "default",
			description:  "default theme should return default theme",
		},
		{
			name:         "light returns light theme",
			input:        "light",
			expectedName: "light",
			description:  "light theme should return light theme",
		},
		{
			name:         "invalid theme returns plain theme",
			input:        "invalid_theme",
			expectedName: "plain",
			description:  "Invalid theme should return plain theme as fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := service.GetThemeByName(tt.input)
			assert.NotNil(t, theme, "GetThemeByName should never return nil")
			assert.Equal(t, tt.expectedName, theme.Name, tt.description)
		})
	}
}

func TestRenderService_GetDefaultTheme(t *testing.T) {
	service := NewRenderService()

	// Test without initialization
	theme := service.GetDefaultTheme()
	assert.NotNil(t, theme)
	assert.Equal(t, "plain", theme.Name)

	// Test with initialization
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	theme = service.GetDefaultTheme()
	assert.NotNil(t, theme)
	assert.Equal(t, "plain", theme.Name)
}

func TestRenderService_ThemeStyles(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test that all themes have the required style fields
	themes := []string{"default", "dark", "light", "plain"}

	for _, themeName := range themes {
		t.Run(themeName, func(t *testing.T) {
			theme := service.GetThemeByName(themeName)
			assert.NotNil(t, theme)
			assert.Equal(t, themeName, theme.Name)

			// Verify all required style fields exist (should not be nil)
			assert.NotNil(t, theme.Keyword, "Keyword style should not be nil")
			assert.NotNil(t, theme.Variable, "Variable style should not be nil")
			assert.NotNil(t, theme.Command, "Command style should not be nil")
			assert.NotNil(t, theme.Success, "Success style should not be nil")
			assert.NotNil(t, theme.Error, "Error style should not be nil")
			assert.NotNil(t, theme.Warning, "Warning style should not be nil")
			assert.NotNil(t, theme.Info, "Info style should not be nil")
			assert.NotNil(t, theme.Highlight, "Highlight style should not be nil")
			assert.NotNil(t, theme.Bold, "Bold style should not be nil")
			assert.NotNil(t, theme.Italic, "Italic style should not be nil")
			assert.NotNil(t, theme.Underline, "Underline style should not be nil")
			assert.NotNil(t, theme.Background, "Background style should not be nil")

			// Test that styles can render text without crashing
			testText := "test"
			assert.NotPanics(t, func() {
				theme.Keyword.Render(testText)
				theme.Variable.Render(testText)
				theme.Command.Render(testText)
				theme.Success.Render(testText)
				theme.Error.Render(testText)
				theme.Warning.Render(testText)
				theme.Info.Render(testText)
				theme.Highlight.Render(testText)
				theme.Bold.Render(testText)
				theme.Italic.Render(testText)
				theme.Underline.Render(testText)
				theme.Background.Render(testText)
			}, "Theme styles should render text without panicking")
		})
	}
}

func TestRenderService_ErrorHandling(t *testing.T) {
	service := NewRenderService()
	// Don't initialize the service

	// Test operations on uninitialized service - should still work gracefully
	theme := service.GetThemeByName("dark1")
	assert.NotNil(t, theme, "Should return default theme even when not initialized")
	assert.Equal(t, "plain", theme.Name, "Should return plain theme when not initialized")

	themes := service.GetAvailableThemes()
	assert.Empty(t, themes, "Should return empty themes when not initialized")

	_, exists := service.GetTheme("default")
	assert.False(t, exists, "Should return false for themes when not initialized")
}

func TestRenderService_PlainThemeRendering(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	plainTheme := service.GetThemeByName("plain")
	assert.NotNil(t, plainTheme)

	// Test that plain theme returns text unchanged
	testText := "Hello World"
	styledText := plainTheme.Keyword.Render(testText)
	assert.Equal(t, testText, styledText, "Plain theme should return text unchanged")

	styledText = plainTheme.Error.Render(testText)
	assert.Equal(t, testText, styledText, "Plain theme should return text unchanged")

	styledText = plainTheme.Success.Render(testText)
	assert.Equal(t, testText, styledText, "Plain theme should return text unchanged")
}

func TestRenderService_ConcurrentAccess(t *testing.T) {
	service := NewRenderService()
	ctx := testutils.NewMockContext()
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test concurrent access to themes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Access various themes concurrently
			themes := []string{"default", "dark", "light", "plain", "dark1", "invalid"}
			for _, themeName := range themes {
				theme := service.GetThemeByName(themeName)
				assert.NotNil(t, theme)

				// Use the theme
				theme.Command.Render("test")
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
