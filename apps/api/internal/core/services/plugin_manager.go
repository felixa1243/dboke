package services

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/hashicorp/go-plugin"

	"dboke-plugins/shared"
)

type PluginInstance struct {
	Client  *plugin.Client
	Feature shared.Feature
}

type PluginManager struct {
	plugins map[string]*PluginInstance
	mu      sync.RWMutex
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*PluginInstance),
	}
}

func (m *PluginManager) LoadPlugin(pluginID string) error {
	cwd, _ := os.Getwd()
	// from apps/api down to apps/plugins? Wait plugin handler used apps/plugins?
	// plugin_handler.go: pluginsDir := filepath.Join(cwd, "..", "plugins") -> d:\projects\dboke\apps\api\.. -> d:\projects\dboke\apps\plugins 
	// Wait, no. plugin_handler uses: cwd, _ := os.Getwd(). If ran from apps/api, cwd is apps/api. cwd, "..", "plugins" is apps/plugins.
	// Let's match plugin_handler.go logic.
	pluginDir := filepath.Join(cwd, "..", "plugins", pluginID)

	// Check if disabled
	if _, err := os.Stat(filepath.Join(pluginDir, ".disabled")); err == nil {
		return fmt.Errorf("plugin %s is disabled", pluginID)
	}

	execPath := filepath.Join(pluginDir, "backend.exe")
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		execPath = filepath.Join(pluginDir, "backend")
		if _, err := os.Stat(execPath); os.IsNotExist(err) {
			return fmt.Errorf("backend executable not found for plugin %s", pluginID)
		}
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.HandshakeConfig,
		Plugins:         shared.PluginMap,
		Cmd:             exec.Command(execPath),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC,
		},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to create plugin client: %v", err)
	}

	raw, err := rpcClient.Dispense("feature")
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense plugin: %v", err)
	}

	feature := raw.(shared.Feature)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Cleanup old if exists
	if old, ok := m.plugins[pluginID]; ok {
		old.Client.Kill()
	}

	m.plugins[pluginID] = &PluginInstance{
		Client:  client,
		Feature: feature,
	}

	slog.Info("Successfully loaded plugin", slog.String("pluginID", pluginID))
	return nil
}

func (m *PluginManager) GetFeature(pluginID string) (shared.Feature, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, ok := m.plugins[pluginID]
	if !ok {
		// Attempt to load it just in time
		m.mu.RUnlock()
		err := m.LoadPlugin(pluginID)
		m.mu.RLock()
		if err != nil {
			return nil, err
		}
		instance, ok = m.plugins[pluginID]
		if !ok {
			return nil, fmt.Errorf("plugin %s failed to load", pluginID)
		}
	}
	return instance.Feature, nil
}

func (m *PluginManager) UnloadPlugin(pluginID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if instance, ok := m.plugins[pluginID]; ok {
		instance.Client.Kill()
		delete(m.plugins, pluginID)
	}
}

func (m *PluginManager) UnloadAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, instance := range m.plugins {
		instance.Client.Kill()
	}
	m.plugins = make(map[string]*PluginInstance)
}
