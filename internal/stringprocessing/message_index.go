// Package stringprocessing provides utilities for text processing and parsing operations.
package stringprocessing

import (
	"fmt"
	"strconv"
	"strings"
)

// MessageIndexResult contains the result of parsing a message index
type MessageIndexResult struct {
	// ZeroBasedIndex is the 0-based array index for accessing messages
	ZeroBasedIndex int
	// PositionDescription is a human-readable description of the position
	PositionDescription string
}

// ParseMessageIndex parses a message index string and converts it to 0-based index.
// Supports dual indexing system:
//   - Reverse order: "1", "2", "3" -> last, second-to-last, third-to-last
//   - Normal order: ".1", ".2", ".3" -> first, second, third
//
// Parameters:
//   - idxStr: The index string to parse (e.g., "1", "2", ".1", ".2")
//   - messageCount: Total number of messages in the session
//
// Returns:
//   - MessageIndexResult with zero-based index and position description
//   - Error if the index is invalid or out of bounds
func ParseMessageIndex(idxStr string, messageCount int) (*MessageIndexResult, error) {
	if strings.HasPrefix(idxStr, ".") {
		// Normal order: .1, .2, .3 -> 0-based: 0, 1, 2
		numStr := idxStr[1:]
		if numStr == "" {
			return nil, fmt.Errorf("invalid normal order index format (use .1, .2, .3, etc.)")
		}

		num, err := strconv.Atoi(numStr)
		if err != nil {
			return nil, fmt.Errorf("invalid normal order index number: %w", err)
		}

		if num < 1 || num > messageCount {
			return nil, fmt.Errorf("normal order index %d is out of bounds (session has %d messages)", num, messageCount)
		}

		zeroBasedIndex := num - 1
		positionDesc := GetOrdinalPosition(num, false)
		return &MessageIndexResult{
			ZeroBasedIndex:      zeroBasedIndex,
			PositionDescription: positionDesc,
		}, nil
	}

	// Reverse order: 1, 2, 3 -> 0-based: last, second-to-last, third-to-last
	num, err := strconv.Atoi(idxStr)
	if err != nil {
		return nil, fmt.Errorf("invalid reverse order index number: %w", err)
	}

	if num < 1 || num > messageCount {
		return nil, fmt.Errorf("reverse order index %d is out of bounds (session has %d messages)", num, messageCount)
	}

	zeroBasedIndex := messageCount - num
	positionDesc := GetOrdinalPosition(num, true)
	return &MessageIndexResult{
		ZeroBasedIndex:      zeroBasedIndex,
		PositionDescription: positionDesc,
	}, nil
}

// GetOrdinalPosition returns a human-readable position description.
// Parameters:
//   - num: The position number (1-based)
//   - reverse: If true, describes reverse order positions (1=last), if false, normal order (1=first)
//
// Returns human-readable position string like "last message", "first message", etc.
func GetOrdinalPosition(num int, reverse bool) string {
	if reverse {
		switch num {
		case 1:
			return "last message"
		case 2:
			return "second-to-last message"
		case 3:
			return "third-to-last message"
		default:
			return fmt.Sprintf("%d%s from last message", num, GetOrdinalSuffix(num))
		}
	} else {
		switch num {
		case 1:
			return "first message"
		case 2:
			return "second message"
		case 3:
			return "third message"
		default:
			return fmt.Sprintf("%d%s message", num, GetOrdinalSuffix(num))
		}
	}
}

// GetOrdinalSuffix returns the appropriate ordinal suffix (st, nd, rd, th) for a number.
// Examples: 1 -> "st", 2 -> "nd", 3 -> "rd", 4 -> "th", 11 -> "th", 21 -> "st"
func GetOrdinalSuffix(num int) string {
	// Special case for numbers ending in 11, 12, 13 (like 11, 12, 13, 111, 112, 113, etc.)
	if num%100 >= 11 && num%100 <= 13 {
		return "th"
	}
	switch num % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
