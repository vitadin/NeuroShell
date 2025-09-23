package llm

import (
	"testing"

	"neuroshell/internal/commands"
	"neuroshell/pkg/neurotypes"
)

func TestPromptPolishCommand_BasicInfo(t *testing.T) {
	cmd := &PromptPolishCommand{}

	if cmd.Name() != "prompt-polish" {
		t.Errorf("Expected name 'prompt-polish', got '%s'", cmd.Name())
	}

	if cmd.ParseMode() != neurotypes.ParseModeKeyValue {
		t.Errorf("Expected ParseModeKeyValue, got %v", cmd.ParseMode())
	}

	if cmd.IsReadOnly() {
		t.Error("Expected IsReadOnly() to be false")
	}

	desc := cmd.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}

	usage := cmd.Usage()
	if usage == "" {
		t.Error("Usage should not be empty")
	}

	// Check help info structure
	helpInfo := cmd.HelpInfo()
	if helpInfo.Command != "prompt-polish" {
		t.Errorf("Expected help command 'prompt-polish', got '%s'", helpInfo.Command)
	}

	if len(helpInfo.Options) == 0 {
		t.Error("Expected options in help info")
	}

	if len(helpInfo.Examples) == 0 {
		t.Error("Expected examples in help info")
	}

	if len(helpInfo.StoredVariables) == 0 {
		t.Error("Expected stored variables in help info")
	}

	// Check for _output variable
	foundOutput := false
	for _, variable := range helpInfo.StoredVariables {
		if variable.Name == "_output" {
			foundOutput = true
			break
		}
	}
	if !foundOutput {
		t.Error("Expected _output variable in stored variables")
	}
}

func TestPromptPolishCommand_EmptyInput(t *testing.T) {
	cmd := &PromptPolishCommand{}
	options := make(map[string]string)

	// Test empty input
	err := cmd.Execute(options, "")
	if err != nil {
		t.Errorf("Execute should not return error for empty input, got: %v", err)
	}

	// Test whitespace-only input
	err = cmd.Execute(options, "   \n\t   ")
	if err != nil {
		t.Errorf("Execute should not return error for whitespace input, got: %v", err)
	}
}

func TestPromptPolishCommand_DelegationLogic(t *testing.T) {
	cmd := &PromptPolishCommand{}

	// Test that the command builds the correct delegation command string
	// We can't easily test the actual delegation without a full service setup,
	// but we can test the command structure logic is sound

	// Test with no options
	options := make(map[string]string)
	err := cmd.Execute(options, "test input")
	// This will fail because stack service isn't available in test,
	// but that's expected - we're testing the delegation pattern works
	if err == nil {
		t.Error("Expected error when stack service not available")
	}
	if err.Error() != "stack service not available: service stack not found" {
		t.Errorf("Expected specific stack service error, got: %v", err)
	}
}

func TestPromptPolishCommand_OptionsHandling(t *testing.T) {
	cmd := &PromptPolishCommand{}

	// Test with options - should attempt delegation
	options := map[string]string{
		"instruction": "Make it formal",
		"model":       "G4OC",
	}

	err := cmd.Execute(options, "test input")
	// Should fail because stack service isn't available, but confirms delegation attempt
	if err == nil {
		t.Error("Expected error when stack service not available")
	}
	if err.Error() != "stack service not available: service stack not found" {
		t.Errorf("Expected specific stack service error, got: %v", err)
	}
}

func TestPromptPolishCommand_RegistrationExists(t *testing.T) {
	// Test that the command is registered
	registry := commands.GetGlobalRegistry()
	cmd, exists := registry.Get("prompt-polish")
	if !exists {
		t.Error("prompt-polish command should be registered")
	}

	if cmd == nil {
		t.Error("Retrieved command should not be nil")
	}

	// Verify it's the correct type
	if _, ok := cmd.(*PromptPolishCommand); !ok {
		t.Error("Retrieved command should be of type *PromptPolishCommand")
	}
}

func TestPromptPolishCommand_HelpInfoCompleteness(t *testing.T) {
	cmd := &PromptPolishCommand{}
	helpInfo := cmd.HelpInfo()

	// Test that help info contains expected options
	expectedOptions := []string{"instruction", "model"}
	foundOptions := make(map[string]bool)

	for _, option := range helpInfo.Options {
		foundOptions[option.Name] = true
	}

	for _, expectedOption := range expectedOptions {
		if !foundOptions[expectedOption] {
			t.Errorf("Expected option '%s' not found in help info", expectedOption)
		}
	}

	// Test that examples are meaningful
	if len(helpInfo.Examples) < 3 {
		t.Error("Expected at least 3 examples in help info")
	}

	// Test that notes mention delegation
	foundDelegationNote := false
	for _, note := range helpInfo.Notes {
		if contains(note, "delegation") || contains(note, "_prompt_polish") || contains(note, "script") {
			foundDelegationNote = true
			break
		}
	}
	if !foundDelegationNote {
		t.Error("Expected help notes to mention delegation or script usage")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
