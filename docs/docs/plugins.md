---
sidebar_position: 7
---

# Plugins

Pathvector can be extended with plugins. Plugins follow the `github.com/natesales/pathvector/pkg/plugin.Plugin` interface.

## Example

```bash
# Create a new plugin package
mkdir /tmp/example-plugin
cd /tmp/example-plugin
go mod init example.com/plugin
# Write your plugin...
cd ..
git clone https://github.com/natesales/pathvector
cd pathvector
echo 'helloworld:example.com/plugin' >> plugin.cfg # Add the plugin's name and package path to Pathvector's plugin config
./plugin-generate.py
go build
```

```go title="main.go"
package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/plugin"
)

type HelloWorld struct{}

var _ plugin.Plugin = (*Plugin)(nil)

func (g *HelloWorld) Description() string {
	return "An example plugin"
}

func (g *HelloWorld) Version() string {
	return "1.0.0"
}

func (g *HelloWorld) Command() *cobra.Command {
	return &cobra.Command{
		Use:   "hello",
		Short: "Show hello world",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello world from the plugin!")
		},
	}
}

func (g *HelloWorld) Modify(c *config.Config) error {
	c.Hostname = "hello-world.example.com"
	return nil
}

func main() {
	plugin.Register("helloworld", &HelloWorld{})
}
```
