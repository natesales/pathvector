package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestIntegrations(t *testing.T) {
	// Make temporary cache directory
	if err := os.Mkdir("/tmp/pathvector-test/", 0755); err != nil && !os.IsExist(err) {
		t.Error(err)
	}

	files, err := filepath.Glob("../tests/*.yml")
	if err != nil {
		t.Error(err)
	}
	for _, testFile := range files {
		args := []string{
			"--verbose",
			"--config", testFile,
			"daemon",
		}
		t.Logf("running integration with args %v", args)
		cmd := NewRootCommand()
		b := bytes.NewBufferString("")
		cmd.SetOut(b)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Error(err)
		}
	}
}
