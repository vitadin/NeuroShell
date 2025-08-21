package services

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptColorService_Name(t *testing.T) {
	service := NewPromptColorService()
	assert.Equal(t, "prompt_color", service.Name())
}

func TestPromptColorService_Initialize(t *testing.T) {
	// Create registry for testing
	registry := NewRegistry()
	SetGlobalRegistry(registry)

	// Register theme service first
	themeService := NewThemeService()
	err := registry.RegisterService(themeService)
	require.NoError(t, err)

	service := NewPromptColorService()
	err = service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
	assert.NotNil(t, service.themeService)
}

func TestPromptColorService_Initialize_WithoutThemeService(t *testing.T) {
	// Create registry without theme service
	registry := NewRegistry()
	SetGlobalRegistry(registry)

	service := NewPromptColorService()
	err := service.Initialize()
	assert.NoError(t, err)
	assert.True(t, service.initialized)
	assert.Nil(t, service.themeService)
}

func TestPromptColorService_GetServiceInfo(t *testing.T) {
	registry := NewRegistry()
	SetGlobalRegistry(registry)

	service := NewPromptColorService()
	err := service.Initialize()
	require.NoError(t, err)

	info := service.GetServiceInfo()
	assert.Equal(t, "prompt_color", info["name"])
	assert.Equal(t, true, info["initialized"])
	assert.Equal(t, "prompt_color", info["type"])
	assert.Contains(t, info["description"].(string), "color markup")
}

func TestPromptColorService_ProcessColorMarkup_NotInitialized(t *testing.T) {
	service := NewPromptColorService()
	result := service.ProcessColorMarkup("{{color:blue}}test{{/color}}")
	assert.Equal(t, "{{color:blue}}test{{/color}}", result)
}

func TestPromptColorService_ProcessColorMarkup_NoMarkup(t *testing.T) {
	service := NewPromptColorService()
	err := service.Initialize()
	require.NoError(t, err)

	input := "plain text prompt"
	result := service.ProcessColorMarkup(input)
	assert.Equal(t, input, result)
}

func TestPromptColorService_ProcessColorMarkup_BasicColors(t *testing.T) {
	// Set color profile to support colors for testing
	originalProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(originalProfile)

	registry := NewRegistry()
	SetGlobalRegistry(registry)

	service := NewPromptColorService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "red color",
			input:    "{{color:red}}error{{/color}}",
			contains: "error",
		},
		{
			name:     "blue color",
			input:    "{{color:blue}}info{{/color}}",
			contains: "info",
		},
		{
			name:     "hex color",
			input:    "{{color:#ff0000}}custom{{/color}}",
			contains: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ProcessColorMarkup(tt.input)
			assert.Contains(t, result, tt.contains)
			assert.NotEqual(t, tt.input, result) // Should be different due to ANSI codes
		})
	}
}

func TestPromptColorService_ProcessColorMarkup_StyleMarkup(t *testing.T) {
	// Set color profile to support colors for testing
	originalProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(originalProfile)

	registry := NewRegistry()
	SetGlobalRegistry(registry)

	service := NewPromptColorService()
	err := service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "bold text",
			input:    "{{bold}}important{{/bold}}",
			contains: "important",
		},
		{
			name:     "italic text",
			input:    "{{italic}}emphasis{{/italic}}",
			contains: "emphasis",
		},
		{
			name:     "underline text",
			input:    "{{underline}}underlined{{/underline}}",
			contains: "underlined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ProcessColorMarkup(tt.input)

			// Check that styling was applied (result should be different and contain ANSI codes)
			assert.NotEqual(t, tt.input, result)
			assert.Contains(t, result, "\x1b[") // Should contain ANSI escape codes

			// For underline, lipgloss may apply per-character, so check each char exists
			if tt.name == "underline text" {
				for _, char := range tt.contains {
					assert.Contains(t, result, string(char))
				}
			} else {
				// For other styles, the text should be contained within the result
				assert.Contains(t, result, tt.contains)
			}
		})
	}
}

func TestPromptColorService_ProcessColorMarkup_MultipleMarkup(t *testing.T) {
	// Set color profile to support colors for testing
	originalProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(originalProfile)

	registry := NewRegistry()
	SetGlobalRegistry(registry)

	service := NewPromptColorService()
	err := service.Initialize()
	require.NoError(t, err)

	input := "{{color:blue}}path{{/color}} {{bold}}{{color:green}}status{{/color}}{{/bold}}"
	result := service.ProcessColorMarkup(input)

	assert.Contains(t, result, "path")
	assert.Contains(t, result, "status")
	assert.NotEqual(t, input, result)
}

