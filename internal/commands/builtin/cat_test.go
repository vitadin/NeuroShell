package builtin

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"neuroshell/internal/context"
	"neuroshell/internal/services"
	"neuroshell/internal/stringprocessing"
	"neuroshell/pkg/neurotypes"
)

// setupCatTestRegistry creates a clean service registry for testing
func setupCatTestRegistry(t *testing.T) {
	ctx := context.New()

	// Create a new registry for testing
	oldRegistry := services.GetGlobalRegistry()
	services.SetGlobalRegistry(services.NewRegistry())

	// Set the test context as global context
	context.SetGlobalContext(ctx)

	// Register variable service
	err := services.GetGlobalRegistry().RegisterService(services.NewVariableService())
	require.NoError(t, err)

	// Initialize services
	err = services.GetGlobalRegistry().InitializeAll()
	require.NoError(t, err)

	// Cleanup function to restore original registry
	t.Cleanup(func() {
		services.SetGlobalRegistry(oldRegistry)
		context.ResetGlobalContext()
	})
}

// createTempFile creates a temporary file with given content for testing
func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "cat_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = tmpFile.Close() // Ignore close error in test cleanup
	}()

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return tmpFile.Name()
}

// createTempBinaryFile creates a temporary binary file for testing
func createTempBinaryFile(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "cat_test_binary_*.bin")
	if err != nil {
		t.Fatalf("Failed to create temp binary file: %v", err)
	}
	defer func() {
		_ = tmpFile.Close() // Ignore close error in test cleanup
	}()

	// Write binary content (null bytes)
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}
	if _, err := tmpFile.Write(binaryContent); err != nil {
		t.Fatalf("Failed to write binary content: %v", err)
	}

	return tmpFile.Name()
}

// TestCatCommand_Name tests the Name method
func TestCatCommand_Name(t *testing.T) {
	cmd := &CatCommand{}
	if cmd.Name() != "cat" {
		t.Errorf("Expected name 'cat', got '%s'", cmd.Name())
	}
}

// TestCatCommand_ParseMode tests the ParseMode method
func TestCatCommand_ParseMode(t *testing.T) {
	cmd := &CatCommand{}
	if cmd.ParseMode() != neurotypes.ParseModeKeyValue {
		t.Errorf("Expected ParseModeKeyValue, got %v", cmd.ParseMode())
	}
}

// TestCatCommand_Description tests the Description method
func TestCatCommand_Description(t *testing.T) {
	cmd := &CatCommand{}
	desc := cmd.Description()
	if !strings.Contains(desc, "Display file contents") {
		t.Errorf("Expected description to contain 'Display file contents', got '%s'", desc)
	}
}

// TestCatCommand_Usage tests the Usage method
func TestCatCommand_Usage(t *testing.T) {
	cmd := &CatCommand{}
	usage := cmd.Usage()
	if !strings.Contains(usage, "\\cat") {
		t.Errorf("Expected usage to contain '\\cat', got '%s'", usage)
	}
}

