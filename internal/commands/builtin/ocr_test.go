package builtin

import (
	"reflect"
	"testing"

	"neuroshell/pkg/neurotypes"
)

func TestOCRCommand_Name(t *testing.T) {
	cmd := &OCRCommand{}
	expected := "ocr"
	if got := cmd.Name(); got != expected {
		t.Errorf("OCRCommand.Name() = %v, want %v", got, expected)
	}
}

func TestOCRCommand_ParseMode(t *testing.T) {
	cmd := &OCRCommand{}
	expected := neurotypes.ParseModeKeyValue
	if got := cmd.ParseMode(); got != expected {
		t.Errorf("OCRCommand.ParseMode() = %v, want %v", got, expected)
	}
}

func TestOCRCommand_Description(t *testing.T) {
	cmd := &OCRCommand{}
	desc := cmd.Description()
	if desc == "" {
		t.Error("OCRCommand.Description() returned empty string")
	}
}

func TestOCRCommand_Usage(t *testing.T) {
	cmd := &OCRCommand{}
	usage := cmd.Usage()
	if usage == "" {
		t.Error("OCRCommand.Usage() returned empty string")
	}
}

func TestOCRCommand_IsReadOnly(t *testing.T) {
	cmd := &OCRCommand{}
	expected := false // OCR creates files and makes HTTP requests
	if got := cmd.IsReadOnly(); got != expected {
		t.Errorf("OCRCommand.IsReadOnly() = %v, want %v", got, expected)
	}
}

func TestOCRCommand_HelpInfo(t *testing.T) {
	cmd := &OCRCommand{}
	helpInfo := cmd.HelpInfo()

	if helpInfo.Command != cmd.Name() {
		t.Errorf("HelpInfo.Command = %v, want %v", helpInfo.Command, cmd.Name())
	}

	if helpInfo.Description != cmd.Description() {
		t.Errorf("HelpInfo.Description = %v, want %v", helpInfo.Description, cmd.Description())
	}

	if helpInfo.ParseMode != cmd.ParseMode() {
		t.Errorf("HelpInfo.ParseMode = %v, want %v", helpInfo.ParseMode, cmd.ParseMode())
	}

	// Check that required options are present
	expectedOptions := []string{"pdf", "output", "api_key", "model", "max_tokens", "temperature", "pages", "to", "silent"}
	actualOptions := make([]string, len(helpInfo.Options))
	for i, opt := range helpInfo.Options {
		actualOptions[i] = opt.Name
	}

	for _, expectedOpt := range expectedOptions {
		found := false
		for _, actualOpt := range actualOptions {
			if actualOpt == expectedOpt {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected option %s not found in help info", expectedOpt)
		}
	}

	// Check that examples are provided
	if len(helpInfo.Examples) == 0 {
		t.Error("HelpInfo should include usage examples")
	}
}

func TestOCRCommand_extractOptions(t *testing.T) {
	cmd := &OCRCommand{}

	tests := []struct {
		name        string
		options     map[string]string
		expectError bool
		checkConfig func(*OCRConfig) bool
	}{
		{
			name:        "missing PDF path",
			options:     map[string]string{},
			expectError: true,
		},
		{
			name: "basic valid options",
			options: map[string]string{
				"pdf":     "test.pdf",
				"api_key": "test-key",
			},
			expectError: false,
			checkConfig: func(config *OCRConfig) bool {
				return config.PDFPath == "test.pdf" &&
					config.APIKey == "test-key" &&
					config.Model == "meta-llama/Llama-3.2-90B-Vision-Instruct" &&
					config.MaxTokens == 4500 &&
					config.Temperature == 0.0 &&
					config.Pages == "all" &&
					config.ToVariable == "_output" &&
					!config.Silent
			},
		},
		{
			name: "custom output path",
			options: map[string]string{
				"pdf":     "test.pdf",
				"api_key": "test-key",
				"output":  "custom.md",
			},
			expectError: false,
			checkConfig: func(config *OCRConfig) bool {
				return config.OutputPath == "custom.md"
			},
		},
		{
			name: "custom model and tokens",
			options: map[string]string{
				"pdf":        "test.pdf",
				"api_key":    "test-key",
				"model":      "custom-model",
				"max_tokens": "8000",
			},
			expectError: false,
			checkConfig: func(config *OCRConfig) bool {
				return config.Model == "custom-model" && config.MaxTokens == 8000
			},
		},
		{
			name: "invalid max_tokens",
			options: map[string]string{
				"pdf":        "test.pdf",
				"api_key":    "test-key",
				"max_tokens": "invalid",
			},
			expectError: true,
		},
		{
			name: "custom temperature",
			options: map[string]string{
				"pdf":         "test.pdf",
				"api_key":     "test-key",
				"temperature": "0.5",
			},
			expectError: false,
			checkConfig: func(config *OCRConfig) bool {
				return config.Temperature == 0.5
			},
		},
		{
			name: "invalid temperature",
			options: map[string]string{
				"pdf":         "test.pdf",
				"api_key":     "test-key",
				"temperature": "invalid",
			},
			expectError: true,
		},
		{
			name: "custom pages",
			options: map[string]string{
				"pdf":     "test.pdf",
				"api_key": "test-key",
				"pages":   "1-3,5",
			},
			expectError: false,
			checkConfig: func(config *OCRConfig) bool {
				return config.Pages == "1-3,5"
			},
		},
		{
			name: "silent mode",
			options: map[string]string{
				"pdf":     "test.pdf",
				"api_key": "test-key",
				"silent":  "true",
			},
			expectError: false,
			checkConfig: func(config *OCRConfig) bool {
				return config.Silent
			},
		},
		{
			name: "invalid silent value",
			options: map[string]string{
				"pdf":     "test.pdf",
				"api_key": "test-key",
				"silent":  "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := cmd.extractOptions(tt.options)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkConfig != nil && !tt.checkConfig(config) {
				t.Error("Config validation failed")
			}
		})
	}
}

