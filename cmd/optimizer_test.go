package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/util"
)

func TestOptimizer(t *testing.T) {
	args := []string{
		"--verbose",
	}
	files, err := filepath.Glob("../tests/probe-*.yml")
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, 1, len(files))

	for _, testFile := range files {
		t.Run(testFile, func(t *testing.T) {
			for _, dir := range []string{"/tmp/test-cache", "/tmp/bird-conf"} {
				if err := util.RemoveFileGlob(dir + "/*"); err != nil {
					t.Errorf("failed to remove %s: %v", dir, err)
				}
			}

			// Run pathvector to generate config first, so there is a config to modify
			rootCmd.SetArgs(append(args, []string{
				"generate",
				"--config", testFile,
			}...))
			t.Logf("Running pre-optimizer generate: %v", args)
			assert.Nil(t, rootCmd.Execute())

			args = append(args, []string{
				"optimizer",
				"--config", testFile,
			}...)
			t.Logf("running probe integration with args %v", args)
			rootCmd.SetArgs(args)
			assert.Nil(t, rootCmd.Execute())

			// Check if local pref is lowered
			checkFile, err := os.ReadFile("/tmp/bird-conf/AS65510_EXAMPLE.conf")
			assert.Nil(t, err)
			if !strings.Contains(string(checkFile), "bgp_local_pref = 80; # pathvector:localpref") {
				t.Errorf("expected bgp_local_pref = 80 but not found in file")
			}
		})
	}
}
