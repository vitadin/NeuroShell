package logger

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

// Global logger instance
var Logger *log.Logger

func init() {
	// Create new logger instance
	Logger = log.New(os.Stderr)
	
	// Remove timestamps as requested
	Logger.SetTimeFormat("")
	
	// Set log level based on environment variable, default to Info
	level := strings.ToLower(os.Getenv("NEURO_LOG_LEVEL"))
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
		Logger.SetLevel(log.InfoLevel) // Default level
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