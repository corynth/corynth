package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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
			// Load gRPC plugin executable
			if err := m.loadGRPCPlugin(entryPath); err != nil {
				fmt.Printf("Warning: failed to load gRPC plugin from %s: %v\n", entryPath, err)
			}
		}
	}

	return nil
}

// loadBuiltinPlugins loads built-in plugins
func (m *Manager) loadBuiltinPlugins() {
	// Load only essential plugins by default
	// Only shell plugin is built-in - all others are remote RPC/subprocess plugins
	m.plugins["shell"] = NewShellPlugin()
	
	// All other plugins (http, file, git, slack, reporting) will be loaded as remote RPC plugins on demand
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
	// Check for plugin.yaml or plugin.yml
	metadataPath := filepath.Join(path, "plugin.yaml")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		metadataPath = filepath.Join(path, "plugin.yml")
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
			cloneOpts.Auth = &http.BasicAuth{
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
			pullOpts.Auth = &http.BasicAuth{
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

	// Look for plugin directory containing plugin.go (most common pattern)
	pluginDirPath := filepath.Join(repoPath, pluginName)
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
		return fmt.Errorf("plugin '%s' not found in repository (looked for .so file, %s/plugin.go, %s.go, and plugins/%s/)", pluginName, pluginName, pluginName, pluginName)
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

// loadGRPCPlugin loads a gRPC plugin executable (simplified for now)
func (m *Manager) loadGRPCPlugin(executablePath string) error {
	// Extract plugin name from executable path
	fileName := filepath.Base(executablePath)
	if !strings.HasPrefix(fileName, "corynth-plugin-") {
		return fmt.Errorf("invalid gRPC plugin name format: %s", fileName)
	}
	
	pluginName := strings.TrimPrefix(fileName, "corynth-plugin-")
	
	// For now, create a simple gRPC plugin placeholder
	// This avoids the import cycle while maintaining functionality
	grpcPlugin := &SimpleGRPCPlugin{
		metadata: Metadata{
			Name:        pluginName,
			Description: fmt.Sprintf("gRPC plugin: %s", pluginName),
			Version:     "1.0.0",
		},
		executablePath: executablePath,
	}
	
	// Store the plugin
	m.plugins[pluginName] = grpcPlugin
	
	return nil
}