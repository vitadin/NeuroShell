// Package statemachine implements try block error boundary management for the stack-based execution engine.
// The TryHandler manages error boundaries, error capture, and try block state transitions.
package statemachine

import (
	"fmt"
	"strings"

	"neuroshell/internal/logger"
	"neuroshell/internal/services"

	"github.com/charmbracelet/log"
)

// TryHandler manages try block error boundaries and error capture.
// It provides a clean interface for managing try block state and error handling.
type TryHandler struct {
	// Services
	stackService    *services.StackService
	variableService *services.VariableService
	// Logger
	logger *log.Logger
}

// NewTryHandler creates a new try handler with the required services.
func NewTryHandler() *TryHandler {
	th := &TryHandler{
		logger: logger.NewStyledLogger("TryHandler"),
	}

	// Initialize services
	var err error
	th.stackService, err = services.GetGlobalStackService()
	if err != nil {
		th.logger.Error("Failed to get stack service", "error", err)
	}

	th.variableService, err = services.GetGlobalVariableService()
	if err != nil {
		th.logger.Error("Failed to get variable service", "error", err)
	}

	return th
}

// GenerateUniqueTryID generates a unique identifier for a try block.
func (th *TryHandler) GenerateUniqueTryID() string {
	if th.stackService == nil {
		return "try_id_0"
	}
	return fmt.Sprintf("try_id_%d", th.stackService.GetCurrentTryDepth())
}

// PushTryBoundary pushes error boundary markers around a target command.
// This sets up the try block structure on the stack.
func (th *TryHandler) PushTryBoundary(tryID string, targetCommand string) {
	if th.stackService == nil {
		return
	}

	th.logger.Debug("Pushing try boundary", "tryID", tryID, "targetCommand", targetCommand)

	// Push error boundary markers around target command (reverse order for LIFO)
	th.stackService.PushCommand("ERROR_BOUNDARY_END:" + tryID)
	th.stackService.PushCommand(targetCommand)
	th.stackService.PushCommand("ERROR_BOUNDARY_START:" + tryID)
}

// HandleTryError handles errors that occur within a try block.
// It sets the appropriate status and error variables.
func (th *TryHandler) HandleTryError(err error) {
	th.logger.Debug("Handling try block error", "error", err)

	if th.variableService == nil {
		return
	}

	// Set error variables exactly like current implementation
	_ = th.variableService.SetSystemVariable("_status", "1")

	// Unwrap error messages to get the original error
	errorMsg := err.Error()
	if strings.HasPrefix(errorMsg, "command execution failed") {
		// Extract the original error message after the colon and space
		if idx := strings.Index(errorMsg, ": "); idx != -1 {
			errorMsg = errorMsg[idx+2:]
		}
	} else if strings.HasPrefix(errorMsg, "command resolution failed") {
		// Extract the original error message after the colon and space
		if idx := strings.Index(errorMsg, ": "); idx != -1 {
			errorMsg = errorMsg[idx+2:]
		}
	}
	_ = th.variableService.SetSystemVariable("_error", errorMsg)

	// Mark the try block as having captured an error
	if th.stackService != nil {
		th.stackService.SetTryErrorCaptured()
	}
}

