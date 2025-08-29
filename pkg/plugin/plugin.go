package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"gopkg.in/yaml.v3"
)

// Plugin represents a Corynth plugin
type Plugin interface {
	// Metadata returns plugin metadata
	Metadata() Metadata
	
	// Execute runs the plugin action
	Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error)
	
	// Validate checks if the plugin configuration is valid
	Validate(params map[string]interface{}) error
	
	// Actions returns available actions for this plugin
	Actions() []Action
}

// Metadata contains plugin metadata
type Metadata struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Description string   `yaml:"description" json:"description"`
	Author      string   `yaml:"author" json:"author"`
	Tags        []string `yaml:"tags" json:"tags"`
	Repository  string   `yaml:"repository" json:"repository"`
	License     string   `yaml:"license" json:"license"`
}

// Action represents a plugin action
type Action struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Inputs      map[string]InputSpec   `yaml:"inputs" json:"inputs"`
	Outputs     map[string]OutputSpec  `yaml:"outputs" json:"outputs"`
	Examples    []Example              `yaml:"examples" json:"examples"`
}

// InputSpec defines an input parameter
type InputSpec struct {
	Type        string      `yaml:"type" json:"type"`
	Description string      `yaml:"description" json:"description"`
	Required    bool        `yaml:"required" json:"required"`
	Default     interface{} `yaml:"default" json:"default"`
	Validation  string      `yaml:"validation" json:"validation"`
}

// OutputSpec defines an output parameter
type OutputSpec struct {
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description" json:"description"`
}

// Example provides usage examples
type Example struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Input       map[string]interface{} `yaml:"input" json:"input"`
}

// Manager manages plugins
type Manager struct {
	mu           sync.RWMutex
	plugins      map[string]Plugin
	repositories []Repository
	localPath    string
	cachePath    string
	autoInstall  bool
}

// Repository represents a plugin repository
type Repository struct {
	Name     string
	URL      string
	Branch   string
	Token    string
	Priority int
}

// NewManager creates a new plugin manager
func NewManager(localPath, cachePath string, autoInstall bool) *Manager {
	return &Manager{
		plugins:     make(map[string]Plugin),
		localPath:   localPath,
		cachePath:   cachePath,
		autoInstall: autoInstall,
	}
}

// AddRepository adds a plugin repository
func (m *Manager) AddRepository(repo Repository) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.repositories = append(m.repositories, repo)
}

// LoadLocal loads plugins from local directory
func (m *Manager) LoadLocal() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create local plugin directory if it doesn't exist
	if err := os.MkdirAll(m.localPath, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Load built-in plugins first
	m.loadBuiltinPlugins()

	// Then load from local directory
	entries, err := os.ReadDir(m.localPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(m.localPath, entry.Name())
		
		if entry.IsDir() {
			// Load plugin from directory
			if err := m.loadPlugin(entryPath); err != nil {
				fmt.Printf("Warning: failed to load plugin from %s: %v\n", entryPath, err)
			}
		} else if strings.HasSuffix(entry.Name(), ".so") {
			// Load compiled plugin from .so file
			if err := m.loadCompiledPlugin(entryPath); err != nil {
				fmt.Printf("Warning: failed to load compiled plugin from %s: %v\n", entryPath, err)
			}
		} else if strings.HasPrefix(entry.Name(), "corynth-plugin-") && !strings.Contains(entry.Name(), ".") {
			// Load JSON protocol plugin executable
			if err := m.loadJSONPluginExecutable(entryPath); err != nil {
				fmt.Printf("Warning: failed to load JSON plugin from %s: %v\n", entryPath, err)
			}
		}
	}

	return nil
}

// loadBuiltinPlugins loads built-in plugins
func (m *Manager) loadBuiltinPlugins() {
	// Only shell plugin is built-in as per design specification
	// All other plugins must be loaded remotely
	m.plugins["shell"] = NewShellPlugin()
}

// tryLoadBuiltinPlugin attempts to lazy load a built-in plugin by name
func (m *Manager) tryLoadBuiltinPlugin(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if already loaded
	if _, exists := m.plugins[name]; exists {
		return true
	}
	
	// Only shell is built-in - all others should be loaded as RPC plugins
	if name == "shell" {
		// Shell is already loaded during initialization
		return true
	}
	
	// All other plugins should be loaded as RPC/subprocess plugins
	return false
}

// tryLoadRPCPlugin attempts to load a plugin as an RPC/subprocess plugin
func (m *Manager) tryLoadRPCPlugin(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Check if already loaded
	if _, exists := m.plugins[name]; exists {
		return true
	}
	
	// Look for plugin script in examples/plugins directory
	scriptPaths := []string{
		fmt.Sprintf("examples/plugins/%s-plugin.py", name),
		fmt.Sprintf("examples/plugins/%s-plugin.js", name),
		fmt.Sprintf("examples/plugins/%s-plugin.sh", name),
		fmt.Sprintf("examples/plugins/%s", name),
	}
	
	for _, scriptPath := range scriptPaths {
		if _, err := os.Stat(scriptPath); err == nil {
			// Found executable script - create RPC plugin
			metadata, actions, err := m.loadRPCPluginMetadata(scriptPath)
			if err != nil {
				continue
			}
			
			plugin := &ScriptPlugin{
				metadata:   metadata,
				scriptPath: scriptPath,
				actions:    actions,
			}
			
			m.plugins[name] = plugin
			return true
		}
	}
	
	return false
}

// loadRPCPluginMetadata loads metadata and actions from an RPC plugin script
func (m *Manager) loadRPCPluginMetadata(scriptPath string) (Metadata, []Action, error) {
	// Execute script with "metadata" action to get plugin info
	cmd := exec.Command(scriptPath, "metadata")
	output, err := cmd.Output()
	if err != nil {
		return Metadata{}, nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	
	var metadata Metadata
	if err := json.Unmarshal(output, &metadata); err != nil {
		return Metadata{}, nil, fmt.Errorf("invalid metadata JSON: %w", err)
	}
	
	// Execute script with "actions" action to get available actions
	cmd = exec.Command(scriptPath, "actions")
	output, err = cmd.Output()
	if err != nil {
		return metadata, nil, fmt.Errorf("failed to get actions: %w", err)
	}
	
	var actionsMap map[string]interface{}
	if err := json.Unmarshal(output, &actionsMap); err != nil {
		return metadata, nil, fmt.Errorf("invalid actions JSON: %w", err)
	}
	
	// Convert actions map to Action structs
	var actions []Action
	for name, actionData := range actionsMap {
		if actionMap, ok := actionData.(map[string]interface{}); ok {
			action := Action{
				Name:        name,
				Description: getStringFromMap(actionMap, "description"),
			}
			actions = append(actions, action)
		}
	}
	
	return metadata, actions, nil
}

// Helper function to safely get string from map
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// PluginConfig represents the complete plugin configuration
type PluginConfig struct {
	Metadata `yaml:",inline"`
	Actions  []Action `yaml:"actions"`
}

// loadPlugin loads a plugin from a directory
func (m *Manager) loadPlugin(path string) error {
	// First, check if it's a remote plugin with 'plugin' executable (JSON protocol)
	pluginExecutablePath := filepath.Join(path, "plugin")
	if _, err := os.Stat(pluginExecutablePath); err == nil {
		return m.loadRemoteJSONPlugin(path, pluginExecutablePath)
	}

	// Check for plugin.yaml or plugin.yml (traditional plugins)
	metadataPath := filepath.Join(path, "plugin.yaml")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		metadataPath = filepath.Join(path, "plugin.yml")
	}

	// If no YAML metadata, this might be a remote plugin without YAML
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf("no plugin metadata found (plugin.yaml/yml) and no 'plugin' executable")
	}

	// Read plugin metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin metadata: %w", err)
	}

	var config PluginConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse plugin metadata: %w", err)
	}

	// Check if it's a Go plugin
	soPath := filepath.Join(path, config.Name+".so")
	if _, err := os.Stat(soPath); err == nil {
		return m.loadGoPlugin(soPath, config.Metadata)
	}

	// Check if it's a script plugin
	scriptPath := filepath.Join(path, "plugin.sh")
	if _, err := os.Stat(scriptPath); err == nil {
		return m.loadScriptPlugin(scriptPath, config.Metadata, config.Actions)
	}

	return fmt.Errorf("no valid plugin implementation found")
}

