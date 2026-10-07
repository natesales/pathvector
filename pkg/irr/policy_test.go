package irr

import (
	"bufio"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example aut-num objects based on natesales/pathvector#165
const (
	// Imports from and exports to an as-set (AS199514:AS-UPSTREAMS = AS924, AS6939) for IPv6 only
	autNumUpstreamsSet = `% This is the RIPE Database query service.

aut-num:        AS207960
as-name:        RAPDODGE
remarks:        # comment only
mp-export:      afi ipv6.unicast to AS199514:AS-UPSTREAMS announce AS-RAPDODGE
mp-import:      afi ipv6.unicast from AS199514:AS-UPSTREAMS accept ANY
source:         RIPE
`
	// Imports from and exports to AS6939 directly for all address families
	autNumAnyUnicast = `aut-num:        AS207960
mp-export:      afi any.unicast to AS6939 announce AS-ROUTE48
mp-import:      afi any.unicast from AS6939 accept ANY
source:         RIPE
`
)

// upstreamsExpander expands AS199514:AS-UPSTREAMS and counts calls
func upstreamsExpander(calls *int) func(string) ([]uint32, error) {
	return func(asSet string) ([]uint32, error) {
		*calls++
		if asSet == "AS199514:AS-UPSTREAMS" {
			return []uint32{924, 6939}, nil
		}
		return nil, nil
	}
}

func TestParseRPSLObjects(t *testing.T) {
	objects := parseRPSLObjects("% server comment\r\n\r\naut-num: AS1\r\nimport: from AS2 # a comment\r\n  accept ANY\r\n+ AND NOT AS3\r\n\r\n\r\nas-set: AS-FOO\nmembers: AS1\n")
	require.Len(t, objects, 2)
	assert.Equal(t, []rpslAttribute{{"aut-num", "AS1"}, {"import", "from AS2 accept ANY AND NOT AS3"}}, objects[0])
	assert.Equal(t, []rpslAttribute{{"as-set", "AS-FOO"}, {"members", "AS1"}}, objects[1])
}

func TestParsePolicyRule(t *testing.T) {
	testCases := []struct {
		attr, value string
		expected    PolicyRule
	}{
		// Examples from natesales/pathvector#165
		{"mp-export", "afi ipv6.unicast to AS199514:AS-UPSTREAMS announce AS-RAPDODGE", PolicyRule{PolicyExport, false, true, []string{"AS199514:AS-UPSTREAMS"}, true}},
		{"mp-import", "afi ipv6.unicast from AS199514:AS-UPSTREAMS accept ANY", PolicyRule{PolicyImport, false, true, []string{"AS199514:AS-UPSTREAMS"}, true}},
		{"mp-export", "afi any.unicast to AS6939 announce AS-ROUTE48", PolicyRule{PolicyExport, true, true, []string{"AS6939"}, true}},
		{"mp-import", "afi any.unicast from AS6939 accept ANY", PolicyRule{PolicyImport, true, true, []string{"AS6939"}, true}},
		// Plain import/export are IPv4 only
		{"import", "from AS6939 accept ANY", PolicyRule{PolicyImport, true, false, []string{"AS6939"}, true}},
		{"export", "to AS6939 announce AS-ROUTE48", PolicyRule{PolicyExport, true, false, []string{"AS6939"}, true}},
		// mp- without afi applies to all families
		{"mp-import", "from AS6939 accept ANY", PolicyRule{PolicyImport, true, true, []string{"AS6939"}, true}},
		// afi variants and lists
		{"mp-import", "afi ipv4 from AS1 accept ANY", PolicyRule{PolicyImport, true, false, []string{"AS1"}, true}},
		{"mp-import", "afi ipv6 from AS1 accept ANY", PolicyRule{PolicyImport, false, true, []string{"AS1"}, true}},
		{"mp-import", "afi any from AS1 accept ANY", PolicyRule{PolicyImport, true, true, []string{"AS1"}, true}},
		{"mp-import", "afi ipv4.unicast, ipv6.unicast from AS1 accept ANY", PolicyRule{PolicyImport, true, true, []string{"AS1"}, true}},
		{"mp-import", "afi ipv4.unicast,ipv6.unicast from AS1 accept ANY", PolicyRule{PolicyImport, true, true, []string{"AS1"}, true}},
		{"mp-import", "afi ipv4.multicast from AS1 accept ANY", PolicyRule{PolicyImport, false, false, []string{"AS1"}, true}},
		{"mp-import", "AFI IPV6.UNICAST FROM as1 ACCEPT any", PolicyRule{PolicyImport, false, true, []string{"AS1"}, true}},
		// accept NOT ANY exchanges nothing
		{"import", "from AS1 accept NOT ANY", PolicyRule{PolicyImport, true, false, []string{"AS1"}, false}},
		// Missing action
		{"import", "from AS1", PolicyRule{PolicyImport, true, false, []string{"AS1"}, false}},
		// Router addresses, actions, protocol, OR and multiple peerings
		{"import", "protocol BGP4 into OSPF from AS1 192.0.2.1 at 192.0.2.2 action pref=100; accept AS1", PolicyRule{PolicyImport, true, false, []string{"AS1"}, true}},
		{"import", "from AS1 OR AS2 accept ANY", PolicyRule{PolicyImport, true, false, []string{"AS1", "AS2"}, true}},
		{"import", "from AS1 action pref=1; from AS2 accept ANY", PolicyRule{PolicyImport, true, false, []string{"AS1", "AS2"}, true}},
		{"mp-import", "afi ipv6.unicast { from AS1 accept ANY; }", PolicyRule{PolicyImport, false, true, []string{"AS1"}, true}},
	}
	for _, tc := range testCases {
		assert.Equalf(t, tc.expected, parsePolicyRule(tc.attr, tc.value), "%s: %s", tc.attr, tc.value)
	}
}

func TestParsePolicy(t *testing.T) {
	rules, found := ParsePolicy(autNumUpstreamsSet, 207960)
	assert.True(t, found)
	assert.Len(t, rules, 2)

	// Wrong ASN
	_, found = ParsePolicy(autNumUpstreamsSet, 65000)
	assert.False(t, found)

	// Multiple objects from different sources are combined, other object types are ignored
	rules, found = ParsePolicy(autNumUpstreamsSet+"\n"+autNumAnyUnicast+"\nas-set: AS-FOO\nmembers: AS207960\nimport: from AS1 accept ANY\n", 207960)
	assert.True(t, found)
	assert.Len(t, rules, 4)

	_, found = ParsePolicy("%  No entries found for the selected source(s).\n", 207960)
	assert.False(t, found)
}

func TestCheckPolicy(t *testing.T) {
	var calls int
	expand := upstreamsExpander(&calls)
	check := func(response string, localASN uint32) PolicyResult {
		rules, found := ParsePolicy(response, 207960)
		require.True(t, found)
		result, err := CheckPolicy(rules, localASN, expand)
		require.NoError(t, err)
		return result
	}

	// AS6939 and AS924 are members of AS199514:AS-UPSTREAMS: IPv6 only
	assert.Equal(t, PolicyResult{Import6: true, Export6: true}, check(autNumUpstreamsSet, 6939))
	assert.Equal(t, PolicyResult{Import6: true, Export6: true}, check(autNumUpstreamsSet, 924))
	assert.Equal(t, PolicyResult{}, check(autNumUpstreamsSet, 34553))

	// Direct any.unicast policy with AS6939: both families, no as-set expansion needed
	calls = 0
	assert.Equal(t, PolicyResult{true, true, true, true}, check(autNumAnyUnicast, 6939))
	assert.Equal(t, 0, calls)
	assert.Equal(t, PolicyResult{}, check(autNumAnyUnicast, 34553))

	// Only an import is not enough for export
	assert.Equal(t, PolicyResult{Import4: true}, check("aut-num: AS207960\nimport: from AS34553 accept ANY\n", 34553))

	// AS-ANY matches everyone
	assert.Equal(t, PolicyResult{Import4: true, Export4: true}, check("aut-num: AS207960\nimport: from AS-ANY accept ANY\nexport: to AS-ANY announce AS207960\n", 34553))

	// Each as-set is only expanded once
	calls = 0
	check("aut-num: AS207960\nimport: from AS199514:AS-UPSTREAMS accept ANY\nexport: to AS199514:AS-UPSTREAMS announce ANY\n", 34553)
	assert.Equal(t, 1, calls)
}

func TestCheckPolicyExpandError(t *testing.T) {
	failing := func(string) ([]uint32, error) { return nil, errors.New("bgpq4 failed") }

	rules, _ := ParsePolicy(autNumUpstreamsSet, 207960)
	_, err := CheckPolicy(rules, 6939, failing)
	assert.Error(t, err, "undetermined result must be reported")

	// A complete result from direct matches doesn't need the as-set
	rules, _ = ParsePolicy(autNumAnyUnicast+"mp-import: afi any from AS-FOO accept ANY\n", 207960)
	result, err := CheckPolicy(rules, 6939, failing)
	assert.NoError(t, err)
	assert.Equal(t, PolicyResult{true, true, true, true}, result)
}

func TestPolicyResultMissing(t *testing.T) {
	assert.Empty(t, PolicyResult{true, true, true, true}.Missing(34553, true, true))
	assert.Empty(t, PolicyResult{Import6: true, Export6: true}.Missing(34553, false, true))
	assert.Equal(t, []string{"import from AS34553 accept (IPv4)", "export to AS34553 announce (IPv4)"}, PolicyResult{Import6: true, Export6: true}.Missing(34553, true, true))
	assert.Equal(t, []string{"mp-export afi ipv6 to AS34553 announce (IPv6)"}, PolicyResult{Import6: true}.Missing(34553, false, true))
}

func TestSourcesFlagValue(t *testing.T) {
	assert.Equal(t, "", sourcesFlagValue(nil))
	assert.Equal(t, "RADB,RIPE", sourcesFlagValue([]string{"-S", "RADB,RIPE"}))
	assert.Equal(t, "RIPE", sourcesFlagValue([]string{"-R", "24", "-SRIPE"}))
	assert.Equal(t, "RIPE", sourcesFlagValue([]string{"-AS", "RIPE"}))
	assert.Equal(t, "", sourcesFlagValue([]string{"-l", "-S"}))
	assert.Equal(t, "", sourcesFlagValue([]string{"-S"}))
}

// fakeWhois starts a whois server that records the query and replies with response
func fakeWhois(t *testing.T, response string) (addr string, queries chan string) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })
	queries = make(chan string, 10)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			line, _ := bufio.NewReader(conn).ReadString('\n')
			queries <- line
			_, _ = conn.Write([]byte(response))
			_ = conn.Close()
		}
	}()
	return l.Addr().String(), queries
}

