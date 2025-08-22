// Package cli provides command-line interface setup for neurotest.
package cli

import (
	"fmt"
	"os"

	"neuroshell/cmd/neurotest/internal/experiments"
	"neuroshell/cmd/neurotest/internal/golden"
	"neuroshell/cmd/neurotest/internal/neurorc"

	"github.com/spf13/cobra"
)

// addGoldenFileCommands adds golden file testing commands
func (app *App) addGoldenFileCommands(rootCmd *cobra.Command) {
	// Record command
	recordCmd := &cobra.Command{
		Use:   "record <testname>",
		Short: "Record a new test case",
		Long: `Record a new test case by running a .neuro script and capturing its output.
The output will be saved as a golden file for future comparisons.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			recorder := golden.NewRecorder(app.Config)
			return recorder.RecordTest(args[0])
		},
	}

	// Run command
	runCmd := &cobra.Command{
		Use:   "run <testname>",
		Short: "Run a specific test case",
		Long: `Run a specific test case and compare its output with the expected golden file.
Returns exit code 0 if the test passes, non-zero if it fails.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			runner := golden.NewRunner(app.Config)
			return runner.RunTest(args[0])
		},
	}

	// Run all command
	runAllCmd := &cobra.Command{
		Use:   "run-all",
		Short: "Run all test cases",
		Long: `Run all test cases in the test directory and report the results.
Returns exit code 0 if all tests pass, non-zero if any fail.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			runner := golden.NewRunner(app.Config)
			return runner.RunAllTests()
		},
	}

	// Accept command
	acceptCmd := &cobra.Command{
		Use:   "accept <testname>",
		Short: "Accept current output as golden",
		Long: `Update the golden file for a test case with the current output.
Use this after verifying that the new behavior is correct.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			recorder := golden.NewRecorder(app.Config)
			return recorder.AcceptTest(args[0])
		},
	}

	// Diff command
	diffCmd := &cobra.Command{
		Use:   "diff <testname>",
		Short: "Show differences between expected and actual output",
		Long: `Show detailed differences between the expected golden file output
and the actual output from running the test.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			differ := golden.NewDiffer(app.Config)
			return differ.ShowDiff(args[0])
		},
	}

	rootCmd.AddCommand(recordCmd, runCmd, runAllCmd, acceptCmd, diffCmd)
}

// addExperimentCommands adds experiment-related commands
func (app *App) addExperimentCommands(rootCmd *cobra.Command) {
	// Record experiment command
	recordExperimentCmd := &cobra.Command{
		Use:   "record-experiment <experiment-name>",
		Short: "Record a real-world experiment execution",
		Long: `Record a real-world experiment by running a .neuro script from examples/experiments/
with actual LLM API calls (no test mode). Creates uniquely identified recordings
with timestamps and metadata in experiments/recordings/<experiment-name>/`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			recorder := experiments.NewRecorder(app.Config)
			return recorder.RecordExperiment(args[0])
		},
	}

	// Run experiment command
	runExperimentCmd := &cobra.Command{
		Use:   "run-experiment <experiment-name> <session-id>",
		Short: "Run an experiment and compare with a specific recording",
		Long: `Run an experiment and compare its output with a specific recording.
Use this to verify reproducibility or debug experiment behavior.`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			runner := experiments.NewRunner(app.Config)
			return runner.RunExperiment(args[0], args[1])
		},
	}

	// Record all experiments command
	recordAllExperimentsCmd := &cobra.Command{
		Use:   "record-all-experiments",
		Short: "Record all available experiments",
		Long: `Record all experiments found in examples/experiments/ directory.
Each experiment will be run with actual LLM API calls and recorded
with unique session IDs and metadata.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			recorder := experiments.NewRecorder(app.Config)
			return recorder.RecordAllExperiments()
		},
	}

	rootCmd.AddCommand(recordExperimentCmd, runExperimentCmd, recordAllExperimentsCmd)
}

