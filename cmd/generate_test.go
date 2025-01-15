package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/util/log"
)

func TestGenerate(t *testing.T) {
	mkTmpCache(t)

	baseArgs := []string{
		"generate",
		"--verbose",
		"--dry-run",
	}

	withGenerateConfigs(t, func(testFile string) {
		args := append(baseArgs, []string{
			"--config", testFile,
		}...)
		t.Logf("running generate integration with args %v", args)
		rootCmd.SetArgs(args)
		_ = log.Capture()
		defer log.ResetCapture()
		assert.Nil(t, rootCmd.Execute())
	})
}
