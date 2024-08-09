package cmd

import (
	"os"
	"testing"
)

func TestDumpTable(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	mkTmpCache(t)

	args := []string{
		"dump",
		"--verbose",
		"--dry-run",
	}

	withGenerateConfigs(t, func(testFile string) {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running dump integration with args %v", args)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}
	})

	w.Close()
	os.Stdout = old
}

func TestDumpYAML(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	mkTmpCache(t)

	args := []string{
		"dump",
		"--yaml",
		"--verbose",
		"--dry-run",
	}
	withGenerateConfigs(t, func(testFile string) {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running dump integration with args %v", args)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}
	})

	w.Close()
	os.Stdout = old
}
