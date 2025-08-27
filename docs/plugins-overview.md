# Plugin System Overview

Corynth's cloud-native plugin system provides a powerful and extensible way to add functionality to your workflows. Focused on modern infrastructure and container orchestration, this guide covers everything you need to know about using and developing plugins.

## Plugin Architecture

### Plugin Types

**Built-in Plugins**:
- Bundled with Corynth installation
- Available immediately after `corynth init`
- No compilation required

**Remote Plugins**:
- Hosted in GitHub repositories
- Automatically downloaded and compiled when first used
- Cached locally for performance

### Plugin Loading Process

1. **Discovery**: Plugin referenced in workflow step
2. **Installation Check**: Check if plugin already installed locally
3. **Auto-Installation**: If not found, download from repository
4. **Compilation**: Compile Go source to shared library (`.so`)
5. **Loading**: Load compiled plugin into Corynth runtime
6. **Caching**: Cache for future use

## Available Plugins

### Built-in Plugins

#### Git Plugin
**Purpose**: Git version control operations  
**Actions**: `clone`, `status`, `commit`

```hcl
step "clone_repo" {
  plugin = "git"
  action = "clone"
  
  params = {
    url = "https://github.com/example/repo.git"
    path = "/tmp/repo"
    branch = "main"
  }
}
```

#### Slack Plugin  
**Purpose**: Slack messaging and notifications  
**Actions**: `message`, `file`

```hcl
step "notify_team" {
  plugin = "slack"
  action = "message"
  
  params = {
    channel = "#deployments"
    text = "Deployment completed successfully!"
    webhook_url = var.slack_webhook
  }
}
```

### Remote Plugins (Auto-installed)

#### HTTP Plugin
**Purpose**: HTTP client for REST API calls  
**Repository**: `https://github.com/corynth/corynthplugins/http`  
**Actions**: `get`, `post`, `put`, `delete`

```hcl
step "api_call" {
  plugin = "http"  # Auto-installed on first use
  action = "get"
  
  params = {
    url = "https://api.github.com/users/octocat"
    timeout = 30
    headers = {
      "User-Agent" = "Corynth/1.2.0"
    }
  }
}
```

#### File Plugin
**Purpose**: File system operations  
**Repository**: `https://github.com/corynth/corynthplugins/file`  
**Actions**: `read`, `write`, `delete`, `exists`

```hcl
step "save_config" {
  plugin = "file"  # Auto-installed on first use
  action = "write"
  
  params = {
    path = "/tmp/config.json"
    content = "${api_call.body}"
  }
}
```

#### Shell Plugin
**Purpose**: Execute shell commands and scripts  
**Repository**: `https://github.com/corynth/corynthplugins/shell`  
**Actions**: `exec`

```hcl
step "build_project" {
  plugin = "shell"  # Auto-installed on first use
  action = "exec"
  
  params = {
    command = "cd /path/to/project && make build"
    timeout = 600
    working_dir = "/path/to/project"
  }
}
```

## Plugin Management

### Installation Commands

```bash
# List all installed plugins
corynth plugin list

# Install specific plugin manually
corynth plugin install http
corynth plugin install file
corynth plugin install shell

# Get plugin information
corynth plugin info git
corynth plugin info http

# Remove plugin (removes compiled binary, not source)
corynth plugin remove http
```

### Plugin Information

```bash
$ corynth plugin info http
Plugin: http
Version: 1.0.0
Description: HTTP client for REST API calls and web requests
Author: Corynth Plugins Team
Tags: http, api, rest, web

Actions:
  • get - Perform HTTP GET request
  • post - Perform HTTP POST request  
  • put - Perform HTTP PUT request
  • delete - Perform HTTP DELETE request

Installation: Auto-installed from GitHub repository
Status: Installed and ready
```

## Plugin Actions Reference

### Git Plugin Actions

#### clone
**Description**: Clone a Git repository  
**Parameters**:
- `url` (string, required): Repository URL
- `path` (string, optional): Local clone path
- `branch` (string, optional): Branch to clone

**Returns**:
- `path` (string): Local repository path
- `commit` (string): Latest commit hash

#### status
**Description**: Get repository status  
**Parameters**:
- `path` (string, optional): Repository path (default: ".")

**Returns**:
- `clean` (bool): Whether repository is clean
- `branch` (string): Current branch name

### HTTP Plugin Actions

#### get
**Description**: Perform HTTP GET request  
**Parameters**:
- `url` (string, required): Request URL
- `timeout` (number, optional): Timeout in seconds (default: 30)
- `headers` (object, optional): Request headers

**Returns**:
- `status_code` (number): HTTP status code
- `body` (string): Response body
- `headers` (object): Response headers

#### post
**Description**: Perform HTTP POST request  
**Parameters**:
- `url` (string, required): Request URL
- `body` (string, optional): Request body
- `headers` (object, optional): Request headers
- `timeout` (number, optional): Timeout in seconds

**Returns**:
- `status_code` (number): HTTP status code
- `body` (string): Response body

### File Plugin Actions

#### read
**Description**: Read file contents  
**Parameters**:
- `path` (string, required): File path to read

**Returns**:
- `content` (string): File contents
- `size` (number): File size in bytes
- `modified` (number): Last modified timestamp

