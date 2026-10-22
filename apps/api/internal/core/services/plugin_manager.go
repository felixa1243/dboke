package services

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// pluginIDPattern restricts plugin IDs to a single safe path segment, since IDs
// are joined into filesystem paths for install, toggle, delete and execution.
var pluginIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

// IsValidPluginID reports whether id is safe to use as a plugin directory name.
func IsValidPluginID(id string) bool {
	return pluginIDPattern.MatchString(id)
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]*PluginInstance),
	}
}

func (m *PluginManager) LoadPlugin(pluginID string) error {
	if !IsValidPluginID(pluginID) {
		return fmt.Errorf("invalid plugin id: %q", pluginID)
	}

	cwd, _ := os.Getwd()
	// Plugins live in apps/plugins, matching plugin_handler.go
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
