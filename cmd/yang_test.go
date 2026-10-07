package cmd

import (
	"os"
	"testing"
)

func TestYang(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	rootCmd.SetArgs([]string{
		"yang",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Error(err)
	}
	w.Close()
	os.Stdout = old
}
