package irr

import (
	"reflect"
	"testing"
)

func TestGetIRRPrefixSet(t *testing.T) {
	testCases := []struct {
		asSet          string
		family         uint8
		expectedOutput []string
		shouldError    bool
	}{
		{"AS112", 4, []string{"192.31.196.0/24", "192.175.48.0/24"}, false},
		{"AS112", 6, []string{"2001:4:112::/48", "2620:4f:8000::/48"}, false},
		{"AS112", 9, []string{"2001:4:112::/48", "2620:4f:8000::/48"}, true}, // Invalid address family
		{"AS-LROOT", 6, []string{"2001:500:3::/48", "2001:500:8c::/48", "2001:500:9c::/47{47,48}", "2001:500:9e::/47", "2001:500:9f::/48", "2620:0:22b0::/48", "2620:0:2ee0::/48"}, false},
		{"AS-FROOT", 4, []string{"192.5.4.0/23{23,24}", "199.212.90.0/23", "199.212.92.0/23", "202.41.142.0/24"}, false},
	}
	for _, tc := range testCases {
		out, err := getIRRPrefixSet(tc.asSet, tc.family, "rr.ntt.net", 10)
		if err != nil && !tc.shouldError {
			t.Error(err)
		} else if err == nil && tc.shouldError {
			t.Errorf("as-set %s family %d should error but didn't", tc.asSet, tc.family)
		}
		if err == nil && !reflect.DeepEqual(out, tc.expectedOutput) {
			t.Errorf("as-set %s family %d failed. expected '%s' got '%s'", tc.asSet, tc.family, tc.expectedOutput, out)
		}
	}
}
