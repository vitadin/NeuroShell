// Package statemachine implements silent block boundary management for the stack-based execution engine.
// The SilentHandler manages silent boundaries and silent block state transitions.
package statemachine

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"neuroshell/internal/logger"
	"neuroshell/internal/services"
)

// SilentHandler manages silent block boundaries and output suppression.
// It provides a clean interface for managing silent block state.
type SilentHandler struct {
	// Services
	stackService *services.StackService
	// Logger
	logger *log.Logger
}

// NewSilentHandler creates a new silent handler with the required services.
func NewSilentHandler() *SilentHandler {
	sh := &SilentHandler{
		logger: logger.NewStyledLogger("SilentHandler"),
	}

	// Initialize services
	var err error
	sh.stackService, err = services.GetGlobalStackService()
	if err != nil {
		sh.logger.Error("Failed to get stack service", "error", err)
	}

	return sh
}

// GenerateUniqueSilentID generates a unique identifier for a silent block.
func (sh *SilentHandler) GenerateUniqueSilentID() string {
	if sh.stackService == nil {
		return "silent_id_0"
	}
	return fmt.Sprintf("silent_id_%d", sh.stackService.GetCurrentSilentDepth())
}

// PushSilentBoundary pushes silent boundary markers around a target command.
// This sets up the silent block structure on the stack.
func (sh *SilentHandler) PushSilentBoundary(silentID string, targetCommand string) {
	if sh.stackService == nil {
		return
	}

	sh.logger.Debug("Pushing silent boundary", "silentID", silentID, "targetCommand", targetCommand)

	// Push silent boundary markers around target command (reverse order for LIFO)
	sh.stackService.PushCommand("SILENT_BOUNDARY_END:" + silentID)
	sh.stackService.PushCommand(targetCommand)
	sh.stackService.PushCommand("SILENT_BOUNDARY_START:" + silentID)
}

// EnterSilentBlock enters a silent block with the given ID.
func (sh *SilentHandler) EnterSilentBlock(silentID string) {
	sh.logger.Debug("Entering silent block", "silentID", silentID)
	if sh.stackService != nil {
		sh.stackService.PushSilentBoundary(silentID)
	}
}

// ExitSilentBlock exits the silent block with the given ID.
func (sh *SilentHandler) ExitSilentBlock(silentID string) {
	sh.logger.Debug("Exiting silent block", "silentID", silentID)

	if sh.stackService != nil {
		sh.stackService.PopSilentBoundary()
	}
}

// IsInSilentBlock returns true if currently inside a silent block.
func (sh *SilentHandler) IsInSilentBlock() bool {
	if sh.stackService == nil {
		return false
	}
	return sh.stackService.IsInSilentBlock()
}

// GetCurrentSilentID returns the ID of the current silent block.
func (sh *SilentHandler) GetCurrentSilentID() string {
	if sh.stackService == nil {
		return ""
	}
	return sh.stackService.GetCurrentSilentID()
}

// IsSilentBoundaryMarker checks if a command is a silent boundary marker.
func (sh *SilentHandler) IsSilentBoundaryMarker(command string) (bool, string, bool) {
	if strings.HasPrefix(command, "SILENT_BOUNDARY_START:") {
		silentID := strings.TrimPrefix(command, "SILENT_BOUNDARY_START:")
		return true, silentID, true // isStart = true
	}
	if strings.HasPrefix(command, "SILENT_BOUNDARY_END:") {
		silentID := strings.TrimPrefix(command, "SILENT_BOUNDARY_END:")
		return true, silentID, false // isStart = false
	}
	return false, "", false
}
