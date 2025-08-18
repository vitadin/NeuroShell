package stringprocessing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMessageIndex(t *testing.T) {
	tests := []struct {
		name         string
		idxStr       string
		messageCount int
		expectIndex  int
		expectDesc   string
		expectError  bool
	}{
		// Reverse order tests
		{
			name:         "reverse order - last message",
			idxStr:       "1",
			messageCount: 5,
			expectIndex:  4,
			expectDesc:   "last message",
			expectError:  false,
		},
		{
			name:         "reverse order - second-to-last message",
			idxStr:       "2",
			messageCount: 5,
			expectIndex:  3,
			expectDesc:   "second-to-last message",
			expectError:  false,
		},
		{
			name:         "reverse order - third-to-last message",
			idxStr:       "3",
			messageCount: 5,
			expectIndex:  2,
			expectDesc:   "third-to-last message",
			expectError:  false,
		},
		{
			name:         "reverse order - single message",
			idxStr:       "1",
			messageCount: 1,
			expectIndex:  0,
			expectDesc:   "last message",
			expectError:  false,
		},

		// Normal order tests
		{
			name:         "normal order - first message",
			idxStr:       ".1",
			messageCount: 5,
			expectIndex:  0,
			expectDesc:   "first message",
			expectError:  false,
		},
		{
			name:         "normal order - second message",
			idxStr:       ".2",
			messageCount: 5,
			expectIndex:  1,
			expectDesc:   "second message",
			expectError:  false,
		},
		{
			name:         "normal order - third message",
			idxStr:       ".3",
			messageCount: 5,
			expectIndex:  2,
			expectDesc:   "third message",
			expectError:  false,
		},
		{
			name:         "normal order - last message in session",
			idxStr:       ".5",
			messageCount: 5,
			expectIndex:  4,
			expectDesc:   "5th message",
			expectError:  false,
		},

		// Error cases
		{
			name:         "reverse order - out of bounds (too high)",
			idxStr:       "6",
			messageCount: 5,
			expectError:  true,
		},
		{
			name:         "reverse order - out of bounds (zero)",
			idxStr:       "0",
			messageCount: 5,
			expectError:  true,
		},
		{
			name:         "normal order - out of bounds (too high)",
			idxStr:       ".6",
			messageCount: 5,
			expectError:  true,
		},
		{
			name:         "normal order - out of bounds (zero)",
			idxStr:       ".0",
			messageCount: 5,
			expectError:  true,
		},
		{
			name:         "invalid format - empty normal order",
			idxStr:       ".",
			messageCount: 5,
			expectError:  true,
		},
		{
			name:         "invalid format - non-numeric",
			idxStr:       "abc",
			messageCount: 5,
			expectError:  true,
		},
		{
			name:         "invalid format - normal order non-numeric",
			idxStr:       ".abc",
			messageCount: 5,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMessageIndex(tt.idxStr, tt.messageCount)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectIndex, result.ZeroBasedIndex)
				assert.Equal(t, tt.expectDesc, result.PositionDescription)
			}
		})
	}
}

func TestGetOrdinalPosition(t *testing.T) {
	tests := []struct {
		name     string
		num      int
		reverse  bool
		expected string
	}{
		// Reverse order tests
		{"reverse order - 1", 1, true, "last message"},
		{"reverse order - 2", 2, true, "second-to-last message"},
		{"reverse order - 3", 3, true, "third-to-last message"},
		{"reverse order - 4", 4, true, "4th from last message"},
		{"reverse order - 5", 5, true, "5th from last message"},
		{"reverse order - 11", 11, true, "11th from last message"},
		{"reverse order - 21", 21, true, "21st from last message"},
		{"reverse order - 22", 22, true, "22nd from last message"},
		{"reverse order - 23", 23, true, "23rd from last message"},

		// Normal order tests
		{"normal order - 1", 1, false, "first message"},
		{"normal order - 2", 2, false, "second message"},
		{"normal order - 3", 3, false, "third message"},
		{"normal order - 4", 4, false, "4th message"},
		{"normal order - 5", 5, false, "5th message"},
		{"normal order - 11", 11, false, "11th message"},
		{"normal order - 21", 21, false, "21st message"},
		{"normal order - 22", 22, false, "22nd message"},
		{"normal order - 23", 23, false, "23rd message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetOrdinalPosition(tt.num, tt.reverse)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetOrdinalSuffix(t *testing.T) {
	tests := []struct {
		num      int
		expected string
	}{
		{1, "st"},
		{2, "nd"},
		{3, "rd"},
		{4, "th"},
		{5, "th"},
		{10, "th"},
		{11, "th"}, // Special case
		{12, "th"}, // Special case
		{13, "th"}, // Special case
		{14, "th"},
		{21, "st"},
		{22, "nd"},
		{23, "rd"},
		{24, "th"},
		{31, "st"},
		{101, "st"},
		{111, "th"}, // Special case
		{112, "th"}, // Special case
		{113, "th"}, // Special case
		{121, "st"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("ordinal_%d", tt.num), func(t *testing.T) {
			result := GetOrdinalSuffix(tt.num)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseMessageIndex_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		idxStr       string
		messageCount int
		expectError  bool
		errorMessage string
	}{
		{
			name:         "zero message count",
			idxStr:       "1",
			messageCount: 0,
			expectError:  true,
		},
		{
			name:         "negative message count",
			idxStr:       "1",
			messageCount: -1,
			expectError:  true,
		},
		{
			name:         "empty index string",
			idxStr:       "",
			messageCount: 5,
			expectError:  true,
		},
		{
			name:         "only dot",
			idxStr:       ".",
			messageCount: 5,
			expectError:  true,
		},
		{
			name:         "multiple dots",
			idxStr:       "..1",
			messageCount: 5,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMessageIndex(tt.idxStr, tt.messageCount)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
