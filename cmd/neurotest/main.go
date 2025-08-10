// Package main provides the neurotest CLI application for end-to-end testing of NeuroShell.
// neurotest uses golden files to record, run, and verify expected behavior of Neuro CLI commands.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"neuroshell/internal/version"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var (
	testDir             = "test/golden"
	neurocmd            = "neuro" // Will be resolved by checkNeuroCommand()
	verbose             bool
	testTimeout         = 30 // seconds
	normalizationEngine *NormalizationEngine
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "neurotest",
	Short: "End-to-end testing tool for Neuro CLI",
	Long: `neurotest is a testing tool for the Neuro CLI that uses golden files to
verify expected behavior. It can record, run, and verify test cases.`,
}

// recordCmd records a new test case
var recordCmd = &cobra.Command{
	Use:   "record <testname>",
	Short: "Record a new test case",
	Long: `Record a new test case by running a .neuro script and capturing its output.
The output will be saved as a golden file for future comparisons.`,
	Args: cobra.ExactArgs(1),
	RunE: recordTest,
}

// runCmd runs a specific test case
var runCmd = &cobra.Command{
	Use:   "run <testname>",
	Short: "Run a specific test case",
	Long: `Run a specific test case and compare its output with the expected golden file.
Returns exit code 0 if the test passes, non-zero if it fails.`,
	Args: cobra.ExactArgs(1),
	RunE: runTest,
}

// runAllCmd runs all test cases
var runAllCmd = &cobra.Command{
	Use:   "run-all",
	Short: "Run all test cases",
	Long: `Run all test cases in the test directory and report the results.
Returns exit code 0 if all tests pass, non-zero if any fail.`,
	RunE: runAllTests,
}

// acceptCmd updates the golden file with current output
var acceptCmd = &cobra.Command{
	Use:   "accept <testname>",
	Short: "Accept current output as golden",
	Long: `Update the golden file for a test case with the current output.
Use this after verifying that the new behavior is correct.`,
	Args: cobra.ExactArgs(1),
	RunE: acceptTest,
}

// diffCmd shows differences between expected and actual output
var diffCmd = &cobra.Command{
	Use:   "diff <testname>",
	Short: "Show differences between expected and actual output",
	Long: `Show detailed differences between the expected golden file output
and the actual output from running the test.`,
	Args: cobra.ExactArgs(1),
	RunE: diffTest,
}

// versionCmd shows version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version of neurotest with build information.`,
	Run: func(cmd *cobra.Command, _ []string) {
		detailed, _ := cmd.Flags().GetBool("detailed")
		if detailed {
			fmt.Printf("neurotest %s\n", version.GetDetailedVersion())
		} else {
			fmt.Printf("neurotest %s\n", version.GetFormattedVersion())
		}
	},
}

// recordExperimentCmd records a new experiment execution
var recordExperimentCmd = &cobra.Command{
	Use:   "record-experiment <experiment-name>",
	Short: "Record a real-world experiment execution",
	Long: `Record a real-world experiment by running a .neuro script from examples/experiments/
with actual LLM API calls (no test mode). Creates uniquely identified recordings
with timestamps and metadata in experiments/recordings/<experiment-name>/`,
	Args: cobra.ExactArgs(1),
	RunE: recordExperiment,
}

// runExperimentCmd runs a specific experiment and compares with a recording
var runExperimentCmd = &cobra.Command{
	Use:   "run-experiment <experiment-name> <session-id>",
	Short: "Run an experiment and compare with a specific recording",
	Long: `Run an experiment and compare its output with a specific recording.
Use this to verify reproducibility or debug experiment behavior.`,
	Args: cobra.ExactArgs(2),
	RunE: runExperiment,
}

// recordAllExperimentsCmd records all available experiments
var recordAllExperimentsCmd = &cobra.Command{
	Use:   "record-all-experiments",
	Short: "Record all available experiments",
	Long: `Record all experiments found in examples/experiments/ directory.
Each experiment will be run with actual LLM API calls and recorded
with unique session IDs and metadata.`,
	RunE: recordAllExperiments,
}