// loadRemoteJSONPlugin loads a remote plugin that uses JSON stdin/stdout protocol
func (m *Manager) loadRemoteJSONPlugin(pluginDir, executablePath string) error {
	// Extract plugin name from directory
	pluginName := filepath.Base(pluginDir)
	
	// Get metadata from the plugin executable
	metadata, err := m.loadJSONPluginMetadata(executablePath)
	if err != nil {
		return fmt.Errorf("failed to load plugin metadata: %w", err)
	}
	
	// Get actions from the plugin executable
	actions, err := m.loadJSONPluginActions(executablePath)
	if err != nil {
		return fmt.Errorf("failed to load plugin actions: %w", err)
	}
	
	// Create script plugin instance
	plugin := &ScriptPlugin{
		metadata:   metadata,
		scriptPath: executablePath,
		actions:    actions,
	}
	
	// Use the plugin name from metadata if available, otherwise use directory name
	if metadata.Name != "" {
		pluginName = metadata.Name
	}
	
	// Store the plugin
	m.plugins[pluginName] = plugin
	
	return nil
}

// loadJSONPluginMetadata gets metadata from a JSON protocol plugin
func (m *Manager) loadJSONPluginMetadata(executablePath string) (Metadata, error) {
	debug := os.Getenv("CORYNTH_DEBUG") != ""
	if debug {
		log.Printf("[DEBUG] Loading metadata from: %s", executablePath)
	}
	
	cmd := exec.Command(executablePath, "metadata")
	output, err := cmd.Output()
	if err != nil {
		if debug {
			log.Printf("[DEBUG] Metadata command failed: %v", err)
		}
		return Metadata{}, fmt.Errorf("failed to get metadata: %w", err)
	}
	
	if debug {
		log.Printf("[DEBUG] Metadata output: %s", string(output))
	}
	
	var metadata Metadata
	if err := json.Unmarshal(output, &metadata); err != nil {
		if debug {
			log.Printf("[DEBUG] JSON unmarshal failed: %v", err)
		}
		return Metadata{}, fmt.Errorf("invalid metadata JSON: %w", err)
	}
	
	return metadata, nil
}

// loadJSONPluginActions gets actions from a JSON protocol plugin
func (m *Manager) loadJSONPluginActions(executablePath string) ([]Action, error) {
	cmd := exec.Command(executablePath, "actions")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get actions: %w", err)
	}
	
	// The output is a map of action names to action specs
	var actionsMap map[string]interface{}
	if err := json.Unmarshal(output, &actionsMap); err != nil {
		return nil, fmt.Errorf("invalid actions JSON: %w", err)
	}
	
	// Convert to Action structs
	var actions []Action
	for name, spec := range actionsMap {
		action := Action{
			Name:        name,
			Description: getStringFromActionSpec(spec, "description"),
		}
		
		// Parse inputs and outputs if available
		if specMap, ok := spec.(map[string]interface{}); ok {
			action.Inputs = parseInputSpecs(specMap["inputs"])
			action.Outputs = parseOutputSpecs(specMap["outputs"])
		}
		
		actions = append(actions, action)
	}
	
	return actions, nil
}

// Helper functions for parsing action specifications
func getStringFromActionSpec(spec interface{}, key string) string {
	if specMap, ok := spec.(map[string]interface{}); ok {
		if val, ok := specMap[key].(string); ok {
			return val
		}
	}
	return ""
}

func parseInputSpecs(inputs interface{}) map[string]InputSpec {
	result := make(map[string]InputSpec)
	if inputsMap, ok := inputs.(map[string]interface{}); ok {
		for name, spec := range inputsMap {
			if specMap, ok := spec.(map[string]interface{}); ok {
				inputSpec := InputSpec{
					Type:        getStringFromMap(specMap, "type"),
					Description: getStringFromMap(specMap, "description"),
					Required:    getBoolFromMap(specMap, "required"),
					Default:     specMap["default"],
				}
				result[name] = inputSpec
			}
		}
	}
	return result
}

func parseOutputSpecs(outputs interface{}) map[string]OutputSpec {
	result := make(map[string]OutputSpec)
	if outputsMap, ok := outputs.(map[string]interface{}); ok {
		for name, spec := range outputsMap {
			if specMap, ok := spec.(map[string]interface{}); ok {
				outputSpec := OutputSpec{
					Type:        getStringFromMap(specMap, "type"),
					Description: getStringFromMap(specMap, "description"),
				}
				result[name] = outputSpec
			}
		}
	}
	return result
}

// Helper function to safely get bool from map
func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

// loadJSONPluginExecutable loads a JSON protocol plugin from an executable file
func (m *Manager) loadJSONPluginExecutable(executablePath string) error {
	// Extract plugin name from executable path (remove "corynth-plugin-" prefix)
	fileName := filepath.Base(executablePath)
	if !strings.HasPrefix(fileName, "corynth-plugin-") {
		return fmt.Errorf("invalid JSON plugin name format: %s", fileName)
	}
	
	pluginName := strings.TrimPrefix(fileName, "corynth-plugin-")
	
	// Get metadata from the plugin executable
	metadata, err := m.loadJSONPluginMetadata(executablePath)
	if err != nil {
		return fmt.Errorf("failed to load plugin metadata: %w", err)
	}
	
	// Get actions from the plugin executable
	actions, err := m.loadJSONPluginActions(executablePath)
	if err != nil {
		return fmt.Errorf("failed to load plugin actions: %w", err)
	}
	
	// Create script plugin instance
	plugin := &ScriptPlugin{
		metadata:   metadata,
		scriptPath: executablePath,
		actions:    actions,
	}
	
	// Use the plugin name from metadata if available, otherwise use extracted name
	if metadata.Name != "" {
		pluginName = metadata.Name
	}
	
	// Store the plugin
	m.plugins[pluginName] = plugin
	
	return nil
}

