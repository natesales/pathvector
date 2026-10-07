package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBirdshCLIAlias(t *testing.T) {
	c, _, err := rootCmd.Find([]string{"cli"})
	assert.Nil(t, err)
	assert.Equal(t, birdshCmd, c)
}

func TestRemoveString(t *testing.T) {
	assert.Equal(t, []string{"a", "c"}, removeString([]string{"a", "cli", "c"}, "cli"))
	assert.Nil(t, removeString([]string{"cli"}, "cli"))
}
