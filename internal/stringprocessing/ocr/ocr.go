// Package ocr provides utility functions for OCR processing.
package ocr

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePageSpec parses a page specification string and returns a slice of page numbers.
// Supported formats:
//   - "all" or "" - returns all pages (1 to totalPages)
//   - "5" - single page
//   - "1-3" - page range
//   - "1-3,5,7-9" - combination of ranges and single pages
//
// Returns an error if the specification is invalid or page numbers are out of range.
func ParsePageSpec(pagesSpec string, totalPages int) ([]int, error) {
	if pagesSpec == "" || pagesSpec == "all" {
		// Return all pages
		pages := make([]int, totalPages)
		for i := 0; i < totalPages; i++ {
			pages[i] = i + 1
		}
		return pages, nil
	}

	var pageNumbers []int
	ranges := strings.Split(pagesSpec, ",")

	for _, rangeStr := range ranges {
		rangeStr = strings.TrimSpace(rangeStr)
		if rangeStr == "" {
			continue
		}

		// Check if it's a range (e.g., "1-3") or single page
		if strings.Contains(rangeStr, "-") {
			parts := strings.Split(rangeStr, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", rangeStr)
			}

			start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start page: %s", parts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end page: %s", parts[1])
			}

			if start < 1 || start > totalPages {
				return nil, fmt.Errorf("start page %d out of range (1-%d)", start, totalPages)
			}

			if end < 1 || end > totalPages {
				return nil, fmt.Errorf("end page %d out of range (1-%d)", end, totalPages)
			}

			if start > end {
				return nil, fmt.Errorf("start page %d greater than end page %d", start, end)
			}

			// Add all pages in range
			for i := start; i <= end; i++ {
				pageNumbers = append(pageNumbers, i)
			}
		} else {
			// Single page
			pageNum, err := strconv.Atoi(rangeStr)
			if err != nil {
				return nil, fmt.Errorf("invalid page number: %s", rangeStr)
			}

			if pageNum < 1 || pageNum > totalPages {
				return nil, fmt.Errorf("page %d out of range (1-%d)", pageNum, totalPages)
			}

			pageNumbers = append(pageNumbers, pageNum)
		}
	}

	// Remove duplicates and sort
	uniquePages := make(map[int]bool)
	var result []int
	for _, page := range pageNumbers {
		if !uniquePages[page] {
			uniquePages[page] = true
			result = append(result, page)
		}
	}

	return result, nil
}

// GetDefaultOCRPrompt returns the default prompt text for OCR processing.
// This prompt instructs the model to:
//   - Convert document text naturally
//   - Convert equations to LaTeX
//   - Convert tables to markdown
//   - Include metadata in front matter
func GetDefaultOCRPrompt() string {
	return "Attached is one page of a document that you must process. " +
		"Just return the plain text representation of this document as if you were reading it naturally. " +
		"Convert equations to LaTeX and tables to markdown.\n" +
		"Return your output as markdown, with a front matter section on top specifying values for the " +
		"primary_language, is_rotation_valid, rotation_correction, is_table, and is_diagram parameters."
}
