package builtin

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"neuroshell/internal/commands"
	"neuroshell/internal/services"
	"neuroshell/pkg/neurotypes"
)

// CatCommand implements the \cat command for displaying file contents.
// It reads and displays the contents of files, similar to the Unix cat command.
type CatCommand struct{}

// Name returns the command name "cat" for registration and lookup.
func (c *CatCommand) Name() string {
	return "cat"
}

// ParseMode returns ParseModeKeyValue for key-value argument parsing.
func (c *CatCommand) ParseMode() neurotypes.ParseMode {
	return neurotypes.ParseModeKeyValue
}

// Description returns a brief description of what the cat command does.
func (c *CatCommand) Description() string {
	return "Display file contents with optional line limiting and variable storage"
}

// Usage returns the syntax and usage examples for the cat command.
func (c *CatCommand) Usage() string {
	return "\\cat[path=file_path, to=var_name, silent=true, lines=10, start=5] or \\cat file_path"
}

// HelpInfo returns structured help information for the cat command.
func (c *CatCommand) HelpInfo() neurotypes.HelpInfo {
	return neurotypes.HelpInfo{
		Command:     c.Name(),
		Description: c.Description(),
		Usage:       c.Usage(),
		ParseMode:   c.ParseMode(),
		Options: []neurotypes.HelpOption{
			{
				Name:        "path",
				Description: "Path to the file to display",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "to",
				Description: "Variable name to store the result",
				Required:    false,
				Type:        "string",
				Default:     "_output",
			},
			{
				Name:        "silent",
				Description: "Suppress console output",
				Required:    false,
				Type:        "bool",
				Default:     "false",
			},
			{
				Name:        "lines",
				Description: "Maximum number of lines to read",
				Required:    false,
				Type:        "int",
			},
			{
				Name:        "start",
				Description: "Starting line number (1-based)",
				Required:    false,
				Type:        "int",
				Default:     "1",
			},
		},
		Examples: []neurotypes.HelpExample{
			{
				Command:     "\\cat[path=config.txt]",
				Description: "Display contents of config.txt using bracket syntax",
			},
			{
				Command:     "\\cat /etc/hosts",
				Description: "Display contents of /etc/hosts using message syntax",
			},
			{
				Command:     "\\cat[path=${config_file}]",
				Description: "Display file using variable interpolation",
			},
			{
				Command:     "\\cat[path=log.txt, lines=20]",
				Description: "Display first 20 lines of log.txt",
			},
			{
				Command:     "\\cat[path=data.csv, start=10, lines=5]",
				Description: "Display 5 lines starting from line 10",
			},
			{
				Command:     "\\cat[path=report.txt, to=report_content, silent=true]",
				Description: "Store file contents in variable without console output",
			},
			{
				Command:     "\\cat[path=../data/report.csv]",
				Description: "Display relative path file contents",
			},
			{
				Command:     "\\set[file=README.md] \\cat[path=${file}]",
				Description: "Store file path in variable and display contents",
			},
		},
		Notes: []string{
			"File path can be specified via bracket option [path=...] or as message",
			"Supports both absolute and relative file paths",
			"Variable interpolation works for file paths",
			"File contents are stored in specified variable (default: ${_output})",
			"Use lines= option to limit output for large files",
			"Use start= option with lines= to read specific line ranges",
			"Use silent=true to store contents without console output",
			"Binary files are detected and handled gracefully",
			"Non-existent files will show appropriate error messages",
		},
	}
}

