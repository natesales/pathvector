package cmd

import "testing"

func TestVersion(t *testing.T) {
	rootCmd.SetArgs([]string{
		"version",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Error(err)
	}
}