func TestPromptColorService_ProcessColorMarkup_SemanticColors(t *testing.T) {
	// Set color profile to support colors for testing
	originalProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(originalProfile)

	registry := NewRegistry()
	SetGlobalRegistry(registry)

	// Register theme service
	themeService := NewThemeService()
	err := registry.RegisterService(themeService)
	require.NoError(t, err)

	service := NewPromptColorService()
	err = service.Initialize()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "info semantic",
			input:    "{{color:info}}information{{/color}}",
			contains: "information",
		},
		{
			name:     "success semantic",
			input:    "{{color:success}}completed{{/color}}",
			contains: "completed",
		},
		{
			name:     "error semantic",
			input:    "{{color:error}}failed{{/color}}",
			contains: "failed",
		},
		{
			name:     "warning semantic",
			input:    "{{color:warning}}caution{{/color}}",
			contains: "caution",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ProcessColorMarkup(tt.input)
			assert.Contains(t, result, tt.contains)
			assert.NotEqual(t, tt.input, result)
		})
	}
}

func TestPromptColorService_ProcessColorMarkup_ASCIIFallback(t *testing.T) {
	// Set color profile to ASCII (no colors)
	originalProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	defer lipgloss.SetColorProfile(originalProfile)

	registry := NewRegistry()
	SetGlobalRegistry(registry)

	service := NewPromptColorService()
	err := service.Initialize()
	require.NoError(t, err)

	input := "{{color:blue}}text{{/color}} {{bold}}bold{{/bold}}"
	result := service.ProcessColorMarkup(input)

	// Should strip markup but preserve text
	expected := "text bold"
	assert.Equal(t, expected, result)
}

func TestPromptColorService_IsColorSupported(t *testing.T) {
	service := NewPromptColorService()

	// Test with color support
	originalProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	assert.True(t, service.IsColorSupported())

	// Test without color support
	lipgloss.SetColorProfile(termenv.Ascii)
	assert.False(t, service.IsColorSupported())

	// Restore original profile
	lipgloss.SetColorProfile(originalProfile)
}

func TestPromptColorService_createDirectColorStyle(t *testing.T) {
	service := NewPromptColorService()

	tests := []struct {
		name      string
		colorSpec string
	}{
		{"hex color", "#ff0000"},
		{"red", "red"},
		{"blue", "blue"},
		{"green", "green"},
		{"yellow", "yellow"},
		{"cyan", "cyan"},
		{"magenta", "magenta"},
		{"purple", "purple"},
		{"white", "white"},
		{"gray", "gray"},
		{"grey", "grey"},
		{"bright-red", "bright-red"},
		{"bright-blue", "bright-blue"},
		{"invalid", "invalid-color"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := service.createDirectColorStyle(tt.colorSpec)
			// Should create a style without panicking
			assert.NotNil(t, style)
		})
	}
}

func TestPromptColorService_stripColorMarkup(t *testing.T) {
	service := NewPromptColorService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single color",
			input:    "{{color:blue}}text{{/color}}",
			expected: "text",
		},
		{
			name:     "multiple colors",
			input:    "{{color:red}}error{{/color}} and {{color:green}}success{{/color}}",
			expected: "error and success",
		},
		{
			name:     "no markup",
			input:    "plain text",
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.stripColorMarkup(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPromptColorService_stripStyleMarkup(t *testing.T) {
	service := NewPromptColorService()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bold markup",
			input:    "{{bold}}text{{/bold}}",
			expected: "text",
		},
		{
			name:     "multiple styles",
			input:    "{{bold}}bold{{/bold}} {{italic}}italic{{/italic}}",
			expected: "bold italic",
		},
		{
			name:     "no markup",
			input:    "plain text",
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.stripStyleMarkup(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkPromptColorService_ProcessColorMarkup_NoMarkup(b *testing.B) {
	registry := NewRegistry()
	SetGlobalRegistry(registry)

	service := NewPromptColorService()
	_ = service.Initialize()

	input := "plain text without any markup"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ProcessColorMarkup(input)
	}
}

func BenchmarkPromptColorService_ProcessColorMarkup_WithMarkup(b *testing.B) {
	registry := NewRegistry()
	SetGlobalRegistry(registry)

	service := NewPromptColorService()
	_ = service.Initialize()

	input := "{{color:blue}}${@pwd}{{/color}} {{bold}}{{color:green}}status{{/color}}{{/bold}}"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ProcessColorMarkup(input)
	}
}
