package process

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/embed"
	"github.com/natesales/pathvector/pkg/templating"
)

// renderConfig loads a YAML config, renders the global and peer templates into a temporary directory and,
// if a BIRD binary is available, validates the result with `bird -p`. It returns the rendered files keyed by peer name ("" for the global config).
func renderConfig(t *testing.T, configYAML string) map[string]string {
	t.Helper()
	c, err := Load([]byte(configYAML))
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	if err := templating.Load(embed.FS); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	out := map[string]string{}

	var global bytes.Buffer
	if err := templating.Template.ExecuteTemplate(&global, "global.tmpl", c); err != nil {
		t.Fatalf("global template: %v", err)
	}
	out[""] = global.String()
	if err := os.WriteFile(path.Join(dir, "bird.conf"), global.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	for name, p := range c.Peers {
		var b bytes.Buffer
		if err := templating.Template.ExecuteTemplate(&b, "peer.tmpl", &templating.Wrapper{Name: name, Peer: *p, Config: *c}); err != nil {
			t.Fatalf("peer template: %v", err)
		}
		out[name] = bird.Reformat(b.String())
		if err := os.WriteFile(path.Join(dir, fmt.Sprintf("AS%d_%s.conf", *p.ASN, *p.ProtocolName)), []byte(out[name]), 0600); err != nil {
			t.Fatal(err)
		}
	}

	if birdBin, err := exec.LookPath("bird"); err == nil {
		cmd := exec.Command(birdBin, "-p", "-c", "bird.conf")
		cmd.Dir = dir
		if o, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("BIRD validation failed: %v\n%s", err, o)
		}
	} else {
		t.Log("bird binary not found, skipping BIRD validation")
	}

	return out
}

// protocolLevel returns the part of a rendered peer config outside of channel ({ipv4,ipv6} { ... }) blocks
func protocolLevel(conf string) string {
	var out strings.Builder
	depth := 0
	for _, line := range strings.Split(conf, "\n") {
		trimmed := strings.TrimSpace(line)
		if depth == 1 && (strings.HasPrefix(trimmed, "ipv4 {") || strings.HasPrefix(trimmed, "ipv6 {")) {
			depth = 2
			continue
		}
		if depth == 0 && strings.HasPrefix(trimmed, "protocol bgp") {
			depth = 1
			continue
		}
		if depth == 1 {
			if trimmed == "}" {
				depth = 0
				continue
			}
			out.WriteString(trimmed + "\n")
			continue
		}
		if depth >= 2 {
			depth += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			if depth < 2 {
				depth = 1
			}
		}
	}
	return out.String()
}

func TestRenderAdvertiseHostname(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: router.example.com
rpki-enable: false
peers:
  AS112:
    asn: 112
    advertise-hostname: true
    neighbors: [192.0.2.112, 2001:db8::112]
`)
	// advertise hostname is a protocol option, it must not be rendered inside a channel block
	if !strings.Contains(protocolLevel(out["AS112"]), "advertise hostname on;") {
		t.Errorf("advertise hostname not found at protocol level:\n%s", out["AS112"])
	}
}

func TestRenderDefaultImportLimits(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
peers:
  Transit:
    asn: 6939
    neighbors: [192.0.2.1, 2001:db8::1]
`)
	// Defaults must match the documented values (import-limit4: 1000000, import-limit6: 300000)
	assert.Contains(t, out["Transit"], "define AS6939_TRANSIT_IMPORT_v4 = 1000000;")
	assert.Contains(t, out["Transit"], "define AS6939_TRANSIT_IMPORT_v6 = 300000;")
}

func TestRenderASPrefsPerAF(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
peers:
  Peer:
    asn: 64496
    mp-unicast-46: true
    as-prefs:
      174: 90
    as-prefs4:
      6939: 80
    as-prefs6:
      6939: 120
    neighbors: [192.0.2.1]
`)
	conf := out["Peer"]
	v4 := conf[strings.Index(conf, "ipv4 {"):strings.Index(conf, "ipv6 {")]
	v6 := conf[strings.Index(conf, "ipv6 {"):]
	assert.Contains(t, v4, "if (174 ~ bgp_path) then { bgp_local_pref = 90; }")
	assert.Contains(t, v6, "if (174 ~ bgp_path) then { bgp_local_pref = 90; }")
	assert.Contains(t, v4, "if (6939 ~ bgp_path) then { bgp_local_pref = 80; }")
	assert.NotContains(t, v4, "bgp_local_pref = 120")
	assert.Contains(t, v6, "if (6939 ~ bgp_path) then { bgp_local_pref = 120; }")
	assert.NotContains(t, v6, "bgp_local_pref = 80")
	// AF-specific prefs are evaluated after as-prefs so they take precedence
	assert.Less(t, strings.Index(v4, "bgp_local_pref = 90"), strings.Index(v4, "bgp_local_pref = 80"))
}

func TestRenderLocalPrefPrecedence(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
peers:
  Peer:
    asn: 64496
    local-pref: 100
    community-prefs:
      "65530:0:200": 95
    as-prefs:
      6939: 150
    neighbors: [192.0.2.1]
`)
	conf := out["Peer"]
	// Last match wins, so as-prefs must be evaluated after community-prefs to take precedence
	localPref := strings.Index(conf, "bgp_local_pref = 100;")
	communityPref := strings.Index(conf, "if ((65530,0,200) ~ bgp_large_community) then { bgp_local_pref = 95; }")
	asPref := strings.Index(conf, "if (6939 ~ bgp_path) then { bgp_local_pref = 150; }")
	assert.True(t, localPref >= 0 && communityPref >= 0 && asPref >= 0, conf)
	assert.Less(t, localPref, communityPref)
	assert.Less(t, communityPref, asPref)
}

