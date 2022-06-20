package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/plugin"
)

type HelloWorld struct{}

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
