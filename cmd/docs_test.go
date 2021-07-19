package cmd

import "testing"

func TestDocs(t *testing.T) {
	rootCmd.SetArgs([]string{
		"docs",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Error(err)
	}
}