// SkipToTryBlockEnd skips all commands until the end of the current try block.
func (th *TryHandler) SkipToTryBlockEnd() {
	if th.stackService == nil {
		return
	}

	currentTryID := th.stackService.GetCurrentTryID()
	if currentTryID == "" {
		return
	}

	th.logger.Debug("Skipping to try block end", "tryID", currentTryID)

	// DEBUG: Show current stack state
	stackContents := th.stackService.PeekStack()
	th.logger.Debug("Current stack contents before skipping", "stack", stackContents, "stackSize", len(stackContents))

	skipCount := 0
	// Pop commands until we find the matching ERROR_BOUNDARY_END
	for !th.stackService.IsEmpty() {
		command, hasCommand := th.stackService.PopCommand()
		if !hasCommand {
			th.logger.Debug("No more commands to pop, breaking")
			break
		}

		skipCount++
		th.logger.Debug("Skipping command", "command", command, "skipCount", skipCount, "lookingFor", "ERROR_BOUNDARY_END:"+currentTryID)

		// CRITICAL FIX: Check for silent boundary markers and process them properly
		// When skipping through commands, we must handle silent boundary markers to avoid state corruption
		switch {
		case strings.HasPrefix(command, "SILENT_BOUNDARY_START:"):
			silentID := strings.TrimPrefix(command, "SILENT_BOUNDARY_START:")
			th.logger.Debug("Processing silent boundary start while skipping", "silentID", silentID)
			th.stackService.PushSilentBoundary(silentID)
		case strings.HasPrefix(command, "SILENT_BOUNDARY_END:"):
			silentID := strings.TrimPrefix(command, "SILENT_BOUNDARY_END:")
			th.logger.Debug("Processing silent boundary end while skipping", "silentID", silentID)
			th.stackService.PopSilentBoundary()
		case command == "ERROR_BOUNDARY_END:"+currentTryID:
			th.logger.Debug("Found matching try block end", "tryID", currentTryID, "totalSkipped", skipCount)
			th.ExitTryBlock(currentTryID)
			return
		}

		// Check if we're skipping too many commands (potential infinite loop detection)
		if skipCount > 1000 {
			th.logger.Error("POTENTIAL INFINITE LOOP: Skipped too many commands looking for try block end", "tryID", currentTryID, "skipCount", skipCount)
			break
		}

		// Skip any other commands in the try block
	}

	th.logger.Error("Never found matching try block end", "tryID", currentTryID, "totalSkipped", skipCount)
}

// EnterTryBlock enters a try block with the given ID.
func (th *TryHandler) EnterTryBlock(tryID string) {
	th.logger.Debug("Entering try block", "tryID", tryID)
	if th.stackService != nil {
		th.stackService.PushErrorBoundary(tryID)
	}
}

// ExitTryBlock exits the try block with the given ID.
func (th *TryHandler) ExitTryBlock(tryID string) {
	th.logger.Debug("Exiting try block", "tryID", tryID)

	// Set success variables if no error was captured
	if th.variableService != nil && th.stackService != nil && !th.stackService.IsTryErrorCaptured() {
		_ = th.variableService.SetSystemVariable("_status", "0")
		_ = th.variableService.SetSystemVariable("_error", "")
	}

	if th.stackService != nil {
		th.stackService.PopErrorBoundary()
	}
}

// SetupEmptyTryCommand sets up variables for an empty try command.
func (th *TryHandler) SetupEmptyTryCommand() {
	if th.variableService == nil {
		return
	}

	// Empty try command - set success variables
	_ = th.variableService.SetSystemVariable("_status", "0")
	_ = th.variableService.SetSystemVariable("_error", "")
	_ = th.variableService.SetSystemVariable("_output", "")
}

// IsInTryBlock returns true if currently inside a try block.
func (th *TryHandler) IsInTryBlock() bool {
	if th.stackService == nil {
		return false
	}
	return th.stackService.IsInTryBlock()
}

// GetCurrentTryID returns the ID of the current try block.
func (th *TryHandler) GetCurrentTryID() string {
	if th.stackService == nil {
		return ""
	}
	return th.stackService.GetCurrentTryID()
}

// IsErrorBoundaryMarker checks if a command is an error boundary marker.
func (th *TryHandler) IsErrorBoundaryMarker(command string) (bool, string, bool) {
	if strings.HasPrefix(command, "ERROR_BOUNDARY_START:") {
		tryID := strings.TrimPrefix(command, "ERROR_BOUNDARY_START:")
		return true, tryID, true // isStart = true
	}
	if strings.HasPrefix(command, "ERROR_BOUNDARY_END:") {
		tryID := strings.TrimPrefix(command, "ERROR_BOUNDARY_END:")
		return true, tryID, false // isStart = false
	}
	return false, "", false
}
