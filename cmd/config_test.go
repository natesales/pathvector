package cmd

import (
	"testing"
)

func TestConfig(t *testing.T) {
	rootCmd.SetArgs([]string{
		"config",
		"-c", "../tests/generate-complex.yml",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Error(err)
	}
}

func TestSanitizeConfig(t *testing.T) {
	rootCmd.SetArgs([]string{
		"config",
		"-c", "../tests/generate-complex.yml",
		"--sanitize",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Error(err)
	}
}