// recordNeuroRCCmd records a new .neurorc startup test case
var recordNeuroRCCmd = &cobra.Command{
	Use:   "record-neurorc <testname>",
	Short: "Record a .neurorc startup test case",
	Long: `Record a .neurorc startup test case by running the shell with specified
.neurorc configuration and capturing the startup behavior. This tests the
auto-startup functionality in an isolated environment.`,
	Args: cobra.ExactArgs(1),
	RunE: recordNeuroRCTest,
}

// runNeuroRCCmd runs a specific .neurorc startup test case
var runNeuroRCCmd = &cobra.Command{
	Use:   "run-neurorc <testname>",
	Short: "Run a .neurorc startup test case",
	Long: `Run a .neurorc startup test case and compare its behavior with the expected
golden file. This tests shell startup behavior and .neurorc auto-execution.`,
	Args: cobra.ExactArgs(1),
	RunE: runNeuroRCTest,
}

// diffNeuroRCCmd shows differences in .neurorc startup test output
var diffNeuroRCCmd = &cobra.Command{
	Use:   "diff-neurorc <testname>",
	Short: "Show differences in .neurorc startup test output",
	Long: `Show detailed differences between the expected and actual output
from a .neurorc startup test case.`,
	Args: cobra.ExactArgs(1),
	RunE: diffNeuroRCTest,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Initialize normalization engine
	normalizationEngine = NewNormalizationEngine()

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&testDir, "test-dir", "test/golden", "Test directory")
	rootCmd.PersistentFlags().StringVar(&neurocmd, "neuro-cmd", "neuro", "Neuro command to test (will try ./bin/neuro, then PATH)")
	rootCmd.PersistentFlags().IntVar(&testTimeout, "timeout", 30, "Test timeout in seconds")

	// Add version command flags
	versionCmd.Flags().Bool("detailed", false, "Show detailed version information")

	// Add subcommands
	rootCmd.AddCommand(recordCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(runAllCmd)
	rootCmd.AddCommand(acceptCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(versionCmd)

	// Add experiment commands
	rootCmd.AddCommand(recordExperimentCmd)
	rootCmd.AddCommand(runExperimentCmd)
	rootCmd.AddCommand(recordAllExperimentsCmd)

	// Add .neurorc startup test commands
	rootCmd.AddCommand(recordNeuroRCCmd)
	rootCmd.AddCommand(runNeuroRCCmd)
	rootCmd.AddCommand(diffNeuroRCCmd)
}

// recordTest records a new test case by running the .neuro script
func recordTest(_ *cobra.Command, args []string) error {
	testName := args[0]

	if verbose {
		fmt.Printf("Recording test case: %s\n", testName)
	}

	// Check if neuro command exists
	if err := checkNeuroCommand(); err != nil {
		return err
	}

	// Find the .neuro script file
	scriptPath, err := findNeuroScript(testName)
	if err != nil {
		return err
	}

	// Run the neuro command and capture output
	output, err := runNeuroScript(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to run neuro script: %w", err)
	}

	// Save the output as expected result (flat structure only)
	expectedFile := filepath.Join(testDir, testName+".expected")
	if err := os.MkdirAll(filepath.Dir(expectedFile), 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	if err := os.WriteFile(expectedFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write expected output: %w", err)
	}

	fmt.Printf("Recorded test case: %s\n", testName)
	fmt.Printf("Expected output saved to: %s\n", expectedFile)

	return nil
}

// runTest runs a specific test case
func runTest(_ *cobra.Command, args []string) error {
	testName := args[0]

	if verbose {
		fmt.Printf("Running test case: %s\n", testName)
	}

	// Check if neuro command exists
	if err := checkNeuroCommand(); err != nil {
		return err
	}

	// Find the .neuro script file
	scriptPath, err := findNeuroScript(testName)
	if err != nil {
		return err
	}

	// Check if expected file exists (flat structure only)
	expectedFile := filepath.Join(testDir, testName+".expected")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		return fmt.Errorf("expected output file not found: %s\nRun 'neurotest record %s' first", expectedFile, testName)
	}

	// Read expected output
	expectedBytes, err := os.ReadFile(expectedFile)
	if err != nil {
		return fmt.Errorf("failed to read expected output: %w", err)
	}
	expected := string(expectedBytes)

	// Run the neuro command and capture output
	actual, err := runNeuroScript(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to run neuro script: %w", err)
	}

	// Smart comparison with normalization and placeholders
	var passed bool
	var normalized bool

	// First try exact match (for backward compatibility)
	if strings.TrimSpace(actual) == strings.TrimSpace(expected) {
		passed = true
	} else {
		// Try smart comparison with placeholders
		if normalizationEngine.CompareWithPlaceholders(expected, actual) {
			passed = true
			normalized = true
		} else {
			// Try normalized comparison (replace known patterns with placeholders)
			normalizedExpected := normalizationEngine.NormalizeOutput(expected)
			normalizedActual := normalizationEngine.NormalizeOutput(actual)
			if strings.TrimSpace(normalizedActual) == strings.TrimSpace(normalizedExpected) {
				passed = true
				normalized = true
			}
		}
	}

	if passed {
		if normalized && verbose {
			fmt.Printf("PASS: %s (using smart comparison)\n", testName)
		} else {
			fmt.Printf("PASS: %s\n", testName)
		}
		return nil
	}

	fmt.Printf("FAIL: %s\n", testName)
	showDetailedDiff(expected, actual, testName)
	return fmt.Errorf("test failed: output mismatch")
}

// runAllTests runs all test cases
func runAllTests(cmd *cobra.Command, _ []string) error {
	if verbose {
		fmt.Println("Running all test cases...")
	}

	// Find all test directories
	testDirs, err := findAllTests()
	if err != nil {
		return err
	}

	if len(testDirs) == 0 {
		fmt.Println("No test cases found")
		return nil
	}

	passed := 0
	failed := 0

	for _, testName := range testDirs {
		fmt.Printf("Running %s... ", testName)
		if err := runTest(cmd, []string{testName}); err != nil {
			fmt.Println("FAIL")
			if verbose {
				fmt.Printf("  Error: %v\n", err)
			}
			failed++
		} else {
			fmt.Println("PASS")
			passed++
		}
	}

	fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)

	if failed > 0 {
		return fmt.Errorf("%d test(s) failed", failed)
	}

	return nil
}

