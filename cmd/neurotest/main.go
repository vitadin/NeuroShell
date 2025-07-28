// Package main provides the neurotest CLI application for end-to-end testing of NeuroShell.
// neurotest uses golden files to record, run, and verify expected behavior of Neuro CLI commands.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var (
	version             = "0.1.0"
	testDir             = "test/golden"
	neurocmd            = "neuro"
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
	Long:  `Display the version of neurotest.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("neurotest v%s\n", version)
	},
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
	rootCmd.PersistentFlags().StringVar(&neurocmd, "neuro-cmd", "./bin/neuro", "Neuro command to test")
	rootCmd.PersistentFlags().IntVar(&testTimeout, "timeout", 30, "Test timeout in seconds")

	// Add subcommands
	rootCmd.AddCommand(recordCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(runAllCmd)
	rootCmd.AddCommand(acceptCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(versionCmd)
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
	_, err := exec.LookPath(neurocmd)
	if err != nil {
		return fmt.Errorf("neuro command not found: %s\nMake sure it's installed and in PATH", neurocmd)
	}
	return nil
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
	cmd := exec.Command(neurocmd, "batch", absPath, "--test-mode")

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
