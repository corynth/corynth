# Corynth Plugin Development Guide

## Quick Start

Creating a new Corynth plugin is simple with the JSON stdin/stdout protocol. Follow this guide to build production-ready plugins.

### 1. Plugin Structure

Every plugin needs these files:
```
my-plugin/
├── plugin          # Bash wrapper script
├── plugin.go       # Go implementation  
├── go.mod          # Go module
└── samples/        # Example workflows (optional)
```

### 2. Bash Wrapper (`plugin`)

```bash
#!/usr/bin/env bash
# Corynth MyPlugin Plugin
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec go run "$DIR/plugin.go" "$@"
```

Make it executable: `chmod +x plugin`

### 3. Go Module (`go.mod`)

```go
module github.com/corynth/plugins/my-plugin

go 1.21

// Add any dependencies you need
require (
    github.com/example/dependency v1.0.0
)
```

### 4. Go Implementation (`plugin.go`)

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
)

// Plugin metadata
type Metadata struct {
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Description string   `json:"description"`
    Author      string   `json:"author"`
    Tags        []string `json:"tags"`
}

// Action input/output specifications
type IOSpec struct {
    Type        string      `json:"type"`
    Required    bool        `json:"required"`
    Default     interface{} `json:"default,omitempty"`
    Description string      `json:"description"`
}

type ActionSpec struct {
    Description string            `json:"description"`
    Inputs      map[string]IOSpec `json:"inputs"`
    Outputs     map[string]IOSpec `json:"outputs"`
}

// Your plugin implementation
type MyPlugin struct {
    // Add any state you need
}

func NewMyPlugin() *MyPlugin {
    return &MyPlugin{}
}

func (p *MyPlugin) GetMetadata() Metadata {
    return Metadata{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "My awesome plugin description",
        Author:      "Your Name",
        Tags:        []string{"category", "functionality"},
    }
}

func (p *MyPlugin) GetActions() map[string]ActionSpec {
    return map[string]ActionSpec{
        "my_action": {
            Description: "Performs my awesome action",
            Inputs: map[string]IOSpec{
                "input_param": {
                    Type:        "string",
                    Required:    true,
                    Description: "Input parameter description",
                },
                "optional_param": {
                    Type:        "number",
                    Required:    false,
                    Default:     42,
                    Description: "Optional parameter with default",
                },
            },
            Outputs: map[string]IOSpec{
                "result": {
                    Type:        "string",
                    Description: "Action result",
                },
                "success": {
                    Type:        "boolean", 
                    Description: "Whether action succeeded",
                },
            },
        },
    }
}

func (p *MyPlugin) Execute(action string, params map[string]interface{}) (map[string]interface{}, error) {
    switch action {
    case "my_action":
        return p.executeMyAction(params)
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}

func (p *MyPlugin) executeMyAction(params map[string]interface{}) (map[string]interface{}, error) {
    // Get required parameter
    inputParam, ok := params["input_param"].(string)
    if !ok || inputParam == "" {
        return map[string]interface{}{
            "error": "input_param is required and must be a string",
        }, nil
    }
    
    // Get optional parameter with default
    optionalParam := 42.0 // default
    if val, ok := params["optional_param"].(float64); ok {
        optionalParam = val
    }
    
    // Perform your action logic here
    result := fmt.Sprintf("Processed '%s' with value %.0f", inputParam, optionalParam)
    
    // Return results
    return map[string]interface{}{
        "result":  result,
        "success": true,
    }, nil
}

// Standard main function for JSON protocol
func main() {
    if len(os.Args) < 2 {
        json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
            "error": "action required",
        })
        os.Exit(1)
    }

    action := os.Args[1]
    plugin := NewMyPlugin()

    var result interface{}

    switch action {
    case "metadata":
        result = plugin.GetMetadata()
    case "actions":
        result = plugin.GetActions()
    default:
        var params map[string]interface{}
        inputData, err := io.ReadAll(os.Stdin)
        if err != nil {
            result = map[string]interface{}{"error": fmt.Sprintf("failed to read input: %v", err)}
        } else if len(inputData) > 0 {
            if err := json.Unmarshal(inputData, &params); err != nil {
                result = map[string]interface{}{"error": fmt.Sprintf("failed to parse JSON: %v", err)}
            } else {
                result, err = plugin.Execute(action, params)
                if err != nil {
                    result = map[string]interface{}{"error": err.Error()}
                }
            }
        } else {
            result, err = plugin.Execute(action, map[string]interface{}{})
            if err != nil {
                result = map[string]interface{}{"error": err.Error()}
            }
        }
    }

    json.NewEncoder(os.Stdout).Encode(result)
}
```

## Testing Your Plugin

### Manual Testing

```bash
# Test metadata
./plugin metadata
{"name":"my-plugin","version":"1.0.0",...}

