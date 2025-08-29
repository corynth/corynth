package plugin

import (
	"crypto/sha256"
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
	Security     SecurityInfo       `json:"security"`
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

// SecurityInfo represents plugin security information
type SecurityInfo struct {
	TrustLevel     string    `json:"trust_level"`     // official, verified, community
	SignatureURL   string    `json:"signature_url"`   // URL to signature file
	ChecksumURL    string    `json:"checksum_url"`    // URL to checksum file
	SHA256         string    `json:"sha256"`          // SHA256 hash
	ScannedAt      string    `json:"scanned_at"`      // Last security scan date
	ScanResults    []string  `json:"scan_results"`    // Security scan findings
	Publisher      string    `json:"publisher"`       // Official publisher
	Verified       bool      `json:"verified"`        // Signature verified
	AuditTrail     []AuditEntry `json:"audit_trail"`  // Change history
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Action    string `json:"action"`     // published, updated, verified
	Timestamp string `json:"timestamp"`  // RFC3339 format
	Actor     string `json:"actor"`      // Who performed the action
	Details   string `json:"details"`    // Additional information
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

// Security verification functions

// VerifyPluginSecurity verifies the security of a plugin before installation
func (m *Manager) VerifyPluginSecurity(plugin *RegistryPlugin, pluginPath string) error {
	security := plugin.Security
	
	// Check trust level - if not specified, warn but allow installation
	if security.TrustLevel == "" {
		fmt.Printf("Warning: Plugin '%s' has no trust level specified - installing as unverified\n", plugin.Name)
		// Continue with installation but skip other security checks
		return nil
	}
	
	// For non-official plugins, require additional verification
	if security.TrustLevel != "official" && !security.Verified {
		fmt.Printf("Warning: Installing unverified plugin '%s' from trust level '%s'\n", 
			plugin.Name, security.TrustLevel)
		
		// Check for basic security scan results
		if len(security.ScanResults) > 0 {
			for _, result := range security.ScanResults {
				if strings.Contains(strings.ToLower(result), "malware") ||
				   strings.Contains(strings.ToLower(result), "virus") ||
				   strings.Contains(strings.ToLower(result), "threat") {
					return fmt.Errorf("security scan detected potential threat: %s", result)
				}
			}
		}
	}
	
	// Verify SHA256 checksum if provided
	if security.SHA256 != "" {
		if err := m.verifyFileChecksum(pluginPath, security.SHA256); err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}
	
	// Check file size (basic sanity check)
	if err := m.checkPluginFileSize(pluginPath, plugin.Size); err != nil {
		return fmt.Errorf("file size check failed: %w", err)
	}
	
	return nil
}

// verifyFileChecksum verifies the SHA256 checksum of a file
func (m *Manager) verifyFileChecksum(filePath, expectedSHA256 string) error {
	if expectedSHA256 == "" {
		return nil // No checksum to verify
	}
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	actualSHA256 := fmt.Sprintf("%x", sha256.Sum256(data))
	if actualSHA256 != expectedSHA256 {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA256, actualSHA256)
	}
	
	return nil
}

// checkPluginFileSize checks if plugin file size is reasonable
func (m *Manager) checkPluginFileSize(filePath, expectedSize string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	
	// Check for suspicious file sizes
	size := info.Size()
	if size > 100*1024*1024 { // 100MB limit
		return fmt.Errorf("plugin file too large: %d bytes", size)
	}
	
	if size < 1024 { // 1KB minimum
		return fmt.Errorf("plugin file too small: %d bytes", size)
	}
	
	return nil
}

// GetPluginSecurityInfo returns detailed security information for a plugin
func (m *Manager) GetPluginSecurityInfo(name string) (*SecurityInfo, error) {
	plugin, err := m.GetPluginInfo(name)
	if err != nil {
		return nil, err
	}
	
	return &plugin.Security, nil
}

// FilterPluginsByTrustLevel filters plugins by security trust level
func (m *Manager) FilterPluginsByTrustLevel(plugins []RegistryPlugin, trustLevel string) []RegistryPlugin {
	var filtered []RegistryPlugin
	
	for _, plugin := range plugins {
		if plugin.Security.TrustLevel == trustLevel {
			filtered = append(filtered, plugin)
		}
	}
	
	return filtered
}

// GetSecurityStats returns security statistics for the registry
func (m *Manager) GetSecurityStats() (map[string]int, error) {
	plugins, err := m.ListAvailable()
	if err != nil {
		return nil, err
	}
	
	stats := map[string]int{
		"total":     len(plugins),
		"official":  0,
		"verified":  0,
		"community": 0,
		"unscanned": 0,
	}
	
	for _, plugin := range plugins {
		switch plugin.Security.TrustLevel {
		case "official":
			stats["official"]++
		case "verified":
			stats["verified"]++
		case "community":
			stats["community"]++
		default:
			stats["unscanned"]++
		}
	}
	
	return stats, nil
}