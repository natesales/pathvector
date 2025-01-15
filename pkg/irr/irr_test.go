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
		{"AS-FROOT", 4, []string{"192.5.4.0/23{23,24}"}, false},
	}
	for _, tc := range testCases {
		out, err := PrefixSet(tc.asSet, tc.family, "rr.ntt.net", irrQueryTimeout, "bgpq4", "")
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
		t.Run(tc.asSet, func(t *testing.T) {
			peer := config.Peer{ASSet: util.Ptr(tc.asSet)}
			err := Update(&peer, "rr.ntt.net", irrQueryTimeout, "bgpq4", "")
			if err != nil && tc.shouldError {
				return
			}
			if tc.shouldError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tc.prefixSet4, *peer.PrefixSet4)
			assert.Equal(t, tc.prefixSet6, *peer.PrefixSet6)
		})
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
