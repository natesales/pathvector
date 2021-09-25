package cmd

import (
	"fmt"
	"testing"
)

func TestMatch(t *testing.T) {
	baseArgs := []string{
		"match",
		"--verbose",
	}

	testCases := []struct {
		asnA uint
		asnB uint
	}{
		{34553, 13335},
		{54113, 13335},
	}
	for _, tc := range testCases {
		rootCmd.SetArgs(append(baseArgs, []string{
			"-l", fmt.Sprintf("%d", tc.asnA),
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
}
