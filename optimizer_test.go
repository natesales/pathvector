package main

import (
	"testing"
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
		out := sameAddressFamily(tc.a, tc.b)
		if out != tc.same {
			t.Errorf("a %s b %s expected same %v got %v", tc.a, tc.b, tc.same, out)
		}
	}
}