# Test actions
./plugin actions  
{"my_action":{"description":"...","inputs":{...}}}

# Test execution
echo '{"input_param":"test","optional_param":100}' | ./plugin my_action
{"result":"Processed 'test' with value 100","success":true}
```

### Workflow Testing

Create a test workflow:

```hcl
# test-my-plugin.hcl
workflow "test-my-plugin" {
  description = "Test my awesome plugin"
  version = "1.0.0"

  step "test_action" {
    plugin = "my-plugin"
    action = "my_action"
    
    params = {
      input_param = "hello world"
      optional_param = 123
    }
  }

  step "confirm" {
    plugin = "shell"
    action = "exec"
    depends_on = ["test_action"]
    
    params = {
      command = "echo 'Plugin test completed!'"
    }
  }
}
```

Test with Corynth:
```bash
corynth apply test-my-plugin.hcl
```

## Best Practices

### ✅ Error Handling

Always return errors in JSON format, never panic:

```go
func (p *MyPlugin) executeMyAction(params map[string]interface{}) (map[string]interface{}, error) {
    // Validate required parameters
    if param, ok := params["required_param"].(string); !ok || param == "" {
        return map[string]interface{}{
            "error": "required_param is required and must be a non-empty string",
        }, nil // Return nil error, put error in response
    }
    
    // Handle external operations that might fail
    result, err := someExternalOperation()
    if err != nil {
        return map[string]interface{}{
            "error": fmt.Sprintf("operation failed: %v", err),
        }, nil
    }
    
    return map[string]interface{}{
        "result": result,
        "success": true,
    }, nil
}
```

### ✅ Input Validation

Validate all inputs and provide clear error messages:

```go
// Validate string parameter
func getStringParam(params map[string]interface{}, key string, required bool) (string, error) {
    val, exists := params[key]
    if !exists {
        if required {
            return "", fmt.Errorf("%s parameter is required", key)
        }
        return "", nil
    }
    
    strVal, ok := val.(string)
    if !ok {
        return "", fmt.Errorf("%s must be a string", key)
    }
    
    return strVal, nil
}

