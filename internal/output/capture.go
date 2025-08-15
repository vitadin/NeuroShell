package output

import (
	"bytes"
	"strings"
)

// CaptureBuffer is a thread-safe buffer for capturing output during tests.
type CaptureBuffer struct {
	buf bytes.Buffer
}

// NewCaptureBuffer creates a new capture buffer.
func NewCaptureBuffer() *CaptureBuffer {
	return &CaptureBuffer{}
}

// Write implements io.Writer for capturing output.
func (c *CaptureBuffer) Write(p []byte) (n int, err error) {
	return c.buf.Write(p)
}

// String returns the captured output as a string.
func (c *CaptureBuffer) String() string {
	return c.buf.String()
}

// Lines returns the captured output split into lines.
func (c *CaptureBuffer) Lines() []string {
	content := c.String()
	if content == "" {
		return []string{}
	}
	return strings.Split(strings.TrimSuffix(content, "\n"), "\n")
}

// Reset clears the captured output.
func (c *CaptureBuffer) Reset() {
	c.buf.Reset()
}

// Len returns the number of bytes captured.
func (c *CaptureBuffer) Len() int {
	return c.buf.Len()
}

// Contains checks if the captured output contains the given text.
func (c *CaptureBuffer) Contains(text string) bool {
	return strings.Contains(c.String(), text)
}

// CaptureOutput captures output from a function that uses a Printer.
// This is a convenience function for testing.
func CaptureOutput(fn func(*Printer)) string {
	buffer := NewCaptureBuffer()
	printer := NewPrinter(WithWriter(buffer), TestMode())
	fn(printer)
	return buffer.String()
}

// CaptureOutputWithStyles captures output from a function using the provided StyleProvider.
func CaptureOutputWithStyles(provider StyleProvider, fn func(*Printer)) string {
	buffer := NewCaptureBuffer()
	printer := NewPrinter(WithWriter(buffer), WithStyles(provider))
	fn(printer)
	return buffer.String()
}

// MockStyleProvider is a simple mock implementation of StyleProvider for testing.
type MockStyleProvider struct {
	available bool
	styles    map[string]TextStyle
}

// NewMockStyleProvider creates a new mock style provider.
func NewMockStyleProvider() *MockStyleProvider {
	return &MockStyleProvider{
		available: true,
		styles:    make(map[string]TextStyle),
	}
}

// SetStyle sets a style for the given semantic type.
func (m *MockStyleProvider) SetStyle(semantic string, style TextStyle) {
	m.styles[semantic] = style
}

// SetAvailable sets whether the provider is available.
func (m *MockStyleProvider) SetAvailable(available bool) {
	m.available = available
}

// GetStyle implements StyleProvider.GetStyle.
func (m *MockStyleProvider) GetStyle(semantic string) TextStyle {
	if style, exists := m.styles[semantic]; exists {
		return style
	}
	// Return a simple mock style that wraps text in brackets
	return &MockTextStyle{semantic: semantic}
}

// IsAvailable implements StyleProvider.IsAvailable.
func (m *MockStyleProvider) IsAvailable() bool {
	return m.available
}

// MockTextStyle is a simple mock implementation of TextStyle for testing.
type MockTextStyle struct {
	semantic string
}

// Render implements TextStyle.Render by wrapping text in brackets with semantic info.
func (m *MockTextStyle) Render(text string) string {
	return "[" + m.semantic + "]" + text + "[/" + m.semantic + "]"
}