// TestCatCommand_HelpInfo tests the HelpInfo method
func TestCatCommand_HelpInfo(t *testing.T) {
	cmd := &CatCommand{}
	helpInfo := cmd.HelpInfo()

	// Check basic structure
	if helpInfo.Command != "cat" {
		t.Errorf("Expected command 'cat', got '%s'", helpInfo.Command)
	}

	// Check that required options are present
	expectedOptions := []string{"path", "to", "silent", "lines", "start"}
	for _, expectedOpt := range expectedOptions {
		found := false
		for _, opt := range helpInfo.Options {
			if opt.Name == expectedOpt {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected option '%s' not found in help info", expectedOpt)
		}
	}

	// Check examples exist
	if len(helpInfo.Examples) == 0 {
		t.Error("Expected examples in help info")
	}

	// Check notes exist
	if len(helpInfo.Notes) == 0 {
		t.Error("Expected notes in help info")
	}
}

// TestCatCommand_Execute_BasicFunctionality tests basic file reading
func TestCatCommand_Execute_BasicFunctionality(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	tests := []struct {
		name     string
		content  string
		args     map[string]string
		input    string
		expected string
	}{
		{
			name:     "simple text file",
			content:  "Hello, World!",
			args:     map[string]string{},
			input:    "", // will be set to temp file path
			expected: "Hello, World!",
		},
		{
			name:     "multiline content",
			content:  "Line 1\nLine 2\nLine 3",
			args:     map[string]string{},
			input:    "",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "empty file",
			content:  "",
			args:     map[string]string{},
			input:    "",
			expected: "",
		},
		{
			name:     "file with special characters",
			content:  "Special: !@#$%^&*()_+-={}[]|\\:;\"'<>?,./ \t",
			args:     map[string]string{},
			input:    "",
			expected: "Special: !@#$%^&*()_+-={}[]|\\:;\"'<>?,./ \t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with content
			tempFile := createTempFile(t, tt.content)
			defer func() { _ = os.Remove(tempFile) }()

			// Set input to temp file path if not using args
			if tt.args["path"] == "" {
				tt.input = tempFile
			} else {
				tt.args["path"] = tempFile
			}

			// Capture output
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, tt.input)
				if err != nil {
					t.Errorf("Execute failed: %v", err)
				}
			})

			// Verify output
			output = strings.TrimSuffix(output, "\n") // Remove trailing newline added by cat
			if output != tt.expected {
				t.Errorf("Expected output '%s', got '%s'", tt.expected, output)
			}

			// Verify variable storage
			variableService, _ := services.GetGlobalVariableService()
			storedValue, err := variableService.Get("_output")
			if err != nil {
				t.Errorf("Failed to get _output variable: %v", err)
			}
			if storedValue != tt.expected {
				t.Errorf("Expected stored value '%s', got '%s'", tt.expected, storedValue)
			}
		})
	}
}

// TestCatCommand_Execute_BracketSyntax tests bracket syntax usage
func TestCatCommand_Execute_BracketSyntax(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	content := "Test content for bracket syntax"
	tempFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tempFile) }()

	tests := []struct {
		name string
		args map[string]string
	}{
		{
			name: "path in args",
			args: map[string]string{"path": tempFile},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, "")
				if err != nil {
					t.Errorf("Execute failed: %v", err)
				}
			})

			output = strings.TrimSuffix(output, "\n")
			if output != content {
				t.Errorf("Expected output '%s', got '%s'", content, output)
			}
		})
	}
}

// TestCatCommand_Execute_ToOption tests the 'to' option for variable storage
func TestCatCommand_Execute_ToOption(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	content := "Content for variable storage test"
	tempFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tempFile) }()

	tests := []struct {
		name        string
		args        map[string]string
		expectedVar string
		expectedVal string
	}{
		{
			name:        "store in custom variable",
			args:        map[string]string{"path": tempFile, "to": "my_content"},
			expectedVar: "my_content",
			expectedVal: content,
		},
		{
			name:        "store in _output (default)",
			args:        map[string]string{"path": tempFile},
			expectedVar: "_output",
			expectedVal: content,
		},
		{
			name:        "store in user variable test_error",
			args:        map[string]string{"path": tempFile, "to": "test_error"},
			expectedVar: "test_error",
			expectedVal: content,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, "")
			if err != nil {
				t.Errorf("Execute failed: %v", err)
			}

			// Check variable storage
			variableService, _ := services.GetGlobalVariableService()
			storedValue, err := variableService.Get(tt.expectedVar)
			if err != nil {
				t.Errorf("Failed to get variable '%s': %v", tt.expectedVar, err)
			}
			if storedValue != tt.expectedVal {
				t.Errorf("Expected variable '%s' to contain '%s', got '%s'", tt.expectedVar, tt.expectedVal, storedValue)
			}
		})
	}
}