// Validate number parameter
func getNumberParam(params map[string]interface{}, key string, defaultVal float64) float64 {
    if val, ok := params[key].(float64); ok {
        return val
    }
    return defaultVal
}
```

### ✅ Type Safety

Use type assertions safely with proper defaults:

```go
func (p *MyPlugin) executeAction(params map[string]interface{}) (map[string]interface{}, error) {
    // Safe string extraction
    url, ok := params["url"].(string)
    if !ok || url == "" {
        return map[string]interface{}{"error": "url is required"}, nil
    }
    
    // Safe number extraction with default
    timeout := 30.0
    if val, ok := params["timeout"].(float64); ok && val > 0 {
        timeout = val
    }
    
    // Safe boolean extraction
    enableFeature := false
    if val, ok := params["enable_feature"].(bool); ok {
        enableFeature = val
    }
    
    // Safe object extraction
    headers := make(map[string]string)
    if headerObj, ok := params["headers"].(map[string]interface{}); ok {
        for key, value := range headerObj {
            if strValue, ok := value.(string); ok {
                headers[key] = strValue
            }
        }
    }
    
    // Use the validated parameters...
}
```

### ✅ Documentation

Document your actions clearly:

```go
func (p *MyPlugin) GetActions() map[string]ActionSpec {
    return map[string]ActionSpec{
        "send_request": {
            Description: "Send HTTP request to specified URL with optional headers and authentication",
            Inputs: map[string]IOSpec{
                "url": {
                    Type:        "string",
                    Required:    true,
                    Description: "Target URL for the HTTP request (must include protocol)",
                },
                "method": {
                    Type:        "string", 
                    Required:    false,
                    Default:     "GET",
                    Description: "HTTP method: GET, POST, PUT, DELETE, etc.",
                },
                "headers": {
                    Type:        "object",
                    Required:    false,
                    Description: "HTTP headers as key-value pairs (e.g. {'Authorization': 'Bearer token'})",
                },
                "timeout": {
                    Type:        "number",
                    Required:    false,
                    Default:     30,
                    Description: "Request timeout in seconds (must be positive)",
                },
            },
            Outputs: map[string]IOSpec{
                "status_code": {
                    Type:        "number",
                    Description: "HTTP response status code (e.g. 200, 404, 500)",
                },
                "content": {
                    Type:        "string", 
                    Description: "Response body content as string",
                },
                "headers": {
                    Type:        "object",
                    Description: "Response headers as key-value pairs",
                },
            },
        },
    }
}
```

## Real-World Examples

Check out these production plugins for reference:

- **HTTP Plugin**: `/official/http/plugin.go` - REST API calls
- **File Plugin**: `/official/file/plugin.go` - File system operations
- **Calculator Plugin**: `/official/calculator/plugin.go` - Math calculations
- **Slack Plugin**: `/official/slack/plugin.go` - Messaging integration

## Distribution

Once your plugin is complete:

1. **Test thoroughly** with real workflows
2. **Add example workflows** in `samples/` directory  
3. **Submit to official repository** via pull request
4. **Update registry.json** with plugin metadata

Your plugin will then be available via:
```bash
corynth plugin install my-plugin
```

## Remote Installation Architecture

### Self-Contained Wrappers

When plugins are installed remotely via `corynth plugin install`, the system creates **self-contained wrapper scripts** that embed your plugin's source code directly. This ensures:

- ✅ **Zero external dependencies** - plugin runs independently  
- ✅ **Automatic compilation** - Go source compiled on-demand
- ✅ **Version consistency** - exact source code from repository
- ✅ **Fast execution** - minimal overhead after first compilation

### Installation Process

1. **Repository Clone**: Corynth clones the plugin repository temporarily
2. **Source Embedding**: Your `plugin.go` and `go.mod` are embedded into a bash wrapper
3. **Self-Contained Creation**: Wrapper script contains all necessary source code
4. **Dynamic Compilation**: On first execution, Go code is compiled to temporary binary
5. **Execution**: Compiled binary executes with full JSON protocol support

### Installation Priority

The plugin installer searches for plugins in this order:

1. **JSON Protocol Scripts** (`plugin`) - **Highest Priority**
2. Platform-specific binaries (`plugin-darwin`, `plugin-linux`)  
3. Generic plugin binaries (`my-plugin-plugin`, `plugin`)
4. Plugin name only (`my-plugin`)

This prioritization ensures JSON protocol plugins (which work reliably) are preferred over compiled binaries that may have compatibility issues.

### Wrapper Script Structure

The generated self-contained wrapper looks like this:

```bash
#!/bin/bash
# Corynth Self-Contained JSON Protocol Plugin: my-plugin

set -e
TEMP_DIR=$(mktemp -d)
trap "rm -rf '$TEMP_DIR'" EXIT

# Embedded source code extracted to temporary directory
# Go compilation happens here if needed
# Binary execution with full argument/stdin forwarding
```

### Development Implications

- **Standard Structure**: Keep using the `plugin`/`plugin.go`/`go.mod` structure
- **Dependency Management**: All go.mod dependencies are preserved and available
- **No Changes Required**: Existing plugins work without modification
- **Performance**: First run compiles (~1s), subsequent runs are immediate

This architecture ensures robust remote plugin installation while maintaining the simplicity of local development.

## Performance Tips

- Keep plugins **lightweight** - they start fresh for each execution
- **Minimize dependencies** - faster startup times
- **Cache expensive operations** when possible
- **Use efficient JSON parsing** for large data sets
- **Validate early** - fail fast on invalid inputs

## Security Considerations

- **Validate all inputs** - never trust workflow parameters
- **Sanitize file paths** - prevent directory traversal
- **Limit network access** - validate URLs and timeouts  
- **Handle secrets properly** - never log sensitive data
- **Use standard libraries** when possible for security-reviewed code

This guide covers everything needed to create production-ready Corynth plugins. The JSON protocol is simple, reliable, and performant for workflow orchestration.