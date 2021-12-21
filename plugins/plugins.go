package plugins

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/config"
)

var plugins = make(map[string]Plugin)

// Plugin defines an interface for plugins to implement
type Plugin interface {
	Description() string
	Execute(c *config.Config) error
}

// Register registers a plugin
func Register(name string, plugin Plugin) {
	plugins[name] = plugin
}

// All runs all plugins
func All(c *config.Config) error {
	for name, plugin := range plugins {
		log.Debugf("running plugin %s", name)
		if err := plugin.Execute(c); err != nil {
			return fmt.Errorf("[plugin %s]: %s", name, err)
		}
	}
	return nil
}

// Get returns the plugins map
func Get() map[string]Plugin {
	return plugins
}