func TestRenderPrefixPrefs(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
peers:
  Peer:
    asn: 64496
    as-prefs:
      6939: 150
    prefix-prefs:
      "198.51.100.0/24+": 200
      "2001:db8:100::/40{40,48}": 210
    neighbors: [192.0.2.1, 2001:db8::1]
`)
	conf := out["Peer"]
	v4 := conf[:strings.Index(conf, "ipv6 {")]
	v6 := conf[strings.Index(conf, "ipv6 {"):]
	assert.Contains(t, v4, "if (net ~ [ 198.51.100.0/24+ ]) then { bgp_local_pref = 200; }")
	assert.NotContains(t, v4, "2001:db8:100::/40")
	assert.Contains(t, v6, "if (net ~ [ 2001:db8:100::/40{40,48} ]) then { bgp_local_pref = 210; }")
	assert.NotContains(t, v6, "198.51.100.0/24")
	// prefix-prefs take precedence over as-prefs (last match wins)
	assert.Less(t, strings.Index(v4, "bgp_local_pref = 150"), strings.Index(v4, "bgp_local_pref = 200"))
}

func TestRenderGateway(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
peers:
  Upstream:
    asn: 64496
    gateway: recursive
    neighbors: ["fe80::1%eth0"]
`)
	assert.Contains(t, out["Upstream"], "gateway recursive;")
	assert.NotContains(t, protocolLevel(out["Upstream"]), "gateway", "gateway is a channel option")

	_, err := Load([]byte(`
asn: 65530
router-id: 192.0.2.1
peers:
  Upstream:
    asn: 64496
    gateway: indirect
    neighbors: [192.0.2.1]
`))
	assert.ErrorContains(t, err, "invalid gateway mode")
}

func TestRenderCost(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
peers:
  Core:
    asn: 65530
    direct: true
    cost: 20
    neighbors: [192.0.2.2]
`)
	assert.Contains(t, out["Core"], "cost 20;")
	assert.NotContains(t, protocolLevel(out["Core"]), "cost", "cost is a channel option")
}

func TestRenderPeerSource(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
source4: 192.0.2.1
source6: 2001:db8::1
peers:
  IX:
    asn: 64496
    source4: 198.51.100.1
    source6: 2001:db8:ffff::1
    neighbors: [198.51.100.2, 2001:db8:ffff::2]
`)
	conf := out["IX"]
	v4 := conf[:strings.Index(conf, "ipv6 {")]
	v6 := conf[strings.Index(conf, "ipv6 {"):]
	assert.Contains(t, v4, "krt_prefsrc = 198.51.100.1;")
	assert.NotContains(t, v4, "2001:db8:ffff::1")
	assert.Contains(t, v6, "krt_prefsrc = 2001:db8:ffff::1;")
	// The global source must not override the per-peer source in the kernel export filter
	assert.Contains(t, out[""], "if !defined(krt_prefsrc) then krt_prefsrc = 192.0.2.1;")

	_, err := Load([]byte(`
asn: 65530
router-id: 192.0.2.1
peers:
  IX:
    asn: 64496
    source4: 2001:db8::1
    neighbors: [198.51.100.2]
`))
	assert.ErrorContains(t, err, "invalid source4 address")
}