// acceptTest updates the golden file with current output
func acceptTest(cmd *cobra.Command, args []string) error {
	testName := args[0]

	if verbose {
		fmt.Printf("Accepting current output for test: %s\n", testName)
	}

	// This is essentially the same as record, but with different messaging
	return recordTest(cmd, args)
}

// diffTest shows differences between expected and actual output
func diffTest(_ *cobra.Command, args []string) error {
	testName := args[0]

	if verbose {
		fmt.Printf("Showing diff for test case: %s\n", testName)
	}

	// Check if neuro command exists
	if err := checkNeuroCommand(); err != nil {
		return err
	}

	// Find the .neuro script file
	scriptPath, err := findNeuroScript(testName)
	if err != nil {
		return err
	}

	// Check if expected file exists (flat structure only)
	expectedFile := filepath.Join(testDir, testName+".expected")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		return fmt.Errorf("expected output file not found: %s", expectedFile)
	}

	// Read expected output
	expectedBytes, err := os.ReadFile(expectedFile)
	if err != nil {
		return fmt.Errorf("failed to read expected output: %w", err)
	}
	expected := string(expectedBytes)

	// Run the neuro command and capture output
	actual, err := runNeuroScript(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to run neuro script: %w", err)
	}

	// Show enhanced diff
	showDetailedDiff(expected, actual, testName)

	return nil
}

