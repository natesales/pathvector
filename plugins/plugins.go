package plugins

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/natesales/pathvector/pkg/config"
)

var hsConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PATHVECTOR",
	MagicCookieValue: "core",
}

var all = map[string]Plug{}

var pluginMap = map[string]plugin.Plugin{
	"plugin": &PlugPlugin{},
}

// Load loads plugins from a directory
func Load(dir string) error {
	files, err := filepath.Glob(path.Join(dir, "/*.pvec"))
	if err != nil {
		return err
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Info,
		Output: hclog.DefaultOutput,
	})

	for _, file := range files {
		fileNoPrefix := strings.TrimPrefix(file, path.Join(dir, "/")+"/")

		client := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: hsConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(file),
			Logger:          logger,
		})

		rpcClient, err := client.Client()
		if err != nil {
			return fmt.Errorf("connecting to %s: %s", fileNoPrefix, err)
		}

		raw, err := rpcClient.Dispense("plugin")
		if err != nil {
			return fmt.Errorf("dispensing from %s: %s", fileNoPrefix, err)
		}

		all[fileNoPrefix] = raw.(Plug)
	}
	return nil
}

// Register registers a plugin with the core
func Register(p Plug) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: hsConfig,
		Plugins: map[string]plugin.Plugin{
			"plugin": &PlugPlugin{Impl: p},
		},
	})
}

// ApplyConfig runs the config through each plugin
func ApplyConfig(c *config.Config) error {
	for name, p := range all {
		if err := p.Modify(c); err != nil {
			return fmt.Errorf("[plugin %s]: %s", name, err)
		}
	}
	return nil
}

// All returns all plugins
func All() map[string]Plug {
	return all
}
