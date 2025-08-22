// Package cli provides command-line interface setup for neurotest.
package cli

import (
	"neuroshell/cmd/neurotest/internal/normalize"
	"neuroshell/cmd/neurotest/shared"

	"github.com/spf13/cobra"
)

// App represents the neurotest CLI application
type App struct {
	Config     *shared.Config
	Normalizer *normalize.NormalizationEngine
}

// NewApp creates a new neurotest CLI application
func NewApp() *App {
	return &App{
		Config:     shared.NewConfig(),
		Normalizer: normalize.NewNormalizationEngine(),
	}
}

// CreateRootCommand creates and configures the root command
func (app *App) CreateRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "neurotest",
		Short: "End-to-end testing tool for Neuro CLI",
		Long: `neurotest is a testing tool for the Neuro CLI that uses golden files to
verify expected behavior. It can record, run, and verify test cases.`,
	}

	// Add global flags
	rootCmd.PersistentFlags().BoolVarP(&app.Config.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().StringVar(&app.Config.TestDir, "test-dir", shared.DefaultTestDir, "Test directory")
	rootCmd.PersistentFlags().StringVar(&app.Config.NeuroCmd, "neuro-cmd", shared.DefaultNeuroCmd, "Neuro command to test (will try ./bin/neuro, then PATH)")
	rootCmd.PersistentFlags().IntVar(&app.Config.TestTimeout, "timeout", shared.DefaultTestTimeout, "Test timeout in seconds")

	// Add all subcommands
	app.addGoldenFileCommands(rootCmd)
	app.addCFlagCommands(rootCmd)
	app.addExperimentCommands(rootCmd)
	app.addNeuroRCCommands(rootCmd)
	app.addVersionCommand(rootCmd)

	return rootCmd
}
