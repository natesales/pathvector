package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptimizer(t *testing.T) {
	// cobra keeps flag values between Execute calls in the same process, so set --dry-run and --no-configure
	// explicitly instead of inheriting them from earlier tests (e.g. TestGenerate runs with --dry-run)
	args := []string{
		"--verbose",
		"--dry-run=false",
		"--no-configure",
	}
	files, err := filepath.Glob("../tests/probe-*.yml")
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, 1, len(files))

	// The probe configs use test-bird/ as the BIRD directory so the active config isn't written to /etc/bird
	assert.Nil(t, os.MkdirAll("test-bird", 0o750))
	defer os.RemoveAll("test-bird")

	for _, testFile := range files {
		// Run pathvector to generate config first, so there is an active config in the BIRD directory to modify
		rootCmd.SetArgs(append(args, []string{
			"generate",
			"--config", testFile,
		}...))
		t.Logf("Running pre-optimizer generate: %v", args)
		assert.Nil(t, rootCmd.Execute())

		optimizerArgs := append(args, []string{
			"optimizer",
			"--config", testFile,
		}...)
		t.Logf("running probe integration with args %v", optimizerArgs)
		rootCmd.SetArgs(optimizerArgs)
		assert.Nil(t, rootCmd.Execute())

		// The optimizer modifies the active config in the BIRD directory (natesales/pathvector#235)
		checkFile, err := os.ReadFile("test-bird/AS65510_EXAMPLE.conf")
		assert.Nil(t, err)
		if !strings.Contains(string(checkFile), "bgp_local_pref = 80; # pathvector:localpref") {
			t.Errorf("expected bgp_local_pref = 80 but not found in file")
		}
	}
}
