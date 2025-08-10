// Package main provides the neurotest CLI application for end-to-end testing of NeuroShell.
// neurotest uses golden files to record, run, and verify expected behavior of Neuro CLI commands.
package main

import (
	"os"

	"neuroshell/cmd/neurotest/internal/cli"
)

func main() {
	app := cli.NewApp()
	rootCmd := app.CreateRootCommand()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