#### write
**Description**: Write content to file  
**Parameters**:
- `path` (string, required): File path to write
- `content` (string, required): Content to write
- `mode` (string, optional): File permissions (default: "0644")

**Returns**:
- `success` (bool): Whether write succeeded
- `size` (number): Bytes written

### Shell Plugin Actions

#### exec
**Description**: Execute shell command  
**Parameters**:
- `command` (string, required): Command to execute
- `working_dir` (string, optional): Working directory
- `timeout` (number, optional): Timeout in seconds (default: 300)
- `env` (object, optional): Environment variables

**Returns**:
- `output` (string): Command output
- `exit_code` (number): Exit code
- `success` (bool): Whether command succeeded

## Plugin Development

### Quick Start

Generate a new plugin scaffold:

```bash
corynth plugin init my-plugin --type http --author "Your Name"
cd my-plugin
```

This creates a complete plugin structure with:
- Plugin implementation (`plugin.go`)
- Comprehensive tests (`plugin_test.go`)
- Documentation (`README.md`)
- Sample workflows (`samples/`)
- Build automation (`Makefile`)

### Plugin Interface

Every plugin must implement:

```go
type Plugin interface {
    Metadata() plugin.Metadata
    Actions() []plugin.Action
    Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error)
    Validate(params map[string]interface{}) error
}
```

### Example Implementation

```go
package main

import (
    "context"
    "fmt"
    "github.com/corynth/corynth/pkg/plugin"
)

type MyPlugin struct{}

func (p *MyPlugin) Metadata() plugin.Metadata {
    return plugin.Metadata{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "Description of plugin functionality",
        Author:      "Your Name",
        Tags:        []string{"category", "functionality"},
        License:     "Apache-2.0",
    }
}

func (p *MyPlugin) Actions() []plugin.Action {
    return []plugin.Action{
        {
            Name:        "my-action",
            Description: "What this action does",
            Inputs: map[string]plugin.InputSpec{
                "param1": {
                    Type:        "string",
                    Description: "Parameter description",
                    Required:    true,
                },
            },
            Outputs: map[string]plugin.OutputSpec{
                "result": {
                    Type:        "string",
                    Description: "Action result",
                },
            },
        },
    }
}

func (p *MyPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
    switch action {
    case "my-action":
        return p.executeMyAction(ctx, params)
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}

func (p *MyPlugin) Validate(params map[string]interface{}) error {
    return nil
}

func (p *MyPlugin) executeMyAction(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    param1, ok := params["param1"].(string)
    if !ok {
        return nil, fmt.Errorf("param1 parameter is required")
    }
    
    // Implementation logic here
    
    return map[string]interface{}{
        "result": fmt.Sprintf("Processed: %s", param1),
    }, nil
}

// Required: Export the plugin
var ExportedPlugin MyPlugin
```

## Plugin Registry

### Plugin Repository Structure

The official plugin repository at `https://github.com/corynth/corynthplugins` contains:

```
corynthplugins/
├── http/                    # HTTP client plugin
├── file/                    # File operations plugin  
├── shell/                   # Shell execution plugin
├── calculator/              # Math operations plugin
├── docker/                  # Docker operations plugin
├── terraform/               # Terraform execution plugin
├── vault/                   # HashiCorp Vault plugin
├── mysql/                   # MySQL database plugin
├── redis/                   # Redis operations plugin
└── registry.json           # Plugin registry metadata
```

### Adding Your Plugin

1. **Fork** the corynthplugins repository
2. **Create** your plugin directory
3. **Implement** following the interface requirements
4. **Test** thoroughly with sample workflows
5. **Document** with comprehensive README
6. **Submit** pull request for review

### Plugin Registry Entry

Add your plugin to `registry.json`:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Brief description of plugin functionality",
  "author": "Your Name",
  "format": "source",
  "installation": "Automatically compiled when first used",
  "tags": ["category", "functionality", "integration"],
  "actions": [
    {
      "name": "action-name",
      "description": "What this action does",
      "example": "Usage example"
    }
  ],
  "requirements": {"corynth": ">=1.2.0"}
}
```

## Advanced Plugin Features

### Parameter Validation

```go
func (p *MyPlugin) Validate(params map[string]interface{}) error {
    if url, ok := params["url"].(string); !ok || url == "" {
        return fmt.Errorf("url parameter is required")
    }
    
    if timeout, ok := params["timeout"].(float64); ok {
        if timeout <= 0 {
            return fmt.Errorf("timeout must be positive")
        }
    }
    
    return nil
}
```

### Context Handling

```go
func (p *MyPlugin) executeWithTimeout(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    // Respect context cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Continue processing
    }
    
    // Long operation with context
    return p.performOperation(ctx, params)
}
```

### Error Handling

```go
func (p *MyPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
    result, err := p.performAction(ctx, action, params)
    if err != nil {
        // Return context-rich errors
        return nil, fmt.Errorf("action %s failed: %w", action, err)
    }
    
    return result, nil
}
```

## Next Steps

- [Plugin Development Guide](development.md) - Detailed development documentation
- [Built-in Plugin Reference](builtin-plugins.md) - Complete reference for built-in plugins
- [Remote Plugin Reference](remote-plugins.md) - Remote plugin documentation
- [Example Plugins](../examples/plugins/) - Real-world plugin examples

---

**Extend Corynth's capabilities with the powerful plugin system!** 🔌