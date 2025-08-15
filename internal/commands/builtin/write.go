package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"neuroshell/internal/commands"
	"neuroshell/internal/output"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// WriteCommand implements the \write command for writing content to files.
// It provides file writing capabilities similar to bash redirection (> and >>) within the NeuroShell environment.
type WriteCommand struct{}

// Name returns the command name "write" for registration and lookup.
func (c *WriteCommand) Name() string {
	return "write"
}

// ParseMode returns ParseModeKeyValue for standard argument parsing.
func (c *WriteCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the write command does.
func (c *WriteCommand) Description() string {
	return "Write content to a file with overwrite or append modes"
}

// Usage returns the syntax and usage examples for the write command.
func (c *WriteCommand) Usage() string {
	return "\\write[file=path, mode=append] content"
}

// HelpInfo returns structured help information for the write command.
func (c *WriteCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "file",
				Description: "Path to the file to write",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "mode",
				Description: "Write mode: 'overwrite' (default) or 'append'",
				Required:    false,
				Type:        "string",
				Default:     "overwrite",
			},
			{
				Name:        "silent",
				Description: "Suppress console output",
				Required:    false,
				Type:        "bool",
				Default:     "false",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\write[file=output.txt] Hello World!",
				Description: "Write content to file (overwrite mode)",
			},
			{
				Command:     "\\write[file=log.txt, mode=append] New entry: ${@date}",
				Description: "Append content to file with variable interpolation",
			},
			{
				Command:     "\\write[file=data/results.txt] ${analysis_results}",
				Description: "Write variable content, creating directories if needed",
			},
			{
				Command:     "\\write[file=temp.txt, silent=true] ${temp_data}",
				Description: "Write content without console feedback",
			},
			{
				Command:     "\\write[file=backup.txt] ${_output}",
				Description: "Write previous command output to file",
			},
		},
		Notes: []string{
			"file parameter is required",
			"Default mode is 'overwrite' (replaces file contents like >)",
			"Use mode='append' to add content to end of file (like >>)",
			"Parent directories are created automatically if they don't exist",
			"Variable interpolation works in both file paths and content",
			"Use silent=true to suppress feedback messages",
			"File permissions are set to 0644, directory permissions to 0755",
		},
	}
}

// Execute writes content to the specified file with the given mode.
// Options:
//   - file: target file path (required)
//   - mode: "overwrite" (default) or "append"
//   - silent: suppress console output (default: false)
func (c *WriteCommand) Execute(args map[string]string, input string) error {
	// Get file path - required parameter
	filePath := strings.TrimSpace(args["file"])
	if filePath == "" {
		return fmt.Errorf("file parameter is required. Usage: \\write[file=path] content")
	}

	// Parse mode option with tolerant default
	mode := strings.ToLower(strings.TrimSpace(args["mode"]))
	if mode == "" {
		mode = "overwrite"
	}
	if mode != "overwrite" && mode != "append" {
		return fmt.Errorf("invalid mode '%s'. Use 'overwrite' or 'append'", mode)
	}

	// Parse silent option with tolerant default
	silent := false
	if silentStr := args["silent"]; silentStr != "" {
		if parsedSilent, err := strconv.ParseBool(silentStr); err == nil {
			silent = parsedSilent
		}
		// If parsing fails, silent remains false (tolerant default)
	}

	// Content to write is exactly what the user provided
	content := input

	// Create parent directories if they don't exist
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", dir, err)
		}
	}

	var err error
	var bytesWritten int

	if mode == "append" {
		// Append mode - open file with append flag
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file '%s' for append: %w", filePath, err)
		}
		defer func() {
			_ = file.Close() // Ignore close error in defer
		}()

		written, err := file.WriteString(content)
		if err != nil {
			return fmt.Errorf("failed to append to file '%s': %w", filePath, err)
		}
		bytesWritten = written
	} else {
		// Overwrite mode - replace file contents
		contentBytes := []byte(content)
		err = os.WriteFile(filePath, contentBytes, 0644)
		if err != nil {
			return fmt.Errorf("failed to write file '%s': %w", filePath, err)
		}
		bytesWritten = len(contentBytes)
	}

	// Provide feedback unless silent mode is enabled
	if !silent {
		printer := c.createPrinter()
		if mode == "append" {
			printer.Success(fmt.Sprintf("Appended %d bytes to '%s'", bytesWritten, filePath))
		} else {
			printer.Success(fmt.Sprintf("Wrote %d bytes to '%s'", bytesWritten, filePath))
		}
	}

	return nil
}

// createPrinter creates a printer with theme service as style provider
func (c *WriteCommand) createPrinter() *output.Printer {
	// Try to get theme service as style provider
	themeService, err := services.GetGlobalThemeService()
	if err != nil {
		// Fall back to plain style provider
		return output.NewPrinter(output.WithStyles(output.NewPlainStyleProvider()))
	}

	return output.NewPrinter(output.WithStyles(themeService))
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&WriteCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register write command: %v", err))
	}
}
