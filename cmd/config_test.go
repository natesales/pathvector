package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/util/log"
)

func TestConfig(t *testing.T) {
	rootCmd.SetArgs([]string{
		"config",
		"-c", "../tests/generate-complex.yml",
	})

	out := log.Capture()
	defer log.ResetCapture()
	assert.Nil(t, rootCmd.Execute())
	assert.Contains(t, out.String(), "# Pathvector devel")
	assert.Contains(t, out.String(), "asn: 65530")
}

func TestSanitizeConfig(t *testing.T) {
	rootCmd.SetArgs([]string{
		"config",
		"-c", "../tests/generate-complex.yml",
		"--sanitize",
	})

	out := log.Capture()
	defer log.ResetCapture()
	assert.Nil(t, rootCmd.Execute())
	assert.Contains(t, out.String(), "# Pathvector devel")
	assert.Contains(t, out.String(), "asn: 65530")
	assert.Contains(t, out.String(), "- 2001:db8:")
}
