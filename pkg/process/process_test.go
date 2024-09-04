package process

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/util"
)

func TestCategorizeCommunity(t *testing.T) {
	testCases := []struct {
		input          string
		expectedOutput string
		shouldError    bool
	}{
		{"34553,0", "standard", false},
		{"1,1", "standard", false},
		{"4242424242:4242424242:0", "large", false},
		{"1:1:0", "large", false},
		{":", "", true},
		{"4242424242,0", "", true},
		{"0,4242424242", "", true},
		{"foo,1", "", true},
		{"1,bar", "", true},
		{"", "", true},
		{":1:1", "", true},
		{"1::1", "", true},
		{"1:1:", "", true},
		{"-1:1:1", "", true},
		{"1:-1:1", "", true},
		{"1:1:-1", "", true},
	}
	for _, tc := range testCases {
		cType := categorizeCommunity(tc.input)
		if tc.shouldError {
			assert.Equal(t, "", cType)
		} else {
			assert.Equal(t, tc.expectedOutput, cType)
		}
	}
}

func TestLoad(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
prefixes:
  - 192.0.2.0/24
  - 2001:db8::/48
kernel:
  statics:
    "203.0.113.0/24" : "192.0.2.10"
    "2001:db8:2::/64" : "2001:db8::1"
vrrp:
 VRRP 1:
    state: primary
    interface: eth0
    priority: 255
    vips:
      - 192.0.2.1/24
      - 2001:db8::1/64
 VRRP 2:
    state: backup
    interface: eth1
    priority: 255
    vips:
      - 192.0.2.2/24
      - 2001:db8::2/64
peers:
  Example:
    asn: 65530
    announce-originated: false
    neighbors:
      - 203.0.113.25
      - 2001:db8:2::25
`

	globalConfig, err := Load([]byte(configFile))
	assert.Nil(t, err)

	assert.Equal(t, 34553, globalConfig.ASN)
	assert.Equal(t, "192.0.2.1", globalConfig.RouterID)
	assert.Equal(t, 1, len(globalConfig.Peers))
	assert.Equal(t, 65530, *globalConfig.Peers["Example"].ASN)
	assert.Equal(t, []string{"203.0.113.25", "2001:db8:2::25"}, *globalConfig.Peers["Example"].NeighborIPs)
}

func TestLoadLocalPref(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
peers:
  Peer 10:
    asn: 65510
    local-pref: 110
    neighbors:
      - 192.0.2.10
  Peer 20:
    asn: 65520
    set-local-pref: false
    default-local-pref: 120
    neighbors:
      - 192.0.2.20
`

	globalConfig, err := Load([]byte(configFile))
	assert.NoError(t, err)

	assert.Len(t, globalConfig.Peers, 2)

	peer10 := globalConfig.Peers["Peer 10"]
	assert.Equal(t, 65510, util.Deref(peer10.ASN))
	assert.Equal(t, 110, util.Deref(peer10.LocalPref))
	assert.True(t, util.Deref(peer10.SetLocalPref))
	assert.Nil(t, peer10.DefaultLocalPref)

	peer20 := globalConfig.Peers["Peer 20"]
	assert.Equal(t, 65520, util.Deref(peer20.ASN))
	assert.Equal(t, 100, util.Deref(peer20.LocalPref))
	assert.False(t, util.Deref(peer20.SetLocalPref))
	assert.Equal(t, 120, util.Deref(peer20.DefaultLocalPref))
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	configFile := "INVALID YAML"
	_, err := Load([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "YAML unmarshal") {
		t.Errorf("expected yaml unmarshal error, got %+v", err)
	}
}

func TestLoadConfigValidationError(t *testing.T) {
	configFile := "router-id: foo"
	_, err := Load([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "validation") {
		t.Errorf("expected validation error, got %+v", err)
	}
}

func TestLoadConfigInvalidOriginPrefix(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
prefixes:
  - foo/24
  - 2001:db8::/48`
	_, err := Load([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "Invalid origin prefix") {
		t.Errorf("expected invalid origin prefix error, got %+v", err)
	}
}

func TestLoadConfigInvalidVRRPState(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
vrrp:
  VRRP 1:
    state: invalid
    interface: eth1
    priority: 255
    vips:
      - 192.0.2.2/24
      - 2001:db8::2/64`
	_, err := Load([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "VRRP state must be") {
		t.Errorf("expected VRRP state error, got %+v", err)
	}
}

func TestLoadConfigInvalidStaticPrefix(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
kernel:
  statics:
    "foo/24" : "192.0.2.10"
    "2001:db8:2::/64" : "2001:db8::1"
`
	_, err := Load([]byte(configFile))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Invalid static prefix")
}

func TestLoadConfigInvalidVIP(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
vrrp:
  VRRP 1:
    state: invalid
    interface: eth1
    priority: 255
    vips:
      - foo/24
      - 2001:db8::2/64`

	_, err := Load([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "Invalid VIP") {
		t.Errorf("expected invalid VIP error, got %+v", err)
	}
}

func TestTemplateInheritance(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
templates:
  upstream:
    local-pref: 90
    filter-irr: false

peers:
  Upstream 1:
    asn: 65510
    template: upstream
    neighbors:
      - 192.0.2.2

  Upstream 2:
    asn: 65520
    template: upstream
    filter-irr: true
    neighbors:
      - 192.0.2.3

  Upstream 3:
    asn: 65530
    local-pref: 2
    filter-irr: false
    neighbors:
      - 192.0.2.4
`
	globalConfig, err := Load([]byte(configFile))
	if err != nil {
		t.Error(err)
	}

	assert.Len(t, globalConfig.Peers, 3)

	upstream1 := globalConfig.Peers["Upstream 1"]
	assert.Equal(t, 65510, *upstream1.ASN)
	assert.Equal(t, 90, *upstream1.LocalPref)
	assert.False(t, *upstream1.FilterIRR)
	assert.True(t, *upstream1.FilterRPKI)

	upstream2 := globalConfig.Peers["Upstream 2"]
	assert.Equal(t, 65520, *upstream2.ASN)
	assert.Equal(t, 90, *upstream2.LocalPref)
	assert.True(t, *upstream2.FilterIRR)
	assert.True(t, *upstream2.FilterRPKI)

	upstream3 := globalConfig.Peers["Upstream 3"]
	assert.Equal(t, 65530, *upstream3.ASN)
	assert.Equal(t, 2, *upstream3.LocalPref)
	assert.False(t, *upstream3.FilterIRR)
	assert.True(t, *upstream3.FilterRPKI)
}
