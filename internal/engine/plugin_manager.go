package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"time"

	"github.com/corynth/corynth/plugins/core/ansible"
	"github.com/corynth/corynth/plugins/core/git"
	"github.com/corynth/corynth/plugins/core/shell"
)

// PluginManagerImpl is responsible for managing plugins
type PluginManagerImpl struct {
	corePlugins map[string]Plugin
	remotePlugins map[string]Plugin
	pluginDir string
	downloader *PluginDownloader
}

// NewPluginManager creates a new PluginManagerImpl
func NewPluginManager(pluginDir string) *PluginManagerImpl {
	return &PluginManagerImpl{
		corePlugins: make(map[string]Plugin),
		remotePlugins: make(map[string]Plugin),
		pluginDir: pluginDir,
		downloader: NewPluginDownloader(pluginDir),
	}
}

// Initialize initializes the plugin manager
func (m *PluginManagerImpl) Initialize() error {
	// Register core plugins
	m.registerCorePlugins()

	// Load remote plugins
	if err := m.loadRemotePlugins(); err != nil {
		return err
	}

	return nil
}

// registerCorePlugins registers the core plugins
func (m *PluginManagerImpl) registerCorePlugins() {
	// Register Git plugin
	m.corePlugins["git"] = &gitPluginAdapter{plugin: git.NewGitPlugin()}

	// Register Shell plugin
	m.corePlugins["shell"] = &shellPluginAdapter{plugin: shell.NewShellPlugin()}

	// Register Ansible plugin
	m.corePlugins["ansible"] = &ansiblePluginAdapter{plugin: ansible.NewAnsiblePlugin()}
}

// gitPluginAdapter adapts the Git plugin to the Plugin interface
type gitPluginAdapter struct {
	plugin *git.GitPlugin
}

func (a *gitPluginAdapter) Name() string {
	return a.plugin.Name()
}

func (a *gitPluginAdapter) Execute(action string, params map[string]interface{}) (Result, error) {
	result, err := a.plugin.Execute(action, params)
	if err != nil {
		return Result{
			Status:    result.Status,
			Output:    result.Output,
			Error:     result.Error,
			StartTime: result.StartTime,
			EndTime:   result.EndTime,
			Duration:  result.Duration,
		}, err
	}

	return Result{
		Status:    result.Status,
		Output:    result.Output,
		Error:     result.Error,
		StartTime: result.StartTime,
		EndTime:   result.EndTime,
		Duration:  result.Duration,
	}, nil
}

// shellPluginAdapter adapts the Shell plugin to the Plugin interface
type shellPluginAdapter struct {
	plugin *shell.ShellPlugin
}

func (a *shellPluginAdapter) Name() string {
	return a.plugin.Name()
}

func (a *shellPluginAdapter) Execute(action string, params map[string]interface{}) (Result, error) {
	result, err := a.plugin.Execute(action, params)
	if err != nil {
		return Result{
			Status:    result.Status,
			Output:    result.Output,
			Error:     result.Error,
			StartTime: result.StartTime,
			EndTime:   result.EndTime,
			Duration:  result.Duration,
		}, err
	}

	return Result{
		Status:    result.Status,
		Output:    result.Output,
		Error:     result.Error,
		StartTime: result.StartTime,
		EndTime:   result.EndTime,
		Duration:  result.Duration,
	}, nil
}

// ansiblePluginAdapter adapts the Ansible plugin to the Plugin interface
type ansiblePluginAdapter struct {
	plugin *ansible.AnsiblePlugin
}

func (a *ansiblePluginAdapter) Name() string {
	return a.plugin.Name()
}

func (a *ansiblePluginAdapter) Execute(action string, params map[string]interface{}) (Result, error) {
	result, err := a.plugin.Execute(action, params)
	if err != nil {
		return Result{
			Status:    result.Status,
			Output:    result.Output,
			Error:     result.Error,
			StartTime: result.StartTime,
			EndTime:   result.EndTime,
			Duration:  result.Duration,
		}, err
	}

	return Result{
		Status:    result.Status,
		Output:    result.Output,
		Error:     result.Error,
		StartTime: result.StartTime,
		EndTime:   result.EndTime,
		Duration:  result.Duration,
	}, nil
}

// loadRemotePlugins loads remote plugins from the plugin directory
func (m *PluginManagerImpl) loadRemotePlugins() error {
	// Look for plugins.yaml in the current directory
	manifestPath := "plugins.yaml"
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// If not found, try in the plugin directory
		manifestPath = filepath.Join(m.pluginDir, "plugins.yaml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			// No manifest found, nothing to do
			return nil
		}
	}

	// Load the plugin manifest
	manifest, err := m.downloader.LoadPluginManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("error loading plugin manifest: %w", err)
	}

	// Download and load each plugin
	for _, pluginInfo := range manifest.Plugins {
		// Check if the plugin is already loaded
		if _, ok := m.remotePlugins[pluginInfo.Name]; ok {
			continue
		}

		// Check if the plugin is already downloaded
		pluginPath := filepath.Join(m.pluginDir, pluginInfo.Name, pluginInfo.Name+".so")
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			// Plugin not downloaded, download it
			if err := m.downloader.DownloadPlugin(pluginInfo); err != nil {
				fmt.Printf("Warning: error downloading plugin %s: %s\n", pluginInfo.Name, err)
				continue
			}
		}

		// Load the plugin
		plugin, err := m.LoadPlugin(pluginInfo.Name)
		if err != nil {
			fmt.Printf("Warning: error loading plugin %s: %s\n", pluginInfo.Name, err)
			continue
		}

		// Store the plugin
		m.remotePlugins[pluginInfo.Name] = plugin
	}

	return nil
}

// LoadPlugin loads a plugin by name
func (m *PluginManagerImpl) LoadPlugin(name string) (Plugin, error) {
	// Check if it's a core plugin
	if plugin, ok := m.corePlugins[name]; ok {
		return plugin, nil
	}

	// Check if it's a remote plugin
	if plugin, ok := m.remotePlugins[name]; ok {
		return plugin, nil
	}

	// Try to load it as a remote plugin
	pluginPath := filepath.Join(m.pluginDir, name, name+".so")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		// Try the old path format for backward compatibility
		pluginPath = filepath.Join(m.pluginDir, name+".so")
	}
	
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("error loading plugin %s: %w", name, err)
	}

	// Look up the "New" symbol
	newSymbol, err := p.Lookup("New")
	if err != nil {
		return nil, fmt.Errorf("error looking up New symbol in plugin %s: %w", name, err)
	}

	// Call the New function to create a new plugin instance
	newFunc, ok := newSymbol.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("New symbol in plugin %s is not a function", name)
	}

	pluginInstance := newFunc()
	m.remotePlugins[name] = pluginInstance

	return pluginInstance, nil
}

// ExecutePluginAction executes a plugin action
func (m *PluginManagerImpl) ExecutePluginAction(plugin Plugin, action string, params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Execute the plugin action
	result, err := plugin.Execute(action, params)
	if err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error executing plugin action: %s", err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing plugin action: %w", err)
	}

	return result, nil
}

// GetAvailablePlugins returns a list of available plugins
func (m *PluginManagerImpl) GetAvailablePlugins() []string {
	var plugins []string

	// Add core plugins
	for name := range m.corePlugins {
		plugins = append(plugins, name)
	}

	// Add remote plugins
	for name := range m.remotePlugins {
		plugins = append(plugins, name)
	}

	return plugins
}