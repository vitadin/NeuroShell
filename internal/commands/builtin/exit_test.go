package builtin

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"neuroshell/pkg/neurotypes"
)

func TestExitCommand_Name(t *testing.T) {
	cmd := &ExitCommand{}
	assert.Equal(t, "exit", cmd.Name())
}

func TestExitCommand_ParseMode(t *testing.T) {
	cmd := &ExitCommand{}
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())
}

func TestExitCommand_Description(t *testing.T) {
	cmd := &ExitCommand{}
	assert.Equal(t, "Exit the shell with optional exit code and message", cmd.Description())
}

func TestExitCommand_Usage(t *testing.T) {
	cmd := &ExitCommand{}
	assert.Equal(t, "\\exit[code=N, message=text]", cmd.Usage())
}

// TestExitCommand_Execute tests the exit command by running it in a subprocess
// since os.Exit() would terminate the test process
func TestExitCommand_Execute(t *testing.T) {
	if os.Getenv("TEST_EXIT_COMMAND") == "1" {
		// This code runs in the subprocess
		cmd := &ExitCommand{}

		// This should call os.Exit(0) and terminate the subprocess
		_ = cmd.Execute(map[string]string{}, "")

		// If we reach here, the test failed because os.Exit wasn't called
		t.Fatal("Expected os.Exit to be called")
		return
	}

	// Run the test in a subprocess
	subCmd := exec.Command(os.Args[0], "-test.run=TestExitCommand_Execute")
	subCmd.Env = append(os.Environ(), "TEST_EXIT_COMMAND=1")

	output, err := subCmd.CombinedOutput()
	outputStr := string(output)

	// The subprocess should exit with code 0 (success)
	if e, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 0, e.ExitCode(), "Expected exit code 0, got %d. Output: %s", e.ExitCode(), outputStr)
	} else {
		// No error means exit code 0
		assert.NoError(t, err, "Expected successful exit (code 0)")
	}

	// Exit command should not print any message - it just exits silently
	// We just verify that the subprocess exited cleanly
}

// TestExitCommand_Execute_WithArgs tests that known args are processed correctly
func TestExitCommand_Execute_WithArgs(t *testing.T) {
	if os.Getenv("TEST_EXIT_COMMAND_ARGS") == "1" {
		// This code runs in the subprocess
		cmd := &ExitCommand{}

		// Test with mixed args - known args are processed, unknown are ignored
		args := map[string]string{"force": "true", "code": "1"}
		_ = cmd.Execute(args, "")

		t.Fatal("Expected os.Exit to be called")
		return
	}

	// Run the test in a subprocess
	subCmd := exec.Command(os.Args[0], "-test.run=TestExitCommand_Execute_WithArgs")
	subCmd.Env = append(os.Environ(), "TEST_EXIT_COMMAND_ARGS=1")

	output, err := subCmd.CombinedOutput()
	outputStr := string(output)

	// Should exit with code 1 (code arg is processed)
	if e, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 1, e.ExitCode(), "Expected exit code 1, got %d. Output: %s", e.ExitCode(), outputStr)
	} else {
		t.Errorf("Expected exit code 1 but subprocess succeeded with code 0. Output: %s", outputStr)
	}
}

// TestExitCommand_Execute_WithInput tests that input is ignored
func TestExitCommand_Execute_WithInput(t *testing.T) {
	if os.Getenv("TEST_EXIT_COMMAND_INPUT") == "1" {
		// This code runs in the subprocess
		cmd := &ExitCommand{}

		// Test with input - should still exit normally
		_ = cmd.Execute(map[string]string{}, "some input text")

		t.Fatal("Expected os.Exit to be called")
		return
	}

	// Run the test in a subprocess
	subCmd := exec.Command(os.Args[0], "-test.run=TestExitCommand_Execute_WithInput")
	subCmd.Env = append(os.Environ(), "TEST_EXIT_COMMAND_INPUT=1")

	output, err := subCmd.CombinedOutput()
	outputStr := string(output)

	// Should still exit with code 0 (input is ignored)
	if e, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 0, e.ExitCode(), "Expected exit code 0, got %d. Output: %s", e.ExitCode(), outputStr)
	} else {
		assert.NoError(t, err, "Expected successful exit (code 0)")
	}

	// Exit command should not print any message - it just exits silently
	// We just verify that the subprocess exited cleanly with code 0
}

