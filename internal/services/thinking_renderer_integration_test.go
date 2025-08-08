package services

import (
	"strings"
	"testing"

	"neuroshell/pkg/neurotypes"
)

// TestThinkingRendererService_Integration tests the full integration with service registry.
func TestThinkingRendererService_Integration(t *testing.T) {
	// Test that thinking-renderer service can be accessed via registry
	registry := GetGlobalRegistry()

	// Register the service for testing (simulating shell initialization)
	if !registry.HasService("thinking-renderer") {
		if err := registry.RegisterService(NewThinkingRendererService()); err != nil {
			t.Fatalf("Failed to register thinking-renderer service: %v", err)
		}
	}

	// Get the thinking-renderer service
	service, err := registry.GetService("thinking-renderer")
	if err != nil {
		t.Fatalf("Failed to get thinking-renderer service: %v", err)
	}

	// Type assert to ThinkingRendererService
	thinkingService, ok := service.(*ThinkingRendererService)
	if !ok {
		t.Fatalf("Service is not a ThinkingRendererService")
	}

	// Verify service name
	if thinkingService.Name() != "thinking-renderer" {
		t.Errorf("Expected service name 'thinking-renderer', got '%s'", thinkingService.Name())
	}

	// Initialize the service
	if err := thinkingService.Initialize(); err != nil {
		t.Fatalf("Failed to initialize thinking renderer service: %v", err)
	}

	// Test rendering with mock config
	mockConfig := NewMockRenderConfig()

	// Test rendering blocks
	blocks := []neurotypes.ThinkingBlock{
		{Provider: "anthropic", Type: "thinking", Content: "Test thinking content"},
		{Provider: "gemini", Type: "thinking", Content: "Gemini test content"},
		{Provider: "openai", Type: "reasoning", Content: "OpenAI test content"},
	}

	result := thinkingService.RenderThinkingBlocks(blocks, mockConfig)
	if result == "" {
		t.Fatal("Expected non-empty rendering result")
	}

	// Verify all provider labels are present
	if !strings.Contains(result, "Claude's internal reasoning:") {
		t.Error("Expected result to contain Anthropic provider label")
	}
	if !strings.Contains(result, "Gemini's thought process:") {
		t.Error("Expected result to contain Gemini provider label")
	}
	if !strings.Contains(result, "OpenAI reasoning trace:") {
		t.Error("Expected result to contain OpenAI provider label")
	}

	// Verify content is present
	if !strings.Contains(result, "Test thinking content") {
		t.Error("Expected result to contain Anthropic thinking content")
	}
	if !strings.Contains(result, "Gemini test content") {
		t.Error("Expected result to contain Gemini thinking content")
	}
	if !strings.Contains(result, "OpenAI test content") {
		t.Error("Expected result to contain OpenAI thinking content")
	}
}

// TestThinkingRendererService_BackwardCompatibility tests that legacy methods still work.
func TestThinkingRendererService_BackwardCompatibility(t *testing.T) {
	// Register the service for testing (simulating shell initialization)
	registry := GetGlobalRegistry()
	if !registry.HasService("thinking-renderer") {
		if err := registry.RegisterService(NewThinkingRendererService()); err != nil {
			t.Fatalf("Failed to register thinking-renderer service: %v", err)
		}
	}

	// Get service from global registry
	service, err := GetGlobalThinkingRendererService()
	if err != nil {
		t.Fatalf("Failed to get thinking renderer service: %v", err)
	}

	// Initialize the service (needed for legacy method)
	if err := service.Initialize(); err != nil {
		t.Fatalf("Failed to initialize thinking renderer service: %v", err)
	}

	// Test legacy rendering method
	blocks := []neurotypes.ThinkingBlock{
		{Provider: "anthropic", Type: "thinking", Content: "Legacy test content"},
	}

	result := service.RenderThinkingBlocksLegacy(blocks)

	// Should contain the content and provider label
	if !strings.Contains(result, "Claude's internal reasoning:") {
		t.Error("Legacy rendering should contain Anthropic provider label")
	}
	if !strings.Contains(result, "Legacy test content") {
		t.Error("Legacy rendering should contain the thinking content")
	}
}

// TestThinkingRendererService_ServiceRegistration tests that the service is properly registered.
func TestThinkingRendererService_ServiceRegistration(t *testing.T) {
	registry := GetGlobalRegistry()

	// Register the service for testing (simulating shell initialization)
	if !registry.HasService("thinking-renderer") {
		if err := registry.RegisterService(NewThinkingRendererService()); err != nil {
			t.Fatalf("Failed to register thinking-renderer service: %v", err)
		}
	}

	// Test that the service exists in the registry
	service, err := registry.GetService("thinking-renderer")
	if err != nil {
		t.Fatalf("thinking-renderer service should be registered: %v", err)
	}

	if service == nil {
		t.Fatal("thinking-renderer service should not be nil")
	}

	// Test that it implements the Service interface
	if service.Name() != "thinking-renderer" {
		t.Errorf("Expected service name 'thinking-renderer', got '%s'", service.Name())
	}

	// Test initialization
	if err := service.Initialize(); err != nil {
		t.Errorf("Service initialization should not fail: %v", err)
	}
}

// TestThinkingRendererService_GlobalAccess tests the global access function.
func TestThinkingRendererService_GlobalAccess(t *testing.T) {
	// Register the service for testing (simulating shell initialization)
	registry := GetGlobalRegistry()
	if !registry.HasService("thinking-renderer") {
		if err := registry.RegisterService(NewThinkingRendererService()); err != nil {
			t.Fatalf("Failed to register thinking-renderer service: %v", err)
		}
	}

	// Test the global access function
	service, err := GetGlobalThinkingRendererService()
	if err != nil {
		t.Fatalf("GetGlobalThinkingRendererService should succeed: %v", err)
	}

	if service == nil {
		t.Fatal("GetGlobalThinkingRendererService should not return nil")
	}

	// Verify it's the right type
	if service.Name() != "thinking-renderer" {
		t.Errorf("Expected service name 'thinking-renderer', got '%s'", service.Name())
	}

	// Test that it implements the neurotypes.ThinkingRenderer interface
	var _ neurotypes.ThinkingRenderer = service

	// Test interface methods
	providers := service.GetSupportedProviders()
	if len(providers) == 0 {
		t.Error("GetSupportedProviders should return non-empty list")
	}

	// Test single block rendering
	block := neurotypes.ThinkingBlock{
		Provider: "anthropic",
		Type:     "thinking",
		Content:  "Interface test content",
	}
	config := NewMockRenderConfig()

	result := service.RenderSingleBlock(block, config)
	if result == "" {
		t.Error("RenderSingleBlock should return non-empty result")
	}
}
