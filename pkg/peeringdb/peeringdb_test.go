package peeringdb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

const peeringDbQueryTimeout = 10 // 10 seconds

func TestPeeringDbQuery(t *testing.T) {
	testCases := []struct {
		asn          int
		asSet        string
		name         string
		importLimit4 int
		importLimit6 int
		shouldError  bool
	}{
		{112, "AS112", "DNS-OARC-112", 2, 2, false},
		{34553, "AS34553:AS-ALL", "Nathan Sales", 10, 20, false},
		{44977, "AS44977", "ARIX", 10, 10, false},
		{65530, "", "", 0, 0, true}, // Private ASN, no PeeringDB page
	}
	for _, tc := range testCases {
		pDbData, err := NetworkInfo(uint32(tc.asn), peeringDbQueryTimeout, "", true)
		if err != nil && !tc.shouldError {
			t.Error(err)
		}

		if tc.shouldError && err == nil {
			t.Errorf("asn %d should have errored but didn't", tc.asn)
		}

		if tc.shouldError && err != nil && !strings.Contains(err.Error(), "doesn't have a PeeringDB page") {
			t.Errorf("asn %d should have thrown a no PeeringDB error but got a different error: %v", tc.asn, err)
		}

		if err == nil {
			assert.Equal(t, tc.asSet, pDbData.ASSet)
			assert.Equal(t, tc.name, pDbData.Name)
			assert.Equal(t, tc.importLimit4, pDbData.ImportLimit4)
			assert.Equal(t, tc.importLimit6, pDbData.ImportLimit6)
		}
	}
}

func TestPeeringDbNoPage(t *testing.T) {
	_, err := NetworkInfo(65530, peeringDbQueryTimeout, "", true)
	if err == nil || !strings.Contains(err.Error(), "doesn't have a PeeringDB page") {
		t.Errorf("expected PeeringDB page not exist error, got %v", err)
	}
}

func TestPeeringDbQueryAndModify(t *testing.T) {
	testCases := []struct {
		asn  int
		auto bool
	}{
		{112, true},
		{112, false},
	}
	for _, tc := range testCases {
		Update(&config.Peer{
			ASN:              util.Ptr(tc.asn),
			AutoImportLimits: util.Ptr(tc.auto),
			AutoASSet:        util.Ptr(tc.auto),
			ImportLimit4:     util.Ptr(0),
			ImportLimit6:     util.Ptr(0),
		}, peeringDbQueryTimeout, "", true)
	}
}

func TestPeeringNeverViaRouteServers(t *testing.T) {
	asns, err := NeverViaRouteServers(peeringDbQueryTimeout, "")
	assert.Nil(t, err)
	assert.Equal(t, len(asns), 3)
}
