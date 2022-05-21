package cmd

import (
	"os"
	"testing"
)

func TestDocs(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	rootCmd.SetArgs([]string{
		"docs",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Error(err)
	}
	w.Close()
	os.Stdout = old
}
