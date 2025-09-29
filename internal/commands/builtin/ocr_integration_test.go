package builtin

import (
	"os"
	"strings"
	"testing"

	"neuroshell/internal/services"
)

// TestOCRCommand_Integration tests the OCR command with real DeepInfra API
// This test requires DEEPINFRA_API_KEY environment variable to be set
func TestOCRCommand_Integration(t *testing.T) {
	// Check if API key is available
	apiKey := os.Getenv("DEEPINFRA_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: DEEPINFRA_API_KEY environment variable not set")
	}

	// Check if sample PDF exists
	// Try both relative path from project root and from test directory
	samplePDF := "../../../assets/sample.pdf"
	if _, err := os.Stat(samplePDF); os.IsNotExist(err) {
		// Try alternative path
		samplePDF = "assets/sample.pdf"
		if _, err := os.Stat(samplePDF); os.IsNotExist(err) {
			t.Skipf("Skipping integration test: sample PDF not found")
		}
	}

	// Initialize HTTP service for integration test
	setupIntegrationTest(t)

	// Create OCR command
	cmd := &OCRCommand{}

	// Test options
	options := map[string]string{
		"pdf":     samplePDF,
		"pages":   "1",
		"output":  "test_integration_output.md",
		"api_key": apiKey,
		"to":      "integration_result",
	}

	// Clean up output file after test
	defer func() {
		_ = os.Remove("test_integration_output.md")
	}()

	// Execute OCR command
	err := cmd.Execute(options, "")
	if err != nil {
		// Variable service might not be available in test environment
		// If the error is only about variable setting, we can continue
		if !strings.Contains(err.Error(), "variable service not found") &&
			!strings.Contains(err.Error(), "failed to get variable service") &&
			!strings.Contains(err.Error(), "failed to set variable") {
			t.Fatalf("OCR command execution failed: %v", err)
		}
		t.Logf("Variable service not available (expected in test environment): %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat("test_integration_output.md"); os.IsNotExist(err) {
		t.Fatal("Output file was not created")
	}

	// Read output file
	content, err := os.ReadFile("test_integration_output.md")
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify the output is not empty
	if len(contentStr) == 0 {
		t.Error("Output file is empty")
	}

	// Since this is the Bitcoin whitepaper, verify key content is present
	// We check for essential elements that should be in the first page
	expectedKeywords := []string{
		"Bitcoin",
		"Peer-to-Peer",
		"Electronic Cash System",
		"Satoshi Nakamoto",
		"Abstract",
		"peer-to-peer",
		"digital signatures",
		"double-spending",
	}

	missingKeywords := []string{}
	for _, keyword := range expectedKeywords {
		if !containsCaseInsensitive(contentStr, keyword) {
			missingKeywords = append(missingKeywords, keyword)
		}
	}

	// Allow up to 2 missing keywords due to OCR randomness/variations
	if len(missingKeywords) > 2 {
		t.Errorf("Too many critical keywords missing from OCR output (%d missing): %v", len(missingKeywords), missingKeywords)
	}

	// Verify front matter is present
	if !strings.Contains(contentStr, "---") {
		t.Error("Front matter section not found in output")
	}

	// Verify some expected front matter fields
	frontMatterFields := []string{
		"primary_language",
		"is_rotation_valid",
		"is_table",
		"is_diagram",
	}

	missingFields := []string{}
	for _, field := range frontMatterFields {
		if !strings.Contains(contentStr, field) {
			missingFields = append(missingFields, field)
		}
	}

	// Allow up to 1 missing field due to variations in output format
	if len(missingFields) > 1 {
		t.Errorf("Too many front matter fields missing (%d missing): %v", len(missingFields), missingFields)
	}

	// Verify content length is reasonable (should be at least 500 characters for first page)
	if len(contentStr) < 500 {
		t.Errorf("Output content too short (got %d characters, expected at least 500)", len(contentStr))
	}

	t.Logf("OCR integration test passed successfully")
	t.Logf("Output length: %d characters", len(contentStr))
	t.Logf("Missing keywords: %v (allowed up to 2)", missingKeywords)
	t.Logf("Missing front matter fields: %v (allowed up to 1)", missingFields)
}