func TestRenderNamedCommunities(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
communities:
  BLACKHOLE: 65510:100
  CUSTOMER: 65530:0:10
  UPSTREAM: 65530:0:20
add-on-export: [UPSTREAM]
templates:
  customer:
    add-on-import: [CUSTOMER]
peers:
  Customer:
    asn: 64496
    template: customer
    announce: [CUSTOMER, "65530:0:30"]
    community-prefs:
      BLACKHOLE: 50
    neighbors: [192.0.2.2]
  Customer 2:
    asn: 64497
    template: customer
    neighbors: [192.0.2.3]
`)
	conf := out["Customer"]
	assert.Contains(t, conf, "bgp_large_community.add((65530,0,10));")
	assert.Contains(t, conf, "if ((65530,0,10) ~ bgp_large_community) then accept;")
	assert.Contains(t, conf, "if ((65530,0,30) ~ bgp_large_community) then accept;")
	assert.Contains(t, conf, "if ((65510,100) ~ bgp_community) then { bgp_local_pref = 50; }")
	assert.Contains(t, conf, "bgp_large_community.add((65530,0,20));")
	assert.Contains(t, out["Customer 2"], "bgp_large_community.add((65530,0,10));")
	assert.NotContains(t, conf, "(CUSTOMER)", "community names must be replaced")
}

func TestLoadInvalidCommunityName(t *testing.T) {
	_, err := Load([]byte("asn: 65530\nrouter-id: 192.0.2.1\ncommunities:\n  65530:1: 65530:2\n"))
	assert.ErrorContains(t, err, "community names must not be communities")
	_, err = Load([]byte("asn: 65530\nrouter-id: 192.0.2.1\ncommunities:\n  FOO: bar\n"))
	assert.ErrorContains(t, err, "invalid community bar")
	_, err = Load([]byte("asn: 65530\nrouter-id: 192.0.2.1\norigin-communities: [UNDEFINED]\n"))
	assert.Error(t, err, "undefined community names are invalid communities")
}

func TestRenderGlobalProtocolOverrides(t *testing.T) {
	out := renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
device-scan-time: 10
direct-check-link: true
disable-protocols: [kernel4, kernel6]
global-config: |
  protocol kernel kernel4 { ipv4 { import all; export where source != RTS_DEVICE; }; }
  protocol kernel kernel6 { ipv6 { import all; export where source != RTS_DEVICE; }; }
`)
	global := out[""]
	assert.Contains(t, global, "scan time 10;")
	assert.Contains(t, global, "check link yes;")
	assert.Equal(t, 1, strings.Count(global, "protocol kernel kernel4"), "only the user defined kernel4 protocol must be rendered")
	assert.Equal(t, 1, strings.Count(global, "protocol kernel kernel6"))

	out = renderConfig(t, `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
`)
	assert.NotContains(t, out[""], "check link")
	assert.Contains(t, out[""], "protocol kernel kernel4")

	_, err := Load([]byte("asn: 65530\nrouter-id: 192.0.2.1\ndisable-protocols: [bgp]\n"))
	assert.ErrorContains(t, err, "invalid disable-protocols entry bgp")
}

func TestRenderKernelTables(t *testing.T) {
	base := `
asn: 65530
router-id: 192.0.2.1
hostname: r1
rpki-enable: false
source4: 192.0.2.1
prefixes: [192.0.2.0/24]
kernel:
  table: 10
  tables: [100, 200]
  statics:
    "198.51.100.0/24": 192.0.2.254
`
	out := renderConfig(t, base)
	global := out[""]
	for _, name := range []string{"kernel4_table100", "kernel6_table100", "kernel4_table200", "kernel6_table200"} {
		assert.Contains(t, global, "protocol kernel "+name+" {")
	}
	assert.Contains(t, global, "kernel table 100;")
	// The additional tables use the same export policy as the main kernel table
	assert.Equal(t, 3, strings.Count(global, `if (proto = "statics4") then accept;`))
	assert.Equal(t, 3, strings.Count(global, "if !defined(krt_prefsrc) then krt_prefsrc = 192.0.2.1;"))

	// Without additional tables, only kernel4 and kernel6 are rendered
	out = renderConfig(t, strings.ReplaceAll(base, "  tables: [100, 200]\n", ""))
	assert.Equal(t, 2, strings.Count(out[""], "protocol kernel "))

	_, err := Load([]byte(strings.ReplaceAll(base, "[100, 200]", "[10]")))
	assert.ErrorContains(t, err, "invalid kernel table 10")
}

func TestRenderL3VPN(t *testing.T) {
	base := `
asn: 65530
router-id: 192.0.2.1
hostname: rr1
rpki-enable: false
peers:
  PE1:
    asn: 65530
    rr-client: true
    l3vpn: true
    neighbors: [192.0.2.11]
  PE2:
    asn: 65530
    rr-client: true
    l3vpn: true
    import: false
    neighbors: [192.0.2.12]
`
	out := renderConfig(t, base)
	assert.Contains(t, out[""], "vpn4 table vpntab4;")
	assert.Contains(t, out[""], "vpn6 table vpntab6;")
	assert.Contains(t, out["PE1"], "vpn4 mpls {")
	assert.Contains(t, out["PE1"], "vpn6 mpls {")
	assert.Contains(t, out["PE1"], "import all;")
	assert.Contains(t, out["PE2"], "import none;", "import: false applies to VPN routes")

	// VPN tables are only defined when a peer uses l3vpn
	out = renderConfig(t, strings.ReplaceAll(base, "    l3vpn: true\n", ""))
	assert.NotContains(t, out[""], "vpntab4")
	assert.NotContains(t, out["PE1"], "mpls")
}
