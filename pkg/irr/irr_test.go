package irr

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

const irrQueryTimeout = 10

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
		{"AS-LROOT", 6, []string{"2001:500:3::/48", "2001:500:8c::/48", "2001:500:9c::/47{47,48}", "2001:500:9e::/47", "2001:500:9f::/48", "2602:800:9004::/47{48,48}", "2620:0:22b0::/48", "2620:0:2ee0::/48"}, false},
	}
	for _, tc := range testCases {
		out, err := PrefixSet(tc.asSet, tc.family, "rr.ntt.net", irrQueryTimeout, "")
		if err != nil && !tc.shouldError {
			t.Error(err)
		} else if err == nil && tc.shouldError {
			t.Errorf("as-set %s family %d should error but didn't", tc.asSet, tc.family)
		}
		if err == nil && !reflect.DeepEqual(out, tc.expectedOutput) {
			assert.Equalf(t, tc.expectedOutput, out, "as-set %s family %d failed. expected '%s' got '%s'", tc.asSet, tc.family, tc.expectedOutput, out)
		}
	}
}

func TestBuildIRRPrefixSet(t *testing.T) {
	testCases := []struct {
		asSet       string
		prefixSet4  []string
		prefixSet6  []string
		shouldError bool
	}{
		{"AS112", []string{"192.31.196.0/24", "192.175.48.0/24"}, []string{"2001:4:112::/48", "2620:4f:8000::/48"}, false},
		{"", []string{}, []string{}, true}, // Empty as-set
	}
	for _, tc := range testCases {
		peer := config.Peer{ASSet: util.Ptr(tc.asSet)}
		err := Update(&peer, "rr.ntt.net", irrQueryTimeout, "")
		if err != nil && tc.shouldError {
			return
		}
		if err != nil && !tc.shouldError {
			t.Error(err)
		} else if err == nil && tc.shouldError {
			t.Errorf("as-set %s should error but didn't", tc.asSet)
		}
		assert.Equal(t, tc.prefixSet4, *peer.PrefixSet4)
		assert.Equal(t, tc.prefixSet6, *peer.PrefixSet6)
	}
}

func TestIRRASMembers(t *testing.T) {
	testCases := []struct {
		asSet       string
		asMembers   []uint32
		shouldError bool
	}{
		{"AS34553:AS-PACKETFRAME", []uint32{112, 34553}, false},
		{"", []uint32{}, true}, // Empty as-set
	}
	for _, tc := range testCases {
		members, err := ASMembers(tc.asSet, "rr.ntt.net", irrQueryTimeout, "")
		if tc.shouldError {
			assert.NotNil(t, err)
			continue
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, 2, len(members))
		assert.Equal(t, members[0], uint32(112))
		assert.Equal(t, members[1], uint32(34553))
	}
}
