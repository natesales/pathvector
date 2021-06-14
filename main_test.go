package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrations(t *testing.T) {
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
	files, err := filepath.Glob("tests/*.yml")
	if err != nil {
		t.Error(err)
	}
	for _, testFile := range files {
		t.Logf("running integration with args %v", args)
		run(append(args, []string{
			"--config", testFile,
		}...))
	}
}