// TestExitCommand_Execute_WithCode tests exit command with custom exit codes
func TestExitCommand_Execute_WithCode(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected int
	}{
		{"code_1", "1", 1},
		{"code_2", "2", 2},
		{"code_255", "255", 255},
		{"code_invalid_negative", "-1", 0},  // Should default to 0
		{"code_invalid_too_high", "256", 0}, // Should default to 0
		{"code_invalid_text", "abc", 0},     // Should default to 0
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if os.Getenv("TEST_EXIT_CODE_"+tc.name) == "1" {
				// This code runs in the subprocess
				cmd := &ExitCommand{}
				args := map[string]string{"code": tc.code}
				_ = cmd.Execute(args, "")
				t.Fatal("Expected os.Exit to be called")
				return
			}

			// Run the test in a subprocess
			subCmd := exec.Command(os.Args[0], "-test.run=TestExitCommand_Execute_WithCode/"+tc.name)
			subCmd.Env = append(os.Environ(), "TEST_EXIT_CODE_"+tc.name+"=1")

			output, err := subCmd.CombinedOutput()
			outputStr := string(output)

			if e, ok := err.(*exec.ExitError); ok {
				assert.Equal(t, tc.expected, e.ExitCode(), "Expected exit code %d, got %d. Output: %s", tc.expected, e.ExitCode(), outputStr)
			} else if tc.expected == 0 {
				// No error means exit code 0
				assert.NoError(t, err, "Expected successful exit (code 0)")
			} else {
				t.Errorf("Expected exit code %d but subprocess succeeded with code 0. Output: %s", tc.expected, outputStr)
			}
		})
	}
}

// TestExitCommand_Execute_WithMessage tests exit command with messages
func TestExitCommand_Execute_WithMessage(t *testing.T) {
	testCases := []struct {
		name    string
		message string
	}{
		{"simple_message", "Goodbye!"},
		{"multi_word", "Task completed successfully"},
		{"empty_message", ""},
		{"special_chars", "Error: File not found!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if os.Getenv("TEST_EXIT_MESSAGE_"+tc.name) == "1" {
				// This code runs in the subprocess
				cmd := &ExitCommand{}
				args := map[string]string{"message": tc.message}
				_ = cmd.Execute(args, "")
				t.Fatal("Expected os.Exit to be called")
				return
			}

			// Run the test in a subprocess
			subCmd := exec.Command(os.Args[0], "-test.run=TestExitCommand_Execute_WithMessage/"+tc.name)
			subCmd.Env = append(os.Environ(), "TEST_EXIT_MESSAGE_"+tc.name+"=1")

			output, err := subCmd.CombinedOutput()
			outputStr := string(output)

			// Should exit with code 0
			if e, ok := err.(*exec.ExitError); ok {
				assert.Equal(t, 0, e.ExitCode(), "Expected exit code 0, got %d. Output: %s", e.ExitCode(), outputStr)
			} else {
				assert.NoError(t, err, "Expected successful exit (code 0)")
			}

			// Check that message was printed (if not empty)
			if tc.message != "" {
				assert.Contains(t, outputStr, tc.message, "Expected message '%s' in output: %s", tc.message, outputStr)
			}
		})
	}
}

// TestExitCommand_Execute_WithBoth tests exit command with both code and message
func TestExitCommand_Execute_WithBoth(t *testing.T) {
	if os.Getenv("TEST_EXIT_BOTH") == "1" {
		// This code runs in the subprocess
		cmd := &ExitCommand{}
		args := map[string]string{
			"code":    "42",
			"message": "Custom exit message",
		}
		_ = cmd.Execute(args, "")
		t.Fatal("Expected os.Exit to be called")
		return
	}

	// Run the test in a subprocess
	subCmd := exec.Command(os.Args[0], "-test.run=TestExitCommand_Execute_WithBoth")
	subCmd.Env = append(os.Environ(), "TEST_EXIT_BOTH=1")

	output, err := subCmd.CombinedOutput()
	outputStr := string(output)

	// Should exit with code 42
	if e, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 42, e.ExitCode(), "Expected exit code 42, got %d. Output: %s", e.ExitCode(), outputStr)
	} else {
		t.Errorf("Expected exit code 42 but subprocess succeeded with code 0. Output: %s", outputStr)
	}

	// Check that message was printed
	assert.Contains(t, outputStr, "Custom exit message", "Expected message in output: %s", outputStr)
}