// loadGoPlugin loads a Go plugin
func (m *Manager) loadGoPlugin(path string, metadata Metadata) error {
	p, err := plugin.Open(path)
	if err != nil {
		// Check for common plugin loading issues and provide helpful error messages
		if strings.Contains(err.Error(), "different version") {
			return fmt.Errorf("plugin '%s' was compiled with incompatible Go module version. "+
				"This usually means the plugin needs to be recompiled with the current Corynth version. "+
				"Error: %w", metadata.Name, err)
		}
		return fmt.Errorf("failed to open plugin '%s': %w", metadata.Name, err)
	}

	symbol, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin '%s' is missing required 'Plugin' symbol. "+
			"This plugin may be corrupted or incorrectly compiled. Error: %w", metadata.Name, err)
	}

	plugin, ok := symbol.(Plugin)
	if !ok {
		return fmt.Errorf("plugin '%s' has invalid plugin type. "+
			"The Plugin symbol does not implement the expected interface", metadata.Name)
	}

	m.plugins[metadata.Name] = plugin
	return nil
}

// loadScriptPlugin loads a script-based plugin
func (m *Manager) loadScriptPlugin(path string, metadata Metadata, actions []Action) error {
	plugin := &ScriptPlugin{
		metadata:   metadata,
		scriptPath: path,
		actions:    actions,
	}
	
	m.plugins[metadata.Name] = plugin
	return nil
}

// loadCompiledPlugin loads a compiled .so plugin file directly
func (m *Manager) loadCompiledPlugin(soPath string) error {
	// Make sure we use absolute path for plugin.Open
	absPath, err := filepath.Abs(soPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for plugin: %w", err)
	}
	
	// Extract plugin name from path for better error messages
	pluginName := filepath.Base(absPath)
	if strings.HasSuffix(pluginName, ".so") {
		pluginName = strings.TrimSuffix(pluginName, ".so")
	}
	if strings.HasPrefix(pluginName, "corynth-plugin-") {
		pluginName = strings.TrimPrefix(pluginName, "corynth-plugin-")
	}
	
	// Load the .so file
	p, err := plugin.Open(absPath)
	if err != nil {
		// Handle version mismatch by attempting automatic recompilation
		if strings.Contains(err.Error(), "different version") {
			// Try to find source and recompile automatically
			if recompileErr := m.attemptPluginRecompile(pluginName, absPath); recompileErr == nil {
				// Retry loading after recompilation
				p, retryErr := plugin.Open(absPath)
				if retryErr == nil {
					return m.processLoadedPlugin(p, pluginName, absPath)
				}
			}
			
			return fmt.Errorf("plugin '%s' was compiled with incompatible Go module version. "+
				"Please rebuild the plugin or use a compatible version. "+
				"Try: go build -buildmode=plugin -o %s plugin.go\n"+
				"Error: %w", pluginName, absPath, err)
		}
		if strings.Contains(err.Error(), "no such file") {
			return fmt.Errorf("plugin file not found: %s. "+
				"Make sure the plugin is installed correctly", absPath)
		}
		return fmt.Errorf("failed to load plugin '%s' from %s: %w", pluginName, absPath, err)
	}

	return m.processLoadedPlugin(p, pluginName, absPath)
}

// processLoadedPlugin handles the common logic for processing a loaded Go plugin
func (m *Manager) processLoadedPlugin(p *plugin.Plugin, pluginName, absPath string) error {
	// Look for the ExportedPlugin symbol (standard for our Go plugins)
	symbol, err := p.Lookup("ExportedPlugin")
	if err != nil {
		return fmt.Errorf("plugin '%s' is missing required 'ExportedPlugin' symbol. "+
			"This plugin may not be a valid Corynth plugin or needs to be recompiled. Error: %w", pluginName, err)
	}

	// Cast to Plugin interface
	pluginInstance, ok := symbol.(Plugin)
	if !ok {
		// Try to work around Go plugin interface identity issue using reflection
		pluginWrapper := &ReflectedPlugin{underlying: symbol}
		
		// Test if the wrapper works by calling Metadata
		if err := pluginWrapper.testMethods(); err != nil {
			return fmt.Errorf("plugin does not implement required methods: %w", err)
		}
		
		pluginInstance = pluginWrapper
	}

	// Get metadata to determine plugin name
	metadata := pluginInstance.Metadata()
	m.plugins[metadata.Name] = pluginInstance
	
	return nil
}

// attemptPluginRecompile tries to recompile a plugin when version mismatch occurs
func (m *Manager) attemptPluginRecompile(pluginName, pluginPath string) error {
	// Find the source directory for this plugin
	// Check if we have cached source from a recent installation
	pluginDir := filepath.Dir(pluginPath)
	possibleSourceDirs := []string{
		filepath.Join(pluginDir, pluginName),
		fmt.Sprintf(".corynth/cache/repos/official/%s", pluginName),
		fmt.Sprintf("examples/plugins/%s", pluginName),
	}
	
	for _, sourceDir := range possibleSourceDirs {
		if _, err := os.Stat(filepath.Join(sourceDir, "plugin.go")); err == nil {
			// Found source, try to recompile
			return m.recompilePlugin(sourceDir, pluginPath)
		}
	}
	
	return fmt.Errorf("no source found to recompile plugin %s", pluginName)
}

// recompilePlugin recompiles a plugin from source
func (m *Manager) recompilePlugin(sourceDir, outputPath string) error {
	// Fix paths in plugin files before compilation
	if err := m.fixPluginPaths(sourceDir); err != nil {
		return fmt.Errorf("failed to fix plugin paths: %w", err)
	}
	
	// Run go mod tidy to ensure dependencies are resolved
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	tidyCmd.Dir = sourceDir
	
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to resolve plugin dependencies: %w\nOutput: %s", err, string(output))
	}

	// Compile plugin
	buildCmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputPath, "plugin.go")
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=1")
	buildCmd.Dir = sourceDir
	
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to recompile plugin: %w\nOutput: %s", err, string(output))
	}
	
	return nil
}

// Get returns a plugin by name
func (m *Manager) Get(name string) (Plugin, error) {
	m.mu.RLock()
	plugin, exists := m.plugins[name]
	m.mu.RUnlock()

	if exists {
		return plugin, nil
	}

	// Try to lazy load built-in plugin first
	if m.tryLoadBuiltinPlugin(name) {
		m.mu.RLock()
		plugin, exists = m.plugins[name]
		m.mu.RUnlock()
		
		if exists {
			return plugin, nil
		}
	}

	// Try to load as RPC plugin from local scripts
	if m.tryLoadRPCPlugin(name) {
		m.mu.RLock()
		plugin, exists = m.plugins[name]
		m.mu.RUnlock()
		
		if exists {
			return plugin, nil
		}
	}

	// If auto-install is enabled, try to install from repositories
	if m.autoInstall {
		if err := m.InstallFromRepository(name); err != nil {
			return nil, fmt.Errorf("plugin '%s' not found and auto-install failed: %w", name, err)
		}

		m.mu.RLock()
		plugin, exists = m.plugins[name]
		m.mu.RUnlock()

		if exists {
			return plugin, nil
		}
	}

	return nil, fmt.Errorf("plugin '%s' not found", name)
}

// InstallFromRepository installs a plugin from configured repositories
func (m *Manager) InstallFromRepository(name string) error {
	var allErrors []string
	
	for _, repo := range m.repositories {
		err := m.installFromGit(repo, name)
		if err == nil {
			return nil
		}
		allErrors = append(allErrors, fmt.Sprintf("repo %s: %v", repo.Name, err))
	}
	
	if len(allErrors) == 0 {
		return fmt.Errorf("plugin '%s' not found: no repositories configured", name)
	}
	
	return fmt.Errorf("plugin '%s' installation failed from all repositories: %s", name, strings.Join(allErrors, "; "))
}

