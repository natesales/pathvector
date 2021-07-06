package main

import (
	"os"
	"path/filepath"
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

func TestMainProbe(t *testing.T) {
	args := []string{
		"probe",
		"--verbose",
	}
	files, err := filepath.Glob("tests/probe-*.yml")
	if err != nil {
		t.Error(err)
	}
	for _, testFile := range files {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running probe integration with args %v", args)
		rootCmd.SetArgs(args)
		rootCmd.Execute()
	}
}
