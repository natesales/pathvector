package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	mkTmpCache(t)

	args := []string{
		"generate",
		"--verbose",
		"--dry-run",
	}

	withGenerateConfigs(t, func(testFile string) {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running generate integration with args %v", args)
		rootCmd.SetArgs(args)
		assert.Nil(t, rootCmd.Execute())
	})
}