// installFromGit installs a plugin from a Git repository
func (m *Manager) installFromGit(repo Repository, pluginName string) error {
	// First try to install from pre-compiled binaries
	if err := m.installPrecompiledPlugin(repo, pluginName); err == nil {
		return nil
	}
	
	// Fallback to source installation
	return m.installSourcePlugin(repo, pluginName)
}

// installPrecompiledPlugin installs a pre-compiled binary plugin
func (m *Manager) installPrecompiledPlugin(repo Repository, pluginName string) error {
	// Detect platform
	platform := runtime.GOOS
	arch := runtime.GOARCH
	
	// Map Go arch names to release arch names
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return fmt.Errorf("unsupported architecture: %s", arch)
	}
	
	// Construct binary name
	binaryName := fmt.Sprintf("corynth-plugin-%s-%s-%s", pluginName, platform, arch)
	if platform == "windows" {
		binaryName += ".exe"
	}
	
	// Try to download from GitHub releases
	releaseURL := fmt.Sprintf("https://github.com/corynth/plugins/releases/latest/download/%s", binaryName)
	
	// Download to temporary file first
	destPath := filepath.Join(m.localPath, fmt.Sprintf("corynth-plugin-%s", pluginName))
	tempPath := destPath + ".tmp"
	
	if err := m.downloadFile(releaseURL, tempPath); err != nil {
		return fmt.Errorf("failed to download precompiled plugin: %w", err)
	}
	
	// Make executable and move to final location
	if err := os.Chmod(tempPath, 0755); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to make plugin executable: %w", err)
	}
	
	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to install plugin: %w", err)
	}
	
	fmt.Printf("✓ Installed precompiled plugin %s for %s/%s\n", pluginName, platform, arch)
	return nil
}

// downloadFile downloads a file from URL to local path
func (m *Manager) downloadFile(url, destPath string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	
	// Create file
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	
	// Download
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}
	
	// Copy to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// installSourcePlugin installs a plugin from source (fallback)
