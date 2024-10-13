package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/util/log"
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
		{34553, 112},
		{112, 44977},
	}
	for _, tc := range testCases {
		out := log.Capture()
		rootCmd.SetArgs(append(baseArgs, []string{
			"-l", fmt.Sprintf("%d", tc.asnA),
			"-c", "../tests/generate-simple.yml",
			fmt.Sprintf("%d", tc.asnB),
		}...))
		assert.Nil(t, rootCmd.Execute())
		assert.Contains(t, out.String(), "Finished loading config")
		log.ResetCapture()

		// Local ASN from config file
		out = log.Capture()
		rootCmd.SetArgs(append(baseArgs, []string{
			"-c", "../tests/generate-simple.yml",
			"-l", "0",
			fmt.Sprintf("%d", tc.asnB),
		}...))
		assert.Nil(t, rootCmd.Execute())
		assert.Contains(t, out.String(), "Finished loading config")
		log.ResetCapture()
	}
}
