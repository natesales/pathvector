package optimizer

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util"
)

func writeTestBIRDDir(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(path.Join(dir, "bird.conf"), []byte("router id 192.0.2.1;\nprotocol device {}\ninclude \"AS*.conf\";\n"), 0600); err != nil {
		t.Fatal(err)
	}
	peerFile := path.Join(dir, "AS65510_EXAMPLE.conf")
	if err := os.WriteFile(peerFile, []byte("filter example_import {\n  bgp_local_pref = 100; # pathvector:localpref\n  accept;\n}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return dir, peerFile
}

func TestModifyPrefDryRun(t *testing.T) {
	dir, peerFile := writeTestBIRDDir(t)
	before, _ := os.ReadFile(peerFile)
	peers := map[string]*config.Peer{"Example": {LocalPref: util.Ptr(100), OptimizeInbound: util.Ptr(true)}}
	modifyPref("65510"+Delimiter+"Example", peers, 20, dir, "", "bird", true, true)
	after, _ := os.ReadFile(peerFile)
	if string(before) != string(after) {
		t.Errorf("dry run modified the peer file:\n%s", after)
	}
}

func TestModifyPref(t *testing.T) {
	birdBin, err := exec.LookPath("bird")
	if err != nil {
		t.Skip("bird binary not found")
	}
	dir, peerFile := writeTestBIRDDir(t)
	peers := map[string]*config.Peer{"Example": {LocalPref: util.Ptr(100), OptimizeInbound: util.Ptr(true)}}

	// Run twice: the optimizer must keep working on the active config in the BIRD directory (natesales/pathvector#235)
	for i := 0; i < 2; i++ {
		modifyPref("65510"+Delimiter+"Example", peers, 20, dir, "", birdBin, true, false)
		contents, err := os.ReadFile(peerFile)
		if err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
		if !strings.Contains(string(contents), "bgp_local_pref = 80; # pathvector:localpref") {
			t.Errorf("run %d: local pref not lowered:\n%s", i, contents)
		}
	}
	if _, err := os.Stat(path.Join(dir, "bird.conf")); err != nil {
		t.Errorf("bird.conf missing after optimization: %v", err)
	}
}

func TestModifyPrefNotOptimized(t *testing.T) {
	dir, peerFile := writeTestBIRDDir(t)
	before, _ := os.ReadFile(peerFile)
	peers := map[string]*config.Peer{"Example": {LocalPref: util.Ptr(100), OptimizeInbound: util.Ptr(false)}}
	modifyPref("65510"+Delimiter+"Example", peers, 20, dir, "", "bird", true, false)
	after, _ := os.ReadFile(peerFile)
	if string(before) != string(after) {
		t.Errorf("peer without optimize-inbound was modified:\n%s", after)
	}
}

func TestModifyPrefClampsAtZero(t *testing.T) {
	dir, peerFile := writeTestBIRDDir(t)
	peers := map[string]*config.Peer{"Example": {LocalPref: util.Ptr(10), OptimizeInbound: util.Ptr(true)}}
	birdBin, err := exec.LookPath("bird")
	if err != nil {
		t.Skip("bird binary not found")
	}
	// A modifier larger than the local pref must not wrap around to a huge unsigned value
	modifyPref("65510"+Delimiter+"Example", peers, 20, dir, "", birdBin, true, false)
	contents, _ := os.ReadFile(peerFile)
	if !strings.Contains(string(contents), "bgp_local_pref = 0; # pathvector:localpref") {
		t.Errorf("local pref not clamped to 0:\n%s", contents)
	}
}
