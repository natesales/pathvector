package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// withGenerateConfigs runs a callback on all generate config files
func withGenerateConfigs(t *testing.T, callback func(string)) {
	files, err := filepath.Glob("../tests/generate-*.yml")
	assert.Nil(t, err)
	assert.Greater(t, len(files), 1)

	for _, testFile := range files {
		t.Run(testFile, func(t *testing.T) {
			callback(testFile)
		})
	}
}

// mkTmpCache makes the test-cache directory
func mkTmpCache(t *testing.T) {
	if err := os.Mkdir("/tmp/test-cache", 0755); !os.IsExist(err) {
		assert.Nil(t, err)
	}
}
