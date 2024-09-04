package plugin

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/util/log"
)

var plugins = make(map[string]Plugin)

// Plugin defines an interface for plugins to implement
type Plugin interface {
	Version() string
	Description() string
	Command() *cobra.Command
	Modify(config *config.Config) error
}

// Register registers a plugin
func Register(name string, plugin Plugin) {
	plugins[name] = plugin
}

// ModifyAll runs all plugins
func ModifyAll(c *config.Config) error {
	for name, plugin := range plugins {
		log.Debugf("running plugin %s", name)
		if err := plugin.Modify(c); err != nil {
			return fmt.Errorf("[plugin %s]: %s", name, err)
		}
	}
	return nil
}

// Get returns the plugins map
func Get() map[string]Plugin {
	return plugins
}
