package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDumpTable(t *testing.T) {
	// Make temporary cache directory
	if err := os.Mkdir("test-cache", 0755); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	args := []string{
		"dump",
		"--verbose",
		"--dry-run",
	}
	files, err := filepath.Glob("../tests/generate-*.yml")
	if err != nil {
		t.Error(err)
	}
	if len(files) < 1 {
		t.Fatal("No test files found")
	}
	for _, testFile := range files {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running dump integration with args %v", args)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}
	}
}

func TestDumpYAML(t *testing.T) {
	// Make temporary cache directory
	if err := os.Mkdir("test-cache", 0755); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	args := []string{
		"dump",
		"--yaml",
		"--verbose",
		"--dry-run",
	}
	files, err := filepath.Glob("../tests/generate-*.yml")
	if err != nil {
		t.Error(err)
	}
	if len(files) < 1 {
		t.Fatal("No test files found")
	}
	for _, testFile := range files {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running dump integration with args %v", args)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}
	}
}
