package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrinterBasicOutput(t *testing.T) {
	buffer := NewCaptureBuffer()
	printer := NewPrinter(WithWriter(buffer), TestMode())

	// Test basic output methods
	printer.Print("hello")
	printer.Println("world")
	printer.Printf("number: %d", 42)

	result := buffer.String()

	if !strings.Contains(result, "hello") {
		t.Errorf("Expected output to contain 'hello', got: %s", result)
	}
	if !strings.Contains(result, "world\n") {
		t.Errorf("Expected output to contain 'world\\n', got: %s", result)
	}
	if !strings.Contains(result, "number: 42") {
		t.Errorf("Expected output to contain 'number: 42', got: %s", result)
	}
}

func TestPrinterSemanticOutput(t *testing.T) {
	buffer := NewCaptureBuffer()
	printer := NewPrinter(WithWriter(buffer), TestMode())

	// Test semantic output methods (should use plain prefixes in test mode)
	printer.Info("information")
	printer.Success("completed")
	printer.Warning("careful")
	printer.Error("failed")

	lines := buffer.Lines()

	// In test mode with plain styling, we expect prefixes
	expectedLines := []string{
		"ℹ information",
		"✓ completed",
		"⚠ careful",
		"✗ failed",
	}

	if len(lines) != len(expectedLines) {
		t.Fatalf("Expected %d lines, got %d: %v", len(expectedLines), len(lines), lines)
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("Line %d: expected '%s', got '%s'", i, expected, lines[i])
		}
	}
}

func TestPrinterWithMockStyleProvider(t *testing.T) {
	buffer := NewCaptureBuffer()
	mockProvider := NewMockStyleProvider()
	printer := NewPrinter(WithWriter(buffer), WithStyles(mockProvider))

	// Test that styles are applied when provider is available
	printer.Info("test message")
	printer.Success("success message")

	output := buffer.String()

	// Mock provider wraps text in [semantic]text[/semantic]
	if !strings.Contains(output, "[info]test message[/info]") {
		t.Errorf("Expected styled info output, got: %s", output)
	}
	if !strings.Contains(output, "[success]success message[/success]") {
		t.Errorf("Expected styled success output, got: %s", output)
	}
}

func TestPrinterWithUnavailableStyleProvider(t *testing.T) {
	buffer := NewCaptureBuffer()
	mockProvider := NewMockStyleProvider()
	mockProvider.SetAvailable(false) // Make provider unavailable

	printer := NewPrinter(WithWriter(buffer), WithStyles(mockProvider))

	printer.Info("test message")

	result := buffer.String()

	// Should fall back to plain style since provider is not available
	if !strings.Contains(result, "ℹ test message") {
		t.Errorf("Expected plain style fallback, got: %s", result)
	}
}

func TestPrinterPlainMode(t *testing.T) {
	buffer := NewCaptureBuffer()
	mockProvider := NewMockStyleProvider()
	printer := NewPrinter(WithWriter(buffer), WithStyles(mockProvider), PlainText())

	printer.Info("test message")
	printer.Success("success message")

	output := buffer.String()

	// Should use plain text even with available style provider
	if !strings.Contains(output, "ℹ test message") {
		t.Errorf("Expected plain text for info, got: %s", output)
	}
	if !strings.Contains(output, "✓ success message") {
		t.Errorf("Expected plain text for success, got: %s", output)
	}

	// Should not contain styled markup
	if strings.Contains(output, "[info]") || strings.Contains(output, "[success]") {
		t.Errorf("Should not contain styled markup in plain mode, got: %s", output)
	}
}

func TestPrinterJSONMode(t *testing.T) {
	buffer := NewCaptureBuffer()
	printer := NewPrinter(WithWriter(buffer), JSON())

	printer.Info("test message")
	printer.Error("error message")

	lines := buffer.Lines()

	// Should output JSON format
	if len(lines) != 2 {
		t.Fatalf("Expected 2 JSON lines, got %d: %v", len(lines), lines)
	}

	// Check that output contains JSON structure
	if !strings.Contains(lines[0], `"type":"info"`) {
		t.Errorf("Expected JSON with type:info, got: %s", lines[0])
	}
	if !strings.Contains(lines[0], `"message":"test message"`) {
		t.Errorf("Expected JSON with message, got: %s", lines[0])
	}
	if !strings.Contains(lines[1], `"type":"error"`) {
		t.Errorf("Expected JSON with type:error, got: %s", lines[1])
	}
}

func TestPrinterSilentMode(t *testing.T) {
	buffer := NewCaptureBuffer()
	printer := NewPrinter(WithWriter(buffer), Silent())

	printer.Info("test message")
	printer.Print("another message")
	printer.Error("error message")

	output := buffer.String()

	// Should produce no output in silent mode
	if output != "" {
		t.Errorf("Expected no output in silent mode, got: '%s'", output)
	}
}

