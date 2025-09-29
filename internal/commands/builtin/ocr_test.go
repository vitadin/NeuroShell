package builtin

import (
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
					config.Model == "allenai/olmOCR-7B-0825" &&
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