// TestCatCommand_Execute_SilentOption tests the 'silent' option
func TestCatCommand_Execute_SilentOption(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	content := "Content for silent test"
	tempFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tempFile) }()

	tests := []struct {
		name           string
		args           map[string]string
		expectOutput   bool
		expectVariable bool
	}{
		{
			name:           "silent true - no output",
			args:           map[string]string{"path": tempFile, "silent": "true"},
			expectOutput:   false,
			expectVariable: true,
		},
		{
			name:           "silent false - with output",
			args:           map[string]string{"path": tempFile, "silent": "false"},
			expectOutput:   true,
			expectVariable: true,
		},
		{
			name:           "default behavior (no silent) - with output",
			args:           map[string]string{"path": tempFile},
			expectOutput:   true,
			expectVariable: true,
		},
		{
			name:           "silent with custom variable",
			args:           map[string]string{"path": tempFile, "silent": "true", "to": "silent_content"},
			expectOutput:   false,
			expectVariable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, "")
				if err != nil {
					t.Errorf("Execute failed: %v", err)
				}
			})

			// Check console output
			if tt.expectOutput {
				if len(output) == 0 {
					t.Error("Expected console output but got none")
				}
			} else if len(output) > 0 {
				t.Errorf("Expected no console output but got: '%s'", output)
			}

			// Check variable storage
			if tt.expectVariable {
				variableService, _ := services.GetGlobalVariableService()
				targetVar := tt.args["to"]
				if targetVar == "" {
					targetVar = "_output"
				}
				storedValue, err := variableService.Get(targetVar)
				if err != nil {
					t.Errorf("Failed to get variable '%s': %v", targetVar, err)
				}
				if storedValue != content {
					t.Errorf("Expected variable to contain '%s', got '%s'", content, storedValue)
				}
			}
		})
	}
}

// TestCatCommand_Execute_LinesOption tests the 'lines' option for limiting output
func TestCatCommand_Execute_LinesOption(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	// Create multi-line content
	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	content := strings.Join(lines, "\n")
	tempFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tempFile) }()

	tests := []struct {
		name           string
		args           map[string]string
		expectedOutput string
		expectedStored string
	}{
		{
			name:           "limit to 3 lines",
			args:           map[string]string{"path": tempFile, "lines": "3"},
			expectedOutput: "Line 1\nLine 2\nLine 3",
			expectedStored: "Line 1\nLine 2\nLine 3", // Store what is displayed
		},
		{
			name:           "limit to 1 line",
			args:           map[string]string{"path": tempFile, "lines": "1"},
			expectedOutput: "Line 1",
			expectedStored: "Line 1",
		},
		{
			name:           "limit larger than content",
			args:           map[string]string{"path": tempFile, "lines": "10"},
			expectedOutput: content,
			expectedStored: content,
		},
		{
			name:           "no limit (default)",
			args:           map[string]string{"path": tempFile},
			expectedOutput: content,
			expectedStored: content,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, "")
				if err != nil {
					t.Errorf("Execute failed: %v", err)
				}
			})

			output = strings.TrimSuffix(output, "\n")
			if output != tt.expectedOutput {
				t.Errorf("Expected output '%s', got '%s'", tt.expectedOutput, output)
			}

			// Check stored content
			variableService, _ := services.GetGlobalVariableService()
			storedValue, err := variableService.Get("_output")
			if err != nil {
				t.Errorf("Failed to get _output variable: %v", err)
			}
			if storedValue != tt.expectedStored {
				t.Errorf("Expected stored value '%s', got '%s'", tt.expectedStored, storedValue)
			}
		})
	}
}

