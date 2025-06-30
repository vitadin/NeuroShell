// Package main provides the NeuroShell CLI application entry point.
// NeuroShell is a specialized shell environment designed for seamless interaction with LLM agents.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/abiosoft/ishell/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	_ "neuroshell/internal/commands/builtin" // Import for side effects (init functions)
	"neuroshell/internal/context"
	"neuroshell/internal/logger"
	"neuroshell/internal/orchestration"
	"neuroshell/internal/shell"
)

var (
	logLevel string
	logFile  string
	testMode bool
	version  = "0.1.0" // This could be set at build time
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "neuro",
	Short: "Neuro Shell - LLM-integrated shell environment",
	Long: `Neuro is a specialized shell environment designed for seamless interaction with LLM agents.
It bridges the gap between traditional command-line interfaces and modern AI assistants.`,
	Run: runShell, // Default behavior is to run the interactive shell
}

// shellCmd represents the shell command (explicit version of default behavior)
var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start interactive shell mode",
	Long:  `Start the interactive Neuro shell for LLM-integrated command execution.`,
	Run:   runShell,
}

// batchCmd represents the batch command for non-interactive script execution
var batchCmd = &cobra.Command{
	Use:   "batch <script.neuro>",
	Short: "Execute a .neuro script file in batch mode",
	Long: `Execute a .neuro script file directly without entering interactive mode.
This is useful for automation, CI/CD pipelines, and running predefined workflows.`,
	Args: cobra.ExactArgs(1),
	Run:  runBatch,
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version of Neuro Shell.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Neuro Shell v%s\n", version)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "Set log level (debug|info|warn|error) [default: info]")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Write logs to file instead of stderr")
	rootCmd.PersistentFlags().BoolVar(&testMode, "test-mode", false, "Run in deterministic test mode")

	// Bind flags to viper
	if err := viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding log-level flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding log-file flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("test-mode", rootCmd.PersistentFlags().Lookup("test-mode")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding test-mode flag: %v\n", err)
		os.Exit(1)
	}

	// Add subcommands
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(batchCmd)
	rootCmd.AddCommand(versionCmd)

	// Configure logger before any command execution
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Configure logger with CLI flags
	if err := logger.Configure(logLevel, logFile, testMode); err != nil {
		fmt.Fprintf(os.Stderr, "Error configuring logger: %v\n", err)
		os.Exit(1)
	}
}

func runShell(_ *cobra.Command, _ []string) {
	logger.Info("Starting NeuroShell", "version", version)

	// Initialize services before starting shell
	if err := shell.InitializeServices(testMode); err != nil {
		logger.Fatal("Failed to initialize services", "error", err)
	}

	logger.Info("Services initialized successfully")

	sh := ishell.New()
	sh.SetPrompt("neuro> ")

	// Remove built-in commands so they become user messages or Neuro commands
	sh.DeleteCmd("exit")
	sh.DeleteCmd("help")

	sh.Println(fmt.Sprintf("Neuro Shell v%s - LLM-integrated shell environment", version))
	sh.Println("Type '\\help' for Neuro commands or '\\exit' to quit.")

	sh.NotFound(shell.ProcessInput)

	sh.Run()
}

func runBatch(_ *cobra.Command, args []string) {
	scriptPath := args[0]

	logger.Info("Starting NeuroShell batch mode", "version", version, "script", scriptPath)

	// Validate script file exists and has correct extension
	if err := validateScriptFile(scriptPath); err != nil {
		logger.Fatal("Script validation failed", "error", err)
	}

	// Initialize services before running script
	if err := shell.InitializeServices(testMode); err != nil {
		logger.Fatal("Failed to initialize services", "error", err)
	}

	logger.Info("Services initialized successfully")

	// Create a context for batch execution
	ctx := context.New()
	ctx.SetTestMode(testMode)

	// Execute the script
	if err := executeBatchScript(scriptPath, ctx); err != nil {
		logger.Fatal("Script execution failed", "error", err)
	}

	logger.Info("Script executed successfully", "script", scriptPath)
}

func validateScriptFile(scriptPath string) error {
	// Check if file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script file does not exist: %s", scriptPath)
	}

	// Check file extension
	if ext := filepath.Ext(scriptPath); ext != ".neuro" {
		return fmt.Errorf("script file must have .neuro extension, got: %s", ext)
	}

	return nil
}

func executeBatchScript(scriptPath string, ctx *context.NeuroContext) error {
	// Execute the script using centralized execution logic
	return orchestration.ExecuteScript(scriptPath, ctx)
}
