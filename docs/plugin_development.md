# 🔌 Plugin Development Guide

## 🌟 Introduction

Corynth's plugin system allows you to extend its functionality by creating custom plugins. This guide will walk you through the process of developing your own plugins for Corynth.

## 🏗️ Plugin Structure

A Corynth plugin consists of the following components:

- **Plugin Interface**: Implements the Plugin interface defined in the engine package
- **Actions**: Functions that perform specific operations
- **Results**: Return values from actions

## 🧩 Plugin Interface

Every plugin must implement the following interface:

```go
type Plugin interface {
    Name() string
    Execute(action string, params map[string]interface{}) (Result, error)
}
```

## 📝 Creating a Basic Plugin

### 1️⃣ Define Your Plugin Structure

```go
package myplugin

import (
    "fmt"
    "time"
)

// MyPlugin is a custom plugin
type MyPlugin struct{}

// Result represents the result of a step execution
type Result struct {
    Status    string
    Output    string
    Error     string
    StartTime time.Time
    EndTime   time.Time
    Duration  time.Duration
}

// NewMyPlugin creates a new MyPlugin
func NewMyPlugin() *MyPlugin {
    return &MyPlugin{}
}
```

### 2️⃣ Implement the Plugin Interface

```go
// Name returns the name of the plugin
func (p *MyPlugin) Name() string {
    return "myplugin"
}

// Execute executes a plugin action
func (p *MyPlugin) Execute(action string, params map[string]interface{}) (Result, error) {
    startTime := time.Now()

    switch action {
    case "action1":
        return p.action1(params)
    case "action2":
        return p.action2(params)
    default:
        endTime := time.Now()
        return Result{
            Status:    "error",
            Error:     fmt.Sprintf("unknown action: %s", action),
            StartTime: startTime,
            EndTime:   endTime,
            Duration:  endTime.Sub(startTime),
        }, fmt.Errorf("unknown action: %s", action)
    }
}
```

### 3️⃣ Implement Actions

```go
// action1 performs the first action
func (p *MyPlugin) action1(params map[string]interface{}) (Result, error) {
    startTime := time.Now()

    // Get parameters
    param1, ok := params["param1"].(string)
    if !ok {
        endTime := time.Now()
        return Result{
            Status:    "error",
            Error:     "param1 parameter is required",
            StartTime: startTime,
            EndTime:   endTime,
            Duration:  endTime.Sub(startTime),
        }, fmt.Errorf("param1 parameter is required")
    }

    // Perform action
    output := fmt.Sprintf("Executed action1 with param1: %s", param1)

    endTime := time.Now()
    return Result{
        Status:    "success",
        Output:    output,
        StartTime: startTime,
        EndTime:   endTime,
        Duration:  endTime.Sub(startTime),
    }, nil
}
```

## 🔄 Plugin Registration

### Core Plugins

Core plugins are built into Corynth and are registered in the `plugin_manager.go` file:

```go
// registerCorePlugins registers the core plugins
func (m *PluginManagerImpl) registerCorePlugins() {
    // Register existing plugins
    m.corePlugins["git"] = &gitPluginAdapter{plugin: git.NewGitPlugin()}
    m.corePlugins["shell"] = &shellPluginAdapter{plugin: shell.NewShellPlugin()}
    m.corePlugins["ansible"] = &ansiblePluginAdapter{plugin: ansible.NewAnsiblePlugin()}
    
    // Register your new plugin
    m.corePlugins["myplugin"] = &myPluginAdapter{plugin: myplugin.NewMyPlugin()}
}
```

### Remote Plugins

Remote plugins are loaded dynamically from the plugin directory. To create a remote plugin:

1. Build your plugin as a Go plugin:
   ```bash
   go build -buildmode=plugin -o myplugin.so myplugin.go
   ```

2. Add your plugin to the plugin manifest:
   ```yaml
   plugins:
     - name: "myplugin"
       repository: "https://github.com/user/plugins"
       version: "v1.0.0"
       path: "myplugin"
   ```

## 🧪 Testing Your Plugin

Create a flow that uses your plugin:

```yaml
flow:
  name: "test_plugin_flow"
  description: "Test my custom plugin"
  steps:
    - name: "test_step"
      plugin: "myplugin"
      action: "action1"
      params:
        param1: "test value"
```

Run the flow:

```bash
corynth init test_plugin
cp test_plugin_flow.yaml test_plugin/flows/
corynth plan test_plugin
corynth apply test_plugin
```

## 📦 Plugin Distribution

To distribute your plugin:

1. Host your plugin in a Git repository
2. Tag releases with semantic versioning
3. Update the plugin manifest to point to your repository

## 🔍 Best Practices

- **Error Handling**: Always return meaningful error messages
- **Parameter Validation**: Validate all parameters before use
- **Documentation**: Document all actions and parameters
- **Testing**: Write tests for your plugin
- **Versioning**: Use semantic versioning for your plugin
- **Security**: Avoid executing arbitrary code from user input

## 🌐 Community Plugins

Share your plugins with the Corynth community by:

1. Publishing your plugin repository
2. Adding documentation
3. Submitting a pull request to the Corynth plugin registry

## 🛠️ Advanced Topics

### Plugin Dependencies

If your plugin depends on external libraries or tools:

1. Document the dependencies
2. Check for dependencies in your plugin's initialization
3. Provide clear error messages if dependencies are missing

### Plugin Configuration

For plugins that require configuration:

1. Use the plugin parameters for per-action configuration
2. Use environment variables for global configuration
3. Document all configuration options

### Plugin Lifecycle

Implement initialization and cleanup if needed:

```go
// Initialize initializes the plugin
func (p *MyPlugin) Initialize() error {
    // Perform initialization
    return nil
}

// Cleanup cleans up resources
func (p *MyPlugin) Cleanup() error {
    // Clean up resources
    return nil
}
```

Call these methods in your adapter:

```go
func (a *myPluginAdapter) Execute(action string, params map[string]interface{}) (engine.Result, error) {
    // Initialize plugin
    if err := a.plugin.Initialize(); err != nil {
        return engine.Result{
            Status: "error",
            Error:  fmt.Sprintf("error initializing plugin: %s", err),
        }, err
    }

    // Execute action
    result, err := a.plugin.Execute(action, params)

    // Clean up
    if cleanupErr := a.plugin.Cleanup(); cleanupErr != nil {
        fmt.Printf("Warning: error cleaning up plugin: %s\n", cleanupErr)
    }

    return engine.Result{
        Status:    result.Status,
        Output:    result.Output,
        Error:     result.Error,
        StartTime: result.StartTime,
        EndTime:   result.EndTime,
        Duration:  result.Duration,
    }, err
}
```

## 🤝 Contributing

We welcome contributions to the Corynth plugin ecosystem! Please follow these guidelines:

1. Follow the plugin development guidelines
2. Write tests for your plugin
3. Document your plugin thoroughly
4. Submit a pull request to the Corynth repository