func TestOCRCommand_parsePageSpec(t *testing.T) {
	cmd := &OCRCommand{}

	tests := []struct {
		name        string
		pagesSpec   string
		totalPages  int
		expected    []int
		expectError bool
	}{
		{
			name:       "all pages",
			pagesSpec:  "all",
			totalPages: 3,
			expected:   []int{1, 2, 3},
		},
		{
			name:       "empty spec defaults to all",
			pagesSpec:  "",
			totalPages: 2,
			expected:   []int{1, 2},
		},
		{
			name:       "single page",
			pagesSpec:  "2",
			totalPages: 5,
			expected:   []int{2},
		},
		{
			name:       "page range",
			pagesSpec:  "1-3",
			totalPages: 5,
			expected:   []int{1, 2, 3},
		},
		{
			name:       "multiple ranges and singles",
			pagesSpec:  "1-2,4,6-7",
			totalPages: 10,
			expected:   []int{1, 2, 4, 6, 7},
		},
		{
			name:        "invalid page number",
			pagesSpec:   "0",
			totalPages:  5,
			expectError: true,
		},
		{
			name:        "page number too high",
			pagesSpec:   "10",
			totalPages:  5,
			expectError: true,
		},
		{
			name:        "invalid range format",
			pagesSpec:   "1-2-3",
			totalPages:  5,
			expectError: true,
		},
		{
			name:        "invalid range order",
			pagesSpec:   "3-1",
			totalPages:  5,
			expectError: true,
		},
		{
			name:        "non-numeric page",
			pagesSpec:   "abc",
			totalPages:  5,
			expectError: true,
		},
		{
			name:       "duplicates removed",
			pagesSpec:  "1,2,1,3,2",
			totalPages: 5,
			expected:   []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cmd.parsePageSpec(tt.pagesSpec, tt.totalPages)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parsePageSpec() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestOCRCommand_getOCRPrompt(t *testing.T) {
	cmd := &OCRCommand{}
	prompt := cmd.getOCRPrompt()

	if prompt == "" {
		t.Error("getOCRPrompt() returned empty string")
	}

	// Check that prompt contains key elements
	expectedElements := []string{
		"document",
		"markdown",
		"LaTeX",
		"front matter",
		"primary_language",
		"is_rotation_valid",
		"rotation_correction",
		"is_table",
		"is_diagram",
	}

	for _, element := range expectedElements {
		if !contains(prompt, element) {
			t.Errorf("OCR prompt missing expected element: %s", element)
		}
	}
}

// Helper function to check if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