// showDetailedDiff displays an enhanced diff using go-diff library
func showDetailedDiff(expected, actual, testName string) {
	// Check for smart comparison results
	exactMatch := strings.TrimSpace(actual) == strings.TrimSpace(expected)
	placeholderMatch := normalizationEngine.CompareWithPlaceholders(expected, actual)

	normalizedExpected := normalizationEngine.NormalizeOutput(expected)
	normalizedActual := normalizationEngine.NormalizeOutput(actual)
	normalizedMatch := strings.TrimSpace(normalizedActual) == strings.TrimSpace(normalizedExpected)

	fmt.Printf("=== Comparison Results for %s ===\n", testName)
	fmt.Printf("Exact match: %t\n", exactMatch)
	fmt.Printf("Placeholder match: %t\n", placeholderMatch)
	fmt.Printf("Normalized match: %t\n", normalizedMatch)
	fmt.Println()

	if exactMatch {
		fmt.Println("=== No differences found ===")
		return
	}

	// Show normalized versions if different from original
	if normalizedExpected != expected || normalizedActual != actual {
		fmt.Println("=== Normalized Expected ===")
		fmt.Println(normalizedExpected)
		fmt.Println("=== Normalized Actual ===")
		fmt.Println(normalizedActual)
		fmt.Println()
	}

	// Create diff using go-diff
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(expected, actual, false)

	fmt.Println("=== Detailed Diff ===")
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			fmt.Print(diff.Text)
		case diffmatchpatch.DiffDelete:
			fmt.Printf("[-]%s", diff.Text)
		case diffmatchpatch.DiffInsert:
			fmt.Printf("[+]%s", diff.Text)
		}
	}
	fmt.Println()

	// Show line-by-line comparison for clarity
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	fmt.Println("=== Line by Line Comparison ===")
	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := 0; i < maxLines; i++ {
		expectedLine := ""
		actualLine := ""

		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine != actualLine {
			fmt.Printf("Line %d:\n", i+1)
			fmt.Printf("  Expected: %q\n", expectedLine)
			fmt.Printf("  Actual:   %q\n", actualLine)

			// Check if this line would match with placeholders
			if normalizationEngine.MatchLineWithPlaceholders(expectedLine, actualLine) {
				fmt.Printf("  (Matches with placeholder support)\n")
			}
			fmt.Println()
		}
	}
}

// Helper functions

// checkNeuroCommand verifies that the neuro command is available
func checkNeuroCommand() error {
	originalCmd := neurocmd

	// If user explicitly provided a command other than "neuro", try it as-is first
	if originalCmd != "neuro" {
		// First try the provided command as-is (handles absolute paths)
		if filepath.IsAbs(originalCmd) {
			if _, err := os.Stat(originalCmd); err == nil {
				return nil
			}
			return fmt.Errorf("neuro command not found: %s", originalCmd)
		}

		// For relative paths, try to find it relative to current working directory
		if _, err := os.Stat(originalCmd); err == nil {
			return nil
		}
	}

	// Try common locations for the neuro binary
	candidates := []string{
		"./bin/neuro", // Local build
		"bin/neuro",   // Local build (alternative path)
		"neuro",       // PATH lookup
	}

	// If user provided a specific command, add it to candidates
	if originalCmd != "neuro" && originalCmd != "./bin/neuro" {
		candidates = append([]string{originalCmd}, candidates...)
	}

	// Try to resolve relative to project root if we're in a subdirectory
	if cwd, err := os.Getwd(); err == nil {
		searchDir := cwd
		for {
			// Check if we're in the NeuroShell project root
			if _, err := os.Stat(filepath.Join(searchDir, "go.mod")); err == nil {
				// Add project-relative candidates
				for _, candidate := range []string{"bin/neuro", "./bin/neuro"} {
					projectPath := filepath.Join(searchDir, candidate)
					candidates = append(candidates, projectPath)
				}
				break
			}
			parent := filepath.Dir(searchDir)
			if parent == searchDir {
				break // Reached filesystem root
			}
			searchDir = parent
		}
	}

	// Try each candidate
	var triedPaths []string
	for _, candidate := range candidates {
		triedPaths = append(triedPaths, candidate)

		// Try as absolute path
		if filepath.IsAbs(candidate) {
			if _, err := os.Stat(candidate); err == nil {
				neurocmd = candidate
				return nil
			}
			continue
		}

		// Try as relative path
		if _, err := os.Stat(candidate); err == nil {
			// Convert to absolute path for consistency
			if absPath, err := filepath.Abs(candidate); err == nil {
				neurocmd = absPath
			} else {
				neurocmd = candidate
			}
			return nil
		}

		// Try using exec.LookPath for PATH resolution (for non-path candidates)
		if !strings.Contains(candidate, "/") {
			if resolvedPath, err := exec.LookPath(candidate); err == nil {
				neurocmd = resolvedPath
				return nil
			}
		}
	}

	return fmt.Errorf("neuro command not found\nTried paths: %v\nMake sure the neuro binary exists and is executable", triedPaths)
}

