package optimizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptimizerSameAddressFamily(t *testing.T) {
	testCases := []struct {
		a    string
		b    string
		same bool
	}{
		{"192.0.2.1", "192.0.2.1", true},
		{"192.0.2.1", "2001:db8::1", false},
		{"2001:db8::1", "2001:db8::1", true},
		{"2001:db8::1", "192.0.2.1", false},
	}
	for _, tc := range testCases {
		t.Run(tc.a+"=="+tc.b, func(t *testing.T) {
			assert.Equal(t, tc.same, sameAddressFamily(tc.a, tc.b))
		})
	}
}
