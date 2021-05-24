package main

import (
	"strings"
	"testing"
)

func TestAsSetFilterName(t *testing.T) {
	testCases := []struct {
		asSet          string
		family         uint8
		expectedOutput string
		shouldError    bool
	}{
		{"AS112", 4, "PFXSET_AS112_IP4", false},
		{"AS112", 6, "PFXSET_AS112_IP6", false},
		{"AS112", 9, "PFXSET_AS112_IP6", true}, // Invalid address family
	}
	for _, tc := range testCases {
		out, err := asSetToFilterName(tc.asSet, tc.family)
		if err != nil && !tc.shouldError {
			t.Error(err)
		} else if err == nil && tc.shouldError {
			t.Errorf("as-set %s family %d should error but didn't", tc.asSet, tc.family)
		}
		if err == nil && out != tc.expectedOutput {
			t.Errorf("as-set %s family %d failed. expected '%s' got '%s'", tc.asSet, tc.family, tc.expectedOutput, out)
		}
	}
}

func TestGetIRRPrefixSet(t *testing.T) {
	testCases := []struct {
		asSet          string
		family         uint8
		expectedOutput string
		shouldError    bool
	}{
		{"AS112", 4, "define PFXSET_AS112_IP4 = [192.31.196.0/24,192.175.48.0/24];", false},
		{"AS112", 6, "define PFXSET_AS112_IP6 = [2001:4:112::/48,2620:4f:8000::/48];", false},
		{"AS112", 9, "define PFXSET_AS112_IP6 = [2001:4:112::/48,2620:4f:8000::/48];", true}, // Invalid address family
	}
	for _, tc := range testCases {
		out, err := getIRRPrefixSet(tc.asSet, tc.family, &config{IRRServer: "rr.ntt.net", IRRQueryTimeout: 10})
		if err != nil && !tc.shouldError {
			t.Error(err)
		} else if err == nil && tc.shouldError {
			t.Errorf("as-set %s family %d should error but didn't", tc.asSet, tc.family)
		}
		out = strings.Replace(out, "\n", "", -1)
		out = strings.Replace(out, "    ", "", -1)

		if err == nil && out != tc.expectedOutput {
			t.Errorf("as-set %s family %d failed. expected '%s' got '%s'", tc.asSet, tc.family, tc.expectedOutput, out)
		}
	}
}
