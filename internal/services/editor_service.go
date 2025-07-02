package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"neuroshell/internal/logger"
	"neuroshell/pkg/neurotypes"
)

// EditorService provides external editor integration for NeuroShell.
type EditorService struct {
	initialized bool
	tempDir     string
}

// NewEditorService creates a new EditorService instance.
func NewEditorService() *EditorService {
	return &EditorService{
		initialized: false,
	}
}

// Name returns the service name "editor" for registration.
func (e *EditorService) Name() string {
	return "editor"
}

// Initialize sets up the EditorService for operation.
func (e *EditorService) Initialize(_ neurotypes.Context) error {
	// Create a temporary directory for editor files
	tempDir, err := os.MkdirTemp("", "neuroshell-editor-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	e.tempDir = tempDir
	e.initialized = true

	logger.Debug("EditorService initialized", "tempDir", tempDir)
	return nil
}

// OpenEditor opens the configured editor with a temporary file and returns the content.
func (e *EditorService) OpenEditor(ctx neurotypes.Context) (string, error) {
	if !e.initialized {
		return "", fmt.Errorf("editor service not initialized")
	}

	// Get the editor command
	editorCmd := e.getEditorCommand(ctx)
	if editorCmd == "" {
		return "", fmt.Errorf("no editor configured or found")
	}

	// Create a temporary file
	tempFile, err := e.createTempFile()
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			logger.Error("Failed to remove temp file", "error", err, "file", tempFile)
		}
	}() // Clean up after use

	logger.Debug("Opening editor", "editor", editorCmd, "file", tempFile)

	// Execute the editor
	if err := e.executeEditor(editorCmd, tempFile); err != nil {
		return "", fmt.Errorf("editor execution failed: %w", err)
	}

	// Read the content from the file
	content, err := os.ReadFile(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to read editor content: %w", err)
	}

	contentStr := strings.TrimSpace(string(content))
	logger.Debug("Editor content retrieved", "length", len(contentStr))

	return contentStr, nil
}

// getEditorCommand determines which editor to use based on configuration and environment.
func (e *EditorService) getEditorCommand(ctx neurotypes.Context) string {
	// First, check if user has set a preferred editor via \set
	if editorVar, err := ctx.GetVariable("@editor"); err == nil && editorVar != "" {
		logger.Debug("Using configured editor", "editor", editorVar)
		return editorVar
	}

	// Check environment variable
	if editor := os.Getenv("EDITOR"); editor != "" {
		logger.Debug("Using EDITOR environment variable", "editor", editor)
		return editor
	}

	// Check for common editors in PATH
	commonEditors := []string{"nvim", "vim", "nano", "code", "emacs"}
	for _, editor := range commonEditors {
		if _, err := exec.LookPath(editor); err == nil {
			logger.Debug("Found editor in PATH", "editor", editor)
			return editor
		}
	}

	logger.Debug("No editor found")
	return ""
}

// createTempFile creates a temporary file for editor content.
func (e *EditorService) createTempFile() (string, error) {
	tempFile := filepath.Join(e.tempDir, "neuroshell-input.txt")

	// Create the file with some helpful content
	initialContent := `# NeuroShell Editor Mode
# Enter your message or command below.
# Save and exit to capture the content.

`

	if err := os.WriteFile(tempFile, []byte(initialContent), 0600); err != nil {
		return "", fmt.Errorf("failed to write initial content: %w", err)
	}

	return tempFile, nil
}

// createTempFileWithContent creates a temporary file with specific initial content.
func (e *EditorService) createTempFileWithContent(content string) (string, error) {
	tempFile := filepath.Join(e.tempDir, "neuroshell-input.txt")

	// If content is empty, add helpful header
	if strings.TrimSpace(content) == "" {
		content = `# NeuroShell Editor Mode
# Enter your message or command below.
# Save and exit to capture the content.

`
	}

	if err := os.WriteFile(tempFile, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("failed to write initial content: %w", err)
	}

	return tempFile, nil
}

// OpenEditorWithContent opens the editor with initial content and returns the edited content.
func (e *EditorService) OpenEditorWithContent(ctx neurotypes.Context, initialContent string) (string, error) {
	if !e.initialized {
		return "", fmt.Errorf("editor service not initialized")
	}

	// Get the editor command
	editorCmd := e.getEditorCommand(ctx)
	if editorCmd == "" {
		return "", fmt.Errorf("no editor configured or found")
	}

	// Create a temporary file with initial content
	tempFile, err := e.createTempFileWithContent(initialContent)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			logger.Error("Failed to remove temp file", "error", err, "file", tempFile)
		}
	}() // Clean up after use

	logger.Debug("Opening editor with content", "editor", editorCmd, "file", tempFile, "contentLength", len(initialContent))

	// Execute the editor
	if err := e.executeEditor(editorCmd, tempFile); err != nil {
		return "", fmt.Errorf("editor execution failed: %w", err)
	}

	// Read the content from the file
	content, err := os.ReadFile(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to read editor content: %w", err)
	}

	contentStr := strings.TrimSpace(string(content))
	logger.Debug("Editor content retrieved", "length", len(contentStr))

	return contentStr, nil
}

// executeEditor runs the editor command and waits for it to complete.
func (e *EditorService) executeEditor(editorCmd, filePath string) error {
	// Split the editor command to handle arguments
	parts := strings.Fields(editorCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty editor command")
	}

	cmd := exec.Command(parts[0], append(parts[1:], filePath)...)

	// Connect stdin, stdout, stderr to allow interactive editing
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute and wait for completion
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor command failed: %w", err)
	}

	return nil
}

// Cleanup removes the temporary directory and files.
func (e *EditorService) Cleanup() error {
	if e.tempDir != "" {
		if err := os.RemoveAll(e.tempDir); err != nil {
			logger.Error("Failed to cleanup editor temp directory", "error", err, "tempDir", e.tempDir)
			return err
		}
		logger.Debug("EditorService cleanup completed", "tempDir", e.tempDir)
	}
	return nil
}