func (m *Manager) installSourcePlugin(repo Repository, pluginName string) error {
	// Clone or update repository
	repoPath := filepath.Join(m.cachePath, "repos", repo.Name)
	
	var gitRepo *git.Repository
	var err error

	if _, statErr := os.Stat(repoPath); os.IsNotExist(statErr) {
		// Clone repository
		cloneOpts := &git.CloneOptions{
			URL:      repo.URL,
			Progress: os.Stdout,
		}

		if repo.Branch != "" {
			cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(repo.Branch)
		}

		if repo.Token != "" {
			cloneOpts.Auth = &githttp.BasicAuth{
				Username: "token",
				Password: repo.Token,
			}
		}

		gitRepo, err = git.PlainClone(repoPath, false, cloneOpts)
		if err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
	} else {
		// Open and pull latest changes
		gitRepo, err = git.PlainOpen(repoPath)
		if err != nil {
			return fmt.Errorf("failed to open repository: %w", err)
		}

		w, err := gitRepo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to get worktree: %w", err)
		}

		pullOpts := &git.PullOptions{
			RemoteName: "origin",
		}

		if repo.Token != "" {
			pullOpts.Auth = &githttp.BasicAuth{
				Username: "token",
				Password: repo.Token,
			}
		}

		err = w.Pull(pullOpts)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("failed to pull repository: %w", err)
		}
	}

	// Look for compiled .so file first (preferred)
	soFilePath := filepath.Join(repoPath, fmt.Sprintf("corynth-plugin-%s.so", pluginName))
	if _, err := os.Stat(soFilePath); err == nil {
		// Copy .so file to local path
		destSoPath := filepath.Join(m.localPath, fmt.Sprintf("corynth-plugin-%s.so", pluginName))
		if err := copyFile(soFilePath, destSoPath); err != nil {
			return fmt.Errorf("failed to install compiled plugin: %w", err)
		}
		
		// Load the compiled plugin directly
		return m.loadCompiledPlugin(destSoPath)
	}

	// Enable debug logging if CORYNTH_DEBUG is set
	debug := os.Getenv("CORYNTH_DEBUG") != ""
	
	// Detect current platform
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	if debug {
		log.Printf("[DEBUG] Installing plugin '%s' for platform: %s", pluginName, platform)
	}
	
	// Try multiple paths, prioritizing JSON protocol plugins over compiled binaries
	binaryPaths := []struct {
		path string
		desc string
	}{
		// JSON protocol plugins (highest priority - these actually work)
		{filepath.Join(repoPath, "official", pluginName, "plugin"), "JSON protocol plugin"},
		
		// Platform-specific binaries
		{filepath.Join(repoPath, "official", pluginName, fmt.Sprintf("%s-plugin-%s", pluginName, platform)), "platform-specific binary"},
		{filepath.Join(repoPath, "official", pluginName, fmt.Sprintf("%s-%s", pluginName, platform)), "platform-specific name"},
		
		// Generic binaries
		{filepath.Join(repoPath, "official", pluginName, fmt.Sprintf("%s-plugin", pluginName)), "generic plugin binary"},
		{filepath.Join(repoPath, "official", pluginName, pluginName), "plugin name only"},
		
		// Fallback to root level
		{filepath.Join(repoPath, pluginName, fmt.Sprintf("%s-plugin", pluginName)), "root level plugin"},
		{filepath.Join(repoPath, pluginName, "plugin"), "root level generic"},
		
		// Additional patterns for compatibility
		{filepath.Join(repoPath, "official", pluginName, "bin", pluginName), "bin directory"},
		{filepath.Join(repoPath, "official", pluginName, "dist", pluginName), "dist directory"},
	}

	var lastError error
	for _, bp := range binaryPaths {
		if debug {
			log.Printf("[DEBUG] Trying path: %s (%s)", bp.path, bp.desc)
		}
		
		if info, err := os.Stat(bp.path); err == nil {
			if debug {
				log.Printf("[DEBUG] Found file at %s, size: %d bytes", bp.path, info.Size())
			}
			
			// Check if it's executable (binary or JSON protocol script)
			if isExecutable, err := isExecutableFile(bp.path); err == nil && isExecutable {
				destPath := filepath.Join(m.localPath, fmt.Sprintf("corynth-plugin-%s", pluginName))
				
				// Check if this is a JSON protocol plugin (script wrapper)
				isScript := isShellScript(bp.path)
				if debug {
					if isScript {
						log.Printf("[DEBUG] Found JSON protocol plugin script, copying to: %s", destPath)
					} else {
						log.Printf("[DEBUG] Found binary plugin, copying to: %s", destPath)
					}
				}
				
				// Ensure destination has execute permissions
				if err := copyFileWithPermissions(bp.path, destPath, 0755); err != nil {
					lastError = fmt.Errorf("copy failed: %w", err)
					if debug {
						log.Printf("[DEBUG] Failed to copy: %v", err)
					}
					continue
				}
				
				// Perform security verification for remote plugins
				if pluginInfo, err := m.GetPluginInfo(pluginName); err == nil {
					if err := m.VerifyPluginSecurity(pluginInfo, destPath); err != nil {
						lastError = fmt.Errorf("security verification failed: %w", err)
						if debug {
							log.Printf("[DEBUG] Security verification failed: %v", err)
						}
						// Remove potentially unsafe plugin
						os.Remove(destPath)
						continue
					}
					if debug {
						log.Printf("[DEBUG] Security verification passed for plugin '%s'", pluginName)
					}
				} else if debug {
					log.Printf("[DEBUG] No security info found for plugin '%s', proceeding with basic checks", pluginName)
				}
				
				// Load the plugin based on its type
				if isScript {
					// For JSON protocol plugins, create a self-contained wrapper
					srcDir := filepath.Dir(bp.path)
					
					if debug {
						log.Printf("[DEBUG] Creating self-contained JSON protocol plugin wrapper")
					}
					
					// Create a self-contained wrapper script that includes the Go plugin
					if err := createSelfContainedWrapper(destPath, srcDir, pluginName); err != nil {
						lastError = fmt.Errorf("failed to create self-contained plugin: %w", err)
						if debug {
							log.Printf("[DEBUG] Failed to create self-contained plugin: %v", err)
						}
						continue
					}
					
					// For JSON protocol plugins, load as script plugin
					if err := m.loadJSONPluginExecutable(destPath); err != nil {
						lastError = fmt.Errorf("failed to load JSON plugin: %w", err)
						if debug {
							log.Printf("[DEBUG] Failed to load JSON plugin: %v", err)
						}
						os.Remove(destPath)
						continue
					}
				} else {
					// For compiled plugins, try to load as compiled plugin
					if err := m.loadCompiledPlugin(destPath); err != nil {
						lastError = fmt.Errorf("failed to load compiled plugin: %w", err)
						if debug {
							log.Printf("[DEBUG] Failed to load compiled plugin: %v", err)
						}
						os.Remove(destPath)
						continue
					}
				}
				
				// Perform health check on the plugin
				if err := m.healthCheckPlugin(pluginName); err != nil {
					lastError = fmt.Errorf("health check failed: %w", err)
					if debug {
						log.Printf("[DEBUG] Plugin health check failed: %v", err)
					}
					// Remove unhealthy plugin
					os.Remove(destPath)
					continue
				}
				
				if debug {
					log.Printf("[DEBUG] Successfully installed plugin '%s' from %s", pluginName, bp.desc)
				}
				return nil // Success!
			} else if debug {
				log.Printf("[DEBUG] File is not a valid binary executable")
			}
		} else if debug && !os.IsNotExist(err) {
			log.Printf("[DEBUG] Error checking path: %v", err)
		}
	}
	
	if debug {
		log.Printf("[DEBUG] No pre-compiled binary found, attempting compilation from source")
	}

	// Look for plugin directory containing plugin.go in official/ subdirectory
	pluginDirPath := filepath.Join(repoPath, "official", pluginName)
	goFilePath := filepath.Join(pluginDirPath, "plugin.go")
	if _, err := os.Stat(goFilePath); err == nil {
		// Compile Go plugin to .so file - use absolute path
		destSoPath, err := filepath.Abs(filepath.Join(m.localPath, fmt.Sprintf("corynth-plugin-%s.so", pluginName)))
		if err != nil {
			return fmt.Errorf("failed to get absolute path for plugin destination: %w", err)
		}
		if err := m.compileGoPlugin(goFilePath, destSoPath); err != nil {
			return fmt.Errorf("failed to compile plugin: %w", err)
		}
		
		// Load the compiled plugin
		return m.loadCompiledPlugin(destSoPath)
	}

	// Look for Go source file in root to compile  
	rootGoFilePath := filepath.Join(repoPath, fmt.Sprintf("%s.go", pluginName))
	if _, err := os.Stat(rootGoFilePath); err == nil {
		// Compile Go plugin to .so file - use absolute path
		destSoPath, err := filepath.Abs(filepath.Join(m.localPath, fmt.Sprintf("corynth-plugin-%s.so", pluginName)))
		if err != nil {
			return fmt.Errorf("failed to get absolute path for plugin destination: %w", err)
		}
		if err := m.compileGoPlugin(rootGoFilePath, destSoPath); err != nil {
			return fmt.Errorf("failed to compile plugin: %w", err)
		}
		
		// Load the compiled plugin
		return m.loadCompiledPlugin(destSoPath)
	}

	// Fallback to legacy source plugin directory structure
	legacyPluginPath := filepath.Join(repoPath, "plugins", pluginName)
	if _, err := os.Stat(legacyPluginPath); os.IsNotExist(err) {
		// Create comprehensive error message with all attempted paths
		var attemptedPaths []string
		attemptedPaths = append(attemptedPaths, fmt.Sprintf(".so file: %s", filepath.Join(repoPath, fmt.Sprintf("corynth-plugin-%s.so", pluginName))))
		
		for _, bp := range binaryPaths {
			attemptedPaths = append(attemptedPaths, fmt.Sprintf("%s: %s", bp.desc, bp.path))
		}
		
		attemptedPaths = append(attemptedPaths, fmt.Sprintf("Go source: %s", goFilePath))
		attemptedPaths = append(attemptedPaths, fmt.Sprintf("Root Go file: %s", rootGoFilePath))
		attemptedPaths = append(attemptedPaths, fmt.Sprintf("Legacy directory: %s", legacyPluginPath))
		
		errorMsg := fmt.Sprintf("plugin '%s' not found in repository.\n\nAttempted paths:\n", pluginName)
		for i, path := range attemptedPaths {
			errorMsg += fmt.Sprintf("  %d. %s\n", i+1, path)
		}
		
		if lastError != nil {
			errorMsg += fmt.Sprintf("\nLast error encountered: %v", lastError)
		}
		
		if debug {
			errorMsg += "\n\nEnable CORYNTH_DEBUG=1 for detailed debugging information."
		}
		
		return fmt.Errorf("%s", errorMsg)
	}
	
	pluginPath := legacyPluginPath

	// Copy plugin source to local path
	destPath := filepath.Join(m.localPath, pluginName)
	if err := copyDir(pluginPath, destPath); err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	// Load the installed plugin
	return m.loadPlugin(destPath)
}

// List returns all available plugins
func (m *Manager) List() []Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var metadata []Metadata
	for _, p := range m.plugins {
		metadata = append(metadata, p.Metadata())
	}
	return metadata
}