// TestCatCommand_Execute_StartAndLinesOptions tests start and lines options together
func TestCatCommand_Execute_StartAndLinesOptions(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	// Create multi-line content
	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5", "Line 6"}
	content := strings.Join(lines, "\n")
	tempFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tempFile) }()

	tests := []struct {
		name           string
		args           map[string]string
		expectedOutput string
	}{
		{
			name:           "start from line 2, take 3 lines",
			args:           map[string]string{"path": tempFile, "start": "2", "lines": "3"},
			expectedOutput: "Line 2\nLine 3\nLine 4",
		},
		{
			name:           "start from line 4, take 2 lines",
			args:           map[string]string{"path": tempFile, "start": "4", "lines": "2"},
			expectedOutput: "Line 4\nLine 5",
		},
		{
			name:           "start from line 5, take more than available",
			args:           map[string]string{"path": tempFile, "start": "5", "lines": "10"},
			expectedOutput: "Line 5\nLine 6",
		},
		{
			name:           "start from line 1 (default), take 2 lines",
			args:           map[string]string{"path": tempFile, "lines": "2"},
			expectedOutput: "Line 1\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, "")
				if err != nil {
					t.Errorf("Execute failed: %v", err)
				}
			})

			output = strings.TrimSuffix(output, "\n")
			if output != tt.expectedOutput {
				t.Errorf("Expected output '%s', got '%s'", tt.expectedOutput, output)
			}
		})
	}
}

// TestCatCommand_Execute_ErrorCases tests various error conditions
func TestCatCommand_Execute_ErrorCases(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	tests := []struct {
		name        string
		args        map[string]string
		input       string
		expectError bool
		errorString string
	}{
		{
			name:        "no file path provided",
			args:        map[string]string{},
			input:       "",
			expectError: true,
			errorString: "file path is required",
		},
		{
			name:        "non-existent file",
			args:        map[string]string{},
			input:       "/path/that/does/not/exist.txt",
			expectError: true,
			errorString: "failed to access file",
		},
		{
			name:        "directory instead of file",
			args:        map[string]string{},
			input:       "/tmp", // This should be a directory
			expectError: true,
			errorString: "is a directory, not a file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmd.Execute(tt.args, tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorString) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorString, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCatCommand_Execute_BinaryFileDetection tests binary file detection
func TestCatCommand_Execute_BinaryFileDetection(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	// Create binary file
	binaryFile := createTempBinaryFile(t)
	defer func() { _ = os.Remove(binaryFile) }()

	err := cmd.Execute(map[string]string{"path": binaryFile}, "")
	if err == nil {
		t.Error("Expected error for binary file but got none")
	}
	if !strings.Contains(err.Error(), "appears to be a binary file") {
		t.Errorf("Expected binary file error, got: %v", err)
	}
}

// TestCatCommand_Execute_InvalidOptions tests handling of invalid option values
func TestCatCommand_Execute_InvalidOptions(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	content := "Test content"
	tempFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tempFile) }()

	tests := []struct {
		name string
		args map[string]string
	}{
		{
			name: "invalid silent value",
			args: map[string]string{"path": tempFile, "silent": "not_a_boolean"},
		},
		{
			name: "invalid lines value",
			args: map[string]string{"path": tempFile, "lines": "not_a_number"},
		},
		{
			name: "negative lines value",
			args: map[string]string{"path": tempFile, "lines": "-5"},
		},
		{
			name: "invalid start value",
			args: map[string]string{"path": tempFile, "start": "not_a_number"},
		},
		{
			name: "zero start value",
			args: map[string]string{"path": tempFile, "start": "0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These should not error due to tolerant parsing
			err := cmd.Execute(tt.args, "")
			if err != nil {
				t.Errorf("Expected tolerant parsing, but got error: %v", err)
			}
		})
	}
}

// TestCatCommand_Execute_CombinedOptions tests various option combinations
func TestCatCommand_Execute_CombinedOptions(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	content := strings.Join(lines, "\n")
	tempFile := createTempFile(t, content)
	defer func() { _ = os.Remove(tempFile) }()

	tests := []struct {
		name           string
		args           map[string]string
		expectOutput   bool
		expectedOutput string
		expectedVar    string
		expectedStored string
	}{
		{
			name:           "silent + custom variable + lines",
			args:           map[string]string{"path": tempFile, "silent": "true", "to": "test_var", "lines": "2"},
			expectOutput:   false,
			expectedOutput: "",
			expectedVar:    "test_var",
			expectedStored: "Line 1\nLine 2", // Store what would be displayed
		},
		{
			name:           "start + lines + custom variable",
			args:           map[string]string{"path": tempFile, "start": "3", "lines": "2", "to": "range_var"},
			expectOutput:   true,
			expectedOutput: "Line 3\nLine 4",
			expectedVar:    "range_var",
			expectedStored: "Line 3\nLine 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(tt.args, "")
				if err != nil {
					t.Errorf("Execute failed: %v", err)
				}
			})

			// Check console output
			if tt.expectOutput {
				output = strings.TrimSuffix(output, "\n")
				if output != tt.expectedOutput {
					t.Errorf("Expected output '%s', got '%s'", tt.expectedOutput, output)
				}
			} else if len(output) > 0 {
				t.Errorf("Expected no output but got: '%s'", output)
			}

			// Check variable storage
			variableService, _ := services.GetGlobalVariableService()
			storedValue, err := variableService.Get(tt.expectedVar)
			if err != nil {
				t.Errorf("Failed to get variable '%s': %v", tt.expectedVar, err)
			}
			if storedValue != tt.expectedStored {
				t.Errorf("Expected variable to contain '%s', got '%s'", tt.expectedStored, storedValue)
			}
		})
	}
}

