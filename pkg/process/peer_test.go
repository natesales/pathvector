package process

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/embed"
	"github.com/natesales/pathvector/pkg/templating"
)

// shimBGPQ4 puts a fake bgpq4 script running body first in PATH for the duration of the test
func shimBGPQ4(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	//nolint:gosec
	require.NoError(t, os.WriteFile(filepath.Join(dir, "bgpq4"), []byte("#!/bin/sh\n"+body+"\n"), 0755))
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// fakeBGPQ4Body answers prefix queries with the given v4/v6 output and member queries with the given JSON
func fakeBGPQ4Body(prefixes4, prefixes6, members string) string {
	return `case "$*" in
  *-Ab4*) printf 'define x = [\n` + prefixes4 + `\n];\n' ;;
  *-Ab6*) printf 'define x = [\n` + prefixes6 + `\n];\n' ;;
  *-tj*) echo '` + members + `' ;;
esac`
}

// loadTestPeers loads a config with the given peers YAML and prepares templates and a temporary cache directory
func loadTestPeers(t *testing.T, peersYAML string) *config.Config {
	t.Helper()
	c, err := Load([]byte(`
asn: 34553
router-id: 192.0.2.1
prefixes:
  - 192.0.2.0/24
peers:
` + peersYAML))
	require.NoError(t, err)
	c.CacheDirectory = t.TempDir()
	require.NoError(t, templating.Load(embed.FS))
	return c
}

// runPeer runs peer() for a single peer and returns its rendered config
func runPeer(t *testing.T, c *config.Config, name string) string {
	t.Helper()
	return runPeerOffline(t, c, name, false)
}

// runPeerOffline runs peer() for a single peer with the given offline mode and returns its rendered config
func runPeerOffline(t *testing.T, c *config.Config, name string, offline bool) string {
	t.Helper()
	wg := new(sync.WaitGroup)
	wg.Add(1)
	peer(name, c.Peers[name], c, offline, wg)
	wg.Wait()
	files, err := filepath.Glob(filepath.Join(c.CacheDirectory, "AS*_"+*c.Peers[name].ProtocolName+".conf"))
	require.NoError(t, err)
	require.Len(t, files, 1)
	b, err := os.ReadFile(files[0])
	require.NoError(t, err)
	return string(b)
}

const irrTestPeers = `
  Good:
    asn: 65510
    as-set: AS-GOOD
    filter-irr: true
    auto-as-set-members: true
    filter-as-set: true
    neighbors:
      - 203.0.113.10
      - 2001:db8::10
  Bad:
    asn: 65520
    as-set: AS-BAD
    filter-irr: true
    auto-as-set-members: true
    filter-as-set: true
    neighbors:
      - 203.0.113.20
`

func TestPeerIRRFailureRejectsImports(t *testing.T) {
	shimBGPQ4(t, `case "$*" in *AS-BAD*) echo "ERROR: something went wrong" >&2; exit 1 ;; esac
`+fakeBGPQ4Body("    198.51.100.0/24", "    2001:db8:1::/48", `{"NN": [65510]}`))
	c := loadTestPeers(t, irrTestPeers)

	good := runPeer(t, c, "Good")
	assert.True(t, *c.Peers["Good"].Import)
	assert.NotContains(t, good, "reject; # import: false")
	assert.Contains(t, good, "198.51.100.0/24")
	assert.Contains(t, good, "2001:db8:1::/48")

	bad := runPeer(t, c, "Bad")
	assert.False(t, *c.Peers["Bad"].Import)
	assert.Contains(t, bad, "reject; # import: false")
	// filter-as-set with no members is disabled to keep the BIRD config valid (imports are rejected anyway)
	assert.False(t, *c.Peers["Bad"].FilterASSet)
	assert.NotContains(t, bad, "AS_SET_MEMBERS = [")
}

func TestPeerIRRNoPrefixesRejectsImports(t *testing.T) {
	shimBGPQ4(t, fakeBGPQ4Body("", "", `{"NN": [65520]}`))
	c := loadTestPeers(t, irrTestPeers)

	out := runPeer(t, c, "Bad")
	assert.False(t, *c.Peers["Bad"].Import)
	assert.Contains(t, out, "reject; # import: false")
}

func TestPeerIRRSingleFamily(t *testing.T) {
	// Only IPv6 prefixes registered: import stays enabled, the IPv4 prefix set is empty and so rejects everything
	shimBGPQ4(t, fakeBGPQ4Body("", "    2001:db8:1::/48", `{"NN": [65510]}`))
	c := loadTestPeers(t, irrTestPeers)

	out := runPeer(t, c, "Good")
	assert.True(t, *c.Peers["Good"].Import)
	assert.True(t, strings.Contains(out, "_PFX_v4 = -empty-;"))
}

func TestPeerIRRCacheFallback(t *testing.T) {
	cacheDir := t.TempDir()
	load := func(peers string) *config.Config {
		c := loadTestPeers(t, peers)
		c.CacheDirectory = cacheDir
		return c
	}

	// A successful run populates the cache
	shimBGPQ4(t, fakeBGPQ4Body("    198.51.100.0/24", "    2001:db8:1::/48", `{"NN": [65510, 65511]}`))
	runPeer(t, load(irrTestPeers), "Good")
	assert.FileExists(t, filepath.Join(cacheDir, "irr", "AS65510_GOOD.json"))

	// IRR is now unreachable: the cached prefix sets and members are used instead of rejecting imports
	shimBGPQ4(t, `echo "ERROR: network unreachable" >&2; exit 1`)
	c := load(irrTestPeers)
	out := runPeer(t, c, "Good")
	assert.True(t, *c.Peers["Good"].Import)
	assert.NotContains(t, out, "reject; # import: false")
	assert.Contains(t, out, "198.51.100.0/24")
	assert.Contains(t, out, "2001:db8:1::/48")
	assert.Equal(t, []uint32{65510, 65511}, *c.Peers["Good"].ASSetMembers)

	// Peers without cached data still fail safe
	runPeer(t, c, "Bad")
	assert.False(t, *c.Peers["Bad"].Import)

	// Cached data for a different as-set is not used
	c = load(strings.ReplaceAll(irrTestPeers, "as-set: AS-GOOD", "as-set: AS-OTHER"))
	runPeer(t, c, "Good")
	assert.False(t, *c.Peers["Good"].Import)
}

func TestPeerOffline(t *testing.T) {
	cacheDir := t.TempDir()
	shimBGPQ4(t, fakeBGPQ4Body("    198.51.100.0/24", "    2001:db8:1::/48", `{"NN": [65510]}`))
	c := loadTestPeers(t, irrTestPeers)
	c.CacheDirectory = cacheDir
	runPeer(t, c, "Good")

	// In offline mode bgpq4 must not be run at all
	marker := filepath.Join(t.TempDir(), "ran")
	shimBGPQ4(t, "touch "+marker+"\n"+fakeBGPQ4Body("    192.0.2.0/24", "    2001:db8:9::/48", `{"NN": [65599]}`))
	c = loadTestPeers(t, irrTestPeers)
	c.CacheDirectory = cacheDir
	out := runPeerOffline(t, c, "Good", true)
	assert.NoFileExists(t, marker)
	assert.True(t, *c.Peers["Good"].Import)
	assert.Contains(t, out, "198.51.100.0/24")

	// Without cached data, offline peers fail safe
	runPeerOffline(t, c, "Bad", true)
	assert.False(t, *c.Peers["Bad"].Import)
	assert.NoFileExists(t, marker)
}

// fakeWhoisServer serves the given aut-num responses by query ("AS65510" etc.) and returns its address
func fakeWhoisServer(t *testing.T, responses map[string]string) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			line, _ := bufio.NewReader(conn).ReadString('\n')
			_, _ = conn.Write([]byte(responses[strings.TrimSpace(line)]))
			_ = conn.Close()
		}
	}()
	return l.Addr().String()
}