// Search searches for plugins by criteria
func (m *Manager) Search(query string, tags []string) []Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []Metadata
	query = strings.ToLower(query)

	for _, p := range m.plugins {
		meta := p.Metadata()
		
		// Check if query matches name or description
		if query != "" {
			nameMatch := strings.Contains(strings.ToLower(meta.Name), query)
			descMatch := strings.Contains(strings.ToLower(meta.Description), query)
			if !nameMatch && !descMatch {
				continue
			}
		}

		// Check if tags match
		if len(tags) > 0 {
			hasTag := false
			for _, tag := range tags {
				for _, pTag := range meta.Tags {
					if strings.EqualFold(tag, pTag) {
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

// Update updates a plugin to the latest version
func (m *Manager) Update(name string) error {
	m.mu.RLock()
	p, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	meta := p.Metadata()
	if meta.Repository == "" {
		return fmt.Errorf("plugin '%s' has no repository information", name)
	}

	// Find repository configuration
	var repo *Repository
	for _, r := range m.repositories {
		if strings.Contains(meta.Repository, r.URL) {
			repo = &r
			break
		}
	}

	if repo == nil {
		return fmt.Errorf("repository for plugin '%s' not configured", name)
	}

	// Reinstall from repository
	return m.installFromGit(*repo, name)
}

// Remove removes a plugin
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[name]; !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	// Remove from memory
	delete(m.plugins, name)

	// Remove from disk - both directory and .so file
	pluginPath := filepath.Join(m.localPath, name)
	soPath := filepath.Join(m.localPath, fmt.Sprintf("corynth-plugin-%s.so", name))
	
	// Remove directory if it exists
	if _, err := os.Stat(pluginPath); err == nil {
		if err := os.RemoveAll(pluginPath); err != nil {
			return fmt.Errorf("failed to remove plugin directory: %w", err)
		}
	}
	
	// Remove .so file if it exists
	if _, err := os.Stat(soPath); err == nil {
		if err := os.Remove(soPath); err != nil {
			return fmt.Errorf("failed to remove plugin .so file: %w", err)
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Create parent directories
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// Copy file permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}

// compileGoPlugin compiles a Go source file to a .so plugin
func (m *Manager) compileGoPlugin(srcPath, destPath string) error {
	// Create parent directories
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Set working directory to the source file's directory for proper module resolution
	workDir := filepath.Dir(srcPath)
	
	// Fix paths in plugin files before compilation
	if err := m.fixPluginPaths(workDir); err != nil {
		return fmt.Errorf("failed to fix plugin paths: %w", err)
	}
	
	// First run go mod tidy to ensure dependencies are resolved
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Env = append(os.Environ(),
		"CGO_ENABLED=1", // Required for plugins
	)
	tidyCmd.Dir = workDir
	
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to resolve plugin dependencies: %w\nOutput: %s", err, string(output))
	}

	// Execute go build with plugin buildmode
	// Use just the filename since we're setting workDir
	fileName := filepath.Base(srcPath)
	
	buildCmd := exec.Command("go", "build", "-buildmode=plugin", "-o", destPath, fileName)
	
	// Set environment variables for the build
	buildCmd.Env = append(os.Environ(),
		"CGO_ENABLED=1", // Required for plugins
	)
	buildCmd.Dir = workDir
	
	// Capture output for error reporting
	output, err := buildCmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("failed to compile plugin: %w\nOutput: %s", err, string(output))
	}
	
	// Verify the destination file was created
	if _, err := os.Stat(destPath); err != nil {
		// Check if it was created in the working directory instead
		localSoPath := filepath.Join(workDir, filepath.Base(destPath))
		if _, err := os.Stat(localSoPath); err == nil {
			// Move it to the correct location
			if err := os.Rename(localSoPath, destPath); err != nil {
				return fmt.Errorf("failed to move .so file from %s to %s: %w", localSoPath, destPath, err)
			}
		} else {
			return fmt.Errorf("compiled plugin not found at expected location: %s", destPath)
		}
	}
	
	return nil
}

// fixPluginPaths fixes common path issues in plugin files before compilation
func (m *Manager) fixPluginPaths(workDir string) error {
	// Fix go.mod file
	goModPath := filepath.Join(workDir, "go.mod")
	if err := m.fixGoModPaths(goModPath); err != nil {
		return fmt.Errorf("failed to fix go.mod: %w", err)
	}
	
	// Fix plugin.go file
	pluginGoPath := filepath.Join(workDir, "plugin.go")
	if err := m.fixPluginGoPaths(pluginGoPath); err != nil {
		return fmt.Errorf("failed to fix plugin.go: %w", err)
	}
	
	return nil
}

// fixGoModPaths fixes the replace directive in go.mod to point to the correct corynth directory
func (m *Manager) fixGoModPaths(goModPath string) error {
	// Read go.mod file
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	
	// Find the corynth source directory by looking for pkg/plugin
	corynthDir, err := m.findCorynthSourceDir()
	if err != nil {
		return fmt.Errorf("failed to find corynth source directory: %w", err)
	}
	
	
	// Replace the incorrect path with the correct corynth source path
	contentStr := string(content)
	
	// Fix common wrong patterns for both corynth and corynth-dist
	patterns := []string{
		"../../../corynth-dist",
		"../../../../corynth-dist", 
		"../../corynth-dist",
		"../corynth-dist",
		"../../../corynth",
		"../../../../corynth", 
		"../../corynth",
		"../corynth",
	}
	
	for _, pattern := range patterns {
		// Fix both corynth and corynth-dist replace directives
		contentStr = strings.ReplaceAll(contentStr, 
			fmt.Sprintf("replace github.com/corynth/corynth => %s", pattern),
			fmt.Sprintf("replace github.com/corynth/corynth => %s", corynthDir))
		contentStr = strings.ReplaceAll(contentStr, 
			fmt.Sprintf("replace github.com/corynth/corynth-dist => %s", pattern),
			fmt.Sprintf("replace github.com/corynth/corynth-dist => %s", corynthDir))
	}
	
	// Fix pkg/plugin specific patterns (convert to full module import)
	pkgPluginPatterns := []string{
		"../../../corynth/pkg/plugin",
		"../../../../corynth/pkg/plugin",
		"../../corynth/pkg/plugin", 
		"../corynth/pkg/plugin",
		"../../../corynth-dist/pkg/plugin",
		"../../../../corynth-dist/pkg/plugin",
		"../../corynth-dist/pkg/plugin", 
		"../corynth-dist/pkg/plugin",
	}
	
	// Also fix the current broken pattern from remote plugins
	currentBrokenPatterns := []string{
		fmt.Sprintf("%s/pkg/plugin", corynthDir),
		"/tmp/corynth-fresh-test/corynth/pkg/plugin",
	}
	
	// First, convert pkg/plugin specific imports to full module imports
	for _, pattern := range append(pkgPluginPatterns, currentBrokenPatterns...) {
		contentStr = strings.ReplaceAll(contentStr,
			fmt.Sprintf("replace github.com/corynth/corynth/pkg/plugin => %s", pattern),
			fmt.Sprintf("replace github.com/corynth/corynth => %s", corynthDir))
	}
	
	// Fix require statements to use full module instead of /pkg/plugin
	contentStr = strings.ReplaceAll(contentStr,
		"require (\n    github.com/corynth/corynth/pkg/plugin v0.0.0-20240101000000-000000000000\n)",
		"require github.com/corynth/corynth v0.0.0-00010101000000-000000000000")
	contentStr = strings.ReplaceAll(contentStr,
		"github.com/corynth/corynth/pkg/plugin v0.0.0-20240101000000-000000000000",
		"github.com/corynth/corynth v0.0.0-00010101000000-000000000000")
	contentStr = strings.ReplaceAll(contentStr,
		"require github.com/corynth/corynth/pkg/plugin",
		"require github.com/corynth/corynth")
	
	// Fix corynth-dist require statements too - use actual working directory for replace
	contentStr = strings.ReplaceAll(contentStr,
		"require github.com/corynth/corynth-dist/pkg/plugin v0.0.0-20240101000000-000000000000",
		"require github.com/corynth/corynth v0.0.0-00010101000000-000000000000")
	contentStr = strings.ReplaceAll(contentStr,
		"github.com/corynth/corynth-dist/pkg/plugin v0.0.0-20240101000000-000000000000",
		"github.com/corynth/corynth v0.0.0-00010101000000-000000000000")
	contentStr = strings.ReplaceAll(contentStr,
		"require github.com/corynth/corynth-dist/pkg/plugin",
		"require github.com/corynth/corynth")
	contentStr = strings.ReplaceAll(contentStr,
		"github.com/corynth/corynth-dist v0.0.0-20240101000000-000000000000",
		"github.com/corynth/corynth v0.0.0-00010101000000-000000000000")
	
	// Remove the fake version requirement entirely and use replace directive only
	contentStr = strings.ReplaceAll(contentStr,
		"require github.com/corynth/corynth v0.0.0-00010101000000-000000000000",
		"")
	contentStr = strings.ReplaceAll(contentStr,
		"github.com/corynth/corynth v0.0.0-00010101000000-000000000000",
		"")
	
	// Write back the fixed content
	return os.WriteFile(goModPath, []byte(contentStr), 0644)
}

// findCorynthSourceDir finds the corynth source directory by looking for pkg/plugin
func (m *Manager) findCorynthSourceDir() (string, error) {
	// Start with current working directory and its parents
	currentDir, _ := os.Getwd()
	
	// Common possible locations relative to current working directory
	candidates := []string{
		".",                                    // Current directory
		"..",                                   // One level up
		"../..",                               // Two levels up
		"../../..",                            // Three levels up
		"../../../..",                         // Four levels up
		"../../../../..",                      // Five levels up
	}
	
	// Also try relative to current working directory
	for _, candidate := range candidates {
		testDir := filepath.Join(currentDir, candidate)
		
		// Check if this directory contains pkg/plugin
		pkgPluginDir := filepath.Join(testDir, "pkg", "plugin")
		if _, err := os.Stat(pkgPluginDir); err == nil {
			abs, err := filepath.Abs(testDir)
			if err == nil {
				return abs, nil
			}
		}
	}
	
	// Try to find corynth binary location and work backwards
	execPath, err := os.Executable()
	if err == nil {
		// If corynth binary exists, try its directory and parents
		execDir := filepath.Dir(execPath)
		execCandidates := []string{
			execDir,
			filepath.Join(execDir, ".."),
			filepath.Join(execDir, "../.."),
			filepath.Join(execDir, "../../.."),
		}
		
		for _, candidate := range execCandidates {
			pkgPluginDir := filepath.Join(candidate, "pkg", "plugin")
			if _, err := os.Stat(pkgPluginDir); err == nil {
				abs, err := filepath.Abs(candidate)
				if err == nil {
					return abs, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("corynth source directory not found (looked for pkg/plugin in various locations from %s)", currentDir)
}

// fixPluginGoPaths fixes import paths in plugin.go file
func (m *Manager) fixPluginGoPaths(pluginGoPath string) error {
	// Read plugin.go file
	content, err := os.ReadFile(pluginGoPath)
	if err != nil {
		return err
	}
	
	contentStr := string(content)
	
	// Fix common import path issues
	contentStr = strings.ReplaceAll(contentStr,
		`"github.com/corynth/corynth/src/pkg/plugin"`,
		`"github.com/corynth/corynth/pkg/plugin"`)
	
	// Fix corynth-dist import paths to use corynth
	contentStr = strings.ReplaceAll(contentStr,
		`"github.com/corynth/corynth-dist/pkg/plugin"`,
		`"github.com/corynth/corynth/pkg/plugin"`)
	
	// Write back the fixed content
	return os.WriteFile(pluginGoPath, []byte(contentStr), 0644)
}

// ReflectedPlugin is a wrapper that uses reflection to work around Go plugin interface identity issues
type ReflectedPlugin struct {
	underlying interface{}
	impl       reflect.Value // cached dereferenced implementation
}

func (rp *ReflectedPlugin) testMethods() error {
	// Test if the underlying plugin has the required methods
	val := reflect.ValueOf(rp.underlying)
	
	// Dereference the interface to get the concrete implementation
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		// If it's a pointer to an interface, dereference to get the actual implementation
		elem := val.Elem()
		if elem.Kind() == reflect.Interface && !elem.IsNil() {
			val = elem.Elem() // Get the actual implementation
		}
	}
	
	// Cache the implementation value
	rp.impl = val
	
	// Check for required methods - use the value directly for method lookup
	requiredMethods := []string{"Metadata", "Execute", "Validate", "Actions"}
	for _, methodName := range requiredMethods {
		method := val.MethodByName(methodName)
		if !method.IsValid() {
			return fmt.Errorf("method %s not found", methodName)
		}
	}
	
	return nil
}

func (rp *ReflectedPlugin) Metadata() Metadata {
	method := rp.impl.MethodByName("Metadata")
	
	if !method.IsValid() {
		return Metadata{Name: "unknown", Description: "method not found"}
	}
	
	results := method.Call([]reflect.Value{})
	if len(results) != 1 {
		return Metadata{Name: "unknown"}
	}
	
	// Convert the result to our Metadata struct
	result := results[0].Interface()
	
	// Use reflection to extract fields from the returned metadata
	metaVal := reflect.ValueOf(result)
	metadata := Metadata{}
	
	if metaVal.Kind() == reflect.Struct {
		// Map fields manually to handle potential type differences
		if nameField := metaVal.FieldByName("Name"); nameField.IsValid() && nameField.Kind() == reflect.String {
			metadata.Name = nameField.String()
		}
		if versionField := metaVal.FieldByName("Version"); versionField.IsValid() && versionField.Kind() == reflect.String {
			metadata.Version = versionField.String()
		}
		if descField := metaVal.FieldByName("Description"); descField.IsValid() && descField.Kind() == reflect.String {
			metadata.Description = descField.String()
		}
		if authorField := metaVal.FieldByName("Author"); authorField.IsValid() && authorField.Kind() == reflect.String {
			metadata.Author = authorField.String()
		}
		if licenseField := metaVal.FieldByName("License"); licenseField.IsValid() && licenseField.Kind() == reflect.String {
			metadata.License = licenseField.String()
		}
		if tagsField := metaVal.FieldByName("Tags"); tagsField.IsValid() && tagsField.Kind() == reflect.Slice {
			for i := 0; i < tagsField.Len(); i++ {
				if tagVal := tagsField.Index(i); tagVal.Kind() == reflect.String {
					metadata.Tags = append(metadata.Tags, tagVal.String())
				}
			}
		}
	}
	
	return metadata
}

func (rp *ReflectedPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	method := rp.impl.MethodByName("Execute")
	
	// Prepare arguments
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(action),
		reflect.ValueOf(params),
	}
	
	results := method.Call(args)
	if len(results) != 2 {
		return nil, fmt.Errorf("unexpected number of return values from Execute")
	}
	
	// Extract result and error
	resultMap := results[0].Interface().(map[string]interface{})
	var err error
	if !results[1].IsNil() {
		err = results[1].Interface().(error)
	}
	
	return resultMap, err
}

func (rp *ReflectedPlugin) Validate(params map[string]interface{}) error {
	method := rp.impl.MethodByName("Validate")
	
	args := []reflect.Value{reflect.ValueOf(params)}
	results := method.Call(args)
	
	if len(results) != 1 {
		return fmt.Errorf("unexpected number of return values from Validate")
	}
	
	if results[0].IsNil() {
		return nil
	}
	
	return results[0].Interface().(error)
}

func (rp *ReflectedPlugin) Actions() []Action {
	method := rp.impl.MethodByName("Actions")
	
	results := method.Call([]reflect.Value{})
	if len(results) != 1 {
		return []Action{}
	}
	
	// This is complex to convert via reflection, so return empty for now
	// The plugin will still work for Execute calls
	return []Action{}
}

// healthCheckPlugin performs a basic health check on an installed plugin
func (m *Manager) healthCheckPlugin(pluginName string) error {
	// Find the plugin
	p, exists := m.plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin '%s' not found after installation", pluginName)
	}
	
	// For JSON protocol plugins, we can verify they respond correctly
	if scriptPlugin, ok := p.(*ScriptPlugin); ok {
		// Simple health check - verify metadata
		metadata := scriptPlugin.Metadata()
		if metadata.Name == "" {
			return fmt.Errorf("plugin returned empty metadata")
		}
		
		// Verify plugin name matches
		if metadata.Name != pluginName {
			// Some flexibility for name format
			if !strings.Contains(strings.ToLower(metadata.Name), strings.ToLower(pluginName)) {
				return fmt.Errorf("plugin name mismatch: expected '%s', got '%s'", pluginName, metadata.Name)
			}
		}
		actions := scriptPlugin.Actions()
		if len(actions) > 0 {
			// Try to validate with empty params
			if err := scriptPlugin.Validate(map[string]interface{}{}); err != nil {
				// Validation errors are okay, we're just checking the plugin responds
				_ = err
			}
		}
		
		return nil
	}
	
	// For other plugin types, just check metadata
	metadata := p.Metadata()
	if metadata.Name == "" {
		return fmt.Errorf("plugin returned empty metadata")
	}
	
	return nil
}

// Helper function to check if file is a binary executable
func isExecutableBinary(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first few bytes to check for binary signatures
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return false, err
	}
	header = header[:n]

	// Check for shell script indicators
	if bytes.HasPrefix(header, []byte("#!/")) {
		return false, nil // Shell script, not binary
	}

	// Check for ELF (Linux), Mach-O (macOS), or PE (Windows) binary signatures
	if len(header) >= 4 {
		// ELF magic number
		if bytes.Equal(header[:4], []byte{0x7F, 'E', 'L', 'F'}) {
			return true, nil
		}
		// Mach-O magic numbers (32-bit and 64-bit, little and big endian)
		machO := [][]byte{
			{0xFE, 0xED, 0xFA, 0xCE}, // 32-bit little endian
			{0xCE, 0xFA, 0xED, 0xFE}, // 32-bit big endian
			{0xFE, 0xED, 0xFA, 0xCF}, // 64-bit little endian
			{0xCF, 0xFA, 0xED, 0xFE}, // 64-bit big endian
		}
		for _, magic := range machO {
			if bytes.Equal(header[:4], magic) {
				return true, nil
			}
		}
		// PE (Windows) - check for MZ header
		if len(header) >= 2 && bytes.Equal(header[:2], []byte{'M', 'Z'}) {
			return true, nil
		}
	}

	// Additional check: binary files typically have null bytes
	if bytes.Contains(header, []byte{0x00}) {
		return true, nil
	}

	return false, nil
}

// isExecutableFile checks if a file is executable (binary or script with execute permissions)
func isExecutableFile(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	
	// Check if file has execute permissions
	if info.Mode()&0111 == 0 {
		return false, nil
	}
	
	return true, nil
}

// isShellScript checks if a file is a shell script (starts with shebang)
func isShellScript(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	
	// Read first two bytes to check for shebang
	header := make([]byte, 2)
	_, err = file.Read(header)
	if err != nil {
		return false
	}
	
	return string(header) == "#!"
}

// Helper function to copy file with specific permissions
func copyFileWithPermissions(src, dst string, perm os.FileMode) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Ensure permissions are set correctly
	return os.Chmod(dst, perm)
}

// copyDirectory recursively copies a directory and all its contents
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Calculate relative path and destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		} else {
			// Copy file
			return copyFileWithPermissions(path, destPath, info.Mode())
		}
	})
}

