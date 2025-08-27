package pluginv2

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/corynth/corynth/pkg/plugin"
)

// Manager manages both built-in and gRPC plugins
type Manager struct {
	builtinPlugins map[string]plugin.Plugin
	grpcManager    *GRPCPluginManager
	mu             sync.RWMutex
}

// NewManager creates a new v2 plugin manager
func NewManager(localPath string, repositories []plugin.Repository) *Manager {
	// Ensure local plugins directory exists
	if err := os.MkdirAll(localPath, 0755); err != nil {
		fmt.Printf("Warning: failed to create plugin directory: %v\n", err)
	}
	
	return &Manager{
		builtinPlugins: make(map[string]plugin.Plugin),
		grpcManager:    NewGRPCPluginManager(localPath, repositories),
	}
}

// Initialize loads built-in plugins
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Load only the shell plugin as built-in
	m.builtinPlugins["shell"] = &plugin.ShellPlugin{}
	
	return nil
}

// Load loads a plugin by name, trying built-in first, then gRPC
func (m *Manager) Load(name string) (plugin.Plugin, error) {
	// Check built-in plugins first
	m.mu.RLock()
	if builtinPlugin, exists := m.builtinPlugins[name]; exists {
		m.mu.RUnlock()
		return builtinPlugin, nil
	}
	m.mu.RUnlock()
	
	// Try to load as gRPC plugin
	return m.grpcManager.Load(name)
}

// List returns metadata for all available plugins
func (m *Manager) List() []plugin.Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var metadata []plugin.Metadata
	
	// Add built-in plugins
	for _, builtinPlugin := range m.builtinPlugins {
		metadata = append(metadata, builtinPlugin.Metadata())
	}
	
	// Add gRPC plugins
	grpcMetadata := m.grpcManager.List()
	metadata = append(metadata, grpcMetadata...)
	
	return metadata
}

// Search searches for plugins by criteria
func (m *Manager) Search(query string, tags []string) []plugin.Metadata {
	allPlugins := m.List()
	var results []plugin.Metadata
	
	for _, meta := range allPlugins {
		// Simple search implementation
		if query != "" {
			nameMatch := contains(meta.Name, query)
			descMatch := contains(meta.Description, query)
			if !nameMatch && !descMatch {
				continue
			}
		}
		
		// Tag filtering
		if len(tags) > 0 {
			hasTag := false
			for _, searchTag := range tags {
				for _, pluginTag := range meta.Tags {
					if contains(pluginTag, searchTag) {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		
		results = append(results, meta)
	}
	
	return results
}

// Install installs a plugin from repositories
func (m *Manager) Install(name string) error {
	return m.grpcManager.installPlugin(name)
}

// Close closes all plugin connections
func (m *Manager) Close() error {
	return m.grpcManager.Close()
}

// IsBuiltin checks if a plugin is built-in
func (m *Manager) IsBuiltin(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.builtinPlugins[name]
	return exists
}

// ListLocal lists locally installed gRPC plugins
func (m *Manager) ListLocal() []string {
	var plugins []string
	
	// List executable files in plugin directory
	entries, err := os.ReadDir(m.grpcManager.localPath)
	if err != nil {
		return plugins
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasPrefix(name, "corynth-plugin-") {
			// Extract plugin name
			pluginName := strings.TrimPrefix(name, "corynth-plugin-")
			pluginName = strings.TrimSuffix(pluginName, ".exe") // Windows
			plugins = append(plugins, pluginName)
		}
	}
	
	return plugins
}

// Helper function for case-insensitive contains
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (len(substr) == 0 || 
		    strings.Contains(strings.ToLower(s), strings.ToLower(substr)))
}