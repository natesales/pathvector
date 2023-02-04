package autodoc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: lint the resulting markdown files
func TestDocumentConfig(t *testing.T) {
	DocumentConfig(false)
}

func TestSanitizeConfigName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"config.foo", "foo"},
		{"*string", "string"},
		{"map[string]*peer", "map[string]peer"},
	}
	for _, tc := range testCases {
		out := sanitizeConfigName(tc.input)
		assert.Equal(t, tc.expected, out)
	}
}