func TestPeerVerifyIRRPolicy(t *testing.T) {
	shimBGPQ4(t, `echo '{"NN": [34553, 65000]}'`)
	whois := fakeWhoisServer(t, map[string]string{
		// Lists us directly for both families
		"AS65510": "aut-num: AS65510\nmp-import: afi any.unicast from AS34553 accept ANY\nmp-export: afi any.unicast to AS34553 announce AS65510\n",
		// Lists us through an as-set, IPv6 only
		"AS65520": "aut-num: AS65520\nmp-import: afi ipv6.unicast from AS65520:AS-UPSTREAMS accept ANY\nmp-export: afi ipv6.unicast to AS65520:AS-UPSTREAMS announce AS65520\n",
		// Doesn't list us
		"AS65530": "aut-num: AS65530\nimport: from AS174 accept ANY\nexport: to AS174 announce AS65530\n",
	})
	c := loadTestPeers(t, `
  Direct:
    asn: 65510
    verify-irr-policy: true
    neighbors: [203.0.113.10, 2001:db8::10]
  SetV6:
    asn: 65520
    verify-irr-policy: true
    neighbors: [2001:db8::20]
  SetDual:
    asn: 65520
    verify-irr-policy: true
    neighbors: [203.0.113.21, 2001:db8::21]
  Gone:
    asn: 65530
    verify-irr-policy: true
    neighbors: [203.0.113.30]
  Unchecked:
    asn: 65530
    neighbors: [203.0.113.31]
`)
	c.IRRServer = whois

	for name, disabled := range map[string]bool{
		"Direct":    false,
		"SetV6":     false,
		"SetDual":   true, // no IPv4 policy, so the whole peer is disabled
		"Gone":      true,
		"Unchecked": false,
	} {
		out := runPeer(t, c, name)
		assert.Equalf(t, disabled, *c.Peers[name].Disabled, "peer %s", name)
		assert.Equalf(t, disabled, strings.Contains(out, "disabled;"), "peer %s", name)
	}
}

func TestPeerVerifyIRRPolicyUnreachable(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())

	c := loadTestPeers(t, `
  Example:
    asn: 65510
    verify-irr-policy: true
    neighbors: [203.0.113.10]
`)
	c.IRRServer = addr
	runPeer(t, c, "Example")
	assert.False(t, *c.Peers["Example"].Disabled, "whois failure must leave the peer unchanged")

	// Offline mode skips the check
	runPeerOffline(t, c, "Example", true)
	assert.False(t, *c.Peers["Example"].Disabled)
}
