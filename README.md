# Corynth - Workflow Orchestration Platform

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/install)
[![Release](https://img.shields.io/badge/Release-v1.2.0-green.svg)](#installation)

Corynth is a production-ready workflow orchestration platform built in Go that enables you to define, execute, and manage multi-step workflows using HCL (HashiCorp Configuration Language). Features a robust plugin-based architecture with bulletproof remote plugin installation.

## 🚀 Features

### Core Capabilities
- **HCL Workflow Definition** - Define workflows using clear, readable HCL syntax
- **Plugin Ecosystem** - Extensible plugin system with built-in and remote plugins
- **Dependency Management** - Support for step dependencies and parallel execution
- **State Management** - Persistent workflow state with execution tracking
- **Auto-Installation** - Remote plugins automatically installed from GitHub

### Plugin System
- **Built-in Plugin**: `shell` - Available immediately for system operations  
- **gRPC Plugins**: `http`, `docker`, `terraform`, `k8s`, `llm` - Production-grade compiled plugins
- **Automated Installation** - Pre-compiled binaries with zero-configuration setup
- **Remote Discovery** - 14+ plugins available from GitHub plugin registry

### Advanced Features
- **Boolean Logic** - Conditional step execution with boolean expressions and functions
- **Loop Processing** - Iterate over lists, tuples, and objects with powerful loop constructs
- **Parallel Execution** - Run independent steps concurrently
- **Complex Dependencies** - Chain workflows with sophisticated dependency management
- **Error Handling** - Robust error handling and recovery
- **Workflow Templates** - Reusable workflow components
- **State Persistence** - JSON-based state storage and retrieval

## 📦 Installation

### Quick Install (Recommended)

Install the latest release with one command:

```bash
curl -sSL https://raw.githubusercontent.com/corynth/corynth/main/install.sh | bash
```

Or download a specific version:

```bash
curl -sSL https://raw.githubusercontent.com/corynth/corynth/main/install.sh | VERSION=v1.2.0 bash
```

### Download Pre-built Binaries

Download binaries for your platform from the [releases page](https://github.com/corynth/corynth/releases):

- **macOS**: `corynth-darwin-amd64` (Intel) or `corynth-darwin-arm64` (Apple Silicon)
- **Linux**: `corynth-linux-amd64` or `corynth-linux-arm64`
- **Windows**: `corynth-windows-amd64.exe`

### Install via Docker

```bash
docker pull ghcr.io/corynth/corynth:latest
docker run --rm ghcr.io/corynth/corynth:latest version
```

### Build from Source

```bash
git clone https://github.com/corynth/corynth.git
cd corynth
make build
sudo make install
```

**Requirements**: Go 1.21+, Git

## 🚀 Quick Start

### 1. Initialize Project

```bash
corynth init
```

### 2. Create Your First Workflow

```hcl
# hello-world.hcl
workflow "hello-world" {
  description = "My first Corynth workflow"
  version = "1.0.0"

  variable "name" {
    type = string
    default = "World"
    description = "Name to greet"
  }

  step "greet" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Hello, {{.Variables.name}}! Welcome to Corynth.'"
    }
  }
}
```

### 3. Execute Workflow

```bash
corynth apply hello-world.hcl
```

### 4. Check Execution History

```bash
corynth state list
```

## 📖 Workflow Examples

### HTTP API and File Operations

```hcl
workflow "api-to-file" {
  description = "Fetch data from API and save to file"
  version = "1.0.0"

  variable "api_url" {
    type = string
    default = "https://api.github.com/users/octocat"
    description = "API endpoint URL"
  }

  step "fetch_data" {
    plugin = "http"
    action = "get"
    
    params = {
      url = "{{.Variables.api_url}}"
      timeout = 30
    }
  }

  step "save_response" {
    plugin = "file"
    action = "write"
    depends_on = ["fetch_data"]
    
    params = {
      path = "/tmp/api-response.json"
      content = "API response saved successfully"
      create_dirs = true
    }
  }

  step "confirm_save" {
    plugin = "shell"
    action = "exec"
    depends_on = ["save_response"]
    
    params = {
      command = "echo 'Data saved to /tmp/api-response.json'"
    }
  }
}
```

### Multi-Step with Dependencies

```hcl
workflow "complex-pipeline" {
  description = "Complex multi-step pipeline"
  version = "1.0.0"

  step "setup" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "mkdir -p /tmp/pipeline"
    }
  }

  step "clone_repo" {
    plugin = "git"
    action = "clone"
    depends_on = ["setup"]
    
    params = {
      url = "https://github.com/example/repo.git"
      path = "/tmp/pipeline/repo"
      branch = "main"
    }
  }

  step "process_data" {
    plugin = "shell"
    action = "exec"
    depends_on = ["clone_repo"]
    
    params = {
      command = "cd /tmp/pipeline/repo && ./process.sh"
    }
  }

  step "upload_results" {
    plugin = "http"
    action = "post"
    depends_on = ["process_data"]
    
    params = {
      url = "https://api.example.com/results"
      body = "Processing complete"
    }
  }
}
```

### Parallel Execution

```hcl
workflow "parallel-tasks" {
  description = "Run tasks in parallel"
  version = "1.0.0"

  step "task_1" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Task 1' && sleep 2"
    }
  }

  step "task_2" {
    plugin = "http"
    action = "get"
    
    params = {
      url = "https://httpbin.org/delay/1"
    }
  }

  step "task_3" {
    plugin = "file"
    action = "write"
    
    params = {
      path = "/tmp/task3.txt"
      content = "Task 3 completed"
    }
  }

  step "finalize" {
    plugin = "shell"
    action = "exec"
    depends_on = ["task_1", "task_2", "task_3"]
    
    params = {
      command = "echo 'All tasks completed'"
    }
  }
}
```

## 🔌 Plugin System

Corynth features a production-ready gRPC plugin system with automated installation and a growing ecosystem of plugins for infrastructure, container orchestration, cloud deployments, and automation tasks.

### Plugin Architecture

**gRPC-Based Design**: Plugins run as separate processes communicating via gRPC, ensuring:
- **Process Isolation**: Plugin crashes don't affect main workflow engine
- **Language Flexibility**: Plugins can be written in any language supporting gRPC  
- **Performance**: Fast binary communication with minimal overhead
- **Reliability**: Automatic plugin discovery and loading

### Plugin Installation Fix (v1.2.0)

The plugin installation system has been completely fixed and now supports:
- **Automated Installation**: `corynth plugin install <name>` works with pre-compiled binaries
- **Path Resolution**: Fixed path resolution for `official/` directory structure  
- **Binary Detection**: Automatically detects and copies pre-compiled plugin binaries
- **Zero Configuration**: No manual compilation or setup required

### Plugin Categories

#### Core Plugins
- **shell** - Built-in command execution and process management ✅ Working

#### gRPC Plugins (Production-grade compiled plugins)
- **http** - HTTP/HTTPS client for REST API calls ✅ Working (Go/gRPC)
- **docker** - Docker container management and operations ✅ Working (Go/gRPC)
- **terraform** - Infrastructure as Code operations ✅ Working (Go/gRPC)  
- **k8s** - Kubernetes cluster management ✅ Working (Go/gRPC)
- **llm** - Large Language Model integration ✅ Working (Go/gRPC)
- **calculator** - Mathematical calculations and unit conversions ✅ Working
- **reporting** - Generate formatted reports and documents ✅ Working
- **slack** - Slack messaging and notifications ✅ Working

#### System Operations (Remote)
- **time** - Time-based operations and delays
- **log** - Logging and output management

#### Network & Communication (Remote)
- **ping** - Network connectivity testing
- **net-scan** - Network scanning and discovery

#### Database & Storage (Remote)
- **mysql** - MySQL database operations
- **redis** - Redis cache and data structure operations

#### Container & Orchestration (Remote)
- **kubernetes** - Complete K8s resource management (apply, get, delete, scale, logs, wait)
- **helm** - Package manager for Kubernetes (install, upgrade, template, repo management)
- **docker** - Docker container management and orchestration

#### Cloud Providers (Remote)
- **aws** - Amazon Web Services (EC2, S3, Lambda, IAM, VPC operations)
- **gcp** - Google Cloud Platform (Compute Engine, GKE, Cloud Storage, Cloud Functions)

#### Infrastructure & DevOps (Remote)
- **terraform** - Infrastructure as Code operations
- **vault** - HashiCorp Vault secret management
- **calculator** - Mathematical calculations for resource planning
- **json-processor** - JSON data processing for API responses and configurations

#### AI & Reporting (Remote)
- **llm** - Large Language Model integration (OpenAI, Anthropic, Ollama) with temperature control
- **reporting** - Generate formatted reports, tables, and charts (Markdown, HTML, PDF)

### Plugin Status  
- **Built-in Plugins**: 1 essential plugin (`shell`) ✅ Working
- **gRPC Plugins**: 8 working plugins (`http`, `docker`, `terraform`, `k8s`, `llm`, `calculator`, `reporting`, `slack`) ✅ Production-grade compiled binaries
- **Template Processing**: ✅ Variable substitution working with `{{.Variables.name}}` syntax
- **Plugin Auto-discovery**: ✅ gRPC plugins automatically discovered from `.corynth/plugins/` directory
- **Automated Installation**: ✅ `corynth plugin install <name>` works with pre-compiled binaries
- **Remote Repository**: ✅ 14+ plugins available from GitHub plugin registry
- **Boolean Logic**: ✅ Conditional step execution with `condition` expressions
- **Loop Processing**: ✅ Iterate over lists, tuples, and objects with `loop` blocks
- **Flow Control**: ✅ Advanced control flow with boolean functions (`and`, `or`, `not`, `if`)
- **Parallel Execution**: ✅ Independent steps execute concurrently with proper dependency management

### Plugin Discovery & Remote Installation

Corynth automatically discovers and installs plugins from remote repositories with **bulletproof automatic installation**:

```bash
# Discover available plugins from registries
corynth plugin discover

# List currently installed plugins
corynth plugin list

# Install remote plugins automatically - zero configuration needed
corynth plugin install calculator    # ✓ Mathematical calculations
corynth plugin install json-processor # ✓ JSON data processing
corynth plugin install aws            # ✓ Amazon Web Services
corynth plugin install kubernetes     # ✓ Kubernetes management

# Get plugin information
corynth plugin info git

# Generate new plugin scaffold
corynth plugin init my-plugin --type http
```

**Remote Plugin Registry**: 14 total plugins available from [official plugin repository](https://github.com/corynth/corynth):
- **http** - HTTP/HTTPS client for REST API calls ✅ Installed
- **docker** - Docker container management and operations ✅ Installed
- **terraform** - Infrastructure as Code operations ✅ Installed  
- **k8s** - Kubernetes cluster management ✅ Installed
- **llm** - Large Language Model integration ✅ Installed
- **calculator** - Mathematical calculations and unit conversions ✅ Installed
- **reporting** - Generate formatted reports and documents ✅ Installed
- **slack** - Slack messaging and notifications ✅ Installed

### Plugin Usage

```hcl
step "example" {
  plugin = "plugin-name"
  action = "action-name"
  
  params = {
    param1 = "value1"
    param2 = "value2"
  }
}
```

## 📋 Commands Reference

### Core Commands

```bash
# Project management
corynth init                    # Initialize new project
corynth validate workflow.hcl   # Validate workflow syntax
corynth plan workflow.hcl       # Preview execution plan
corynth apply workflow.hcl      # Execute workflow

# State management  
corynth state list             # List recent executions
corynth state show <id>        # Show execution details
corynth state clean            # Clean old state files

# Plugin management
corynth plugin list            # List installed plugins
corynth plugin install <name>  # Install plugin
corynth plugin info <name>     # Get plugin info
corynth plugin init <name>     # Generate new plugin

# Utilities
corynth sample                 # Generate sample workflows
corynth version               # Show version info
```

### Execution Options

```bash
# Auto-approve execution
corynth apply --auto-approve workflow.hcl

# Quiet mode (minimal output)
corynth apply --quiet workflow.hcl

# Limit parallel execution
corynth apply --parallel 2 workflow.hcl

# Verbose output
corynth apply --verbose workflow.hcl
```

## 🏗️ Workflow Syntax

### Basic Structure

```hcl
workflow "name" {
  description = "Workflow description"
  version     = "1.0.0"

  step "step_name" {
    plugin = "plugin-name"
    action = "action-name"
    
    params = {
      key = "value"
    }
  }
}
```

### Advanced Features

#### Variables
```hcl
workflow "with-variables" {
  variable "api_url" {
    type = string
    default = "https://api.example.com"
  }

  step "call_api" {
    plugin = "http"
    action = "get"
    
    params = {
      url = var.api_url
    }
  }
}
```

#### Dependencies
```hcl
step "second_step" {
  plugin = "shell"
  action = "exec"
  depends_on = ["first_step"]
  
  params = {
    command = "echo '${first_step.output}'"
  }
}
```

#### Error Handling
```hcl
step "risky_operation" {
  plugin = "shell"
  action = "exec"
  
  params = {
    command = "potentially-failing-command"
  }
  
  retry {
    max_attempts = 3
    delay = "5s"
  }
  
  continue_on {
    error = true
  }
}
```

#### Boolean Logic & Conditional Execution
```hcl
step "conditional_step" {
  plugin = "shell"
  action = "exec"
  condition = "var.environment == \"production\""
  
  params = {
    command = "echo 'Running in production'"
  }
}

step "complex_condition" {
  plugin = "shell"
  action = "exec"
  condition = "and(var.deploy_enabled, equal(var.environment, \"production\"))"
  
  params = {
    command = "echo 'Complex condition met'"
  }
}
```

#### Loop Processing
```hcl
step "process_servers" {
  plugin = "shell"
  action = "exec"
  
  loop {
    over = "[\"web1\", \"web2\", \"web3\"]"
    variable = "server"
  }
  
  params = {
    command = "echo 'Processing server'"
  }
}

step "environment_deploy" {
  plugin = "shell"
  action = "exec"
  condition = "var.deploy_enabled"
  
  loop {
    over = "[\"dev\", \"staging\", \"prod\"]"
    variable = "env"
  }
  
  params = {
    command = "echo 'Deploying to environment'"
  }
}
```

#### Available Boolean Functions
- `and(a, b)` - Logical AND
- `or(a, b)` - Logical OR  
- `not(a)` - Logical NOT
- `if(condition, true_value, false_value)` - Conditional expression
- `equal(a, b)` - Equality comparison
- `notequal(a, b)` - Inequality comparison
- `lessthan(a, b)` - Less than comparison
- `greaterthan(a, b)` - Greater than comparison

## 🛠️ Configuration

### Project Configuration

Corynth uses `corynth.hcl` for project configuration:

```hcl
# corynth.hcl
project {
  name = "my-project"
  version = "1.0.0"
}

orchestration {
  max_concurrent_workflows = 10
  default_timeout = "30m"
  state_backend = "local"
  
  retry {
    default_attempts = 3
    backoff_strategy = "exponential"
  }
}

logging {
  level = "info"
  format = "structured"
  output = "stdout"
  file = "~/.corynth/logs/corynth.log"
  include_stack_traces = true
  component_logging = {
    plugin = "debug"
    state = "info"
    workflow = "info"
  }
}
```

### Plugin Repository Configuration

Corynth supports multiple plugin repositories with flexible configuration options:

#### Method 1: Environment Variable (Quick Setup)
```bash
# Single repository URL
export CORYNTH_PLUGIN_REPO="https://github.com/yourcompany/corynth-plugins.git"
```

#### Method 2: Configuration File (Recommended)
Configure multiple repositories with priorities in `corynth.hcl`:

```hcl
# corynth.hcl
plugins {
  auto_install = true
  local_path   = "bin/plugins"
  
  # Corporate plugins (highest priority)
  repositories {
    name     = "corporate"
    url      = "https://github.com/mycompany/corynth-plugins"
    branch   = "main"
    priority = 1
  }
  
  # Community plugins
  repositories {
    name     = "community"
    url      = "https://github.com/community/corynth-plugins"
    branch   = "main" 
    priority = 2
  }
  
  # Official plugins (fallback)
  repositories {
    name     = "official"
    url      = "https://github.com/corynth/plugins"
    branch   = "main"
    priority = 3
  }
  
  cache {
    enabled  = true
    path     = "~/.corynth/cache"
    ttl      = "24h"
    max_size = "1GB"
  }
}
```

#### Method 3: Private Repositories
```hcl
plugins {
  repositories {
    name     = "private"
    url      = "https://github.com/mycompany/private-plugins"
    branch   = "main"
    priority = 1
    
    # Authentication via environment variables:
    # GITHUB_TOKEN, GIT_USERNAME, GIT_PASSWORD
  }
}
```

#### Repository Discovery Process
1. **Search by Priority**: Repositories searched in priority order (1 = highest)
2. **Plugin Lookup**: Searches for plugin in each repository
3. **Auto-Installation**: Downloads and compiles plugin on first use
4. **Caching**: Compiled plugins cached for faster subsequent use
5. **Fallback**: Falls back to next repository if plugin not found

#### Use Cases
- **Corporate Environment**: Point to internal plugin repositories
- **Multi-Team**: Different teams maintain different plugin collections
- **Development**: Use development repos for testing new plugins
- **Hybrid**: Mix official, community, and corporate plugins

See [examples/plugin-repository-configs.hcl](examples/plugin-repository-configs.hcl) for complete configuration examples.

### Environment Variables

```bash
# Plugin repository URL (simple setup)
export CORYNTH_PLUGIN_REPO="https://github.com/yourcompany/plugins.git"

# Git authentication for private repos
export GITHUB_TOKEN="your_github_token"
export GIT_USERNAME="your_username"
export GIT_PASSWORD="your_password"

# State directory
export CORYNTH_STATE_DIR="~/.corynth/state"

# Plugin cache directory
export CORYNTH_PLUGIN_CACHE="~/.corynth/cache"

# Log level
export CORYNTH_LOG_LEVEL="info"

# Disable color output
export CORYNTH_NO_COLOR="true"

# Log file output
export CORYNTH_LOG_FILE="~/.corynth/logs/corynth.log"

# Enable debug logging for plugins
export CORYNTH_PLUGIN_LOG_LEVEL="debug"
```

### Logging Configuration

Corynth provides comprehensive logging capabilities for debugging, monitoring, and auditing:

#### Log Levels
- **`debug`** - Detailed information for debugging (includes plugin operations, state changes)
- **`info`** - General information about workflow execution (default)
- **`warn`** - Warning messages for non-fatal issues
- **`error`** - Error messages for failed operations
- **`fatal`** - Critical errors that stop execution

#### Configuration Methods

**Method 1: Environment Variables (Quick Setup)**
```bash
# Global log level
export CORYNTH_LOG_LEVEL="debug"

# Log to file
export CORYNTH_LOG_FILE="~/.corynth/logs/corynth.log"

# Component-specific logging
export CORYNTH_PLUGIN_LOG_LEVEL="debug"
export CORYNTH_STATE_LOG_LEVEL="info" 
export CORYNTH_WORKFLOW_LOG_LEVEL="info"
```

**Method 2: Configuration File (Recommended)**
```hcl
# corynth.hcl
logging {
  # Global settings
  level  = "info"
  format = "structured"  # or "text"
  output = "both"        # stdout, file, or both
  
  # File output
  file                 = "~/.corynth/logs/corynth.log"
  max_size            = "100MB"
  max_backups         = 5
  max_age_days        = 30
  compress_backups    = true
  
  # Advanced options
  include_stack_traces = true
  show_colors         = true
  timestamp_format    = "2006-01-02 15:04:05"
  
  # Component-specific log levels
  component_logging = {
    plugin    = "debug"   # Plugin operations
    state     = "info"    # State management
    workflow  = "info"    # Workflow execution
    http      = "warn"    # HTTP requests
    retry     = "debug"   # Retry operations
    timeout   = "info"    # Timeout handling
  }
  
  # Audit logging
  audit = {
    enabled     = true
    file        = "~/.corynth/logs/audit.log"
    log_success = true
    log_failures = true
    include_params = false  # Don't log sensitive parameters
  }
}
```

#### Log Formats

**Structured Format (JSON)**
```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "info",
  "component": "workflow",
  "message": "Step 'api_call' completed successfully",
  "workflow": "data-pipeline",
  "step": "api_call",
  "duration": "2.3s",
  "metadata": {
    "plugin": "http",
    "action": "get",
    "status_code": 200
  }
}
```

**Text Format (Human Readable)**
```
2024-01-15 10:30:45 [INFO] [workflow] Step 'api_call' completed successfully
2024-01-15 10:30:46 [DEBUG] [plugin] Loading plugin: http
2024-01-15 10:30:47 [ERROR] [state] Failed to save state: connection timeout
```

#### Common Logging Scenarios

**Development & Debugging**
```hcl
logging {
  level = "debug"
  format = "text"
  output = "stdout"
  show_colors = true
  
  component_logging = {
    plugin = "debug"
    workflow = "debug"
    state = "debug"
  }
}
```

**Production Monitoring**
```hcl
logging {
  level = "info"
  format = "structured"
  output = "file"
  file = "/var/log/corynth/corynth.log"
  
  max_size = "1GB"
  max_backups = 10
  compress_backups = true
  
  audit = {
    enabled = true
    file = "/var/log/corynth/audit.log"
    log_success = false    # Only log failures in production
    log_failures = true
  }
}
```

**Troubleshooting Plugin Issues**
```bash
# Enable debug logging for specific components
export CORYNTH_PLUGIN_LOG_LEVEL="debug"
export CORYNTH_HTTP_LOG_LEVEL="debug"

# Run workflow with verbose logging
corynth apply --log-level debug my-workflow.hcl

# Check plugin-specific logs
tail -f ~/.corynth/logs/corynth.log | grep "\[plugin\]"
```

#### Log File Management

**Automatic Rotation**
- **Size-based**: Rotate when file reaches `max_size`
- **Time-based**: Keep logs for `max_age_days`
- **Count-based**: Retain `max_backups` files
- **Compression**: Gzip old log files to save space

**Manual Log Management**
```bash
# View recent logs
tail -f ~/.corynth/logs/corynth.log

# Search for errors
grep "ERROR" ~/.corynth/logs/corynth.log

# Filter by component
grep "\[plugin\]" ~/.corynth/logs/corynth.log

# View structured logs (JSON format)
jq '.level, .message' ~/.corynth/logs/corynth.log

# Monitor workflow execution
tail -f ~/.corynth/logs/corynth.log | grep -E "(workflow|step)"
```

#### Integration with External Tools

**ELK Stack (Elasticsearch, Logstash, Kibana)**
```hcl
logging {
  format = "structured"
  output = "file"
  file = "/var/log/corynth/corynth.json"
  
  # Logstash will parse this JSON format
}
```

**Splunk**
```hcl
logging {
  format = "structured"
  timestamp_format = "2006-01-02T15:04:05.000Z"
  include_metadata = true
}
```

**Prometheus/Grafana (via log parsing)**
```hcl
logging {
  format = "structured"
  
  # Include metrics in logs
  include_performance = true
  include_counters = true
}
```

#### Troubleshooting with Logs

**Common Issues**

1. **Plugin Not Loading**
   ```bash
   # Check plugin logs
   CORYNTH_PLUGIN_LOG_LEVEL=debug corynth apply workflow.hcl
   ```

2. **State Management Issues**
   ```bash
   # Debug state operations
   CORYNTH_STATE_LOG_LEVEL=debug corynth apply workflow.hcl
   ```

3. **Network/HTTP Problems**
   ```bash
   # Debug HTTP plugin calls
   CORYNTH_HTTP_LOG_LEVEL=debug corynth apply workflow.hcl
   ```

4. **Performance Issues**
   ```hcl
   logging {
     level = "info"
     include_performance = true
     include_timing = true
   }
   ```

**Log Analysis**
```bash
# Count errors by type
grep "ERROR" ~/.corynth/logs/corynth.log | cut -d']' -f3 | sort | uniq -c

# Find slow operations
grep "duration" ~/.corynth/logs/corynth.log | grep -E "[5-9]\.[0-9]+s|[1-9][0-9]+\.[0-9]+s"

# Monitor plugin usage
grep "\[plugin\]" ~/.corynth/logs/corynth.log | grep "Loading plugin" | cut -d'"' -f2 | sort | uniq -c
```

See [examples/logging-configs.hcl](examples/logging-configs.hcl) for complete logging configuration examples covering development, production, CI/CD, monitoring, and security scenarios.

## 🧪 Development

### Building from Source

```bash
git clone https://github.com/corynth/corynth.git
cd corynth

# Install dependencies
go mod tidy

# Run tests
make test

# Build binary
make build

# Install locally
make install
```

### Running Tests

```bash
# Run all tests
make test

# Run specific test package
go test ./pkg/workflow/...

# Run with coverage
make test-coverage
```

### Plugin Development

```bash
# Generate new plugin scaffold
corynth plugin init my-plugin --type http --author "Your Name"

# Navigate to plugin directory
cd my-plugin

# Implement plugin logic in plugin.go
# Test with sample workflows
corynth apply samples/basic-usage.hcl

# Submit to plugin repository
git add . && git commit -m "Add my-plugin"
git push origin add-my-plugin
```

See [docs/PLUGIN_DEVELOPMENT_GUIDE.md](docs/PLUGIN_DEVELOPMENT_GUIDE.md) for comprehensive plugin development documentation.

## 📊 Performance

### Benchmarks

- **Simple workflow execution**: < 1 second
- **Complex multi-step (5+ steps)**: < 5 seconds  
- **Plugin auto-installation**: 8-15 seconds (cached thereafter)
- **State persistence**: < 50ms
- **Parallel step execution**: Concurrent with proper synchronization

### Scalability

- **Maximum concurrent workflows**: Configurable (default: 10)
- **Maximum steps per workflow**: No hard limit
- **State storage**: Local JSON files (extensible to databases)
- **Plugin ecosystem**: Unlimited plugins via git repositories

## 🔐 Security

### Security Features

- **Input validation** - All parameters validated before plugin execution
- **Path traversal protection** - File operations restricted to safe paths
- **Timeout enforcement** - All operations have configurable timeouts
- **Resource limits** - Memory and CPU usage monitoring
- **Secure plugin loading** - Plugins compiled and loaded safely

### Best Practices

- Use environment variables for sensitive data (API keys, passwords)
- Validate all external inputs in custom plugins
- Use specific versions for production workflows
- Regularly update plugins and Corynth core
- Monitor workflow execution logs for suspicious activity

## 🤝 Contributing

Contributions are welcome! Corynth is now in production-ready status with a fully functional plugin system:

### Priority Areas for Contribution
- **Plugin Ecosystem**: Create new plugins for additional services and tools
- **Testing**: Add comprehensive unit and integration tests
- **Documentation**: Expand examples and use-case documentation
- **Performance**: Optimize workflow execution and plugin loading
- **CI/CD**: Enhance automated testing and release pipelines

### How to Contribute

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** and add tests where possible
4. **Build and test**: `make build && make test`
5. **Commit your changes**: `git commit -m 'Add amazing feature'`
6. **Push to the branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Development Setup

```bash
git clone https://github.com/corynth/corynth.git
cd corynth
go mod download
make build
./corynth version
```

### Areas for Enhancement
- Additional plugin development (AWS, GCP, monitoring tools)
- Comprehensive test coverage expansion
- Performance optimizations for large workflows
- Enhanced error reporting and debugging tools

### Development Guidelines

- Follow Go best practices and idioms
- Add tests for new functionality when possible
- Update documentation for user-facing changes
- Use conventional commit messages
- Test your changes with `make build`

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙋 Support

### Documentation

- **Plugin Development**: [docs/PLUGIN_DEVELOPMENT_GUIDE.md](docs/PLUGIN_DEVELOPMENT_GUIDE.md)
- **Plugin System**: [docs/PLUGIN_SYSTEM.md](docs/PLUGIN_SYSTEM.md)
- **Quick Reference**: [docs/PLUGIN_QUICK_REFERENCE.md](docs/PLUGIN_QUICK_REFERENCE.md)

### Community

- **Issues**: [GitHub Issues](https://github.com/corynth/corynth/issues)
- **Discussions**: [GitHub Discussions](https://github.com/corynth/corynth/discussions)
- **Plugin Registry**: [Corynth Plugins](https://github.com/corynth/plugins)

### Getting Help

1. Check the documentation and examples above
2. Search existing GitHub issues
3. Create a new issue with:
   - Corynth version (`corynth version`)
   - Operating system and architecture
   - Complete workflow file (if applicable)
   - Full error message and logs

---

**Built with ❤️ for workflow automation**

*Corynth empowers developers to create reliable, maintainable, and scalable workflow automation solutions.*