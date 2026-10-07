package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceFlag(t *testing.T) {
	flags := rootCmd.PersistentFlags()
	// Flag values persist across rootCmd executions in other tests, so start from a clean state
	reset := func() {
		_ = flags.Set("trace", "false")
		_ = flags.Set("verbose", "false")
	}
	reset()
	t.Cleanup(reset)

	assert.Nil(t, flags.Parse([]string{"-t"}))
	assert.True(t, trace, "--trace should set trace")
	assert.False(t, verbose, "--trace should not set verbose")

	assert.Nil(t, flags.Parse([]string{"-v"}))
	assert.True(t, verbose)
}
