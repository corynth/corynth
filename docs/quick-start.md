# Quick Start Guide

Get up and running with Corynth in minutes. This guide walks you through creating and executing your first workflow.

## Prerequisites

- [Corynth installed](installation.md) and available in your PATH
- Git installed (for plugin auto-installation)
- Internet connection (for remote plugin downloads)

## Step 1: Initialize Project

Create a new directory for your Corynth project:

```bash
mkdir my-corynth-project
cd my-corynth-project
corynth init
```

Expected output:
```
Initializing Corynth...
✓ Configuration validated
✓ Created .corynth directory structure
✓ Plugin system initialized
✓ State backend initialized
✓ Corynth has been successfully initialized!
```

## Step 2: Create Your First Workflow

Create a simple workflow file:

```bash
cat > hello-world.hcl << 'EOF'
workflow "hello-world" {
  description = "My first Corynth workflow"
  version = "1.0.0"

  step "greet" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Hello, World from Corynth!'"
    }
  }
}
EOF
```

## Step 3: Validate Workflow

Check that your workflow syntax is correct:

```bash
corynth validate hello-world.hcl
```

Expected output:
```
• Validating hello-world.hcl...
✓ Workflow is valid
Name: hello-world
Description: My first Corynth workflow
Steps: 1
```

## Step 4: Preview Execution Plan

See what Corynth will do before executing:

```bash
corynth plan hello-world.hcl
```

Expected output:
```
Workflow: hello-world
Description: My first Corynth workflow
Steps to execute: 1

Steps:
  • 1. greet (shell.exec)

Estimated duration: 30s
```

## Step 5: Execute Workflow

Run your first workflow:

```bash
corynth apply hello-world.hcl
```

Interactive mode will prompt for confirmation:
```
Workflow: hello-world
Description: My first Corynth workflow
Steps to execute: 1

Steps:
  • 1. greet (shell.exec)

Do you want to proceed? (y/N): y
```

For automatic approval:
```bash
corynth apply --auto-approve hello-world.hcl
```

Expected output:
```
• Loading workflow from hello-world.hcl...
Starting workflow execution...
Hello, World from Corynth!

✓ ✓ Workflow completed: hello-world
Duration: 0s
Steps: 1/1
```

## Step 6: Check Execution History

View your workflow execution history:

```bash
corynth state list
```

Expected output:
```
Recent executions (1)
  • 5d2d2966: hello-world (success) - 0s
```

Get detailed information about an execution:
```bash
corynth state show 5d2d2966
```

## What Just Happened?

1. **Initialization**: Corynth created a local project structure
2. **Plugin Loading**: The `shell` plugin was automatically loaded
3. **Workflow Parsing**: Your HCL file was parsed and validated
4. **Execution**: The workflow step was executed successfully
5. **State Persistence**: Execution results were saved locally

## Next Steps

### Try More Plugins

Install and use remote plugins:

```bash
# Install HTTP plugin for API calls
corynth plugin install http

# Create API workflow
cat > api-test.hcl << 'EOF'
workflow "api-test" {
  description = "Test API endpoint"
  version = "1.0.0"

  step "get_user" {
    plugin = "http"
    action = "get"
    
    params = {
      url = "https://api.github.com/users/octocat"
      timeout = 30
    }
  }
}
EOF

# Execute API workflow
corynth apply --auto-approve api-test.hcl
```

### Create Multi-Step Workflows

```bash
cat > multi-step.hcl << 'EOF'
workflow "multi-step" {
  description = "Multi-step workflow example"
  version = "1.0.0"

  step "setup" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "mkdir -p /tmp/corynth-demo"
    }
  }

  step "create_file" {
    plugin = "file"
    action = "write"
    depends_on = ["setup"]
    
    params = {
      path = "/tmp/corynth-demo/data.txt"
      content = "Hello from multi-step workflow!"
    }
  }

  step "read_file" {
    plugin = "shell"
    action = "exec"
    depends_on = ["create_file"]
    
    params = {
      command = "cat /tmp/corynth-demo/data.txt"
    }
  }

  step "cleanup" {
    plugin = "shell"
    action = "exec"
    depends_on = ["read_file"]
    
    params = {
      command = "rm -rf /tmp/corynth-demo"
    }
  }
}
EOF

corynth apply --auto-approve multi-step.hcl
```

### Explore More Examples

- [Simple Workflows](../examples/simple/)
- [API Integration](../examples/integration/)
- [File Processing](../examples/data-processing/)

### Learn Advanced Features

- [Workflow Syntax Guide](workflow-syntax.md)
- [Dependencies & Parallel Execution](dependencies.md)
- [Variables & Configuration](configuration.md)

## Common Commands Summary

```bash
# Project management
corynth init                    # Initialize project
corynth validate workflow.hcl   # Validate workflow
corynth plan workflow.hcl       # Preview execution
corynth apply workflow.hcl      # Execute workflow

# Plugin management
corynth plugin list             # List plugins
corynth plugin install <name>   # Install plugin
corynth plugin info <name>      # Plugin details

# State management
corynth state list              # Execution history
corynth state show <id>         # Execution details
corynth state clean             # Clean old executions

# Utilities
corynth sample                  # Generate samples
corynth version                 # Show version
```

## Need Help?

- **Documentation**: Continue with the [User Guide](README.md)
- **Examples**: Browse [example workflows](../examples/)
- **Issues**: [GitHub Issues](https://github.com/corynth/corynth/issues)
- **Community**: [GitHub Discussions](https://github.com/corynth/corynth/discussions)

---

**You're ready to start building powerful workflows with Corynth!** 🚀