// findNeuroScript finds the .neuro script file for a test
func findNeuroScript(testName string) (string, error) {
	// Try different possible locations (flat structure only)
	candidates := []string{
		filepath.Join(testDir, testName+".neuro"),        // Flat structure
		filepath.Join("test/scripts", testName+".neuro"), // Alternative scripts location
		testName + ".neuro",                              // Current directory
		testName,                                         // As provided
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("neuro script not found for test: %s\nTried: %v", testName, candidates)
}

// runNeuroScript executes a .neuro script and returns its output
func runNeuroScript(scriptPath string) (string, error) {
	// Convert to absolute path for consistent execution
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Use neuro batch command to run the script
	cmd := exec.Command(neurocmd, "batch", absPath, "--test-mode", "--no-color")

	// Set environment variables for consistent testing
	cmd.Env = append(os.Environ(),
		"NEURO_LOG_LEVEL=fatal", // Minimize log noise in tests
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	// Clean up the output by removing shell startup messages
	cleanOutput := cleanNeuroOutput(string(output))
	return cleanOutput, nil
}

// findAllTests finds all available test cases
func findAllTests() ([]string, error) {
	var tests []string
	testSet := make(map[string]bool) // To avoid duplicates

	// Look for entries in testDir
	entries, err := os.ReadDir(testDir)
	if err != nil {
		if os.IsNotExist(err) {
			return tests, nil // No test directory yet
		}
		return nil, fmt.Errorf("failed to read test directory: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".neuro") {
			// Check for flat structure .neuro files
			testName := strings.TrimSuffix(entry.Name(), ".neuro")
			// Verify corresponding .expected file exists (flat structure only)
			expectedFile := filepath.Join(testDir, testName+".expected")
			if _, err := os.Stat(expectedFile); err == nil {
				if !testSet[testName] {
					tests = append(tests, testName)
					testSet[testName] = true
				}
			}
		}
	}

	return tests, nil
}

// cleanNeuroOutput removes shell startup messages and other noise from output
func cleanNeuroOutput(output string) string {
	lines := strings.Split(output, "\n")
	var cleanLines []string

	skipPatterns := []string{
		"INFO ",
		"DEBUG ",
		"WARN ",
		"ERROR ",
		"Neuro Shell v",
		"Type '\\help' for Neuro commands",
		"Goodbye!",
	}

	for _, line := range lines {
		shouldSkip := false
		for _, pattern := range skipPatterns {
			if strings.Contains(line, pattern) {
				shouldSkip = true
				break
			}
		}

		if !shouldSkip && strings.TrimSpace(line) != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// recordExperiment records a new experiment execution with real API calls
func recordExperiment(_ *cobra.Command, args []string) error {
	experimentName := args[0]

	if verbose {
		fmt.Printf("Recording experiment: %s\n", experimentName)
	}

	// Check if neuro command exists
	if err := checkNeuroCommand(); err != nil {
		return err
	}

	// Find the experiment script
	scriptPath, err := findExperimentScript(experimentName)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("Found experiment script: %s\n", scriptPath)
	}

	// Run the experiment (with real API calls)
	output, duration, err := runExperimentScript(scriptPath)
	exitCode := 0
	if err != nil {
		exitCode = 1
		// Don't return error immediately - we still want to save the recording
		if verbose {
			fmt.Printf("Experiment failed: %v\n", err)
		}
	}

	// Save the experiment recording
	sessionID, err := saveExperimentRecording(experimentName, scriptPath, output, duration, exitCode)
	if err != nil {
		return fmt.Errorf("failed to save experiment recording: %w", err)
	}

	fmt.Printf("Recorded experiment: %s\n", experimentName)
	fmt.Printf("Session ID: %s\n", sessionID)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Recording saved to: experiments/recordings/%s/%s.expected\n", experimentName, sessionID)

	// If the experiment failed, return error after saving
	if exitCode != 0 {
		return fmt.Errorf("experiment execution failed but recording was saved")
	}

	return nil
}

// runExperiment runs an experiment and compares with a specific recording
func runExperiment(_ *cobra.Command, args []string) error {
	experimentName := args[0]
	sessionID := args[1]

	if verbose {
		fmt.Printf("Running experiment: %s (comparing with session: %s)\n", experimentName, sessionID)
	}

	// Check if neuro command exists
	if err := checkNeuroCommand(); err != nil {
		return err
	}

	// Find the experiment script
	scriptPath, err := findExperimentScript(experimentName)
	if err != nil {
		return err
	}

	// Check if the recording exists
	recordingDir := filepath.Join(recordingsDir, experimentName)
	expectedFile := filepath.Join(recordingDir, sessionID+".expected")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		return fmt.Errorf("recording not found: %s\nUse 'neurotest record-experiment %s' to create recordings", expectedFile, experimentName)
	}

	// Read expected output
	expectedBytes, err := os.ReadFile(expectedFile)
	if err != nil {
		return fmt.Errorf("failed to read expected output: %w", err)
	}
	expected := string(expectedBytes)

	// Run the experiment
	actual, duration, err := runExperimentScript(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to run experiment: %w", err)
	}

	// Compare outputs using the same logic as regular tests
	var passed bool
	var normalized bool

	// First try exact match
	if strings.TrimSpace(actual) == strings.TrimSpace(expected) {
		passed = true
	} else {
		// Try smart comparison with placeholders
		if normalizationEngine.CompareWithPlaceholders(expected, actual) {
			passed = true
			normalized = true
		} else {
			// Try normalized comparison
			normalizedExpected := normalizationEngine.NormalizeOutput(expected)
			normalizedActual := normalizationEngine.NormalizeOutput(actual)
			if strings.TrimSpace(normalizedActual) == strings.TrimSpace(normalizedExpected) {
				passed = true
				normalized = true
			}
		}
	}

	if passed {
		if normalized && verbose {
			fmt.Printf("PASS: %s (session: %s, duration: %v, using smart comparison)\n", experimentName, sessionID, duration)
		} else {
			fmt.Printf("PASS: %s (session: %s, duration: %v)\n", experimentName, sessionID, duration)
		}
		return nil
	}

	fmt.Printf("FAIL: %s (session: %s, duration: %v)\n", experimentName, sessionID, duration)
	showDetailedDiff(expected, actual, fmt.Sprintf("%s:%s", experimentName, sessionID))
	return fmt.Errorf("experiment output mismatch")
}

// recordAllExperiments records all available experiments
func recordAllExperiments(_ *cobra.Command, _ []string) error {
	if verbose {
		fmt.Println("Recording all experiments...")
	}

	// Find all experiments
	experiments, err := findAllExperiments()
	if err != nil {
		return err
	}

	if len(experiments) == 0 {
		fmt.Println("No experiments found in examples/experiments/")
		return nil
	}

	recorded := 0
	failed := 0

	for _, experimentName := range experiments {
		fmt.Printf("Recording %s... ", experimentName)
		if err := recordExperiment(nil, []string{experimentName}); err != nil {
			fmt.Println("FAIL")
			if verbose {
				fmt.Printf("  Error: %v\n", err)
			}
			failed++
		} else {
			fmt.Println("RECORDED")
			recorded++
		}
	}

	fmt.Printf("\nResults: %d recorded, %d failed\n", recorded, failed)

	if failed > 0 {
		return fmt.Errorf("%d experiment(s) failed to record", failed)
	}

	return nil
}

// .neurorc startup testing functionality

// NeuroRCTestConfig defines the configuration for a .neurorc startup test
type NeuroRCTestConfig struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Setup       NeuroRCTestSetup  `json:"setup"`
	CLIFlags    []string          `json:"cli_flags"`
	EnvVars     map[string]string `json:"env_vars"`
	InputSeq    []string          `json:"input_sequence"`
	ExpectedOut []string          `json:"expected_contains"`
	ExpectedNot []string          `json:"expected_not_contains"`
	Timeout     int               `json:"timeout_seconds"`
}

// NeuroRCTestSetup defines the test environment setup
type NeuroRCTestSetup struct {
	WorkingDirNeuroRC string            `json:"working_dir_neurorc"`
	HomeDirNeuroRC    string            `json:"home_dir_neurorc"`
	CustomFiles       map[string]string `json:"custom_files"` // filename -> content
}

// NeuroRCTestEnvironment represents an isolated test environment
type NeuroRCTestEnvironment struct {
	TempDir     string
	WorkingDir  string
	HomeDir     string
	ConfigFiles map[string]string
}

// recordNeuroRCTest records a new .neurorc startup test case
func recordNeuroRCTest(_ *cobra.Command, args []string) error {
	testName := args[0]

	if verbose {
		fmt.Printf("Recording .neurorc startup test: %s\n", testName)
	}

	// Check if neuro command exists
	if err := checkNeuroCommand(); err != nil {
		return err
	}

	// Find the test configuration file
	configFile, err := findNeuroRCTestConfig(testName)
	if err != nil {
		return err
	}

	// Parse test configuration
	config, err := parseNeuroRCTestConfig(configFile)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("Loaded test config: %s\n", config.Description)
	}

	// Create isolated test environment
	testEnv, err := createNeuroRCTestEnvironment(config)
	if err != nil {
		return err
	}
	defer cleanupNeuroRCTestEnvironment(testEnv)

	// Run the .neurorc startup test
	output, err := runNeuroRCShellTest(testEnv, config)
	if err != nil {
		return fmt.Errorf("failed to run .neurorc startup test: %w", err)
	}

	// Save the output as expected result
	expectedFile := filepath.Join(testDir, "neurorc", testName+".expected")
	if err := os.MkdirAll(filepath.Dir(expectedFile), 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	if err := os.WriteFile(expectedFile, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write expected output: %w", err)
	}

	fmt.Printf("Recorded .neurorc startup test: %s\n", testName)
	fmt.Printf("Expected output saved to: %s\n", expectedFile)

	return nil
}

// runNeuroRCTest runs a specific .neurorc startup test case
func runNeuroRCTest(_ *cobra.Command, args []string) error {
	testName := args[0]

	if verbose {
		fmt.Printf("Running .neurorc startup test: %s\n", testName)
	}

	// Check if neuro command exists
	if err := checkNeuroCommand(); err != nil {
		return err
	}

	// Find the test configuration file
	configFile, err := findNeuroRCTestConfig(testName)
	if err != nil {
		return err
	}

	// Parse test configuration
	config, err := parseNeuroRCTestConfig(configFile)
	if err != nil {
		return err
	}

	// Check if expected file exists
	expectedFile := filepath.Join(testDir, "neurorc", testName+".expected")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		return fmt.Errorf("expected output file not found: %s\nRun 'neurotest record-neurorc %s' first", expectedFile, testName)
	}

	// Read expected output
	expectedBytes, err := os.ReadFile(expectedFile)
	if err != nil {
		return fmt.Errorf("failed to read expected output: %w", err)
	}
	expected := string(expectedBytes)

	// Create isolated test environment
	testEnv, err := createNeuroRCTestEnvironment(config)
	if err != nil {
		return err
	}
	defer cleanupNeuroRCTestEnvironment(testEnv)

	// Run the .neurorc startup test
	actual, err := runNeuroRCShellTest(testEnv, config)
	if err != nil {
		return fmt.Errorf("failed to run .neurorc startup test: %w", err)
	}

	// Smart comparison with normalization
	var passed bool
	var normalized bool

	// First try exact match
	if strings.TrimSpace(actual) == strings.TrimSpace(expected) {
		passed = true
	} else {
		// Try smart comparison with placeholders
		if normalizationEngine.CompareWithPlaceholders(expected, actual) {
			passed = true
			normalized = true
		} else {
			// Try normalized comparison
			normalizedExpected := normalizationEngine.NormalizeOutput(expected)
			normalizedActual := normalizationEngine.NormalizeOutput(actual)
			if strings.TrimSpace(normalizedActual) == strings.TrimSpace(normalizedExpected) {
				passed = true
				normalized = true
			}
		}
	}

	if passed {
		if normalized && verbose {
			fmt.Printf("PASS: %s (using smart comparison)\n", testName)
		} else {
			fmt.Printf("PASS: %s\n", testName)
		}
		return nil
	}

	fmt.Printf("FAIL: %s\n", testName)
	showDetailedDiff(expected, actual, testName)
	return fmt.Errorf("test failed: output mismatch")
}
