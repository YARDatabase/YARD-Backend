package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseItemTypesHelper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "comma separated",
			input:    "SWORD,AXE,BOW",
			expected: []string{"SWORD", "AXE", "BOW"},
		},
		{
			name:     "with spaces",
			input:    "SWORD, AXE, BOW",
			expected: []string{"SWORD", "AXE", "BOW"},
		},
		{
			name:     "single item",
			input:    "SWORD",
			expected: []string{"SWORD"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseItemTypesHelper(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func parseItemTypesHelper(itemTypes string) []string {
	parts := strings.Split(itemTypes, ",")
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(strings.ToUpper(part))
	}
	return result
}

