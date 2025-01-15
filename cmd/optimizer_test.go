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

			baseArgs := []string{
				"--verbose",
				"--config", testFile,
			}

			// Run pathvector to generate config first, so there is a config to modify
			rootCmd.SetArgs(append(baseArgs, "generate"))
			t.Log("Running pre-optimizer generate")
			assert.Nil(t, rootCmd.Execute())

			rootCmd.SetArgs(append(baseArgs, "optimizer"))
			t.Log("Running probe integration")
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