func TestPrinterWithPrefix(t *testing.T) {
	buffer := NewCaptureBuffer()
	printer := NewPrinter(WithWriter(buffer), WithPrefix("[TEST] "), TestMode())

	printer.Info("message")

	output := buffer.String()

	if !strings.Contains(output, "[TEST] ℹ message") {
		t.Errorf("Expected prefixed output, got: %s", output)
	}
}

func TestCaptureOutput(t *testing.T) {
	// Test the convenience function
	output := CaptureOutput(func(p *Printer) {
		p.Info("captured message")
		p.Success("another message")
	})

	if !strings.Contains(output, "ℹ captured message") {
		t.Errorf("Expected captured info message, got: %s", output)
	}
	if !strings.Contains(output, "✓ another message") {
		t.Errorf("Expected captured success message, got: %s", output)
	}
}

func TestCaptureOutputWithStyles(t *testing.T) {
	mockProvider := NewMockStyleProvider()

	output := CaptureOutputWithStyles(mockProvider, func(p *Printer) {
		p.Info("styled message")
	})

	if !strings.Contains(output, "[info]styled message[/info]") {
		t.Errorf("Expected styled output, got: %s", output)
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Save original global printer
	originalPrinter := GetGlobalPrinter()
	defer SetGlobalPrinter(originalPrinter)

	// Configure global printer for testing
	buffer := NewCaptureBuffer()
	ConfigureGlobal(WithWriter(buffer), TestMode())

	// Test global functions
	Print("hello")
	Println("world")
	Info("info message")
	Success("success message")

	output := buffer.String()

	if !strings.Contains(output, "hello") {
		t.Errorf("Expected 'hello' in output, got: %s", output)
	}
	if !strings.Contains(output, "world\n") {
		t.Errorf("Expected 'world\\n' in output, got: %s", output)
	}
	if !strings.Contains(output, "ℹ info message") {
		t.Errorf("Expected info message in output, got: %s", output)
	}
	if !strings.Contains(output, "✓ success message") {
		t.Errorf("Expected success message in output, got: %s", output)
	}
}

func TestMockStyleProvider(t *testing.T) {
	provider := NewMockStyleProvider()

	// Test default behavior
	if !provider.IsAvailable() {
		t.Error("Mock provider should be available by default")
	}

	style := provider.GetStyle("info")
	result := style.Render("test")
	expected := "[info]test[/info]"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test setting availability
	provider.SetAvailable(false)
	if provider.IsAvailable() {
		t.Error("Mock provider should not be available after SetAvailable(false)")
	}

	// Test custom styles
	customStyle := &MockTextStyle{semantic: "custom"}
	provider.SetStyle("test", customStyle)

	retrievedStyle := provider.GetStyle("test")
	customResult := retrievedStyle.Render("message")
	expectedCustom := "[custom]message[/custom]"

	if customResult != expectedCustom {
		t.Errorf("Expected custom style '%s', got '%s'", expectedCustom, customResult)
	}
}

func TestCaptureBufferMethods(t *testing.T) {
	buffer := NewCaptureBuffer()

	// Test empty buffer
	if buffer.String() != "" {
		t.Error("New buffer should be empty")
	}
	if len(buffer.Lines()) != 0 {
		t.Error("New buffer should have no lines")
	}
	if buffer.Len() != 0 {
		t.Error("New buffer should have zero length")
	}

	// Write some data
	_, err := buffer.Write([]byte("line1\nline2\nline3"))
	if err != nil {
		t.Fatalf("Failed to write to buffer: %v", err)
	}

	if buffer.Len() == 0 {
		t.Error("Buffer should have length after writing")
	}

	lines := buffer.Lines()
	expectedLines := []string{"line1", "line2", "line3"}

	if len(lines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(lines))
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("Line %d: expected '%s', got '%s'", i, expected, lines[i])
		}
	}

	if !buffer.Contains("line2") {
		t.Error("Buffer should contain 'line2'")
	}
	if buffer.Contains("nonexistent") {
		t.Error("Buffer should not contain 'nonexistent'")
	}

	// Test reset
	buffer.Reset()
	if buffer.String() != "" {
		t.Error("Buffer should be empty after reset")
	}
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkPrinterPlainOutput(b *testing.B) {
	buffer := &bytes.Buffer{}
	printer := NewPrinter(WithWriter(buffer), PlainText())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		printer.Info("benchmark message")
		buffer.Reset()
	}
}

func BenchmarkPrinterStyledOutput(b *testing.B) {
	buffer := &bytes.Buffer{}
	mockProvider := NewMockStyleProvider()
	printer := NewPrinter(WithWriter(buffer), WithStyles(mockProvider))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		printer.Info("benchmark message")
		buffer.Reset()
	}
}
