package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Registry represents the plugin registry
type Registry struct {
	Version    string              `json:"version"`
	Updated    string              `json:"updated"`
	Repository string              `json:"repository"`
	Plugins    []RegistryPlugin    `json:"plugins"`
	Categories map[string][]string `json:"categories"`
	Featured   []string            `json:"featured"`
	New        []string            `json:"new"`
	Popular    []string            `json:"popular"`
}

// RegistryPlugin represents a plugin in the registry
type RegistryPlugin struct {
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	Description  string             `json:"description"`
	Author       string             `json:"author"`
	Size         string             `json:"size"`
	Format       string             `json:"format"`
	Filename     string             `json:"filename"`
	Tags         []string           `json:"tags"`
	Actions      []RegistryAction   `json:"actions"`
	Requirements RegistryRequirements `json:"requirements"`
}

// RegistryAction represents a plugin action in the registry
type RegistryAction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// RegistryRequirements represents plugin requirements
type RegistryRequirements struct {
	Corynth string   `json:"corynth"`
	OS      []string `json:"os"`
	Arch    []string `json:"arch"`
}

// FetchRegistry fetches the plugin registry from a repository
func (m *Manager) FetchRegistry(repo Repository) (*Registry, error) {
	// Try multiple registry locations
	registryURLs := []string{
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/registry.json", 
			getRepoOwnerAndName(repo.URL), repo.Branch),
		fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/plugins/registry.json",
			getRepoOwnerAndName(repo.URL), repo.Branch),
	}

	var lastErr error
	for _, url := range registryURLs {
		registry, err := fetchRegistryFromURL(url)
		if err == nil {
			return registry, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("failed to fetch registry: %w", lastErr)
}

// FetchAllRegistries fetches registries from all configured repositories
func (m *Manager) FetchAllRegistries() ([]*Registry, error) {
	var registries []*Registry
	
	for _, repo := range m.repositories {
		registry, err := m.FetchRegistry(repo)
		if err != nil {
			// Log warning but continue with other repositories
			fmt.Printf("Warning: Failed to fetch registry from %s: %v\n", repo.Name, err)
			continue
		}
		registries = append(registries, registry)
	}
	
	if len(registries) == 0 {
		return nil, fmt.Errorf("no registries available")
	}
	
	return registries, nil
}

// SearchRegistry searches for plugins in the registry
func (m *Manager) SearchRegistry(query string, tags []string) ([]RegistryPlugin, error) {
	registries, err := m.FetchAllRegistries()
	if err != nil {
		return nil, err
	}

	var results []RegistryPlugin
	query = strings.ToLower(query)
	
	for _, registry := range registries {
		for _, plugin := range registry.Plugins {
			// Check name and description match
			if query != "" {
				nameMatch := strings.Contains(strings.ToLower(plugin.Name), query)
				descMatch := strings.Contains(strings.ToLower(plugin.Description), query)
				if !nameMatch && !descMatch {
					continue
				}
			}
			
			// Check tag match
			if len(tags) > 0 {
				hasTag := false
				for _, searchTag := range tags {
					for _, pluginTag := range plugin.Tags {
						if strings.EqualFold(searchTag, pluginTag) {
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
			
			results = append(results, plugin)
		}
	}
	
	return results, nil
}

// ListAvailable lists all available plugins from registries
func (m *Manager) ListAvailable() ([]RegistryPlugin, error) {
	registries, err := m.FetchAllRegistries()
	if err != nil {
		return nil, err
	}

	var plugins []RegistryPlugin
	seen := make(map[string]bool)
	
	for _, registry := range registries {
		for _, plugin := range registry.Plugins {
			// Deduplicate by name
			if !seen[plugin.Name] {
				plugins = append(plugins, plugin)
				seen[plugin.Name] = true
			}
		}
	}
	
	return plugins, nil
}

// GetPluginInfo gets detailed information about a plugin from the registry
func (m *Manager) GetPluginInfo(name string) (*RegistryPlugin, error) {
	registries, err := m.FetchAllRegistries()
	if err != nil {
		return nil, err
	}

	for _, registry := range registries {
		for _, plugin := range registry.Plugins {
			if plugin.Name == name {
				return &plugin, nil
			}
		}
	}
	
	return nil, fmt.Errorf("plugin '%s' not found in registry", name)
}

// GetCategories returns all plugin categories
func (m *Manager) GetCategories() (map[string][]string, error) {
	registries, err := m.FetchAllRegistries()
	if err != nil {
		return nil, err
	}

	categories := make(map[string][]string)
	
	for _, registry := range registries {
		for category, plugins := range registry.Categories {
			if existing, ok := categories[category]; ok {
				// Merge and deduplicate
				seen := make(map[string]bool)
				for _, p := range existing {
					seen[p] = true
				}
				for _, p := range plugins {
					if !seen[p] {
						existing = append(existing, p)
					}
				}
				categories[category] = existing
			} else {
				categories[category] = plugins
			}
		}
	}
	
	return categories, nil
}

// GetFeaturedPlugins returns featured plugins
func (m *Manager) GetFeaturedPlugins() ([]string, error) {
	registries, err := m.FetchAllRegistries()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var featured []string
	
	for _, registry := range registries {
		for _, plugin := range registry.Featured {
			if !seen[plugin] {
				featured = append(featured, plugin)
				seen[plugin] = true
			}
		}
	}
	
	return featured, nil
}

// CacheRegistry caches the registry locally
func (m *Manager) CacheRegistry(registry *Registry) error {
	cacheDir := filepath.Join(m.cachePath, "registry")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	
	cacheFile := filepath.Join(cacheDir, "registry.json")
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(cacheFile, data, 0644)
}

// LoadCachedRegistry loads the cached registry
func (m *Manager) LoadCachedRegistry() (*Registry, error) {
	cacheFile := filepath.Join(m.cachePath, "registry", "registry.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}
	
	var registry Registry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, err
	}
	
	// Check if cache is stale (older than 24 hours)
	info, err := os.Stat(cacheFile)
	if err == nil && time.Since(info.ModTime()) > 24*time.Hour {
		return nil, fmt.Errorf("cache is stale")
	}
	
	return &registry, nil
}

// Helper functions

func fetchRegistryFromURL(url string) (*Registry, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry not found: %s", resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var registry Registry
	if err := json.Unmarshal(body, &registry); err != nil {
		return nil, err
	}
	
	return &registry, nil
}

func getRepoOwnerAndName(repoURL string) string {
	// Extract owner/repo from GitHub URL
	// https://github.com/corynth/corynthplugins -> corynth/corynthplugins
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) >= 2 {
		return fmt.Sprintf("%s/%s", parts[len(parts)-2], parts[len(parts)-1])
	}
	return ""
}