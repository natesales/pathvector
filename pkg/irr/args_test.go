package irr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBGPQ4 installs a fake bgpq4 shell script for the duration of the test.
// The script appends its arguments (one invocation per line) to the returned log file and then runs body.
func fakeBGPQ4(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	logFile := filepath.Join(dir, "args.log")
	script := filepath.Join(dir, "bgpq4")
	content := "#!/bin/sh\necho \"$@\" >> " + logFile + "\n" + body + "\n"
	//nolint:gosec
	require.NoError(t, os.WriteFile(script, []byte(content), 0755))

	orig := bgpq4Command
	bgpq4Command = script
	t.Cleanup(func() { bgpq4Command = orig })
	return logFile
}

// invocations returns the recorded fake bgpq4 invocations
func invocations(t *testing.T, logFile string) []string {
	t.Helper()
	b, err := os.ReadFile(logFile)
	require.NoError(t, err)
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

func TestHasSourcesFlag(t *testing.T) {
	testCases := []struct {
		args     string
		expected bool
	}{
		{"", false},
		{"-S RPKI,AFRINIC,ARIN,APNIC,LACNIC,RADB,RIPE", true},
		{"-SRIPE", true},
		{"-AS RIPE", true},
		{"-4 -S RIPE", true},
		{"-R 24", false},
		{"-h rr.ntt.net", false},
		{"-l AS-SET", false},   // argument of -l is not a flag
		{"-lS NAME", false},    // S is the argument of -l
		{"--S", false},         // not a short flag
		{"-w -p -A", false},    // no sources
		{"-L 2 -S RADB", true}, // sources after another flag with an argument
	}
	for _, tc := range testCases {
		assert.Equalf(t, tc.expected, hasSourcesFlag(strings.Fields(tc.args)), "args %q", tc.args)
	}
}

func TestBuildArgs(t *testing.T) {
	testCases := []struct {
		bgpqArgs  string
		asSet     string
		queryArgs []string
		expected  []string
	}{
		// No user args, plain as-set
		{"", "AS-SERVERFORGE", []string{"-h", "rr.ntt.net", "-Ab4"}, []string{"-h", "rr.ntt.net", "-Ab4", "AS-SERVERFORGE"}},
		// No user -S: source from as-set is used
		{"", "RIPE::AS-SERVERFORGE", []string{"-h", "rr.ntt.net", "-Ab4"}, []string{"-h", "rr.ntt.net", "-Ab4", "-S", "RIPE", "AS-SERVERFORGE"}},
		{"-R 24", "RIPE::AS-SERVERFORGE", []string{"-h", "rr.ntt.net", "-Ab4"}, []string{"-R", "24", "-h", "rr.ntt.net", "-Ab4", "-S", "RIPE", "AS-SERVERFORGE"}},
		// User -S: as-set source is stripped (natesales/pathvector#192)
		{"-S RPKI,AFRINIC,ARIN,APNIC,LACNIC,RADB,RIPE", "RIPE::AS-SERVERFORGE", []string{"-h", "rr.ntt.net", "-Ab4"}, []string{"-S", "RPKI,AFRINIC,ARIN,APNIC,LACNIC,RADB,RIPE", "-h", "rr.ntt.net", "-Ab4", "AS-SERVERFORGE"}},
		{"-S RPKI,AFRINIC,ARIN,APNIC,LACNIC,RADB,RIPE", "RIPE::AS-SERVERFORGE", []string{"-h", "rr.ntt.net", "-tj"}, []string{"-S", "RPKI,AFRINIC,ARIN,APNIC,LACNIC,RADB,RIPE", "-h", "rr.ntt.net", "-tj", "AS-SERVERFORGE"}},
		{"-S RADB", "AS112", []string{"-h", "rr.ntt.net", "-Ab6"}, []string{"-S", "RADB", "-h", "rr.ntt.net", "-Ab6", "AS112"}},
		// Extra whitespace in user args doesn't produce empty arguments
		{"  -S  RADB   -R 24 ", "RIPE::AS1", []string{"-tj"}, []string{"-S", "RADB", "-R", "24", "-tj", "AS1"}},
	}
	for _, tc := range testCases {
		assert.Equalf(t, tc.expected, buildArgs(tc.bgpqArgs, tc.asSet, tc.queryArgs...), "bgpq-args %q as-set %q", tc.bgpqArgs, tc.asSet)
	}
}

func TestPrefixSetSourceArgs(t *testing.T) {
	logFile := fakeBGPQ4(t, `printf 'define x = [\n    192.0.2.0/24,\n    198.51.100.0/24\n];\n'`)

	prefixes, err := PrefixSet("RIPE::AS-SERVERFORGE", 4, "rr.ntt.net", 5, "-S RPKI,AFRINIC,ARIN,APNIC,LACNIC,RADB,RIPE")
	require.NoError(t, err)
	assert.Equal(t, []string{"192.0.2.0/24", "198.51.100.0/24"}, prefixes)

	prefixes, err = PrefixSet("RIPE::AS-SERVERFORGE", 6, "rr.ntt.net", 5, "")
	require.NoError(t, err)
	assert.Len(t, prefixes, 2)

	assert.Equal(t, []string{
		"-S RPKI,AFRINIC,ARIN,APNIC,LACNIC,RADB,RIPE -h rr.ntt.net -Ab4 AS-SERVERFORGE",
		"-h rr.ntt.net -Ab6 -S RIPE AS-SERVERFORGE",
	}, invocations(t, logFile))
}

func TestASMembersSourceArgs(t *testing.T) {
	logFile := fakeBGPQ4(t, `echo '{"NN": [112, 34553]}'`)

	members, err := ASMembers("RIPE::AS-SERVERFORGE", "rr.ntt.net", 5, "-S RADB,RIPE")
	require.NoError(t, err)
	assert.Equal(t, []uint32{112, 34553}, members)

	members, err = ASMembers("RIPE::AS-SERVERFORGE", "rr.ntt.net", 5, "")
	require.NoError(t, err)
	assert.Equal(t, []uint32{112, 34553}, members)

	assert.Equal(t, []string{
		"-S RADB,RIPE -h rr.ntt.net -tj AS-SERVERFORGE",
		"-h rr.ntt.net -tj -S RIPE AS-SERVERFORGE",
	}, invocations(t, logFile))
}

func TestParseBirdPrefixList(t *testing.T) {
	testCases := []struct {
		out         string
		expected    []string
		shouldError bool
	}{
		{"NN = [\n    192.31.196.0/24,\n    192.175.48.0/24\n];\n", []string{"192.31.196.0/24", "192.175.48.0/24"}, false},
		{"NN = [\n    2001:500:9c::/47{47,48},\n    2001:500:9e::/47\n];\n", []string{"2001:500:9c::/47{47,48}", "2001:500:9e::/47"}, false},
		{"NN = [ ];\n", nil, false}, // empty list
		{"NN = [\n\n];\n", nil, false},
		{"", nil, true},
		{"garbage", nil, true},
	}
	for _, tc := range testCases {
		out, err := parseBirdPrefixList(tc.out)
		if tc.shouldError {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, out)
	}
}