// TestExitCommand_Execute_MessageOnly tests the basic metadata without calling os.Exit
// Since the exit command now supports parameters, we test the updated interface
func TestExitCommand_Execute_MessageOnly(t *testing.T) {
	// We can't easily test the full Execute method due to os.Exit,
	// but we can test the command interface methods

	cmd := &ExitCommand{}

	// Test that all interface methods work correctly
	assert.Equal(t, "exit", cmd.Name())
	assert.Equal(t, "Exit the shell with optional exit code and message", cmd.Description())
	assert.Equal(t, "\\exit[code=N, message=text]", cmd.Usage())
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())

	// Test HelpInfo
	helpInfo := cmd.HelpInfo()
	assert.Equal(t, "exit", helpInfo.Command)
	assert.Equal(t, "Exit the shell with optional exit code and message", helpInfo.Description)
	assert.Equal(t, "\\exit[code=N, message=text]", helpInfo.Usage)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)
	assert.NotEmpty(t, helpInfo.Options)

	// Test that we have the expected options
	assert.Len(t, helpInfo.Options, 2)

	// Find code and message options
	var codeOption, messageOption *neurotypes.HelpOption
	for i, opt := range helpInfo.Options {
		switch opt.Name {
		case "code":
			codeOption = &helpInfo.Options[i]
		case "message":
			messageOption = &helpInfo.Options[i]
		}
	}

	assert.NotNil(t, codeOption, "Should have code option")
	assert.NotNil(t, messageOption, "Should have message option")

	if codeOption != nil {
		assert.Equal(t, "int", codeOption.Type)
		assert.Equal(t, "0", codeOption.Default)
		assert.False(t, codeOption.Required)
	}

	if messageOption != nil {
		assert.Equal(t, "string", messageOption.Type)
		assert.False(t, messageOption.Required)
	}
}

// BenchmarkExitCommand tests the performance characteristics of the exit command
// Note: This doesn't actually call Execute since that would terminate the process
func BenchmarkExitCommand_Metadata(b *testing.B) {
	cmd := &ExitCommand{}

	b.Run("Name", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cmd.Name()
		}
	})

	b.Run("Description", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cmd.Description()
		}
	})

	b.Run("Usage", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cmd.Usage()
		}
	})

	b.Run("ParseMode", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cmd.ParseMode()
		}
	})
}

// TestExitCommand_Interface tests that ExitCommand properly implements the Command interface
func TestExitCommand_Interface(t *testing.T) {
	var _ neurotypes.Command = &ExitCommand{}

	cmd := &ExitCommand{}

	// Test all interface methods return reasonable values
	assert.NotEmpty(t, cmd.Name())
	assert.NotEmpty(t, cmd.Description())
	assert.NotEmpty(t, cmd.Usage())

	// ParseMode should be valid
	mode := cmd.ParseMode()
	assert.True(t, mode == neurotypes.ParseModeKeyValue || mode == neurotypes.ParseModeRaw)
}

// TestExitCommand_ConsistentMetadata tests that metadata methods return consistent values
func TestExitCommand_ConsistentMetadata(t *testing.T) {
	cmd := &ExitCommand{}

	// Test that multiple calls return the same values
	name1 := cmd.Name()
	name2 := cmd.Name()
	assert.Equal(t, name1, name2)

	desc1 := cmd.Description()
	desc2 := cmd.Description()
	assert.Equal(t, desc1, desc2)

	usage1 := cmd.Usage()
	usage2 := cmd.Usage()
	assert.Equal(t, usage1, usage2)

	mode1 := cmd.ParseMode()
	mode2 := cmd.ParseMode()
	assert.Equal(t, mode1, mode2)
}
