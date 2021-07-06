package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMainGenerate(t *testing.T) {
	// Make temporary cache directory
	if err := os.Mkdir("test-cache", 0755); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	args := []string{
		"--verbose",
		"--dry-run",
		"--cache-directory", "test-cache",
		"--web-ui-file", "test-cache/ui.html",
	}
	files, err := filepath.Glob("tests/generate-*.yml")
	if err != nil {
		t.Error(err)
	}
	for _, testFile := range files {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running generate integration with args %v", args)
		rootCmd.SetArgs(args)
		rootCmd.Execute()
	}
}

func TestMainOptimizer(t *testing.T) {
	args := []string{
		"--verbose",
		"--dry-run",
		"--cache-directory", "test-cache",
	}
	files, err := filepath.Glob("tests/probe-*.yml")
	if err != nil {
		t.Error(err)
	}
	for _, testFile := range files {
		// Run pathvector to generate config first, so there is a config to modify
		rootCmd.SetArgs(append(args, []string{
			"--config", testFile,
		}...))
		rootCmd.Execute()

		// Disable the optimizer after it's ran for a bit
		go func() {
			time.Sleep(5 * time.Second)
			t.Log("disabling optimizer")
			globalOptimizer.Disable = true
		}()

		args = append(args, []string{
			"probe",
			"--config", testFile,
		}...)
		t.Logf("running probe integration with args %v", args)
		rootCmd.SetArgs(args)
		rootCmd.Execute()

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
