package ocr

import (
	"reflect"
	"testing"
)

func TestParsePageSpec(t *testing.T) {
	tests := []struct {
		name        string
		pagesSpec   string
		totalPages  int
		expected    []int
		expectError bool
	}{
		{
			name:       "all pages",
			pagesSpec:  "all",
			totalPages: 3,
			expected:   []int{1, 2, 3},
		},
		{
			name:       "empty spec defaults to all",
			pagesSpec:  "",
			totalPages: 2,
			expected:   []int{1, 2},
		},
		{
			name:       "single page",
			pagesSpec:  "2",
			totalPages: 5,
			expected:   []int{2},
		},
		{
			name:       "page range",
			pagesSpec:  "1-3",
			totalPages: 5,
			expected:   []int{1, 2, 3},
		},
		{
			name:       "multiple ranges and singles",
			pagesSpec:  "1-2,4,6-7",
			totalPages: 10,
			expected:   []int{1, 2, 4, 6, 7},
		},
		{
			name:        "invalid page number",
			pagesSpec:   "0",
			totalPages:  5,
			expectError: true,
		},
		{
			name:        "page number too high",
			pagesSpec:   "10",
			totalPages:  5,
			expectError: true,
		},
		{
			name:        "invalid range format",
			pagesSpec:   "1-2-3",
			totalPages:  5,
			expectError: true,
		},
		{
			name:        "invalid range order",
			pagesSpec:   "3-1",
			totalPages:  5,
			expectError: true,
		},
		{
			name:        "non-numeric page",
			pagesSpec:   "abc",
			totalPages:  5,
			expectError: true,
		},
		{
			name:       "duplicates removed",
			pagesSpec:  "1,2,1,3,2",
			totalPages: 5,
			expected:   []int{1, 2, 3},
		},
		{
			name:       "whitespace handling",
			pagesSpec:  " 1 - 3 , 5 ",
			totalPages: 10,
			expected:   []int{1, 2, 3, 5},
		},
		{
			name:       "complex specification",
			pagesSpec:  "1,3-5,2,7-9,6",
			totalPages: 10,
			expected:   []int{1, 3, 4, 5, 2, 7, 8, 9, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePageSpec(tt.pagesSpec, tt.totalPages)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParsePageSpec() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetDefaultOCRPrompt(t *testing.T) {
	prompt := GetDefaultOCRPrompt()

	if prompt == "" {
		t.Error("GetDefaultOCRPrompt() returned empty string")
	}

	// Check that prompt contains key elements
	expectedElements := []string{
		"document",
		"markdown",
		"LaTeX",
		"front matter",
		"primary_language",
		"is_rotation_valid",
		"rotation_correction",
		"is_table",
		"is_diagram",
	}

	for _, element := range expectedElements {
		if !contains(prompt, element) {
			t.Errorf("OCR prompt missing expected element: %s", element)
		}
	}

	// Verify prompt is of reasonable length
	if len(prompt) < 100 {
		t.Errorf("OCR prompt too short: %d characters", len(prompt))
	}
}

// Helper function to check if a string contains a substring (case insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark for ParsePageSpec with various specifications
func BenchmarkParsePageSpec(b *testing.B) {
	tests := []struct {
		name       string
		pagesSpec  string
		totalPages int
	}{
		{"all pages", "all", 100},
		{"single page", "50", 100},
		{"simple range", "1-10", 100},
		{"complex spec", "1-10,20,30-40,50,60-70,80,90-100", 100},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ParsePageSpec(tt.pagesSpec, tt.totalPages)
			}
		})
	}
}
