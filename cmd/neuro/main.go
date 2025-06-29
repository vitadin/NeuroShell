package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/abiosoft/ishell/v2"
	"neuroshell/internal/logger"
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

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version of Neuro Shell.`,
	Run: func(cmd *cobra.Command, args []string) {
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
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("test-mode", rootCmd.PersistentFlags().Lookup("test-mode"))

	// Add subcommands
	rootCmd.AddCommand(shellCmd)
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

func runShell(cmd *cobra.Command, args []string) {
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