// Package golden provides golden file testing functionality.
package golden

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"neuroshell/cmd/neurotest/internal/normalize"
	"neuroshell/cmd/neurotest/shared"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// Differ handles diff operations for golden file tests
type Differ struct {
	config     *shared.Config
	normalizer *normalize.NormalizationEngine
}

// NewDiffer creates a new golden file differ
func NewDiffer(config *shared.Config) *Differ {
	return &Differ{
		config:     config,
		normalizer: normalize.NewNormalizationEngine(),
	}
}

// ShowDiff displays the differences between expected and actual output
func (d *Differ) ShowDiff(testName string) error {
	scriptPath, err := shared.FindScript(testName, d.config.TestDir)
	if err != nil {
		return fmt.Errorf("failed to find script: %w", err)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("test script not found: %s", scriptPath)
	}

	// Get actual output
	output, err := shared.RunNeuroScript(scriptPath, d.config.NeuroCmd, d.config.TestTimeout)
	if err != nil && d.config.Verbose {
		fmt.Printf("Command failed with error: %v\nOutput: %s\n", err, output)
	}

	actualOutput := d.cleanOutput(output)

	// Get expected output
	expectedPath := filepath.Join(d.config.TestDir, testName+".expected")
	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		return fmt.Errorf("failed to read expected file %s: %w", expectedPath, err)
	}

	expectedOutput := strings.TrimRight(string(expectedContent), "\n")

	// Show detailed diff
	d.ShowDetailedDiff(expectedOutput, actualOutput, testName)

	return nil
}

// ShowCFlagDiff displays the differences between expected and actual output for -c flag tests
func (d *Differ) ShowCFlagDiff(testName string) error {
	scriptPath, err := shared.FindScript(testName, d.config.TestDir)
	if err != nil {
		return fmt.Errorf("failed to find script: %w", err)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("test script not found: %s", scriptPath)
	}

	// Get actual output using -c flag
	output, err := shared.RunNeuroCFlag(scriptPath, d.config.NeuroCmd, d.config.TestTimeout)
	if err != nil && d.config.Verbose {
		fmt.Printf("Command failed with error: %v\nOutput: %s\n", err, output)
	}

	actualOutput := d.cleanOutput(output)

	// Get expected output from .c.expected file
	expectedPath := filepath.Join(d.config.TestDir, testName+".c.expected")
	expectedContent, err := os.ReadFile(expectedPath)
	if err != nil {
		return fmt.Errorf("failed to read -c expected file %s: %w", expectedPath, err)
	}

	expectedOutput := strings.TrimRight(string(expectedContent), "\n")

	// Show detailed diff
	d.ShowDetailedDiff(expectedOutput, actualOutput, testName+" (-c flag)")

	return nil
}

// ShowDetailedDiff displays a detailed comparison between expected and actual output
func (d *Differ) ShowDetailedDiff(expected, actual, testName string) {
	fmt.Printf("=== Test: %s ===\n", testName)

	if expected == actual {
		fmt.Println("No differences found - test passes!")
		return
	}

	fmt.Println("\n--- Expected ---")
	d.printNumberedLines(expected)

	fmt.Println("\n--- Actual ---")
	d.printNumberedLines(actual)

	fmt.Println("\n--- Diff ---")
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(expected, actual, false)

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffDelete:
			fmt.Printf("- %q\n", diff.Text)
		case diffmatchpatch.DiffInsert:
			fmt.Printf("+ %q\n", diff.Text)
		case diffmatchpatch.DiffEqual:
			// Don't print unchanged parts for brevity
			if len(diff.Text) > 50 {
				fmt.Printf("  %q...\n", diff.Text[:47])
			} else {
				fmt.Printf("  %q\n", diff.Text)
			}
		}
	}
}

// printNumberedLines prints lines with line numbers
func (d *Differ) printNumberedLines(content string) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		fmt.Printf("%4d→%s\n", i+1, line)
	}
}

// cleanOutput normalizes output for comparison
func (d *Differ) cleanOutput(output string) string {
	cleaned := shared.CleanOutput(output)
	return d.normalizer.NormalizeOutput(cleaned)
}
