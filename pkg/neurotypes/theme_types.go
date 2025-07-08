// Package neurotypes defines theme-related data structures for NeuroShell's rendering system.
// This file contains the core types for representing and managing theme configurations.
package neurotypes

// ThemeConfig represents a theme configuration loaded from YAML.
// It defines the color and style settings for various semantic elements in NeuroShell.
type ThemeConfig struct {
	// Name is the theme identifier (e.g., "default", "dark", "light", "plain")
	Name string `yaml:"name" json:"name"`

	// Description provides a brief description of the theme
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Styles contains the color and style definitions for different semantic elements
	Styles ThemeStyles `yaml:"styles" json:"styles"`
}

// ThemeStyles defines the styling configuration for different semantic elements.
// Each style can specify foreground color, background color, and text decorations.
type ThemeStyles struct {
	// Keyword style for NeuroShell keywords and built-in functions
	Keyword StyleConfig `yaml:"keyword" json:"keyword"`

	// Variable style for variable names and references
	Variable StyleConfig `yaml:"variable" json:"variable"`

	// Command style for command names
	Command StyleConfig `yaml:"command" json:"command"`

	// Success style for success messages and positive feedback
	Success StyleConfig `yaml:"success" json:"success"`

	// Error style for error messages and warnings
	Error StyleConfig `yaml:"error" json:"error"`

	// Warning style for warning messages
	Warning StyleConfig `yaml:"warning" json:"warning"`

	// Info style for informational messages
	Info StyleConfig `yaml:"info" json:"info"`

	// Highlight style for emphasized text and selections
	Highlight StyleConfig `yaml:"highlight" json:"highlight"`

	// Bold style for bold text
	Bold StyleConfig `yaml:"bold" json:"bold"`

	// Italic style for italic text
	Italic StyleConfig `yaml:"italic" json:"italic"`

	// Underline style for underlined text
	Underline StyleConfig `yaml:"underline" json:"underline"`

	// Background style for background elements
	Background StyleConfig `yaml:"background" json:"background"`

	// List style for list enumerators and bullet points
	List StyleConfig `yaml:"list" json:"list"`
}

// StyleConfig defines the visual styling for a semantic element.
// It supports both simple color specifications and adaptive colors for light/dark terminals.
type StyleConfig struct {
	// Foreground color - can be hex color, named color, or adaptive color object
	Foreground interface{} `yaml:"foreground,omitempty" json:"foreground,omitempty"`

	// Background color - can be hex color, named color, or adaptive color object
	Background interface{} `yaml:"background,omitempty" json:"background,omitempty"`

	// Bold text decoration
	Bold *bool `yaml:"bold,omitempty" json:"bold,omitempty"`

	// Italic text decoration
	Italic *bool `yaml:"italic,omitempty" json:"italic,omitempty"`

	// Underline text decoration
	Underline *bool `yaml:"underline,omitempty" json:"underline,omitempty"`

	// Strikethrough text decoration
	Strikethrough *bool `yaml:"strikethrough,omitempty" json:"strikethrough,omitempty"`
}

// AdaptiveColor defines colors that adapt to light and dark terminal backgrounds.
// This allows themes to work well in both light and dark environments.
type AdaptiveColor struct {
	// Light color for light terminal backgrounds
	Light string `yaml:"light" json:"light"`

	// Dark color for dark terminal backgrounds
	Dark string `yaml:"dark" json:"dark"`
}

// ThemeFile represents a complete theme file loaded from YAML.
// Each theme file contains a single ThemeConfig with all styling information.
type ThemeFile struct {
	ThemeConfig `yaml:",inline" json:",inline"`
}

// ThemeValidationError represents validation errors that occur during theme loading.
type ThemeValidationError struct {
	Field   string `json:"field"`   // The field that failed validation
	Value   string `json:"value"`   // The invalid value
	Message string `json:"message"` // Human-readable error message
}

// Error implements the error interface for ThemeValidationError.
func (e ThemeValidationError) Error() string {
	return e.Message
}
