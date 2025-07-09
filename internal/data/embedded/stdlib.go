// Package embedded provides access to embedded stdlib script files.
// This file enables loading NeuroShell stdlib scripts from the embedded filesystem,
// supporting the enhanced command resolution system with script-based commands.
package embedded

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"neuroshell/pkg/neurotypes"
)

// StdlibFS contains all embedded stdlib script files.
//
//go:embed stdlib/*.neuro
var StdlibFS embed.FS

// StdlibLoader implements the ScriptLoader interface for embedded stdlib scripts.
// It provides access to .neuro scripts embedded in the binary at compile time.
type StdlibLoader struct{}

// NewStdlibLoader creates a new StdlibLoader for accessing embedded stdlib scripts.
func NewStdlibLoader() *StdlibLoader {
	return &StdlibLoader{}
}

// LoadScript loads the content of an embedded stdlib script by filename.
// The filename should include the .neuro extension.
func (s *StdlibLoader) LoadScript(filename string) (string, error) {
	// Ensure filename has .neuro extension
	if !strings.HasSuffix(filename, ".neuro") {
		filename += ".neuro"
	}

	// Construct the embedded path
	scriptPath := filepath.Join("stdlib", filename)

	// Read from embedded filesystem
	content, err := StdlibFS.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("stdlib script not found: %s", filename)
	}

	return string(content), nil
}

// ListAvailableScripts returns a list of all available stdlib script names
// (without the .neuro extension) that can be loaded.
func (s *StdlibLoader) ListAvailableScripts() ([]string, error) {
	entries, err := StdlibFS.ReadDir("stdlib")
	if err != nil {
		return nil, fmt.Errorf("failed to read stdlib directory: %w", err)
	}

	var scripts []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".neuro") {
			// Remove .neuro extension for script name
			scriptName := strings.TrimSuffix(name, ".neuro")
			scripts = append(scripts, scriptName)
		}
	}

	return scripts, nil
}

// ScriptExists checks if a script with the given name exists in the embedded stdlib.
func (s *StdlibLoader) ScriptExists(name string) bool {
	// Ensure name has .neuro extension for checking
	filename := name
	if !strings.HasSuffix(filename, ".neuro") {
		filename += ".neuro"
	}

	scriptPath := filepath.Join("stdlib", filename)

	// Try to read the file to check existence
	_, err := StdlibFS.ReadFile(scriptPath)
	return err == nil
}

// GetScriptPath returns the embedded path for a script.
// For embedded scripts, this returns a virtual path indicating the embedded location.
func (s *StdlibLoader) GetScriptPath(name string) string {
	filename := name
	if !strings.HasSuffix(filename, ".neuro") {
		filename += ".neuro"
	}

	return fmt.Sprintf("embedded://stdlib/%s", filename)
}

// StdlibLoaderService provides the StdlibLoader as a NeuroShell service.
// This enables the loader to be registered with the service registry and
// accessed by other components that need embedded script loading capabilities.
type StdlibLoaderService struct {
	loader *StdlibLoader
}

// NewStdlibLoaderService creates a new service wrapper for the StdlibLoader.
func NewStdlibLoaderService() *StdlibLoaderService {
	return &StdlibLoaderService{
		loader: NewStdlibLoader(),
	}
}

// Name returns the service name for registration.
func (s *StdlibLoaderService) Name() string {
	return "stdlib-loader"
}

// Initialize sets up the StdlibLoaderService.
func (s *StdlibLoaderService) Initialize() error {
	// No initialization needed for embedded filesystem
	return nil
}

// GetLoader returns the underlying StdlibLoader for use by other components.
func (s *StdlibLoaderService) GetLoader() neurotypes.ScriptLoader {
	return s.loader
}

// LoadAllStdlibScripts loads all available stdlib scripts and returns them
// as a map of script names to their content. This is useful for bulk loading
// during system initialization.
func (s *StdlibLoaderService) LoadAllStdlibScripts() (map[string]string, error) {
	scripts := make(map[string]string)

	scriptNames, err := s.loader.ListAvailableScripts()
	if err != nil {
		return nil, fmt.Errorf("failed to list stdlib scripts: %w", err)
	}

	for _, name := range scriptNames {
		content, err := s.loader.LoadScript(name)
		if err != nil {
			// Log the error but continue with other scripts
			continue
		}

		scripts[name] = content
	}

	return scripts, nil
}
