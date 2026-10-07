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
		if cType != "" && tc.shouldError {
			t.Errorf("categorizeCommunity should have errored on '%s' but didn't. expected error, got '%s'", tc.input, cType)
		} else if cType == "" && !tc.shouldError {
			t.Errorf("categorizeCommunity shouldn't have errored on '%s' but did. expected '%s'", tc.input, tc.expectedOutput)
		} else if cType != tc.expectedOutput {
			t.Errorf("categorizeCommunity %s failed. expected '%v' got '%v'", tc.input, tc.expectedOutput, cType)
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
	for peerName, peerData := range globalConfig.Peers {
		switch peerName {
		case "Peer 10":
			assert.Equal(t, 65510, util.Deref(peerData.ASN))
			assert.Equal(t, 110, util.Deref(peerData.LocalPref))
			assert.True(t, util.Deref(peerData.SetLocalPref))
			assert.Nil(t, peerData.DefaultLocalPref)
		case "Peer 20":
			assert.Equal(t, 65520, util.Deref(peerData.ASN))
			assert.Equal(t, 100, util.Deref(peerData.LocalPref))
			assert.False(t, util.Deref(peerData.SetLocalPref))
			assert.Equal(t, 120, util.Deref(peerData.DefaultLocalPref))
		default:
			t.Errorf("peer %s unexpected", peerName)
		}
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	configFile := "INVALID YAML"
	_, err := Load([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "YAML unmarshal") {
		t.Errorf("expected yaml unmarshal error, got %+v", err)
	}
}

func TestLoadConfigUnknownFieldHint(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
peers:
  Example:
    asn: 65510
    neighbors:
      - 203.0.113.12
    peeringdb-cache: false
`
	_, err := Load([]byte(configFile))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "field peeringdb-cache not found in type config.Peer")
	assert.Contains(t, err.Error(), "(peeringdb-cache is a global option, not a per-peer option)")

	configFile = `
asn: 34553
router-id: 192.0.2.1
auto-import-limits: true
`
	_, err = Load([]byte(configFile))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "(auto-import-limits is a per-peer option, set it under a peer or template)")

	configFile = `
asn: 34553
router-id: 192.0.2.1
peers:
  Example:
    asn: 65510
    neighbors:
      - 203.0.113.12
    not-a-real-option: true
`
	_, err = Load([]byte(configFile))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "field not-a-real-option not found in type config.Peer")
	assert.NotContains(t, err.Error(), "option, ")
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
	if err == nil || !strings.Contains(err.Error(), "Invalid static prefix") {
		t.Errorf("expected invalid static prefix error, got %+v", err)
	}
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
    as-set: AS-EXAMPLE
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

	for peerName, peerData := range globalConfig.Peers {
		if peerName == "Upstream 1" {
			if *peerData.ASN != 65510 {
				t.Errorf("peer %s expected ASN 65510 got %d", peerName, *peerData.ASN)
			}
			if *peerData.LocalPref != 90 {
				t.Errorf("peer %s expected local-pref 90 got %d", peerName, *peerData.LocalPref)
			}
			if *peerData.FilterIRR != false {
				t.Errorf("peer %s expected filter-irr false got %v", peerName, *peerData.FilterIRR)
			}
			if *peerData.FilterRPKI != true {
				t.Errorf("peer %s expected filter-rpki true got %v", peerName, *peerData.FilterIRR)
			}
		} else if peerName == "Upstream 2" {
			if *peerData.ASN != 65520 {
				t.Errorf("peer %s expected ASN 65520 got %d", peerName, *peerData.ASN)
			}
			if *peerData.LocalPref != 90 {
				t.Errorf("peer %s expected local-pref 90 got %d", peerName, *peerData.LocalPref)
			}
			if *peerData.FilterIRR != true {
				t.Errorf("peer %s expected filter-irr true got %v", peerName, *peerData.FilterIRR)
			}
			if *peerData.FilterRPKI != true {
				t.Errorf("peer %s expected filter-rpki true got %v", peerName, *peerData.FilterIRR)
			}
		} else if peerName == "Upstream 3" {
			if *peerData.ASN != 65530 {
				t.Errorf("peer %s expected ASN 65530 got %d", peerName, *peerData.ASN)
			}
			if *peerData.LocalPref != 2 {
				t.Errorf("peer %s expected local-pref 2 got %d", peerName, *peerData.LocalPref)
			}
			if *peerData.FilterIRR != false {
				t.Errorf("peer %s expected filter-irr false got %v", peerName, *peerData.FilterIRR)
			}
			if *peerData.FilterRPKI != true {
				t.Errorf("peer %s expected filter-rpki true got %v", peerName, *peerData.FilterIRR)
			}
		} else {
			t.Errorf("")
		}
	}
}

func TestSplitPrefixesByAF(t *testing.T) {
	v4, v6, err := splitPrefixesByAF(&[]string{"2001:db8::/32", "192.0.2.0/24", "198.51.100.0/24{24,32}", "2001:db8:1::/48+"})
	assert.Nil(t, err)
	assert.Equal(t, []string{"192.0.2.0/24", "198.51.100.0/24{24,32}"}, *v4)
	assert.Equal(t, []string{"2001:db8::/32", "2001:db8:1::/48+"}, *v6)

	v4, v6, err = splitPrefixesByAF(nil)
	assert.Nil(t, err)
	assert.Nil(t, v4)
	assert.Nil(t, v6)

	_, _, err = splitPrefixesByAF(&[]string{"not-a-prefix"})
	assert.NotNil(t, err)
}

func TestLoadAnnounceOriginatedOriginCommunities(t *testing.T) {
	base := `
asn: 34553
router-id: 192.0.2.1
peers:
  iBGP:
    asn: 34553
    neighbors:
      - 192.0.2.10
`
	c, err := Load([]byte(base))
	assert.NoError(t, err)
	assert.False(t, util.Deref(c.Peers["iBGP"].AnnounceOriginated), "no prefixes or origin communities, nothing to originate")

	c, err = Load([]byte("origin-communities: [\"34553:1:1\"]\n" + base))
	assert.NoError(t, err)
	assert.True(t, util.Deref(c.Peers["iBGP"].AnnounceOriginated), "origin communities define locally originated routes")
}

func TestLoadIBGPLocalPref(t *testing.T) {
	c, err := Load([]byte(`
asn: 34553
router-id: 192.0.2.1
templates:
  core:
    local-pref: 90
peers:
  iBGP:
    asn: 34553
    neighbors: [192.0.2.10]
  iBGP explicit:
    asn: 34553
    local-pref: 120
    neighbors: [192.0.2.11]
  iBGP explicit set:
    asn: 34553
    set-local-pref: true
    neighbors: [192.0.2.12]
  iBGP template:
    asn: 34553
    template: core
    neighbors: [192.0.2.13]
  eBGP:
    asn: 65510
    neighbors: [192.0.2.14]
`))
	assert.NoError(t, err)
	assert.False(t, util.Deref(c.Peers["iBGP"].SetLocalPref))
	assert.True(t, util.Deref(c.Peers["iBGP explicit"].SetLocalPref))
	assert.True(t, util.Deref(c.Peers["iBGP explicit set"].SetLocalPref))
	assert.True(t, util.Deref(c.Peers["iBGP template"].SetLocalPref))
	assert.Equal(t, 90, util.Deref(c.Peers["iBGP template"].LocalPref))
	assert.True(t, util.Deref(c.Peers["eBGP"].SetLocalPref))
}

func TestLoadJSON(t *testing.T) {
	// YAML is a superset of JSON, so a JSON document with the YAML structure loads as-is
	configFile := `{
  "asn": 34553,
  "router-id": "192.0.2.1",
  "prefixes": ["192.0.2.0/24"],
  "templates": {"upstream": {"local-pref": 80}},
  "peers": {
    "Example": {"asn": 65530, "template": "upstream", "neighbors": ["203.0.113.25"]}
  }
}`
	globalConfig, err := Load([]byte(configFile))
	assert.NoError(t, err)
	assert.Equal(t, 34553, globalConfig.ASN)
	assert.Equal(t, 65530, *globalConfig.Peers["Example"].ASN)
	assert.Equal(t, 80, *globalConfig.Peers["Example"].LocalPref)
}

func TestLoadRFC7951JSON(t *testing.T) {
	// RFC 7951 JSON instance of the pathvector YANG module
	configFile := `{
  "pathvector:asn": "34553",
  "pathvector:router-id": "192.0.2.1",
  "pathvector:prefixes": ["192.0.2.0/24"],
  "pathvector:templates": [{"name": "upstream", "local-pref": "80"}],
  "pathvector:peers": [
    {"name": "Example", "asn": "65530", "template": "upstream", "neighbors": ["203.0.113.25"]}
  ]
}`
	globalConfig, err := Load([]byte(configFile))
	assert.NoError(t, err)
	assert.Equal(t, 34553, globalConfig.ASN)
	assert.Equal(t, 65530, *globalConfig.Peers["Example"].ASN)
	assert.Equal(t, 80, *globalConfig.Peers["Example"].LocalPref)
	assert.Equal(t, []string{"203.0.113.25"}, *globalConfig.Peers["Example"].NeighborIPs)
}

func TestLoadASSetRequired(t *testing.T) {
	for _, option := range []string{"filter-irr", "auto-as-set-members"} {
		base := `
asn: 34553
router-id: 192.0.2.1
peers:
  Example:
    asn: 65510
    neighbors: [192.0.2.10]
    ` + option + `: true
`
		_, err := Load([]byte(base))
		assert.ErrorContains(t, err, option+" requires as-set or auto-as-set")

		_, err = Load([]byte(base + "    as-set: AS-EXAMPLE\n"))
		assert.NoError(t, err)

		_, err = Load([]byte(base + "    auto-as-set: true\n"))
		assert.NoError(t, err)
	}
}

func TestTemplateParentInheritance(t *testing.T) {
	c, err := Load([]byte(`
asn: 34553
router-id: 192.0.2.1
templates:
  base:
    local-pref: 90
    filter-transit-asns: true
    add-on-import: ["34553:0:1"]
  ix:
    template: base
    local-pref: 110
    add-on-import: ["34553:0:2"]
  ix-merge:
    template: base
    merge-template-lists: true
    add-on-import: ["34553:0:2"]
peers:
  Peer 1:
    asn: 65510
    template: ix
    neighbors: [192.0.2.2]
  Peer 2:
    asn: 65520
    template: ix-merge
    merge-template-lists: true
    add-on-import: ["34553:0:3"]
    neighbors: [192.0.2.3]
  Peer 3:
    asn: 65530
    template: ix-merge
    neighbors: [192.0.2.4]
`))
	assert.NoError(t, err)
	p1 := c.Peers["Peer 1"]
	assert.Equal(t, 110, *p1.LocalPref, "child template overrides parent")
	assert.True(t, *p1.FilterTransitASNs, "inherited from parent template")
	assert.Equal(t, []string{"34553:0:2"}, *p1.ImportCommunities, "lists replace by default")

	p2 := c.Peers["Peer 2"]
	assert.Equal(t, 90, *p2.LocalPref)
	assert.Equal(t, []string{"34553:0:1", "34553:0:2", "34553:0:3"}, *p2.ImportCommunities, "lists merged through the template chain")

	p3 := c.Peers["Peer 3"]
	assert.Equal(t, []string{"34553:0:1", "34553:0:2"}, *p3.ImportCommunities, "merging into one peer doesn't modify the template")
}

func TestTemplateInheritanceLoop(t *testing.T) {
	_, err := Load([]byte(`
asn: 34553
router-id: 192.0.2.1
templates:
  a:
    template: b
  b:
    template: a
`))
	assert.ErrorContains(t, err, "template inheritance loop")

	_, err = Load([]byte(`
asn: 34553
router-id: 192.0.2.1
templates:
  a:
    template: missing
`))
	assert.ErrorContains(t, err, "parent template missing which is not defined")
}
