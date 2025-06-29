package logger

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

// Global logger instance
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
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
		Logger.SetTimeFormat("") // No timestamps
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

// Convenience functions for different log levels
func Debug(msg interface{}, keyvals ...interface{}) {
	Logger.Debug(msg, keyvals...)
}

func Info(msg interface{}, keyvals ...interface{}) {
	Logger.Info(msg, keyvals...)
}

func Warn(msg interface{}, keyvals ...interface{}) {
	Logger.Warn(msg, keyvals...)
}

func Error(msg interface{}, keyvals ...interface{}) {
	Logger.Error(msg, keyvals...)
}

func Fatal(msg interface{}, keyvals ...interface{}) {
	Logger.Fatal(msg, keyvals...)
}

// Structured logging helpers for common debugging scenarios
func CommandExecution(command string, args map[string]string) {
	Debug("Executing command", "command", command, "args", args)
}

func ServiceOperation(service string, operation string, details ...interface{}) {
	Debug("Service operation", "service", service, "operation", operation, "details", details)
}

func VariableOperation(operation string, key string, value string) {
	Debug("Variable operation", "operation", operation, "key", key, "value", value)
}

func InterpolationStep(text string, result string) {
	Debug("Variable interpolation", "input", text, "output", result)
}