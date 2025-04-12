package engine

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"gopkg.in/yaml.v3"
)

// PluginManifest represents the plugin manifest file
type PluginManifest struct {
	Plugins []PluginInfo `yaml:"plugins"`
}

// PluginInfo represents information about a plugin
type PluginInfo struct {
	Name       string `yaml:"name"`
	Repository string `yaml:"repository"`
	Version    string `yaml:"version"`
	Path       string `yaml:"path"`
}

// PluginDownloader is responsible for downloading plugins
type PluginDownloader struct {
	pluginDir string
	httpClient *http.Client
}

// NewPluginDownloader creates a new PluginDownloader
func NewPluginDownloader(pluginDir string) *PluginDownloader {
	return &PluginDownloader{
		pluginDir: pluginDir,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// LoadPluginManifest loads the plugin manifest from a file
func (d *PluginDownloader) LoadPluginManifest(path string) (*PluginManifest, error) {
	// Read the manifest file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading plugin manifest: %w", err)
	}

	// Parse the manifest
	var manifest PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("error parsing plugin manifest: %w", err)
	}

	return &manifest, nil
}

// DownloadPlugin downloads a plugin from a repository
func (d *PluginDownloader) DownloadPlugin(info PluginInfo) error {
	// Create the plugin directory if it doesn't exist
	pluginPath := filepath.Join(d.pluginDir, info.Name)
	if err := os.MkdirAll(pluginPath, 0755); err != nil {
		return fmt.Errorf("error creating plugin directory: %w", err)
	}

	// Construct the download URL
	url := constructDownloadURL(info)

	// Download the plugin
	fmt.Printf("Downloading plugin %s from %s...\n", info.Name, url)
	resp, err := d.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading plugin: HTTP status %d", resp.StatusCode)
	}

	// Create a temporary file to download to
	tmpFile, err := os.CreateTemp("", "plugin-*.tar.gz")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy the response body to the temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error downloading plugin: %w", err)
	}

	// Extract the plugin
	if err := extractPlugin(tmpFile.Name(), pluginPath); err != nil {
		return fmt.Errorf("error extracting plugin: %w", err)
	}

	fmt.Printf("Plugin %s downloaded and extracted successfully.\n", info.Name)
	return nil
}

// constructDownloadURL constructs the download URL for a plugin
func constructDownloadURL(info PluginInfo) string {
	// Handle GitHub repositories
	if strings.Contains(info.Repository, "github.com") {
		return fmt.Sprintf("%s/releases/download/%s/%s.tar.gz", info.Repository, info.Version, info.Name)
	}

	// Handle other repositories
	return fmt.Sprintf("%s/%s/%s.tar.gz", info.Repository, info.Version, info.Name)
}

// extractPlugin extracts a plugin from a tar.gz file
func extractPlugin(tarPath, destPath string) error {
	// Use the shell plugin to extract the tar.gz file
	cmd := fmt.Sprintf("tar -xzf %s -C %s", tarPath, destPath)
	_, err := executeCommand(cmd)
	return err
}

// executeCommand executes a shell command
func executeCommand(command string) (string, error) {
	// Create a temporary script file
	tmpFile, err := os.CreateTemp("", "plugin-script-*.sh")
	if err != nil {
		return "", fmt.Errorf("error creating temporary script: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write the command to the script file
	if _, err := tmpFile.WriteString("#!/bin/sh\n" + command); err != nil {
		return "", fmt.Errorf("error writing to temporary script: %w", err)
	}

	// Make the script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return "", fmt.Errorf("error making script executable: %w", err)
	}

	// Execute the script
	output, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("error executing command: %w", err)
	}

	return string(output), nil
}