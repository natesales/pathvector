package plugins

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/config"
)

// Plug is the interface that we're exposing as a plugin
type Plug interface {
	Description() string
	Version() string
	Command() *cobra.Command
	Modify(config *config.Config) error
}

// PlugRPC is an implementation that talks over RPC
type PlugRPC struct{ client *rpc.Client }

func (g *PlugRPC) Description() string {
	var resp string
	err := g.client.Call("Plugin.Description", new(interface{}), &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

func (g *PlugRPC) Version() string {
	var resp string
	err := g.client.Call("Plugin.Version", new(interface{}), &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

func (g *PlugRPC) Command() *cobra.Command {
	var resp *cobra.Command
	err := g.client.Call("Plugin.Command", new(interface{}), &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

func (g *PlugRPC) Modify(c *config.Config) error {
	var resp error
	err := g.client.Call("Plugin.Modify", c, &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

// PlugRPCServer is the RPC server that PlugRPC talks to, conforming to the requirements of net/rpc
type PlugRPCServer struct {
	Impl Plug
}

func (s *PlugRPCServer) Description(args interface{}, resp *string) error {
	*resp = s.Impl.Description()
	return nil
}

func (s *PlugRPCServer) Version(args interface{}, resp *string) error {
	*resp = s.Impl.Version()
	return nil
}

func (s *PlugRPCServer) Command(args interface{}, resp **cobra.Command) error {
	*resp = s.Impl.Command()
	return nil
}

func (s *PlugRPCServer) Modify(c *config.Config, resp *error) error {
	*resp = s.Impl.Modify(c)
	return nil
}

// PlugPlugin is the implementation of plugin.Plugin
type PlugPlugin struct {
	Impl Plug
}

func (p *PlugPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PlugRPCServer{Impl: p.Impl}, nil
}

func (PlugPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PlugRPC{client: c}, nil
}
