package services

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TemporalDisplayService provides temporary console display functionality.
// It can show dynamic information like timers that automatically clean up after conditions are met.
type TemporalDisplayService struct {
	initialized    bool
	activeDisplays map[string]*Display
	mu             sync.RWMutex
}

// Display represents an active temporal display with its own goroutine and rendering logic.
type Display struct {
	id        string
	stopCh    chan struct{}
	ticker    *time.Ticker
	startTime time.Time
	condition func(elapsed time.Duration) bool
	renderer  func(elapsed time.Duration) string
	running   bool
	lastWidth int // Track width of last output for proper cleanup
}

// NewTemporalDisplayService creates a new TemporalDisplayService instance.
func NewTemporalDisplayService() *TemporalDisplayService {
	return &TemporalDisplayService{
		initialized:    false,
		activeDisplays: make(map[string]*Display),
	}
}

// Name returns the service name "temporal-display" for registration.
func (t *TemporalDisplayService) Name() string {
	return "temporal-display"
}

// Initialize sets up the TemporalDisplayService for operation.
func (t *TemporalDisplayService) Initialize() error {
	t.initialized = true
	return nil
}

// StartTimer starts a counting timer display that shows elapsed seconds until maxDuration is reached.
// The timer displays in green color and automatically stops and cleans up when maxDuration is reached.
func (t *TemporalDisplayService) StartTimer(id string, maxDuration time.Duration) error {
	if !t.initialized {
		return fmt.Errorf("temporal display service not initialized")
	}

	condition := func(elapsed time.Duration) bool {
		return elapsed >= maxDuration
	}

	renderer := func(elapsed time.Duration) string {
		seconds := int(elapsed.Seconds())
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
		return style.Render(fmt.Sprintf("%ds", seconds))
	}

	return t.StartCustomDisplay(id, condition, renderer)
}

// StartCustomDisplay starts a custom temporal display with user-defined condition and renderer functions.
// The condition function determines when to stop the display.
// The renderer function defines what content to show based on elapsed time.
func (t *TemporalDisplayService) StartCustomDisplay(id string, condition func(time.Duration) bool, renderer func(time.Duration) string) error {
	if !t.initialized {
		return fmt.Errorf("temporal display service not initialized")
	}

	if condition == nil || renderer == nil {
		return fmt.Errorf("condition and renderer functions cannot be nil")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Stop existing display with same ID if it exists
	if existing, exists := t.activeDisplays[id]; exists {
		t.stopDisplayUnsafe(existing)
	}

	// Create new display
	display := &Display{
		id:        id,
		stopCh:    make(chan struct{}),
		ticker:    time.NewTicker(100 * time.Millisecond), // Update every 100ms for smooth display
		startTime: time.Now(),
		condition: condition,
		renderer:  renderer,
		running:   true,
		lastWidth: 0,
	}

	t.activeDisplays[id] = display

	// Start display goroutine
	go t.runDisplay(display)

	return nil
}

// Stop stops and cleans up a specific temporal display by ID.
func (t *TemporalDisplayService) Stop(id string) error {
	if !t.initialized {
		return fmt.Errorf("temporal display service not initialized")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	display, exists := t.activeDisplays[id]
	if !exists {
		return fmt.Errorf("display with id '%s' not found", id)
	}

	t.stopDisplayUnsafe(display)
	return nil
}

// StopAll stops and cleans up all active temporal displays.
func (t *TemporalDisplayService) StopAll() error {
	if !t.initialized {
		return fmt.Errorf("temporal display service not initialized")
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for _, display := range t.activeDisplays {
		t.stopDisplayUnsafe(display)
	}

	return nil
}

// IsActive checks if a temporal display with the given ID is currently running.
func (t *TemporalDisplayService) IsActive(id string) bool {
	if !t.initialized {
		return false
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	display, exists := t.activeDisplays[id]
	return exists && display.running
}

// runDisplay runs the main display loop for a temporal display in its own goroutine.
func (t *TemporalDisplayService) runDisplay(display *Display) {
	defer func() {
		display.ticker.Stop()
		t.cleanupDisplay(display)

		// Remove from active displays
		t.mu.Lock()
		delete(t.activeDisplays, display.id)
		t.mu.Unlock()
	}()

	for {
		select {
		case <-display.stopCh:
			return
		case <-display.ticker.C:
			elapsed := time.Since(display.startTime)

			// Check if condition is met to stop
			if display.condition(elapsed) {
				return
			}

			// Render and display content
			content := display.renderer(elapsed)
			t.displayContent(display, content)
		}
	}
}

// displayContent outputs the content to the console, overwriting the previous line.
func (t *TemporalDisplayService) displayContent(display *Display, content string) {
	// Clear previous content if it was longer
	if display.lastWidth > 0 {
		clearLine := "\r" + strings.Repeat(" ", display.lastWidth) + "\r"
		_, _ = fmt.Fprint(os.Stdout, clearLine)
	}

	// Display new content
	_, _ = fmt.Fprint(os.Stdout, "\r"+content)

	// Track width for next cleanup
	display.lastWidth = len(content)
}

// cleanupDisplay clears the display line and restores normal output.
func (t *TemporalDisplayService) cleanupDisplay(display *Display) {
	if display.lastWidth > 0 {
		// Clear the line completely
		clearLine := "\r" + strings.Repeat(" ", display.lastWidth) + "\r"
		_, _ = fmt.Fprint(os.Stdout, clearLine)
	}
}

// stopDisplayUnsafe stops a display without acquiring locks (internal use only).
func (t *TemporalDisplayService) stopDisplayUnsafe(display *Display) {
	if display.running {
		display.running = false
		close(display.stopCh)
	}
}

func init() {
	// Register the TemporalDisplayService with the global registry
	if err := GlobalRegistry.RegisterService(NewTemporalDisplayService()); err != nil {
		panic(fmt.Sprintf("failed to register temporal display service: %v", err))
	}
}
