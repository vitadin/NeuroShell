package llm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/services"
)

// TestMain checks for API key availability and skips all tests if not available
func TestMain(m *testing.M) {
	// Check for ZAI API key availability
	apiKey := os.Getenv("Z_DOT_AI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ZAI_API_KEY")
	}

	if apiKey == "" {
		// Gracefully skip all integration tests if no API key
		os.Exit(0)
	}

	// Run the tests
	os.Exit(m.Run())
}

// skipIfNoAPIKey is a helper function to skip individual tests if no API key is available
func skipIfNoAPIKey(t *testing.T) {
	apiKey := os.Getenv("Z_DOT_AI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ZAI_API_KEY")
	}
	if apiKey == "" {
		t.Skip("Skipping integration test: no ZAI API key found")
	}
}

// withTimeout creates a test context with 60 second timeout for API calls
func withTimeout(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// setupTestEnvironment creates a test environment with required services
func setupTestEnvironment(t *testing.T) (*ZaiTranslateCommand, *services.VariableService, func()) {
	// Save current registry
	originalRegistry := services.GetGlobalRegistry()

	// Set up test registry with required services
	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(t, err)
	err = testRegistry.RegisterService(services.NewHTTPRequestService())
	require.NoError(t, err)
	services.SetGlobalRegistry(testRegistry)

	// Initialize services
	err = testRegistry.InitializeAll()
	require.NoError(t, err)

	// Get variable service
	variableService, err := testRegistry.GetService("variable")
	require.NoError(t, err)

	// Set API key from environment
	apiKey := os.Getenv("Z_DOT_AI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ZAI_API_KEY")
	}
	require.NotEmpty(t, apiKey, "API key required for integration tests")

	err = variableService.(*services.VariableService).Set("os.Z_DOT_AI_API_KEY", apiKey)
	require.NoError(t, err)

	cmd := &ZaiTranslateCommand{}

	// Return cleanup function
	cleanup := func() {
		services.SetGlobalRegistry(originalRegistry)
	}

	return cmd, variableService.(*services.VariableService), cleanup
}

// verifyTranslationVariables checks that all expected variables are set after translation
func verifyTranslationVariables(t *testing.T, variableService *services.VariableService, expectTranslation bool) {
	// Check language variables are always set
	sourceLanguages, err := variableService.Get("_zai_source_languages")
	assert.NoError(t, err)
	assert.NotEmpty(t, sourceLanguages)
	assert.Contains(t, sourceLanguages, "auto")
	assert.Contains(t, sourceLanguages, "en")
	assert.Contains(t, sourceLanguages, "zh-CN")

	targetLanguages, err := variableService.Get("_zai_target_languages")
	assert.NoError(t, err)
	assert.NotEmpty(t, targetLanguages)
	assert.Contains(t, targetLanguages, "en")
	assert.Contains(t, targetLanguages, "zh-CN")

	if expectTranslation {
		// Check translation output variables
		output, err := variableService.Get("_output")
		assert.NoError(t, err)
		assert.NotEmpty(t, output)

		translationId, err := variableService.Get("_translation_id")
		assert.NoError(t, err)
		assert.NotEmpty(t, translationId)

		tokensUsed, err := variableService.Get("_tokens_used")
		assert.NoError(t, err)
		assert.NotEmpty(t, tokensUsed)
	}
}

func TestZaiTranslateIntegration_GeneralStrategy_EnglishToChinese(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test general translation strategy with the Robert Frost quote
	options := map[string]string{
		"source":   "en",
		"target":   "zh-CN",
		"strategy": "general",
	}
	input := "Two roads diverged in a wood, and I took the one less traveled by, and that has made all the difference."

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	// Verify translation variables
	verifyTranslationVariables(t, variableService, true)

	// Check that output contains Chinese characters
	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	// Should contain Chinese characters (basic check for Chinese translation)
	assert.True(t, containsChinese(output), "Output should contain Chinese characters")
}

func TestZaiTranslateIntegration_ParaphraseStrategy_EnglishToChinese(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":   "en",
		"target":   "zh-CN",
		"strategy": "paraphrase",
	}
	input := "Two roads diverged in a wood, and I took the one less traveled by, and that has made all the difference."

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Paraphrase output should contain Chinese characters")
}

func TestZaiTranslateIntegration_TwoStepStrategy_EnglishToChinese(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":   "en",
		"target":   "zh-CN",
		"strategy": "two_step",
	}
	input := "The quick brown fox jumps over the lazy dog."

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Two-step output should contain Chinese characters")
}

func TestZaiTranslateIntegration_ThreeStepStrategy_EnglishToChinese(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":   "en",
		"target":   "zh-CN",
		"strategy": "three_step",
	}
	input := "In the midst of winter, I found there was, within me, an invincible summer."

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Three-step output should contain Chinese characters")
}

func TestZaiTranslateIntegration_ReflectionStrategy_EnglishToChinese(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":   "en",
		"target":   "zh-CN",
		"strategy": "reflection",
	}
	input := "The only way to do great work is to love what you do."

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Reflection output should contain Chinese characters")
}

