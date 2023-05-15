package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	//nolint:golint,gosec
	err := os.WriteFile("/etc/bird/protocols.json", []byte(`{"EXAMPLE_AS65510_v4":{"Name":"Example","Tags":null},"EXAMPLE_AS65510_v6":{"Name":"Example","Tags":null}}`), 0644)
	assert.Nil(t, err)

	rootCmd.SetArgs([]string{
		"status",
		"-c", "../tests/generate-simple.yml",
	})
	if err := rootCmd.Execute(); err != nil {
		t.Error(err)
	}
}
