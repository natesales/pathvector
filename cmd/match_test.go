package cmd

import (
	"fmt"
	"os"
	"testing"
)

func TestMatch(t *testing.T) {
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	baseArgs := []string{
		"match",
		"--verbose",
	}

	testCases := []struct {
		asnA uint
		asnB uint
	}{
		{34553, 112},
		{112, 44977},
	}
	for _, tc := range testCases {
		rootCmd.SetArgs(append(baseArgs, []string{
			"-l", fmt.Sprintf("%d", tc.asnA),
			"-c", "../tests/generate-simple.yml",
			fmt.Sprintf("%d", tc.asnB),
		}...))
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}

		// Local ASN from config file
		rootCmd.SetArgs(append(baseArgs, []string{
			"-c", "../tests/generate-simple.yml",
			"-l", "0",
			fmt.Sprintf("%d", tc.asnB),
		}...))
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}
	}
	w.Close()
	os.Stdout = old
}
