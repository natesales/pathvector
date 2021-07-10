package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

		args = append(args, []string{
			"optimizer",
			"--udp",
			"--exit-on-cache-full",
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

func TestMainDumpTable(t *testing.T) {
	// Make temporary cache directory
	if err := os.Mkdir("test-cache", 0755); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	args := []string{
		"dump",
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
		t.Logf("running dump integration with args %v", args)
		rootCmd.SetArgs(args)
		rootCmd.Execute()
	}
}

func TestMainDumpYAML(t *testing.T) {
	// Make temporary cache directory
	if err := os.Mkdir("test-cache", 0755); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	args := []string{
		"dump",
		"--yaml",
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
		t.Logf("running dump integration with args %v", args)
		rootCmd.SetArgs(args)
		rootCmd.Execute()
	}
}