// addNeuroRCCommands adds .neurorc startup testing commands
func (app *App) addNeuroRCCommands(rootCmd *cobra.Command) {
	// Record .neurorc command
	recordNeuroRCCmd := &cobra.Command{
		Use:   "record-neurorc <testname>",
		Short: "Record a .neurorc startup test case",
		Long: `Record a .neurorc startup test case by running the shell with specified
.neurorc configuration and capturing the startup behavior. This tests the
auto-startup functionality in an isolated environment.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return app.recordNeuroRCTest(args[0])
		},
	}

	// Run .neurorc command
	runNeuroRCCmd := &cobra.Command{
		Use:   "run-neurorc <testname>",
		Short: "Run a .neurorc startup test case",
		Long: `Run a .neurorc startup test case and compare its behavior with the expected
golden file. This tests shell startup behavior and .neurorc auto-execution.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			runner := neurorc.NewRunner(app.Config)
			return runner.RunTest(args[0])
		},
	}

	// Diff .neurorc command
	diffNeuroRCCmd := &cobra.Command{
		Use:   "diff-neurorc <testname>",
		Short: "Show differences in .neurorc startup test output",
		Long: `Show detailed differences between the expected and actual output
from a .neurorc startup test case.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return app.diffNeuroRCTest(args[0])
		},
	}

	rootCmd.AddCommand(recordNeuroRCCmd, runNeuroRCCmd, diffNeuroRCCmd)
}

// recordNeuroRCTest records a new .neurorc startup test case
func (app *App) recordNeuroRCTest(testName string) error {
	if app.Config.Verbose {
		fmt.Printf("Recording .neurorc startup test: %s\n", testName)
	}

	configFile, err := neurorc.FindTestConfig(testName)
	if err != nil {
		return fmt.Errorf("failed to find test config: %w", err)
	}

	testConfig, err := neurorc.ParseTestConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to parse test config: %w", err)
	}

	// Create isolated test environment
	testEnv, err := neurorc.CreateTestEnvironment(testConfig)
	if err != nil {
		return fmt.Errorf("failed to create test environment: %w", err)
	}
	defer neurorc.CleanupTestEnvironment(testEnv)

	// Run the test and capture output
	actualOutput, err := neurorc.RunShellTest(testEnv, testConfig)
	if err != nil {
		return fmt.Errorf("failed to run shell test: %w", err)
	}

	actualOutput = neurorc.CleanOutput(actualOutput)

	// Save the expected output
	expectedPath := fmt.Sprintf("%s/neurorc/%s.expected", app.Config.TestDir, testName)
	if err := os.WriteFile(expectedPath, []byte(actualOutput), 0644); err != nil {
		return fmt.Errorf("failed to write expected file: %w", err)
	}

	if app.Config.Verbose {
		fmt.Printf("Recorded expected output for .neurorc test: %s\n", testName)
	}

	return nil
}

// diffNeuroRCTest shows differences in .neurorc startup test output
func (app *App) diffNeuroRCTest(testName string) error {
	if app.Config.Verbose {
		fmt.Printf("Showing diff for .neurorc startup test: %s\n", testName)
	}

	configFile, err := neurorc.FindTestConfig(testName)
	if err != nil {
		return fmt.Errorf("failed to find test config: %w", err)
	}

	testConfig, err := neurorc.ParseTestConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to parse test config: %w", err)
	}

	// Read expected output
	expectedPath := fmt.Sprintf("%s/neurorc/%s.expected", app.Config.TestDir, testName)
	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		return fmt.Errorf("failed to read expected file %s: %w", expectedPath, err)
	}
	expected := string(expectedBytes)

	// Create isolated test environment
	testEnv, err := neurorc.CreateTestEnvironment(testConfig)
	if err != nil {
		return fmt.Errorf("failed to create test environment: %w", err)
	}
	defer neurorc.CleanupTestEnvironment(testEnv)

	// Run the test
	actual, err := neurorc.RunShellTest(testEnv, testConfig)
	if err != nil {
		return fmt.Errorf("failed to run .neurorc startup test: %w", err)
	}

	// Show enhanced diff using golden differ
	differ := golden.NewDiffer(app.Config)
	differ.ShowDetailedDiff(expected, actual, testName)

	return nil
}

// addCFlagCommands adds -c flag testing commands
func (app *App) addCFlagCommands(rootCmd *cobra.Command) {
	// Record -c flag test command
	recordCCmd := &cobra.Command{
		Use:   "record-c <testname>",
		Short: "Record a test case using -c flag",
		Long:  `Record a test case by running neuro with -c flag and save output as .c.expected file`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			recorder := golden.NewRecorder(app.Config)
			return recorder.RecordCFlagTest(args[0])
		},
	}

	// Run -c flag test command
	runCCmd := &cobra.Command{
		Use:   "run-c <testname>",
		Short: "Run a test case using -c flag",
		Long:  `Run a test case using -c flag and compare with .c.expected file`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			runner := golden.NewRunner(app.Config)
			return runner.RunCFlagTest(args[0])
		},
	}

	// Compare modes command
	compareModesCmd := &cobra.Command{
		Use:   "compare-modes <testname>",
		Short: "Compare batch mode vs -c flag output",
		Long:  `Compare the output of batch mode (.expected) with -c flag mode (.c.expected)`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return app.compareModes(args[0])
		},
	}

	rootCmd.AddCommand(recordCCmd, runCCmd, compareModesCmd)
}

// compareModes compares batch mode output (.expected) with -c flag output (.c.expected)
func (app *App) compareModes(testName string) error {
	if app.Config.Verbose {
		fmt.Printf("Comparing modes for test: %s\n", testName)
	}

	// Read batch mode expected output
	batchExpectedPath := fmt.Sprintf("%s/%s.expected", app.Config.TestDir, testName)
	batchExpected, err := os.ReadFile(batchExpectedPath)
	if err != nil {
		return fmt.Errorf("failed to read batch expected file %s: %w", batchExpectedPath, err)
	}

	// Read -c flag expected output
	cExpectedPath := fmt.Sprintf("%s/%s.c.expected", app.Config.TestDir, testName)
	cExpected, err := os.ReadFile(cExpectedPath)
	if err != nil {
		return fmt.Errorf("failed to read -c expected file %s: %w", cExpectedPath, err)
	}

	batchOutput := string(batchExpected)
	cOutput := string(cExpected)

	// Simple comparison - exact match required
	if batchOutput == cOutput {
		fmt.Printf("✅ IDENTICAL: batch mode and -c flag produce identical output for %s\n", testName)
		return nil
	}

	// Show differences using the golden differ
	fmt.Printf("❌ DIFFERENT: batch mode and -c flag produce different output for %s\n", testName)
	fmt.Println("\n=== Batch Mode Output (.expected) ===")
	fmt.Print(batchOutput)
	fmt.Println("\n=== -c Flag Output (.c.expected) ===")
	fmt.Print(cOutput)
	fmt.Println("\n=== Detailed Differences ===")

	differ := golden.NewDiffer(app.Config)
	differ.ShowDetailedDiff(batchOutput, cOutput, testName)

	return fmt.Errorf("outputs differ between batch mode and -c flag")
}
