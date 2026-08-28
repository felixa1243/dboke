package shared

import (
	"net/rpc"
	"github.com/hashicorp/go-plugin"
)

// Handshake is a common handshake that is shared by plugin and host.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DBOKE_PLUGIN",
	MagicCookieValue: "hello",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"feature": &FeaturePlugin{},
}

// Feature is the interface that we're exposing as a plugin.
// Both the Dboke main app and the plugins will use this interface.
type Feature interface {
	// BuildQuery takes a JSON payload of visual graph nodes (from the UI) and returns a SQL string
	BuildQuery(visualPayload string) (string, error)
	
	// GetFrontendComponent returns the URL or raw JS bundle for the React component
	GetFrontendComponent() (string, error)
}

// FeatureRPC handles communication FROM Dboke TO the plugin
type FeatureRPC struct{ client *rpc.Client }

func (g *FeatureRPC) BuildQuery(payload string) (string, error) {
	var resp string
	err := g.client.Call("Plugin.BuildQuery", payload, &resp)
	return resp, err
}

func (g *FeatureRPC) GetFrontendComponent() (string, error) {
	var resp string
	err := g.client.Call("Plugin.GetFrontendComponent", new(interface{}), &resp)
	return resp, err
}

// FeatureRPCServer handles communication FROM the plugin TO Dboke
type FeatureRPCServer struct {
	Impl Feature
}

func (s *FeatureRPCServer) BuildQuery(args string, resp *string) error {
	var err error
	*resp, err = s.Impl.BuildQuery(args)
	return err
}

func (s *FeatureRPCServer) GetFrontendComponent(args interface{}, resp *string) error {
	var err error
	*resp, err = s.Impl.GetFrontendComponent()
	return err
}

// FeaturePlugin is the boilerplate implementation of plugin.Plugin 
type FeaturePlugin struct {
	Impl Feature
}

func (p *FeaturePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &FeatureRPCServer{Impl: p.Impl}, nil
}

func (FeaturePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &FeatureRPC{client: c}, nil
}
