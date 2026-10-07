package yang

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

func TestGenerateContainsExpectedNodes(t *testing.T) {
	module, err := Generate()
	require.NoError(t, err)

	for _, expected := range []string{
		"module pathvector {",
		"yang-version 1.1;",
		`namespace "urn:pathvector:config";`,
		"prefix pv;",
		// Groupings for map[string]*Struct element types
		"grouping peer {",
		"grouping vrrp-instance {",
		"grouping bfd-instance {",
		"grouping mrt-instance {",
		// Named lists
		"list peers {\n    key \"name\";",
		"list templates {\n    key \"name\";",
		"list vrrp {",
		"list bfd {",
		"list mrt {",
		"uses peer;",
		"uses vrrp-instance;",
		// Containers from struct pointers
		"container kernel {",
		"container optimizer {",
		// Scalars, defaults and descriptions
		"leaf asn {\n    type int64;\n    default \"0\";\n    description \"Autonomous System Number\";",
		"leaf peeringdb-cache {\n    type boolean;\n    default \"true\";",
		"leaf peeringdb-query-timeout {\n    type uint64;",
		"type decimal64 {\n        fraction-digits 4;\n      }\n      default \"0.5\";",
		// Leaf-lists from slices
		"leaf-list prefixes {\n    type string;",
		"leaf-list transit-asns {\n    type uint32;",
		"leaf-list prepend-path {\n      type uint32;",
		// Other maps as key/value lists
		"list authorized-providers {\n    key \"key\";",
		"list as-prefs {\n      key \"key\";",
		"list statics {\n      key \"key\";",
		"list plugins {\n    key \"key\";",
	} {
		assert.Contains(t, module, expected)
	}

	// Internal fields with yaml:"-" are skipped
	for _, unexpected := range []string{"protocol-name", "ProtocolName", "prefixes4", "vips4", "Db", "rtr-server-host"} {
		assert.NotContains(t, module, unexpected)
	}

	// Braces are balanced
	assert.Equal(t, strings.Count(module, "{"), strings.Count(module, "}"))
}

func TestGenerateDeterministic(t *testing.T) {
	a, err := Generate()
	require.NoError(t, err)
	b, err := Generate()
	require.NoError(t, err)
	assert.Equal(t, a, b)
}

func TestGroupingName(t *testing.T) {
	assert.Equal(t, "peer", groupingName(reflect.TypeOf(config.Peer{})))
	assert.Equal(t, "vrrp-instance", groupingName(reflect.TypeOf(&config.VRRPInstance{})))
	assert.Equal(t, "bfd-instance", groupingName(reflect.TypeOf(config.BFDInstance{})))
	assert.Equal(t, "mrt-instance", groupingName(reflect.TypeOf(config.MRTInstance{})))
}

func TestQuote(t *testing.T) {
	assert.Equal(t, `"plain"`, quote("plain"))
	assert.Equal(t, `"say \"hi\" it's"`, quote(`say "hi" it's`))
	assert.Equal(t, `"back\\slash\nnewline"`, quote("back\\slash\nnewline"))
}

// TestCommittedModuleUpToDate ensures docs/static/pathvector.yang matches the generator output
func TestCommittedModuleUpToDate(t *testing.T) {
	module, err := Generate()
	require.NoError(t, err)
	committed, err := os.ReadFile(filepath.Join("..", "..", "docs", "static", "pathvector.yang"))
	require.NoError(t, err)
	assert.Equal(t, module, string(committed), "regenerate with: go run . yang > docs/static/pathvector.yang")
}

// TestPyang validates the module with pyang if it is installed
func TestPyang(t *testing.T) {
	if _, err := exec.LookPath("pyang"); err != nil {
		t.Skip("pyang not installed")
	}
	module, err := Generate()
	require.NoError(t, err)
	file := filepath.Join(t.TempDir(), "pathvector.yang")
	require.NoError(t, os.WriteFile(file, []byte(module), 0o600))
	out, err := exec.Command("pyang", "--strict", file).CombinedOutput()
	assert.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(string(out)))
}

func TestIsRFC7951(t *testing.T) {
	assert.True(t, IsRFC7951([]byte(` {"pathvector:asn": "65530"}`)))
	assert.False(t, IsRFC7951([]byte(`{"asn": 65530}`)))
	assert.False(t, IsRFC7951([]byte("asn: 65530\n")))
	assert.False(t, IsRFC7951([]byte(`{invalid`)))
}

func TestFromRFC7951(t *testing.T) {
	instance := `{
  "pathvector:asn": "65530",
  "pathvector:router-id": "192.0.2.1",
  "pathvector:prefixes": ["192.0.2.0/24", "2001:db8::/48"],
  "pathvector:authorized-providers": [
    {"key": 65510, "value": [65511, 65512]}
  ],
  "pathvector:peers": [
    {
      "name": "Example",
      "asn": "65510",
      "neighbors": ["203.0.113.25"],
      "as-prefs": [{"key": 65520, "value": 90}],
      "prefix-communities": [{"key": "192.0.2.0/24", "value": ["65530:1"]}]
    }
  ],
  "pathvector:templates": [
    {"name": "upstream", "local-pref": "80"}
  ],
  "pathvector:kernel": {
    "statics": [{"key": "203.0.113.0/24", "value": "192.0.2.10"}]
  },
  "pathvector:optimizer": {
    "packet-loss-threshold": "0.25"
  },
  "pathvector:plugins": [{"key": "example", "value": "on"}]
}`
	out, err := FromRFC7951([]byte(instance))
	require.NoError(t, err)

	var c config.Config
	require.NoError(t, util.YAMLUnmarshalStrict(out, &c), string(out))

	assert.Equal(t, 65530, c.ASN)
	assert.Equal(t, "192.0.2.1", c.RouterID)
	assert.Equal(t, []string{"192.0.2.0/24", "2001:db8::/48"}, c.Prefixes)
	assert.Equal(t, map[uint32][]uint32{65510: {65511, 65512}}, c.AuthorizedProviders)
	require.Contains(t, c.Peers, "Example")
	assert.Equal(t, 65510, *c.Peers["Example"].ASN)
	assert.Equal(t, []string{"203.0.113.25"}, *c.Peers["Example"].NeighborIPs)
	assert.Equal(t, map[uint32]uint32{65520: 90}, *c.Peers["Example"].ASPrefs)
	assert.Equal(t, map[string][]string{"192.0.2.0/24": {"65530:1"}}, *c.Peers["Example"].PrefixCommunities)
	require.Contains(t, c.Templates, "upstream")
	assert.Equal(t, 80, *c.Templates["upstream"].LocalPref)
	assert.Equal(t, map[string]string{"203.0.113.0/24": "192.0.2.10"}, c.Kernel.Statics)
	assert.Equal(t, 0.25, c.Optimizer.PacketLossThreshold)
	assert.Equal(t, map[string]string{"example": "on"}, c.Plugins)
}

func TestFromRFC7951Errors(t *testing.T) {
	for _, tc := range []string{
		`[]`,
		`{"pathvector:asn": "not-a-number"}`,
		`{"pathvector:peers": [{"asn": 65510}]}`,
		`{"pathvector:peers": "invalid"}`,
		`{"pathvector:authorized-providers": [{"value": [1]}]}`,
	} {
		_, err := FromRFC7951([]byte(tc))
		assert.Error(t, err, tc)
	}
}