func TestQueryAutNum(t *testing.T) {
	addr, queries := fakeWhois(t, autNumAnyUnicast)

	out, err := QueryAutNum(addr, 207960, 5, "")
	require.NoError(t, err)
	assert.Equal(t, autNumAnyUnicast, out)
	assert.Equal(t, "AS207960\r\n", <-queries)

	// IRR sources from bgpq-args are used for the whois query too
	_, err = QueryAutNum(addr, 207960, 5, "-S RADB,RIPE -R 24")
	require.NoError(t, err)
	assert.Equal(t, "-s RADB,RIPE AS207960\r\n", <-queries)

	// Unreachable server
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	closedAddr := l.Addr().String()
	require.NoError(t, l.Close())
	_, err = QueryAutNum(closedAddr, 207960, 1, "")
	assert.Error(t, err)
}

func TestVerifyPolicy(t *testing.T) {
	fakeBGPQ4(t, `echo '{"NN": [924, 6939]}'`)

	addr, _ := fakeWhois(t, autNumUpstreamsSet)
	missing, err := VerifyPolicy(207960, 6939, false, true, addr, 5, "")
	require.NoError(t, err)
	assert.Empty(t, missing)

	missing, err = VerifyPolicy(207960, 6939, true, true, addr, 5, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"import from AS6939 accept (IPv4)", "export to AS6939 announce (IPv4)"}, missing)

	// No aut-num object
	addr, _ = fakeWhois(t, "%  No entries found\n")
	missing, err = VerifyPolicy(207960, 6939, true, false, addr, 5, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"aut-num AS207960 object (not found)"}, missing)

	// as-set expansion fails: undetermined
	fakeBGPQ4(t, "exit 1")
	addr, _ = fakeWhois(t, autNumUpstreamsSet)
	_, err = VerifyPolicy(207960, 6939, false, true, addr, 5, "")
	assert.Error(t, err)
}
