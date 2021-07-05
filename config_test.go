package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
prefixes:
  - 192.0.2.0/24
  - 2001:db8::/48
augments:
  statics:
    "203.0.113.0/24" : "192.0.2.10"
    "2001:db8:2::/64" : "2001:db8::1"
vrrp:
  - state: primary
    interface: eth0
    priority: 255
    vips:
      - 192.0.2.1/24
      - 2001:db8::1/64
  - state: backup
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

	globalConfig, err := loadConfig([]byte(configFile))
	if err != nil {
		t.Error(err)
	}

	if globalConfig.ASN != 34553 {
		t.Errorf("expected ASN 34553 got %d", globalConfig.ASN)
	}
	if globalConfig.RouterID != "192.0.2.1" {
		t.Errorf("expected router-id 192.0.2.1 got %s", globalConfig.RouterID)
	}
	if len(globalConfig.Peers) != 1 {
		t.Errorf("expected 1 peer, got %d", len(globalConfig.Peers))
	}
	if *globalConfig.Peers["Example"].ASN != 65530 {
		t.Errorf("expected peer ASN 34553 got %d", globalConfig.Peers["Example"].ASN)
	}
	if !reflect.DeepEqual(*globalConfig.Peers["Example"].NeighborIPs, []string{"203.0.113.25", "2001:db8:2::25"}) {
		t.Errorf("expected neighbor ips [203.0.113.25 2001:db8:2::25] got %v", globalConfig.Peers["Example"].NeighborIPs)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	configFile := "INVALID YAML"
	_, err := loadConfig([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "YAML unmarshal") {
		t.Errorf("expected yaml unmarshal error, got %+v", err)
	}
}

func TestLoadConfigValidationError(t *testing.T) {
	configFile := "router-id: foo"
	_, err := loadConfig([]byte(configFile))
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
	_, err := loadConfig([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "Invalid origin prefix") {
		t.Errorf("expected invalid origin prefix error, got %+v", err)
	}
}

func TestLoadConfigInvalidVRRPState(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
vrrp:
  - state: invalid
    interface: eth1
    priority: 255
    vips:
      - 192.0.2.2/24
      - 2001:db8::2/64`
	_, err := loadConfig([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "VRRP state must be") {
		t.Errorf("expected VRRP state error, got %+v", err)
	}
}

func TestLoadConfigInvalidStaticPrefix(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
augments:
  statics:
    "foo/24" : "192.0.2.10"
    "2001:db8:2::/64" : "2001:db8::1"
`
	_, err := loadConfig([]byte(configFile))
	if err == nil || !strings.Contains(err.Error(), "Invalid static prefix") {
		t.Errorf("expected invalid static prefix error, got %+v", err)
	}
}

func TestLoadConfigInvalidVIP(t *testing.T) {
	configFile := `
asn: 34553
router-id: 192.0.2.1
vrrp:
  - state: invalid
    interface: eth1
    priority: 255
    vips:
      - foo/24
      - 2001:db8::2/64`

	_, err := loadConfig([]byte(configFile))
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
	globalConfig, err := loadConfig([]byte(configFile))
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

// TODO: lint the resulting markdown files
func TestDocumentConfig(t *testing.T) {
	documentConfig()
}

func TestSanitizeConfigName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"main.foo", "foo"},
		{"*string", "string"},
		{"map[string]*peer", "map[string]peer"},
	}
	for _, tc := range testCases {
		out := sanitizeConfigName(tc.input)
		if out != tc.expected {
			t.Errorf("expected %s got %s", tc.expected, tc.input)
		}
	}
}
