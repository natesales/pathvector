package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/util/log"
)

func TestDumpTable(t *testing.T) {
	mkTmpCache(t)

	args := []string{
		"dump",
		"--verbose",
		"--dry-run",
	}

	withGenerateConfigs(t, func(testFile string) {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running dump integration with args %v", args)

		out := log.Capture()
		defer log.ResetCapture()

		rootCmd.SetArgs(args)
		assert.Nil(t, rootCmd.Execute())
		assert.Contains(t, out.String(), "PREPENDS")
		assert.Contains(t, out.String(), "NAME")
		assert.Contains(t, out.String(), "ASN")
	})
}

func TestDumpYAML(t *testing.T) {
	mkTmpCache(t)

	args := []string{
		"dump",
		"--yaml",
		"--verbose",
		"--dry-run",
	}
	withGenerateConfigs(t, func(testFile string) {
		args = append(args, []string{
			"--config", testFile,
		}...)
		t.Logf("running dump integration with args %v", args)

		out := log.Capture()
		defer log.ResetCapture()

		rootCmd.SetArgs(args)
		assert.Nil(t, rootCmd.Execute())
		assert.Contains(t, out.String(), "global-config: \"\"")
	})
}
