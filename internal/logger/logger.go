// Package logger provides centralized logging functionality for NeuroShell.
// It configures structured logging with support for different output formats and log levels.
package logger

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// Logger is the global logger instance used throughout NeuroShell.
var Logger *log.Logger

func init() {
	// Create new logger instance with default settings
	Logger = log.New(os.Stderr)

	// Remove timestamps as requested
	Logger.SetTimeFormat("")

	// Set default log level
	Logger.SetLevel(log.InfoLevel)
}

// Configure sets up the logger based on CLI flags and environment variables
// CLI flags take precedence over environment variables
func Configure(logLevel string, logFile string, testMode bool) error {
	// Set log level with precedence: CLI flag > env var > default
	level := logLevel
	if level == "" {
		level = strings.ToLower(os.Getenv("NEURO_LOG_LEVEL"))
	}
	if level == "" {
		level = "info" // default
	}

	switch level {
	case "debug":
		Logger.SetLevel(log.DebugLevel)
	case "info":
		Logger.SetLevel(log.InfoLevel)
	case "warn":
		Logger.SetLevel(log.WarnLevel)
	case "error":
		Logger.SetLevel(log.ErrorLevel)
	case "fatal":
		Logger.SetLevel(log.FatalLevel)
	default:
		Logger.SetLevel(log.InfoLevel)
	}

	// Set log output destination
	var output io.Writer = os.Stderr
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return err
		}
		output = file
	}

	// Create new logger with configured output
	Logger = log.New(output)
	Logger.SetTimeFormat("")
	Logger.SetLevel(parseLogLevel(level))

	// Configure for test mode
	if testMode {
		// In test mode, ensure deterministic output
		Logger.SetTimeFormat("")       // No timestamps
		Logger.SetLevel(log.InfoLevel) // Consistent level
	}

	return nil
}

// parseLogLevel converts string to log level
func parseLogLevel(level string) log.Level {
	switch strings.ToLower(level) {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	case "fatal":
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}

// Debug logs a debug message with optional key-value pairs.
func Debug(msg interface{}, keyvals ...interface{}) {
	Logger.Debug(msg, keyvals...)
}

// Info logs an info message with optional key-value pairs.
func Info(msg interface{}, keyvals ...interface{}) {
	Logger.Info(msg, keyvals...)
}

// Warn logs a warning message with optional key-value pairs.
func Warn(msg interface{}, keyvals ...interface{}) {
	Logger.Warn(msg, keyvals...)
}

// Error logs an error message with optional key-value pairs.
func Error(msg interface{}, keyvals ...interface{}) {
	Logger.Error(msg, keyvals...)
}

// Fatal logs a fatal message with optional key-value pairs and exits.
func Fatal(msg interface{}, keyvals ...interface{}) {
	Logger.Fatal(msg, keyvals...)
}

// CommandExecution logs command execution details for debugging.
func CommandExecution(command string, args map[string]string) {
	Debug("Executing command", "command", command, "args", args)
}

// ServiceOperation logs service operation details for debugging.
func ServiceOperation(service string, operation string, details ...interface{}) {
	Debug("Service operation", "service", service, "operation", operation, "details", details)
}

// VariableOperation logs variable operation details for debugging.
func VariableOperation(operation string, key string, value string) {
	Debug("Variable operation", "operation", operation, "key", key, "value", value)
}

// InterpolationStep logs variable interpolation steps for debugging.
func InterpolationStep(text string, result string) {
	Debug("Variable interpolation", "input", text, "output", result)
}

// NewStyledLogger creates a new logger with custom styles and prefix for component-specific logging.
// The prefix parameter is used to create a component-specific logger (e.g., "StateMachine", "Parser", etc.)
func NewStyledLogger(prefix string) *log.Logger {
	// Create custom styles for component logger
	styles := log.DefaultStyles()

	// Custom level styling without prefix (already added via log.Options)
	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("33")). // Blue background
		Foreground(lipgloss.Color("15"))  // White text

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("196")). // Red background
		Foreground(lipgloss.Color("15"))   // White text

	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("240")). // Gray background
		Foreground(lipgloss.Color("15"))   // White text

	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
		SetString("WARN").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("214")). // Orange background
		Foreground(lipgloss.Color("15"))   // White text

	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().
		SetString("FATAL").
		Padding(0, 1, 0, 1).
		Background(lipgloss.Color("88")). // Dark red background
		Foreground(lipgloss.Color("15"))  // White text

	// Custom key styling for common component keys
	styles.Keys["state"] = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))     // Purple
	styles.Keys["input"] = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))     // Blue
	styles.Keys["depth"] = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))    // Orange
	styles.Keys["error"] = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))    // Red
	styles.Keys["command"] = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))   // Green
	styles.Keys["component"] = lipgloss.NewStyle().Foreground(lipgloss.Color("51")) // Cyan

	// Custom value styling
	styles.Values["state"] = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	styles.Values["error"] = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))

	// Create logger with same output destination as global logger
	componentLogger := log.NewWithOptions(os.Stderr, log.Options{
		Prefix: prefix + " ",
	})

	// Apply custom styles
	componentLogger.SetStyles(styles)

	// Match the global logger's level
	componentLogger.SetLevel(Logger.GetLevel())

	return componentLogger
}