// createPluginWrapper creates a simple wrapper script that calls the real plugin
func createPluginWrapper(wrapperPath, pluginExecutable string) error {
	wrapperContent := fmt.Sprintf(`#!/bin/bash
# Corynth JSON Protocol Plugin Wrapper
exec "%s" "$@"
`, pluginExecutable)
	
	if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil {
		return err
	}
	
	return nil
}

// createSelfContainedWrapper creates a self-contained wrapper script that embeds the Go plugin source
func createSelfContainedWrapper(destPath, srcDir, pluginName string) error {
	// Read the plugin.go source file
	pluginGoPath := filepath.Join(srcDir, "plugin.go")
	pluginGoContent, err := os.ReadFile(pluginGoPath)
	if err != nil {
		return fmt.Errorf("failed to read plugin.go: %w", err)
	}

	// Read go.mod if it exists
	goModPath := filepath.Join(srcDir, "go.mod")
	var goModContent []byte
	if _, err := os.Stat(goModPath); err == nil {
		goModContent, _ = os.ReadFile(goModPath)
	}

	// Create a self-contained shell script that contains the Go source embedded
	wrapperContent := fmt.Sprintf(`#!/bin/bash
# Corynth Self-Contained JSON Protocol Plugin: %s
# This script contains the embedded Go plugin source and compiles/runs it on demand

set -e

PLUGIN_NAME="%s"
TEMP_DIR=$(mktemp -d)
PLUGIN_DIR="$TEMP_DIR/$PLUGIN_NAME"
PLUGIN_BINARY="$PLUGIN_DIR/${PLUGIN_NAME}-plugin"
PLUGIN_GO="$PLUGIN_DIR/plugin.go"
GO_MOD="$PLUGIN_DIR/go.mod"

cleanup() {
    rm -rf "$TEMP_DIR" 2>/dev/null || true
}
trap cleanup EXIT

# Create plugin directory
mkdir -p "$PLUGIN_DIR"

# Write embedded Go source
cat > "$PLUGIN_GO" << 'EOF_PLUGIN_GO'
%s
EOF_PLUGIN_GO

# Write embedded go.mod if available
if [ -n "%s" ]; then
cat > "$GO_MOD" << 'EOF_GO_MOD'
%s
EOF_GO_MOD
fi

# Change to plugin directory
cd "$PLUGIN_DIR"

# Compile the plugin if binary doesn't exist or source is newer
if [[ ! -f "$PLUGIN_BINARY" ]] || [[ "$PLUGIN_GO" -nt "$PLUGIN_BINARY" ]]; then
    if ! go build -o "${PLUGIN_NAME}-plugin" plugin.go 2>&1; then
        echo "Failed to compile plugin" >&2
        exit 1
    fi
fi

# Execute the compiled plugin with all arguments and stdin
exec "$PLUGIN_BINARY" "$@"
`,
		pluginName,                          // Plugin name for comment
		pluginName,                          // PLUGIN_NAME variable
		string(pluginGoContent),             // Embedded plugin.go content
		string(goModContent),                // Check if go.mod content exists
		string(goModContent))                // Embedded go.mod content

	// Write the wrapper script
	if err := os.WriteFile(destPath, []byte(wrapperContent), 0755); err != nil {
		return fmt.Errorf("failed to write wrapper script: %w", err)
	}

	return nil
}