// TestCatCommand_isTextFile tests the binary file detection logic
func TestCatCommand_isTextFile(t *testing.T) {
	cmd := &CatCommand{}

	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "plain text",
			content:  []byte("Hello, World!"),
			expected: true,
		},
		{
			name:     "text with newlines",
			content:  []byte("Line 1\nLine 2\nLine 3"),
			expected: true,
		},
		{
			name:     "text with tabs",
			content:  []byte("Column1\tColumn2\tColumn3"),
			expected: true,
		},
		{
			name:     "empty content",
			content:  []byte(""),
			expected: true,
		},
		{
			name:     "binary with null bytes",
			content:  []byte{0x00, 0x01, 0x02, 0x03},
			expected: false,
		},
		{
			name:     "mixed text and null",
			content:  []byte("Hello\x00World"),
			expected: false,
		},
		{
			name:     "invalid UTF-8",
			content:  []byte{0xFF, 0xFE, 0xFD},
			expected: false,
		},
		{
			name:     "lots of control characters",
			content:  []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}, // 90% non-printable
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.isTextFile(tt.content)
			if result != tt.expected {
				t.Errorf("Expected %v for content %v, got %v", tt.expected, tt.content, result)
			}
		})
	}
}

// TestCatCommand_Execute_NewlineHandling tests proper newline handling
func TestCatCommand_Execute_NewlineHandling(t *testing.T) {
	setupCatTestRegistry(t)
	cmd := &CatCommand{}

	tests := []struct {
		name           string
		content        string
		expectedOutput string
	}{
		{
			name:           "content without trailing newline",
			content:        "No newline at end",
			expectedOutput: "No newline at end\n", // Cat should add newline
		},
		{
			name:           "content with trailing newline",
			content:        "Has newline at end\n",
			expectedOutput: "Has newline at end\n", // Should preserve existing newline
		},
		{
			name:           "empty content",
			content:        "",
			expectedOutput: "", // Empty content should remain empty
		},
		{
			name:           "only newline",
			content:        "\n",
			expectedOutput: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := createTempFile(t, tt.content)
			defer func() { _ = os.Remove(tempFile) }()

			output := stringprocessing.CaptureOutput(func() {
				err := cmd.Execute(map[string]string{"path": tempFile}, "")
				if err != nil {
					t.Errorf("Execute failed: %v", err)
				}
			})

			if output != tt.expectedOutput {
				t.Errorf("Expected output %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}
