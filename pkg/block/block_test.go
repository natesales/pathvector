package block

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockParse(t *testing.T) {
	tcs := []struct {
		input       []string
		asns        []uint32
		prefixes    []string
		shouldError bool
	}{
		{[]string{"AS65530", "AS65520", "192.0.2.0/24", "10.0.50.2"}, []uint32{65530, 65520}, []string{"192.0.2.0/24", "10.0.50.2/32"}, false},
		{[]string{"invalid", "AS65520", "192.0.2.0/24", "10.0.50.2"}, []uint32{}, []string{}, true},
		{[]string{"AS65530", "invalid", "192.0.2.0/24", "10.0.50.2"}, []uint32{}, []string{}, true},
		{[]string{"AS65530", "AS65520", "invalid", "10.0.50.2"}, []uint32{}, []string{}, true},
		{[]string{"AS65530", "AS65520", "192.0.2.0/24", "invalid"}, []uint32{}, []string{}, true},
	}

	for _, tc := range tcs {
		asns, prefixes, err := Parse(tc.input)
		if tc.shouldError {
			assert.NotNil(t, err)
			continue
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, tc.asns, asns)
		assert.Equal(t, tc.prefixes, prefixes)
	}
}

func TestBlockCombine(t *testing.T) {
	combined := Combine([]string{"AS65530"}, []string{"https://raw.githubusercontent.com/natesales/pathvector/main/tests/blocklist.txt"}, []string{"../../tests/blocklist.txt"})
	assert.Len(t, combined, 19) // This only combines, doesn't sanitize so newlines and comments are included
}
