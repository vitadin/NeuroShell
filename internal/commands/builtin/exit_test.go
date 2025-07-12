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
	assert.Equal(t, "Exit the shell", cmd.Description())
}

func TestExitCommand_Usage(t *testing.T) {
	cmd := &ExitCommand{}
	assert.Equal(t, "\\exit", cmd.Usage())
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

// TestExitCommand_Execute_WithArgs tests that args are ignored
func TestExitCommand_Execute_WithArgs(t *testing.T) {
	if os.Getenv("TEST_EXIT_COMMAND_ARGS") == "1" {
		// This code runs in the subprocess
		cmd := &ExitCommand{}

		// Test with args - should still exit normally
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

	// Should still exit with code 0 (args are ignored)
	if e, ok := err.(*exec.ExitError); ok {
		assert.Equal(t, 0, e.ExitCode(), "Expected exit code 0, got %d. Output: %s", e.ExitCode(), outputStr)
	} else {
		assert.NoError(t, err, "Expected successful exit (code 0)")
	}

	// Exit command should not print any message - it just exits silently
	// We just verify that the subprocess exited cleanly with code 0
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

// TestExitCommand_Execute_MessageOnly tests the basic metadata without calling os.Exit
// Since the exit command doesn't print any messages, we just test interface compliance
func TestExitCommand_Execute_MessageOnly(t *testing.T) {
	// We can't easily test the full Execute method due to os.Exit,
	// but we can test the command interface methods

	cmd := &ExitCommand{}

	// Test that all interface methods work correctly
	assert.Equal(t, "exit", cmd.Name())
	assert.Equal(t, "Exit the shell", cmd.Description())
	assert.Equal(t, "\\exit", cmd.Usage())
	assert.Equal(t, neurotypes.ParseModeKeyValue, cmd.ParseMode())

	// Test HelpInfo
	helpInfo := cmd.HelpInfo()
	assert.Equal(t, "exit", helpInfo.Command)
	assert.Equal(t, "Exit the shell", helpInfo.Description)
	assert.Equal(t, "\\exit", helpInfo.Usage)
	assert.NotEmpty(t, helpInfo.Examples)
	assert.NotEmpty(t, helpInfo.Notes)
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
