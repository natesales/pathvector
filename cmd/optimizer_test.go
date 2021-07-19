package cmd

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptimizer(t *testing.T) {
	args := []string{
		"--verbose",
		"--dry-run",
	}
	files, err := filepath.Glob("../tests/probe-*.yml")
	if err != nil {
		t.Error(err)
	}
	if len(files) < 1 {
		t.Fatal("No test files found")
	}
	for _, testFile := range files {
		// Run pathvector to generate config first, so there is a config to modify
		rootCmd.SetArgs(append(args, []string{
			"--config", testFile,
		}...))
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}

		args = append(args, []string{
			"optimizer",
			"--config", testFile,
		}...)
		t.Logf("running probe integration with args %v", args)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}

		// Check if local pref is lowered
		checkFile, err := ioutil.ReadFile("test-cache/AS65510_EXAMPLE.conf")
		if err != nil {
			t.Error(err)
		}
		if !strings.Contains(string(checkFile), "bgp_local_pref = 80; # pathvector:localpref") {
			t.Errorf("expected bgp_local_pref = 80 but not found in file")
		}
	}
}
