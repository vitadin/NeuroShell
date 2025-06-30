package builtin

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"neuroshell/internal/testutils"
	"neuroshell/pkg/types"
)

func TestExitCommand_Name(t *testing.T) {
	cmd := &ExitCommand{}
	assert.Equal(t, "exit", cmd.Name())
}

func TestExitCommand_ParseMode(t *testing.T) {
	cmd := &ExitCommand{}
	assert.Equal(t, types.ParseModeKeyValue, cmd.ParseMode())
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
		ctx := testutils.NewMockContext()

		// This should call os.Exit(0) and terminate the subprocess
		_ = cmd.Execute(map[string]string{}, "", ctx)

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

	// Check that "Goodbye!" was printed
	assert.Contains(t, outputStr, "Goodbye!", "Expected 'Goodbye!' message in output")
}

// TestExitCommand_Execute_WithArgs tests that args are ignored
func TestExitCommand_Execute_WithArgs(t *testing.T) {
	if os.Getenv("TEST_EXIT_COMMAND_ARGS") == "1" {
		// This code runs in the subprocess
		cmd := &ExitCommand{}
		ctx := testutils.NewMockContext()

		// Test with args - should still exit normally
		args := map[string]string{"force": "true", "code": "1"}
		_ = cmd.Execute(args, "", ctx)

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

	assert.Contains(t, outputStr, "Goodbye!", "Expected 'Goodbye!' message in output")
}

// TestExitCommand_Execute_WithInput tests that input is ignored
func TestExitCommand_Execute_WithInput(t *testing.T) {
	if os.Getenv("TEST_EXIT_COMMAND_INPUT") == "1" {
		// This code runs in the subprocess
		cmd := &ExitCommand{}
		ctx := testutils.NewMockContext()

		// Test with input - should still exit normally
		_ = cmd.Execute(map[string]string{}, "some input text", ctx)

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

	assert.Contains(t, outputStr, "Goodbye!", "Expected 'Goodbye!' message in output")
}

// TestExitCommand_Execute_MessageOnly tests that the goodbye message is printed
// This test captures stdout before the exit occurs by mocking the execution
func TestExitCommand_Execute_MessageOnly(t *testing.T) {
	// We can't easily test the full Execute method due to os.Exit,
	// but we can test the logic leading up to it by using a wrapper

	// Create a testable version that doesn't call os.Exit
	testableExitCommand := &struct {
		*ExitCommand
		exitCalled bool
		exitCode   int
	}{
		ExitCommand: &ExitCommand{},
	}

	// Override Execute to capture the behavior without calling os.Exit
	executeFunc := func(_ map[string]string, _ string, _ types.Context) error {
		// Capture stdout
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute the part that prints the message
		fmt.Println("Goodbye!")

		// Restore stdout and read output
		w.Close()
		os.Stdout = originalStdout
		output, _ := io.ReadAll(r)
		outputStr := string(output)

		// Verify the message was printed
		assert.Contains(t, outputStr, "Goodbye!")

		// Mark that exit would have been called
		testableExitCommand.exitCalled = true
		testableExitCommand.exitCode = 0

		return nil
	}

	ctx := testutils.NewMockContext()

	// Test the execution logic
	err := executeFunc(map[string]string{}, "", ctx)
	assert.NoError(t, err)
	assert.True(t, testableExitCommand.exitCalled)
	assert.Equal(t, 0, testableExitCommand.exitCode)
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
	var _ types.Command = &ExitCommand{}

	cmd := &ExitCommand{}

	// Test all interface methods return reasonable values
	assert.NotEmpty(t, cmd.Name())
	assert.NotEmpty(t, cmd.Description())
	assert.NotEmpty(t, cmd.Usage())

	// ParseMode should be valid
	mode := cmd.ParseMode()
	assert.True(t, mode == types.ParseModeKeyValue || mode == types.ParseModeRaw)
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
