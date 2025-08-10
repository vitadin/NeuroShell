// Package cli provides command-line interface setup for neurotest.
package cli

import (
	"fmt"

	"neuroshell/internal/version"

	"github.com/spf13/cobra"
)

// addVersionCommand adds the version command
func (app *App) addVersionCommand(rootCmd *cobra.Command) {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the version of neurotest with build information.`,
		Run: func(cmd *cobra.Command, _ []string) {
			detailed, _ := cmd.Flags().GetBool("detailed")
			if detailed {
				fmt.Printf("neurotest %s\n", version.GetDetailedVersion())
			} else {
				fmt.Printf("neurotest %s\n", version.GetVersion())
			}
		},
	}

	versionCmd.Flags().Bool("detailed", false, "Show detailed version information")
	rootCmd.AddCommand(versionCmd)
}