// Execute reads and displays the contents of a file with enhanced options.
// Supports both bracket syntax (\cat[path=file]) and message syntax (\cat file).
// Options:
//   - path: file path to read (can also be specified as message)
//   - to: store result in specified variable (default: ${_output})
//   - silent: suppress console output (default: false)
//   - lines: maximum number of lines to read
//   - start: starting line number (1-based, default: 1)
func (c *CatCommand) Execute(args map[string]string, input string) error {
	// Get file path from either args["path"] or input
	filePath := strings.TrimSpace(args["path"])
	if filePath == "" {
		filePath = strings.TrimSpace(input)
	}

	if filePath == "" {
		return fmt.Errorf("file path is required")
	}

	// Parse options with tolerant defaults
	targetVar := args["to"]
	if targetVar == "" {
		targetVar = "_output" // Default to system output variable
	}

	// Parse silent option with tolerant default
	silent := false
	if silentStr := args["silent"]; silentStr != "" {
		if parsedSilent, err := strconv.ParseBool(silentStr); err == nil {
			silent = parsedSilent
		}
		// If parsing fails, silent remains false (tolerant default)
	}

	// Parse lines option
	var maxLines int
	if linesStr := args["lines"]; linesStr != "" {
		if parsedLines, err := strconv.Atoi(linesStr); err == nil && parsedLines > 0 {
			maxLines = parsedLines
		}
		// If parsing fails or negative, maxLines remains 0 (no limit)
	}

	// Parse start option
	startLine := 1
	if startStr := args["start"]; startStr != "" {
		if parsedStart, err := strconv.Atoi(startStr); err == nil && parsedStart > 0 {
			startLine = parsedStart
		}
		// If parsing fails or invalid, startLine remains 1
	}

	// Check if file exists and is readable
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access file '%s': %w", filePath, err)
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a file", filePath)
	}

	var contentStr string
	var displayContent string

	// Read file contents with line limiting if specified
	if maxLines > 0 || startLine > 1 {
		contentStr, displayContent, err = c.readFileWithLineOptions(filePath, startLine, maxLines)
	} else {
		// Read entire file
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file '%s': %w", filePath, err)
		}

		// Check if file appears to be binary
		if !c.isTextFile(content) {
			return fmt.Errorf("'%s' appears to be a binary file (use a binary file viewer)", filePath)
		}

		contentStr = string(content)
		displayContent = contentStr
	}

	if err != nil {
		return err
	}

	// Get variable service - if not available, continue without storing (graceful degradation)
	if variableService, err := services.GetGlobalVariableService(); err == nil {
		// Store result in target variable
		if targetVar == "_output" || targetVar == "_error" || targetVar == "_status" {
			// Store in system variable (only for specific system variables)
			_ = variableService.SetSystemVariable(targetVar, contentStr)
		} else {
			// Store in user variable (including custom variables with _ prefix)
			_ = variableService.Set(targetVar, contentStr)
		}
		// Ignore storage errors to ensure cat never fails due to variable issues
	}

	// Output to console unless silent mode is enabled
	if !silent {
		fmt.Print(displayContent)
		// Ensure output ends with newline if it doesn't already
		if len(displayContent) > 0 && displayContent[len(displayContent)-1] != '\n' {
			fmt.Println()
		}
	}

	return nil
}

// readFileWithLineOptions reads a file with line range support
func (c *CatCommand) readFileWithLineOptions(filePath string, startLine, maxLines int) (string, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open file '%s': %w", filePath, err)
	}
	defer func() {
		_ = file.Close() // Ignore close error in defer
	}()

	scanner := bufio.NewScanner(file)
	var displayLines []string
	currentLine := 1

	// Read lines and collect them for display
	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line contains non-text content (simple binary detection)
		if !utf8.ValidString(line) {
			return "", "", fmt.Errorf("'%s' appears to be a binary file (use a binary file viewer)", filePath)
		}

		// Collect lines for display based on start/max parameters
		if currentLine >= startLine {
			displayLines = append(displayLines, line)
			if maxLines > 0 && len(displayLines) >= maxLines {
				break
			}
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("error reading file '%s': %w", filePath, err)
	}

	// Join display lines for both storage and console output
	displayContent := strings.Join(displayLines, "\n")

	return displayContent, displayContent, nil
}

// isTextFile performs basic binary file detection
func (c *CatCommand) isTextFile(content []byte) bool {
	// Check for null bytes (common in binary files)
	for _, b := range content {
		if b == 0 {
			return false
		}
	}

	// Check if content is valid UTF-8
	if !utf8.Valid(content) {
		return false
	}

	// Additional check: if more than 30% of bytes are non-printable, likely binary
	nonPrintable := 0
	for _, b := range content {
		if b < 32 && b != '\n' && b != '\r' && b != '\t' {
			nonPrintable++
		}
	}

	if len(content) > 0 && float64(nonPrintable)/float64(len(content)) > 0.3 {
		return false
	}

	return true
}

func init() {
	if err := commands.GetGlobalRegistry().Register(&CatCommand{}); err != nil {
		panic(fmt.Sprintf("failed to register cat command: %v", err))
	}
}
