package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/util/log"
)

func TestDocs(t *testing.T) {
	rootCmd.SetArgs([]string{
		"docs",
	})
	out := log.Capture()
	defer log.ResetCapture()
	assert.Nil(t, rootCmd.Execute())
	assert.Contains(t, out.String(), "# Configuration\n")
}