func TestZaiTranslateIntegration_ChineseToEnglish(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":   "zh-CN",
		"target":   "en",
		"strategy": "general",
	}
	input := "林中有两条路分叉，而我选择了那条少有人走的路，而这让一切变得不同。"

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	// Should be in English
	assert.True(t, isEnglishText(output), "Output should be primarily English text")
}

func TestZaiTranslateIntegration_WithFormalInstruction(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":      "en",
		"target":      "zh-CN",
		"strategy":    "general",
		"instruction": "formal business tone, suitable for professional communication",
	}
	input := "We would like to schedule a meeting to discuss the quarterly results."

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Formal instruction output should contain Chinese characters")
}

func TestZaiTranslateIntegration_WithCasualInstruction(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":      "en",
		"target":      "zh-CN",
		"strategy":    "general",
		"instruction": "casual and friendly tone, like talking to a friend",
	}
	input := "Hey, want to grab some coffee later?"

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Casual instruction output should contain Chinese characters")
}

func TestZaiTranslateIntegration_AutoDetection(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test auto-detection with English input
	options := map[string]string{
		"source": "auto",
		"target": "zh-CN",
	}
	input := "Hello, how are you today?"

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Auto-detection output should contain Chinese characters")

	// Check that source detected variable is set
	sourceDetected, err := variableService.Get("_source_detected")
	assert.NoError(t, err)
	assert.Equal(t, "auto", sourceDetected) // Current implementation stores the source parameter
}

func TestZaiTranslateIntegration_DefaultTargetLanguage(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test with no target specified (should default to English)
	options := map[string]string{
		"source": "zh-CN",
	}
	input := "你好世界"

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, isEnglishText(output), "Default target should produce English output")
}

func TestZaiTranslateIntegration_EmptyInputSetsLanguageVariables(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Test with empty input - should set language variables but not translation variables
	options := map[string]string{}
	input := ""

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	// Should set language variables even with empty input
	verifyTranslationVariables(t, variableService, false)
}

func TestZaiTranslateIntegration_TechnicalText(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":      "en",
		"target":      "zh-CN",
		"strategy":    "two_step",
		"instruction": "technical documentation style, preserve technical terms",
	}
	input := "The API endpoint returns a JSON response with authentication tokens and user metadata."

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Technical text output should contain Chinese characters")
}

func TestZaiTranslateIntegration_LiteraryText(t *testing.T) {
	t.Parallel() // Enable parallel execution
	skipIfNoAPIKey(t)

	cmd, variableService, cleanup := setupTestEnvironment(t)
	defer cleanup()

	options := map[string]string{
		"source":   "en",
		"target":   "zh-CN",
		"strategy": "three_step",
	}
	input := "It was the best of times, it was the worst of times, it was the age of wisdom, it was the age of foolishness."

	err := cmd.Execute(options, input)
	assert.NoError(t, err)

	verifyTranslationVariables(t, variableService, true)

	output, err := variableService.Get("_output")
	assert.NoError(t, err)
	assert.True(t, containsChinese(output), "Literary text output should contain Chinese characters")
}

// Helper function to check if text contains Chinese characters
func containsChinese(text string) bool {
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

// Helper function to check if text is primarily English
func isEnglishText(text string) bool {
	// Simple heuristic: check if text contains mostly ASCII characters and common English words
	asciiCount := 0
	totalCount := 0

	for _, r := range text {
		totalCount++
		if r <= 127 {
			asciiCount++
		}
	}

	// If more than 70% ASCII characters, consider it English
	return totalCount > 0 && float64(asciiCount)/float64(totalCount) > 0.7
}

// Benchmark test for performance measurement
func BenchmarkZaiTranslateIntegration_GeneralStrategy(b *testing.B) {
	apiKey := os.Getenv("Z_DOT_AI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ZAI_API_KEY")
	}
	if apiKey == "" {
		b.Skip("Skipping benchmark: no ZAI API key found")
	}

	// Set up test environment
	originalRegistry := services.GetGlobalRegistry()
	defer services.SetGlobalRegistry(originalRegistry)

	testRegistry := services.NewRegistry()
	err := testRegistry.RegisterService(services.NewVariableService())
	require.NoError(b, err)
	err = testRegistry.RegisterService(services.NewHTTPRequestService())
	require.NoError(b, err)
	services.SetGlobalRegistry(testRegistry)

	err = testRegistry.InitializeAll()
	require.NoError(b, err)

	variableService, err := testRegistry.GetService("variable")
	require.NoError(b, err)

	err = variableService.(*services.VariableService).Set("os.Z_DOT_AI_API_KEY", apiKey)
	require.NoError(b, err)

	cmd := &ZaiTranslateCommand{}
	options := map[string]string{
		"source": "en",
		"target": "zh-CN",
		"strategy": "general",
	}
	input := "Hello, this is a test message for benchmarking."

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := cmd.Execute(options, input)
		if err != nil {
			b.Fatalf("Translation failed: %v", err)
		}
		// No artificial delay needed - let the API handle rate limiting naturally
	}
}