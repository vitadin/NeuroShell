// Package main provides the NeuroShell CLI application entry point.
// NeuroShell is a specialized shell environment designed for seamless interaction with LLM agents.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "neuroshell/internal/commands/assert"  // Import assert commands (init functions)
	_ "neuroshell/internal/commands/builtin" // Import for side effects (init functions)
	_ "neuroshell/internal/commands/render"  // Import render commands (init functions)
	_ "neuroshell/internal/commands/session" // Import session commands (init functions)
	_ "neuroshell/internal/commands/shell"   // Import shell commands (init functions)
	"neuroshell/internal/context"
	"neuroshell/internal/data/embedded"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
	"neuroshell/internal/shell"
	"neuroshell/internal/statemachine"
	"neuroshell/internal/version"

	"github.com/abiosoft/ishell/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/chzyer/readline"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logLevel    string
	logFile     string
	testMode    bool
	noColor     bool
	showVersion bool
	// .neurorc control flags
	noRC      bool
	rcFile    string
	confirmRC bool
	// Command execution flag
	commandString string
	// Global shell instance for prompt updates
	globalShell *ishell.Shell
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "neuro",
	Short: "Neuro Shell - LLM-integrated shell environment",
	Long: `Neuro is a specialized shell environment designed for seamless interaction with LLM agents.
It bridges the gap between traditional command-line interfaces and modern AI assistants.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle --version flag
		if showVersion {
			fmt.Println(version.GetFormattedVersion())
			return
		}

		// Handle -c flag
		if commandString != "" {
			runCommand(cmd, commandString)
			return
		}

		// Default behavior is to run the interactive shell
		runShell(cmd, args)
	},
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
	Long:  `Display the version of Neuro Shell with build information.`,
	Run: func(cmd *cobra.Command, _ []string) {
		detailed, _ := cmd.Flags().GetBool("detailed")
		if detailed {
			fmt.Println(version.GetDetailedVersion())
		} else {
			fmt.Println(version.GetFormattedVersion())
		}
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
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Show version information")

	// Add .neurorc control flags
	rootCmd.PersistentFlags().BoolVar(&noRC, "no-rc", false, "Skip .neurorc startup files")
	rootCmd.PersistentFlags().StringVar(&rcFile, "rc-file", "", "Use specific startup script instead of .neurorc")
	rootCmd.PersistentFlags().BoolVar(&confirmRC, "confirm-rc", false, "Prompt before executing .neurorc files")

	// Add command execution flag
	rootCmd.PersistentFlags().StringVarP(&commandString, "command", "c", "", "Execute command(s) and exit (use \\n for multiple commands)")

	// Add version command flags
	versionCmd.Flags().Bool("detailed", false, "Show detailed version information")

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

	// Bind .neurorc control flags to viper
	if err := viper.BindPFlag("no-rc", rootCmd.PersistentFlags().Lookup("no-rc")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding no-rc flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("rc-file", rootCmd.PersistentFlags().Lookup("rc-file")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding rc-file flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("confirm-rc", rootCmd.PersistentFlags().Lookup("confirm-rc")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding confirm-rc flag: %v\n", err)
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

	// Configure lipgloss color output based on CLI flags, environment, and test mode
	if noColor || testMode || os.Getenv("NO_COLOR") != "" {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

// createCustomReadlineConfig creates a readline configuration with custom key bindings.
func createCustomReadlineConfig() *readline.Config {
	cfg := &readline.Config{
		Prompt:      generateDynamicPrompt(), // Use dynamic prompt generation
		HistoryFile: "/tmp/neuro_history",
	}

	// Set up command highlighting using PromptColorService
	if colorService, err := services.GetGlobalRegistry().GetService("prompt_color"); err == nil {
		if promptColor, ok := colorService.(*services.PromptColorService); ok {
			cfg.Painter = promptColor.CreateCommandHighlighter()
		}
	}

	// Set up custom key listener for Ctrl+E editor shortcut
	cfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		switch key {
		case 5: // Ctrl+E - Open external editor
			logger.Debug("Ctrl+E pressed - opening editor", "currentLine", string(line))

			// Get editor content with current line as initial content
			content, err := openEditorAndGetContent(string(line))
			if err != nil {
				// Log error but don't break the input flow
				logger.Error("Editor operation failed", "error", err)
				// Return original line unchanged
				return line, pos, true
			}

			// Replace current line with editor content
			newLine := []rune(content)
			logger.Debug("Editor content applied", "newContent", content)
			return newLine, len(newLine), true

		default:
			// Let readline handle other keys normally
			return line, pos, false
		}
	})

	return cfg
}

// generateDynamicPrompt creates the current prompt with variable interpolation.
// This function retrieves prompt templates from the ShellPromptService and
// performs interpolation using the context's InterpolateVariables method.
// For multi-line prompts, only returns the last line for readline.
// processPromptLine applies variable interpolation and color processing to a prompt line.
func processPromptLine(template string, ctx *context.NeuroContext) string {
	// First, interpolate variables
	interpolated := ctx.InterpolateVariables(template)

	// Then, apply color processing if available
	colorService, err := services.GetGlobalRegistry().GetService("prompt_color")
	if err != nil {
		// Color service not available, return interpolated text
		return interpolated
	}

	promptColor := colorService.(*services.PromptColorService)
	return promptColor.ProcessColorMarkup(interpolated)
}

func generateDynamicPrompt() string {
	// Get prompt service
	promptService, err := services.GetGlobalRegistry().GetService("shell_prompt")
	if err != nil {
		logger.Debug("Shell prompt service not available, using default", "error", err)
		return "neuro> "
	}

	shellPrompt := promptService.(*services.ShellPromptService)
	lines, err := shellPrompt.GetPromptLines()
	if err != nil {
		logger.Debug("Failed to get prompt lines, using default", "error", err)
		return "neuro> "
	}

	// Get context for interpolation
	ctx := shell.GetGlobalContext()
	if ctx == nil {
		logger.Debug("Global context not available, using default prompt")
		return "neuro> "
	}

	// Process all lines (interpolation + color processing)
	var processedLines []string
	for _, template := range lines {
		processed := processPromptLine(template, ctx)
		processedLines = append(processedLines, processed)
	}

	// Return only the last line for readline
	if len(processedLines) > 0 {
		return processedLines[len(processedLines)-1]
	}

	return "neuro> "
}

// generatePromptPrefix creates the prefix lines for multi-line prompts.
// Returns the first N-1 lines that should be printed before the readline prompt.
func generatePromptPrefix() []string {
	// Get prompt service
	promptService, err := services.GetGlobalRegistry().GetService("shell_prompt")
	if err != nil {
		return nil
	}

	shellPrompt := promptService.(*services.ShellPromptService)
	lines, err := shellPrompt.GetPromptLines()
	if err != nil {
		return nil
	}

	// Get context for interpolation
	ctx := shell.GetGlobalContext()
	if ctx == nil {
		return nil
	}

	// If only one line, no prefix needed
	if len(lines) <= 1 {
		return nil
	}

	// Process the first N-1 lines for the prefix (interpolation + color processing)
	var prefixLines []string
	for i := 0; i < len(lines)-1; i++ {
		processed := processPromptLine(lines[i], ctx)
		prefixLines = append(prefixLines, processed)
	}

	return prefixLines
}

// updateShellPrompt updates the shell prompt with current context.
// This should be called after command execution to refresh the prompt display.
func updateShellPrompt(sh *ishell.Shell) {
	if sh == nil {
		return
	}

	// Set prefix lines for multi-line prompts using new ishell functionality
	prefixLines := generatePromptPrefix()
	sh.SetPromptPrefix(prefixLines)

	// Set only the last line as the readline prompt
	newPrompt := generateDynamicPrompt()
	sh.SetPrompt(newPrompt)
	logger.Debug("Shell prompt updated", "prefixLines", len(prefixLines), "prompt", newPrompt)
}

// openEditorAndGetContent opens the external editor with initial content and returns the edited content.
func openEditorAndGetContent(initialContent string) (string, error) {
	// Get the editor service
	editorService, err := services.GetGlobalRegistry().GetService("editor")
	if err != nil {
		return "", fmt.Errorf("editor service not available: %w", err)
	}

	es := editorService.(*services.EditorService)

	// Create context for the editor operation
	ctx := shell.GetGlobalContext()

	// Set global context for service access
	context.SetGlobalContext(ctx)

	// Open editor with initial content
	content, err := es.OpenEditorWithContent(initialContent)
	if err != nil {
		return "", fmt.Errorf("editor operation failed: %w", err)
	}

	return content, nil
}

// setupAutoComplete configures the shell with custom autocomplete functionality.
func setupAutoComplete(sh *ishell.Shell) error {
	// Get the autocomplete service
	autocompleteService, err := services.GetGlobalRegistry().GetService("autocomplete")
	if err != nil {
		return fmt.Errorf("autocomplete service not available: %w", err)
	}

	// Cast to AutoCompleteService
	autoCompleter := autocompleteService.(*services.AutoCompleteService)

	// Set up custom completer
	sh.CustomCompleter(autoCompleter)

	logger.Debug("Autocomplete service configured successfully")
	return nil
}

func runShell(_ *cobra.Command, _ []string) {
	logger.Info("Starting NeuroShell", "version", version.GetVersion())

	// Initialize services before starting shell
	if err := shell.InitializeServices(testMode); err != nil {
		logger.Fatal("Failed to initialize services", "error", err)
	}

	logger.Debug("Services initialized successfully")

	// Execute system initialization script first (before user .neurorc)
	if err := executeSystemInit(); err != nil {
		logger.Error("Failed to execute system initialization script", "error", err)
		// Don't exit - just log the error and continue with startup
	}

	// Execute .neurorc startup script if found
	if err := executeNeuroRC(); err != nil {
		logger.Error("Failed to execute .neurorc startup script", "error", err)
		// Don't exit - just log the error and continue with shell startup
	}

	// Create shell with custom readline configuration
	cfg := createCustomReadlineConfig()
	sh := ishell.NewWithConfig(cfg)

	// Store shell instance globally for prompt updates
	globalShell = sh

	// Set up autocomplete
	if err := setupAutoComplete(sh); err != nil {
		logger.Error("Failed to setup autocomplete", "error", err)
		// Don't fail startup, just log the error
	}

	// Initialize dynamic prompt
	updateShellPrompt(sh)

	// Set up prompt update callback for the shell handler
	shell.PromptUpdateCallback = func() {
		updateShellPrompt(globalShell)
	}

	// Remove built-in commands so they become user messages or Neuro commands
	sh.DeleteCmd("exit")
	sh.DeleteCmd("help")

	sh.Println(fmt.Sprintf("Neuro Shell v%s - LLM-integrated shell environment", version.GetVersion()))
	sh.Println("Licensed under LGPL v3 (\\license for details)")
	sh.Println("Type '\\help' for Neuro commands, Ctrl+E for editor, or '\\exit' to quit.")

	sh.NotFound(shell.ProcessInput)

	sh.Run()
}

func runBatch(_ *cobra.Command, args []string) {
	scriptPath := args[0]

	logger.Info("Starting NeuroShell batch mode", "version", version.GetVersion(), "script", scriptPath)

	// Validate script file exists and has correct extension
	if err := validateScriptFile(scriptPath); err != nil {
		logger.Fatal("Script validation failed", "error", err)
	}

	// Initialize services before running script
	if err := shell.InitializeServices(testMode); err != nil {
		logger.Fatal("Failed to initialize services", "error", err)
	}

	logger.Info("Services initialized successfully")

	// Execute system initialization script first (before user script)
	if err := executeSystemInit(); err != nil {
		logger.Error("Failed to execute system initialization script", "error", err)
		// Don't exit - just log the error and continue with script execution
	}

	// Get the global context singleton for batch execution
	ctx := shell.GetGlobalContext()
	ctx.SetTestMode(testMode)

	// Execute the script
	if err := executeBatchScript(scriptPath, ctx); err != nil {
		logger.Fatal("Script execution failed", "error", err)
	}

	logger.Debug("Script executed successfully", "script", scriptPath)
}

func runCommand(_ *cobra.Command, cmdString string) {
	logger.Info("Executing command", "command", cmdString)

	// Initialize services before running command
	if err := shell.InitializeServices(testMode); err != nil {
		logger.Fatal("Failed to initialize services", "error", err)
	}

	logger.Info("Services initialized successfully")

	// Execute system initialization script first (unless --no-rc)
	if !noRC {
		if err := executeSystemInit(); err != nil {
			logger.Error("Failed to execute system initialization script", "error", err)
			// Don't exit - just log the error and continue with command execution
		}
	}

	// Get the global context singleton for command execution
	ctx := shell.GetGlobalContext()
	ctx.SetTestMode(testMode)

	// Process escape sequences (\n → newline, but preserve \\n as literal)
	processedCmd := processEscapeSequences(cmdString)

	// Split into lines (same as script processing)
	lines := strings.Split(processedCmd, "\n")
	commandLines := make([]string, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comments
		if trimmed != "" && !strings.HasPrefix(trimmed, "%%") {
			commandLines = append(commandLines, trimmed)
		}
	}

	if len(commandLines) == 0 {
		// No commands to execute
		logger.Debug("No commands to execute")
		return
	}

	// Get stack service to push commands
	registry := services.GetGlobalRegistry()
	stackServiceInterface, err := registry.GetService("stack")
	if err != nil {
		logger.Fatal("Stack service not available", "error", err)
	}
	stackService := stackServiceInterface.(*services.StackService)

	// Push commands in reverse order (LIFO - last in, first out)
	for i := len(commandLines) - 1; i >= 0; i-- {
		stackService.PushCommand(commandLines[i])
	}

	// Create state machine and execute
	sm := statemachine.NewStateMachineWithDefaults(ctx)

	// Execute all commands from the stack
	// The state machine will automatically pop and execute commands from the stack
	if err := sm.Execute(""); err != nil {
		logger.Error("Command execution failed", "error", err)
		os.Exit(1)
	}

	logger.Debug("Command executed successfully")
}

// processEscapeSequences converts \n to actual newlines while preserving \\n as literal \n
// This allows users to separate commands with \n while still being able to use \\n in command arguments
func processEscapeSequences(input string) string {
	// Replace \\n with a temporary placeholder to protect it
	protected := strings.ReplaceAll(input, "\\\\n", "\x00PROTECTED_NEWLINE\x00")

	// Replace \n with actual newlines
	processed := strings.ReplaceAll(protected, "\\n", "\n")

	// Restore \\n as literal \n
	result := strings.ReplaceAll(processed, "\x00PROTECTED_NEWLINE\x00", "\\n")

	return result
}

func validateScriptFile(scriptPath string) error {
	// Check if file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script file does not exist: %s", scriptPath)
	}

	// Check file extension for regular scripts, but allow .neurorc files
	baseName := filepath.Base(scriptPath)
	if baseName == ".neurorc" {
		// .neurorc files are valid startup scripts
		return nil
	}

	if ext := filepath.Ext(scriptPath); ext != ".neuro" {
		return fmt.Errorf("script file must have .neuro extension or be named .neurorc, got: %s", scriptPath)
	}

	return nil
}

func executeBatchScript(scriptPath string, ctx *context.NeuroContext) error {
	// Set global context for services to use
	context.SetGlobalContext(ctx)

	// Execute script using state machine
	logger.Debug("Executing script via state machine", "script", scriptPath)
	sm := statemachine.NewStateMachineWithDefaults(ctx)
	// Add backslash prefix so state machine recognizes it as a file path command
	commandInput := "\\" + scriptPath
	return sm.Execute(commandInput)
}

// executeNeuroRC looks for and executes .neurorc startup scripts.
//
// Priority Order for Configuration:
//  1. CLI flags (highest priority)
//     - --no-rc: Skip all .neurorc files
//     - --rc-file=PATH: Use specific startup script
//     - --confirm-rc: Prompt before executing any .neurorc file
//  2. Environment variables (medium priority)
//     - NEURO_RC=0: Disable .neurorc processing
//     - NEURO_RC_FILE=PATH: Use specific startup script
//  3. Default file search (lowest priority)
//     - Current directory .neurorc
//     - User home directory .neurorc
func executeNeuroRC() error {
	// Priority 1: Check CLI flags first (--no-rc overrides everything)
	if noRC {
		logger.Debug("Skipping .neurorc - disabled by --no-rc flag")
		return nil
	}

	// Priority 2: Check environment variables (NEURO_RC=0 overrides default search)
	if os.Getenv("NEURO_RC") == "0" {
		logger.Debug("Skipping .neurorc - disabled by NEURO_RC=0")
		return nil
	}

	// Handle custom .neurorc file: CLI flag --rc-file takes priority over NEURO_RC_FILE
	var customRC string
	if rcFile != "" {
		customRC = rcFile
		logger.Debug("Using custom .neurorc from CLI flag", "path", customRC)
	} else if envRC := os.Getenv("NEURO_RC_FILE"); envRC != "" {
		customRC = envRC
		logger.Debug("Using custom .neurorc from environment", "path", customRC)
	}

	if customRC != "" {
		if _, err := os.Stat(customRC); err == nil {
			if confirmRC && !confirmRCExecution(customRC) {
				logger.Info("Skipping .neurorc - user declined confirmation")
				return nil
			}
			logger.Info("Executing custom .neurorc file", "path", customRC)
			return executeNeuroRCFile(customRC)
		}
		logger.Warn("Custom .neurorc file not found", "path", customRC)
		return nil // Don't fall through to default search if custom file was specified
	}

	// Priority 3: Default file search - Current directory .neurorc (highest priority in search)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	currentDirRC := filepath.Join(cwd, ".neurorc")
	if _, err := os.Stat(currentDirRC); err == nil {
		if confirmRC && !confirmRCExecution(currentDirRC) {
			logger.Info("Skipping .neurorc - user declined confirmation")
			return nil
		}
		logger.Info("Executing .neurorc from current directory", "path", currentDirRC)
		return executeNeuroRCFile(currentDirRC)
	}

	// Priority 3: Default file search - User home directory .neurorc (fallback)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Debug("Failed to get home directory", "error", err)
		return nil // Not fatal - just skip home directory .neurorc
	}

	homeDirRC := filepath.Join(homeDir, ".neurorc")
	if _, err := os.Stat(homeDirRC); err == nil {
		if confirmRC && !confirmRCExecution(homeDirRC) {
			logger.Info("Skipping .neurorc - user declined confirmation")
			return nil
		}
		logger.Info("Executing .neurorc from home directory", "path", homeDirRC)
		return executeNeuroRCFile(homeDirRC)
	}

	logger.Debug("No .neurorc file found in current or home directory")
	return nil
}

// executeNeuroRCFile executes a specific .neurorc file.
func executeNeuroRCFile(rcPath string) error {
	// Validate the file
	if err := validateScriptFile(rcPath); err != nil {
		return fmt.Errorf("invalid .neurorc file: %w", err)
	}

	// Get the global context for execution
	ctx := shell.GetGlobalContext()
	ctx.SetTestMode(testMode)

	// Set global context for services to use
	context.SetGlobalContext(ctx)

	// Store .neurorc execution information in system variables
	if err := ctx.SetSystemVariable("#neurorc_path", rcPath); err != nil {
		logger.Error("Failed to set #neurorc_path system variable", "error", err)
	}
	if err := ctx.SetSystemVariable("#neurorc_executed", "true"); err != nil {
		logger.Error("Failed to set #neurorc_executed system variable", "error", err)
	}

	// Execute the .neurorc script using state machine
	logger.Debug("Executing .neurorc via state machine", "script", rcPath)
	sm := statemachine.NewStateMachineWithDefaults(ctx)
	// Add backslash prefix so state machine recognizes it as a file path command
	commandInput := "\\" + rcPath
	return sm.Execute(commandInput)
}

// executeSystemInit loads and executes the system initialization script.
// This script runs before user .neurorc files and contains system-level initialization commands.
func executeSystemInit() error {
	// Create stdlib loader for accessing embedded system scripts
	stdlibLoader := embedded.NewStdlibLoader()

	// Load the system initialization script
	scriptContent, err := stdlibLoader.LoadScript("system-init")
	if err != nil {
		// If system-init.neuro doesn't exist, that's okay - just skip it
		logger.Debug("No system initialization script found", "error", err)
		return nil
	}

	// Get the global context for execution
	ctx := shell.GetGlobalContext()
	ctx.SetTestMode(testMode)

	// Set global context for services to use
	context.SetGlobalContext(ctx)

	// Store system initialization information in system variables
	systemInitPath := stdlibLoader.GetScriptPath("system-init")
	if err := ctx.SetSystemVariable("#system_init_path", systemInitPath); err != nil {
		logger.Error("Failed to set #system_init_path system variable", "error", err)
	}
	if err := ctx.SetSystemVariable("#system_init_executed", "true"); err != nil {
		logger.Error("Failed to set #system_init_executed system variable", "error", err)
	}

	// Execute the system initialization script using state machine
	logger.Debug("Executing system initialization script", "path", systemInitPath)
	sm := statemachine.NewStateMachineWithDefaults(ctx)

	// Execute script content directly (not as a file path)
	lines := strings.Split(scriptContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%%") {
			// Skip empty lines and comments
			continue
		}

		// Execute each command silently - any errors are logged but don't stop execution
		if err := sm.Execute(line); err != nil {
			logger.Debug("System init command failed (continuing)", "command", line, "error", err)
		}
	}

	logger.Debug("System initialization completed")
	return nil
}

// confirmRCExecution prompts the user for confirmation before executing a .neurorc file.
func confirmRCExecution(rcPath string) bool {
	fmt.Printf("Found .neurorc file: %s\n", rcPath)
	fmt.Print("Execute startup script? [y/N]: ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// If there's an error reading input (e.g., non-interactive terminal), default to no
		return false
	}

	// Accept y, Y, yes, YES as confirmation
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