// TestOCRCommand_Integration_MultiplePages tests OCR with multiple pages
func TestOCRCommand_Integration_MultiplePages(t *testing.T) {
	// Check if API key is available
	apiKey := os.Getenv("DEEPINFRA_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: DEEPINFRA_API_KEY environment variable not set")
	}

	// Check if sample PDF exists
	// Try both relative path from project root and from test directory
	samplePDF := "../../../assets/sample.pdf"
	if _, err := os.Stat(samplePDF); os.IsNotExist(err) {
		// Try alternative path
		samplePDF = "assets/sample.pdf"
		if _, err := os.Stat(samplePDF); os.IsNotExist(err) {
			t.Skipf("Skipping integration test: sample PDF not found")
		}
	}

	// Initialize HTTP service for integration test
	setupIntegrationTest(t)

	// Create OCR command
	cmd := &OCRCommand{}

	// Test options - process first 2 pages
	options := map[string]string{
		"pdf":     samplePDF,
		"pages":   "1-2",
		"output":  "test_multipage_output.md",
		"api_key": apiKey,
		"to":      "multipage_result",
	}

	// Clean up output file after test
	defer func() {
		_ = os.Remove("test_multipage_output.md")
	}()

	// Execute OCR command
	err := cmd.Execute(options, "")
	if err != nil {
		// Variable service might not be available in test environment
		// If the error is only about variable setting, we can continue
		if !strings.Contains(err.Error(), "variable service not found") &&
			!strings.Contains(err.Error(), "failed to get variable service") &&
			!strings.Contains(err.Error(), "failed to set variable") {
			t.Fatalf("OCR command execution failed: %v", err)
		}
		t.Logf("Variable service not available (expected in test environment): %v", err)
	}

	// Read output file
	content, err := os.ReadFile("test_multipage_output.md")
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify output contains page separator
	pageCount := strings.Count(contentStr, "---\n\n")
	if pageCount < 1 {
		t.Errorf("Expected multiple page sections, but found %d separators", pageCount)
	}

	// Verify content is longer than single page (should be at least 1000 characters)
	if len(contentStr) < 1000 {
		t.Errorf("Multi-page output too short (got %d characters, expected at least 1000)", len(contentStr))
	}

	// Check for content from both pages
	// First page should have abstract/introduction
	// Second page should have more technical content
	page1Keywords := []string{"Abstract", "Bitcoin"}
	page2Keywords := []string{"transaction", "network"} // Common in page 2

	foundPage1 := 0
	for _, keyword := range page1Keywords {
		if containsCaseInsensitive(contentStr, keyword) {
			foundPage1++
		}
	}

	foundPage2 := 0
	for _, keyword := range page2Keywords {
		if containsCaseInsensitive(contentStr, keyword) {
			foundPage2++
		}
	}

	// Should have content from both pages (at least 1 keyword from each)
	if foundPage1 == 0 || foundPage2 == 0 {
		t.Errorf("Missing content from pages: page1=%d/2, page2=%d/2", foundPage1, foundPage2)
	}

	t.Logf("Multi-page OCR integration test passed successfully")
	t.Logf("Output length: %d characters", len(contentStr))
	t.Logf("Page separators found: %d", pageCount)
}

// TestOCRCommand_Integration_CustomModel tests OCR with custom model parameter
func TestOCRCommand_Integration_CustomModel(t *testing.T) {
	// Check if API key is available
	apiKey := os.Getenv("DEEPINFRA_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: DEEPINFRA_API_KEY environment variable not set")
	}

	// Check if sample PDF exists
	// Try both relative path from project root and from test directory
	samplePDF := "../../../assets/sample.pdf"
	if _, err := os.Stat(samplePDF); os.IsNotExist(err) {
		// Try alternative path
		samplePDF = "assets/sample.pdf"
		if _, err := os.Stat(samplePDF); os.IsNotExist(err) {
			t.Skipf("Skipping integration test: sample PDF not found")
		}
	}

	// Initialize HTTP service for integration test
	setupIntegrationTest(t)

	// Create OCR command
	cmd := &OCRCommand{}

	// Test with explicit model specification
	options := map[string]string{
		"pdf":     samplePDF,
		"pages":   "1",
		"output":  "test_custom_model_output.md",
		"api_key": apiKey,
		"model":   "allenai/olmOCR-7B-0825", // Explicitly specify the model
		"to":      "custom_model_result",
	}

	// Clean up output file after test
	defer func() {
		_ = os.Remove("test_custom_model_output.md")
	}()

	// Execute OCR command
	err := cmd.Execute(options, "")
	if err != nil {
		// Variable service might not be available in test environment
		// If the error is only about variable setting, we can continue
		if !strings.Contains(err.Error(), "variable service not found") &&
			!strings.Contains(err.Error(), "failed to get variable service") &&
			!strings.Contains(err.Error(), "failed to set variable") {
			t.Fatalf("OCR command with custom model failed: %v", err)
		}
		t.Logf("Variable service not available (expected in test environment): %v", err)
	}

	// Verify output file exists and has content
	content, err := os.ReadFile("test_custom_model_output.md")
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) < 100 {
		t.Errorf("Custom model output too short: %d bytes", len(content))
	}

	t.Logf("Custom model OCR integration test passed successfully")
	t.Logf("Output length: %d characters", len(content))
}

// containsCaseInsensitive checks if a string contains a substring (case insensitive)
func containsCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// setupIntegrationTest initializes the services needed for integration testing
func setupIntegrationTest(t *testing.T) {
	t.Helper()

	registry := services.GetGlobalRegistry()

	// Register HTTP service - this is required for OCR API calls
	httpService := services.NewHTTPRequestService()
	if err := httpService.Initialize(); err != nil {
		t.Fatalf("Failed to initialize HTTP service: %v", err)
	}
	if err := registry.RegisterService(httpService); err != nil {
		// Service might already be registered, which is OK
		if !strings.Contains(err.Error(), "already registered") {
			t.Fatalf("Failed to register HTTP service: %v", err)
		}
	}
}
