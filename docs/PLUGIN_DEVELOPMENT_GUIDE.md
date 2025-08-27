# Plugin Development Guide

Comprehensive guide for developing, testing, and distributing Corynth plugins with best practices, troubleshooting, and real-world examples.

## Overview

Corynth uses a git-based plugin architecture where plugins are Go modules compiled as shared libraries. This guide covers everything needed to create production-ready plugins.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Plugin Architecture](#plugin-architecture)
3. [Development Environment](#development-environment)
4. [Plugin Interface](#plugin-interface)
5. [Implementation Patterns](#implementation-patterns)
6. [Testing Framework](#testing-framework)
7. [Documentation Standards](#documentation-standards)
8. [Distribution Process](#distribution-process)
9. [Best Practices](#best-practices)
10. [Examples](#examples)
11. [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites
- Go 1.21 or later
- Git
- Basic understanding of Go interfaces

### Create Your First Plugin

1. **Clone the plugin template**:
```bash
git clone https://github.com/corynth/corynthplugins.git
cd corynthplugins
mkdir my-plugin
cd my-plugin
```

2. **Create the basic structure**:
```
my-plugin/
├── plugin.go          # Main plugin implementation
├── go.mod            # Go module definition
├── README.md         # Plugin documentation
└── samples/          # Sample workflows
    └── example.hcl
```

3. **Initialize Go module**:
```bash
go mod init github.com/corynth/corynthplugins/my-plugin
go mod tidy
```

## Plugin Architecture

### Core Concepts

Corynth plugins are:
- **Go shared libraries** compiled with `-buildmode=plugin`
- **Git-distributed** from repositories
- **Lazy-loaded** on first use
- **Interface-compliant** with the Corynth Plugin API

### Plugin Lifecycle

1. **Discovery**: Plugin requested in workflow
2. **Download**: Git clone from repository
3. **Compilation**: On-demand build as shared library
4. **Loading**: Dynamic loading into Corynth runtime
5. **Execution**: Action invocation with parameters
6. **Cleanup**: Resource management and cleanup

## Development Environment

### Required Dependencies

Add to your `go.mod`:
```go
module github.com/corynth/corynthplugins/my-plugin

go 1.21

require (
    github.com/corynth/corynth v1.0.0
)
```

### Directory Structure
```
my-plugin/
├── plugin.go                    # Plugin implementation
├── go.mod                      # Module definition
├── README.md                   # Plugin documentation
├── samples/                    # Example workflows
│   ├── basic-usage.hcl
│   └── advanced-features.hcl
└── tests/                      # Test files (optional)
    └── plugin_test.go
```

## Plugin Interface

### Required Interface Implementation

Every plugin must implement the `Plugin` interface:

```go
package main

import (
    "context"
    "github.com/corynth/corynth/pkg/plugin"
)

type MyPlugin struct{}

// Metadata returns plugin information
func (p *MyPlugin) Metadata() plugin.Metadata {
    return plugin.Metadata{
        Name:        "my-plugin",
        Version:     "1.0.0", 
        Description: "Description of what this plugin does",
        Author:      "Your Name",
        Tags:        []string{"tag1", "tag2", "category"},
        License:     "Apache-2.0",
    }
}

// Actions returns available plugin actions
func (p *MyPlugin) Actions() []plugin.Action {
    return []plugin.Action{
        {
            Name:        "my-action",
            Description: "Description of the action",
            Inputs:      getInputSpecs(),
            Outputs:     getOutputSpecs(),
        },
    }
}

// Execute performs the plugin action
func (p *MyPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
    switch action {
    case "my-action":
        return p.executeMyAction(ctx, params)
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}

// Validate checks parameter validity
func (p *MyPlugin) Validate(params map[string]interface{}) error {
    // Implement parameter validation
    return nil
}

// Required: Export the plugin
var ExportedPlugin MyPlugin
```

### Input/Output Specifications

Define clear parameter specifications:

```go
func getInputSpecs() map[string]plugin.InputSpec {
    return map[string]plugin.InputSpec{
        "required_param": {
            Type:        "string",
            Description: "A required string parameter",
            Required:    true,
        },
        "optional_param": {
            Type:        "number",
            Description: "An optional number with default",
            Required:    false,
            Default:     42,
        },
        "list_param": {
            Type:        "array",
            Description: "A list of items",
            Required:    false,
        },
        "object_param": {
            Type:        "object", 
            Description: "A nested object parameter",
            Required:    false,
        },
    }
}

func getOutputSpecs() map[string]plugin.OutputSpec {
    return map[string]plugin.OutputSpec{
        "result": {
            Type:        "string",
            Description: "The operation result",
        },
        "status_code": {
            Type:        "number",
            Description: "Status code of the operation",
        },
        "metadata": {
            Type:        "object",
            Description: "Additional metadata",
        },
    }
}
```

## Implementation Patterns

### 1. HTTP Client Plugin Pattern

```go
func (p *HttpPlugin) executeRequest(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    // Extract parameters
    url, ok := params["url"].(string)
    if !ok {
        return nil, fmt.Errorf("url parameter is required")
    }
    
    method := getStringParam(params, "method", "GET")
    timeout := getIntParam(params, "timeout", 30)
    
    // Create HTTP client with timeout
    client := &http.Client{
        Timeout: time.Duration(timeout) * time.Second,
    }
    
    // Create request
    req, err := http.NewRequestWithContext(ctx, method, url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Add headers if provided
    if headers, ok := params["headers"].(map[string]interface{}); ok {
        for key, value := range headers {
            req.Header.Set(key, fmt.Sprintf("%v", value))
        }
    }
    
    // Execute request
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    // Return structured response
    return map[string]interface{}{
        "status_code": resp.StatusCode,
        "body":        string(body),
        "headers":     resp.Header,
    }, nil
}
```

### 2. Command Execution Plugin Pattern

```go
func (p *ShellPlugin) executeCommand(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    command, ok := params["command"].(string)
    if !ok {
        return nil, fmt.Errorf("command parameter is required")
    }
    
    // Get optional parameters
    workingDir := getStringParam(params, "working_dir", "")
    timeout := getIntParam(params, "timeout", 300)
    shell := getStringParam(params, "shell", "/bin/bash")
    
    // Create context with timeout
    timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
    defer cancel()
    
    // Create command
    var cmd *exec.Cmd
    if shell != "" {
        cmd = exec.CommandContext(timeoutCtx, shell, "-c", command)
    } else {
        parts := strings.Fields(command)
        cmd = exec.CommandContext(timeoutCtx, parts[0], parts[1:]...)
    }
    
    // Set working directory
    if workingDir != "" {
        cmd.Dir = workingDir
    }
    
    // Set environment variables
    cmd.Env = os.Environ()
    if env, ok := params["env"].(map[string]interface{}); ok {
        for key, value := range env {
            cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", key, value))
        }
    }
    
    // Execute command
    output, err := cmd.CombinedOutput()
    exitCode := 0
    if err != nil {
        if exitError, ok := err.(*exec.ExitError); ok {
            exitCode = exitError.ExitCode()
        } else {
            return nil, fmt.Errorf("command execution failed: %w", err)
        }
    }
    
    return map[string]interface{}{
        "output":    string(output),
        "exit_code": exitCode,
        "success":   exitCode == 0,
    }, nil
}
```

### 3. File Operations Plugin Pattern

```go
func (p *FilePlugin) executeRead(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    path, ok := params["path"].(string)
    if !ok {
        return nil, fmt.Errorf("path parameter is required")
    }
    
    // Security check - prevent path traversal
    if strings.Contains(path, "..") {
        return nil, fmt.Errorf("path traversal not allowed")
    }
    
    // Read file
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    // Get file info
    info, err := os.Stat(path)
    if err != nil {
        return nil, fmt.Errorf("failed to get file info: %w", err)
    }
    
    return map[string]interface{}{
        "content":  string(content),
        "size":     info.Size(),
        "modified": info.ModTime().Unix(),
        "mode":     info.Mode().String(),
    }, nil
}
```

### 4. Database Plugin Pattern

```go
func (p *DatabasePlugin) executeQuery(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    query, ok := params["query"].(string)
    if !ok {
        return nil, fmt.Errorf("query parameter is required")
    }
    
    // Get connection parameters
    connectionString := getStringParam(params, "connection_string", "")
    if connectionString == "" {
        return nil, fmt.Errorf("connection_string is required")
    }
    
    // Open database connection
    db, err := sql.Open("mysql", connectionString)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    defer db.Close()
    
    // Set connection context
    db = db.WithContext(ctx)
    
    // Execute query
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("query execution failed: %w", err)
    }
    defer rows.Close()
    
    // Get column names
    columns, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("failed to get columns: %w", err)
    }
    
    // Read results
    var results []map[string]interface{}
    for rows.Next() {
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))
        for i := range columns {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }
        
        row := make(map[string]interface{})
        for i, col := range columns {
            row[col] = values[i]
        }
        results = append(results, row)
    }
    
    return map[string]interface{}{
        "rows":    results,
        "count":   len(results),
        "columns": columns,
    }, nil
}
```

## Testing Framework

### Unit Testing

Create comprehensive tests for your plugin:

```go
package main

import (
    "context"
    "testing"
    "time"
    
    "github.com/corynth/corynth/pkg/plugin"
)

func TestPluginMetadata(t *testing.T) {
    p := &MyPlugin{}
    meta := p.Metadata()
    
    if meta.Name == "" {
        t.Error("Plugin name cannot be empty")
    }
    
    if meta.Version == "" {
        t.Error("Plugin version cannot be empty")
    }
    
    if len(meta.Tags) == 0 {
        t.Error("Plugin should have at least one tag")
    }
}

func TestPluginActions(t *testing.T) {
    p := &MyPlugin{}
    actions := p.Actions()
    
    if len(actions) == 0 {
        t.Error("Plugin should have at least one action")
    }
    
    for _, action := range actions {
        if action.Name == "" {
            t.Error("Action name cannot be empty")
        }
        if action.Description == "" {
            t.Error("Action description cannot be empty")
        }
    }
}

func TestPluginExecution(t *testing.T) {
    p := &MyPlugin{}
    ctx := context.Background()
    
    params := map[string]interface{}{
        "required_param": "test_value",
        "optional_param": 42,
    }
    
    result, err := p.Execute(ctx, "my-action", params)
    if err != nil {
        t.Fatalf("Plugin execution failed: %v", err)
    }
    
    if result == nil {
        t.Error("Plugin should return a result")
    }
}

func TestPluginValidation(t *testing.T) {
    p := &MyPlugin{}
    
    // Test valid parameters
    validParams := map[string]interface{}{
        "required_param": "test",
    }
    
    if err := p.Validate(validParams); err != nil {
        t.Errorf("Valid parameters should pass validation: %v", err)
    }
    
    // Test invalid parameters
    invalidParams := map[string]interface{}{
        "wrong_param": "test",
    }
    
    if err := p.Validate(invalidParams); err == nil {
        t.Error("Invalid parameters should fail validation")
    }
}

func TestPluginTimeout(t *testing.T) {
    p := &MyPlugin{}
    
    // Create context with short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    params := map[string]interface{}{
        "sleep_duration": 1000, // Sleep longer than timeout
    }
    
    _, err := p.Execute(ctx, "sleep-action", params)
    if err == nil {
        t.Error("Plugin should respect context timeout")
    }
}
```

### Integration Testing

Test the plugin with Corynth workflows:

```bash
# Create test workflow
cat > test-workflow.hcl << EOF
workflow "test-my-plugin" {
  description = "Test my custom plugin"
  version     = "1.0.0"

  step "test_action" {
    plugin = "my-plugin"
    action = "my-action"
    
    params = {
      required_param = "test_value"
      optional_param = 42
    }
  }
}
EOF

# Test with Corynth
corynth validate test-workflow.hcl
corynth plan test-workflow.hcl
corynth apply test-workflow.hcl --auto-approve
```

## Documentation Standards

### README.md Template

```markdown
# My Plugin

Brief description of what the plugin does and its primary use cases.

## Installation

This plugin is automatically downloaded and compiled when first used in a Corynth workflow.

## Actions

### my-action

Description of the action and what it accomplishes.

**Parameters:**
- `required_param` (string, required): Description of the parameter
- `optional_param` (number, optional, default: 42): Description with default
- `list_param` (array, optional): Description of list parameter

**Returns:**
- `result` (string): Description of the result
- `status_code` (number): Operation status code
- `metadata` (object): Additional information

**Example:**
```hcl
step "example" {
  plugin = "my-plugin"
  action = "my-action"
  
  params = {
    required_param = "example_value"
    optional_param = 100
    list_param = ["item1", "item2"]
  }
}
```

## Configuration

### Environment Variables
- `MY_PLUGIN_API_KEY`: API key for authentication
- `MY_PLUGIN_TIMEOUT`: Default timeout in seconds

### Authentication
Description of how to configure authentication if needed.

## Error Handling

Common errors and how to resolve them:

- **Error: "required_param is missing"**: Ensure all required parameters are provided
- **Error: "connection timeout"**: Check network connectivity and timeout settings

## Examples

See the [samples/](samples/) directory for complete workflow examples.

## Contributing

Instructions for contributing to the plugin development.

## License

Apache-2.0 License
```

### Sample Workflows

Create comprehensive examples in the `samples/` directory:

```hcl
# samples/basic-usage.hcl
workflow "basic-my-plugin-usage" {
  description = "Basic usage example for my-plugin"
  version     = "1.0.0"

  step "simple_action" {
    plugin = "my-plugin"
    action = "my-action"
    
    params = {
      required_param = "hello world"
    }
  }

  step "use_result" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["simple_action"]
    
    params = {
      command = "echo 'Result: ${simple_action.result}'"
    }
  }
}
```

```hcl
# samples/advanced-features.hcl
workflow "advanced-my-plugin-features" {
  description = "Advanced features and error handling"
  version     = "1.0.0"

  variable "api_endpoint" {
    type        = string
    default     = "https://api.example.com"
    description = "API endpoint URL"
  }

  step "complex_action" {
    plugin = "my-plugin"
    action = "my-action"
    
    params = {
      required_param = var.api_endpoint
      optional_param = 30
      list_param     = ["option1", "option2", "option3"]
      object_param = {
        nested_key = "nested_value"
        timeout    = 60
      }
    }
  }

  step "conditional_step" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["complex_action"]
    
    params = {
      command = "echo 'Success: ${complex_action.status_code == 200}'"
    }
  }
}
```

## Distribution Process

### 1. Development Phase

1. **Create plugin** following the interface requirements
2. **Write comprehensive tests** using the testing framework
3. **Create documentation** following the standards
4. **Test locally** with sample workflows

### 2. Repository Submission

#### Option A: Submit to Official Repository

1. **Fork the official repository**:
```bash
git clone https://github.com/corynth/corynthplugins.git
cd corynthplugins
git checkout -b add-my-plugin
```

2. **Add your plugin**:
```bash
mkdir my-plugin
# Copy your plugin files
cp -r /path/to/my-plugin/* my-plugin/
```

3. **Update registry.json**:
```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "Brief description of the plugin",
  "author": "Your Name",
  "format": "source",
  "installation": "Automatically compiled when first used",
  "tags": ["category", "functionality", "integration"],
  "actions": [
    {
      "name": "my-action",
      "description": "Description of the action",
      "example": "Example usage"
    }
  ],
  "requirements": {"corynth": ">=1.2.0"}
}
```

4. **Submit pull request**:
```bash
git add .
git commit -m "Add my-plugin: Brief description"
git push origin add-my-plugin
# Create pull request on GitHub
```

#### Option B: Create Your Own Plugin Repository

1. **Create repository structure**:
```bash
mkdir my-corynth-plugins
cd my-corynth-plugins
git init

# Create registry.json
cat > registry.json << EOF
{
  "name": "My Plugin Collection",
  "description": "Custom plugins for specific use cases",
  "version": "1.0.0",
  "plugins": [
    {
      "name": "my-plugin",
      "version": "1.0.0",
      "description": "Custom plugin functionality",
      "author": "Your Name",
      "path": "my-plugin",
      "tags": ["custom", "specific-domain"]
    }
  ]
}
EOF

# Add your plugin
mkdir my-plugin
cp -r /path/to/my-plugin/* my-plugin/
```

2. **Configure users to use your repository**:

Users can then configure your repository in their `corynth.hcl`:
```hcl
plugins {
  repositories {
    name     = "my-plugins"
    url      = "https://github.com/yourname/my-corynth-plugins"
    branch   = "main"
    priority = 1  # Higher priority than official
  }
  
  repositories {
    name     = "official"
    url      = "https://github.com/corynth/corynthplugins"
    branch   = "main"
    priority = 2  # Fallback
  }
}
```

Or via environment variable:
```bash
export CORYNTH_PLUGIN_REPO="https://github.com/yourname/my-corynth-plugins.git"
```

#### Option C: Private/Corporate Repository

1. **Create private repository**:
```bash
# Private GitHub repository
gh repo create mycompany/corynth-plugins --private

# GitLab, Bitbucket, or on-premise Git
git clone https://git.company.com/devtools/corynth-plugins.git
```

2. **Configure authentication**:
```bash
# For GitHub
export GITHUB_TOKEN="ghp_your_token_here"

# For general Git authentication
export GIT_USERNAME="your-username"
export GIT_PASSWORD="your-password"

# For SSH key authentication
# Ensure SSH keys are configured for the Git server
```

3. **Repository structure**:
```
my-company-plugins/
├── registry.json           # Plugin registry
├── database/              # Database plugins
│   ├── plugin.go
│   ├── README.md
│   └── samples/
├── monitoring/            # Monitoring plugins
│   ├── plugin.go
│   ├── README.md
│   └── samples/
└── deployment/           # Deployment plugins
    ├── plugin.go
    ├── README.md
    └── samples/
```

### 3. Review Process

The plugin will be reviewed for:
- **Code quality** and security
- **Interface compliance**
- **Documentation completeness**
- **Test coverage**
- **Performance considerations**

### 4. Publication

Once approved:
- Plugin is merged into the main branch
- Registry is updated automatically
- Plugin becomes available for all Corynth users
- Automatic testing ensures compatibility

## Best Practices

### Security
- **Validate all inputs** thoroughly
- **Sanitize file paths** to prevent traversal attacks
- **Use secure defaults** for all parameters
- **Handle credentials securely** (environment variables, not parameters)
- **Limit resource usage** with timeouts and size limits

### Performance
- **Implement context cancellation** for long-running operations
- **Use connection pooling** for database/API plugins
- **Cache expensive operations** when appropriate
- **Provide progress feedback** for long operations
- **Optimize memory usage** for large data processing

### Error Handling
- **Return meaningful error messages** with context
- **Use structured errors** when possible
- **Implement retry logic** for transient failures
- **Validate parameters early** before expensive operations
- **Clean up resources** properly in all code paths

### Compatibility
- **Support multiple OS/architectures** when possible
- **Use standard Go libraries** for portability
- **Version your plugin APIs** for backward compatibility
- **Test on different Go versions**
- **Document external dependencies** clearly

### Code Organization
```go
// Organize code into logical sections
type MyPlugin struct {
    // Plugin state (avoid global state)
    config *Config
    client *http.Client
}

// Group related functionality
func (p *MyPlugin) executeAPIAction(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    // Implementation
}

func (p *MyPlugin) executeDatabaseAction(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    // Implementation
}

// Helper functions for common operations
func (p *MyPlugin) validateAPIParams(params map[string]interface{}) error {
    // Validation logic
}

func (p *MyPlugin) buildAPIRequest(params map[string]interface{}) (*http.Request, error) {
    // Request building logic
}
```

## Troubleshooting

### Common Issues

#### Plugin Not Found
```
Error: plugin 'my-plugin' not found in any repository
```
**Solution**: Ensure plugin is in the corynthplugins repository and registry.json is updated.

#### Compilation Errors
```
Error: failed to compile plugin: undefined symbol
```
**Solution**: Check import paths and ensure all dependencies are properly defined in go.mod.

#### Interface Compliance
```
Error: ExportedPlugin does not implement Plugin interface
```
**Solution**: Verify all interface methods are implemented with correct signatures.

#### Runtime Errors
```
Error: panic: runtime error: invalid memory address
```
**Solution**: Add nil checks and proper error handling throughout the code.

### Debugging Tips

1. **Use Corynth logging system for structured, consistent logging**:

```go
import (
    "context"
    "time"
    "github.com/corynth/corynth/pkg/logging"
)

type MyPlugin struct {
    logger *logging.Logger
}

func NewMyPlugin() *MyPlugin {
    return &MyPlugin{
        logger: logging.NewDefaultLogger("my-plugin"),
    }
}

func (p *MyPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
    p.logger.Info("Executing action: %s", action)
    p.logger.Debug("Parameters: %+v", params)
    
    start := time.Now()
    result, err := p.doWork(ctx, params)
    duration := time.Since(start)
    
    if err != nil {
        p.logger.Error("Action '%s' failed after %v: %v", action, duration, err)
        return nil, err
    }
    
    p.logger.Info("Action '%s' completed successfully in %v", action, duration)
    p.logger.Debug("Result: %+v", result)
    return result, nil
}

// Use appropriate log levels:
// - Debug: Detailed debugging information
// - Info: General operational information  
// - Warn: Warning conditions (non-fatal)
// - Error: Error conditions (operation failed)
// - Fatal: Critical errors (will exit)
```

Enable debug logging for your plugin during development:
```bash
export CORYNTH_PLUGIN_LOG_LEVEL="debug"
corynth apply your-workflow.hcl
```

2. **Test locally before distribution**:
```bash
# Build plugin manually
go build -buildmode=plugin -o my-plugin.so plugin.go

# Test plugin loading
go run test_plugin_loading.go
```

3. **Use the validation workflow**:
```bash
corynth validate my-test-workflow.hcl
corynth plan my-test-workflow.hcl --detailed
```

### Getting Help

- **Documentation**: Check this guide and existing plugin examples
- **Issues**: Create issues in the corynthplugins repository
- **Community**: Join the Corynth community discussions
- **Examples**: Study existing plugins for patterns and best practices

---

This guide provides everything needed to create production-ready Corynth plugins. Follow the patterns, test thoroughly, and contribute to the growing ecosystem of workflow automation tools.