package cmd

import "testing"

func TestStatus(t *testing.T) {
	rootCmd.SetArgs([]string{
		"status",
		"-c", "../tests/generate-simple.yml",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Error(err)
	}
}
