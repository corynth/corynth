# Corynth

Corynth is a powerful automation orchestration tool that enables users to define, plan, and execute sequential workflows through declarative YAML configurations. Built with Go, Corynth provides a familiar experience to Terraform users by following the init/plan/apply pattern while offering advanced automation capabilities.

![Corynth Logo](docs/images/corynth-logo.png)

## 🎯 Core Value Proposition

Corynth simplifies complex automation tasks by allowing users to:

* Define clear sequences of actions as flows
* Create dependencies between steps
* Chain flows together for complex operations
* Use a familiar Terraform-like workflow
* Leverage a plugin ecosystem for extensibility

## 🏗️ Architecture Overview

Corynth is built with a modular architecture consisting of:

1. **Core Engine**: The main execution engine written in Go
2. **Plugin System**: Extensible plugin architecture
3. **Flow Parser**: YAML parser for flow definitions
4. **Dependency Resolver**: Handles step dependencies
5. **Execution Manager**: Orchestrates flow execution

## 🔍 Key Features

### Command Structure

```
corynth init [dir]                 # Initialize a new Corynth project
corynth plan [dir]                 # Plan execution and validate flows
corynth apply [dir] [flow-name]    # Execute all flows or a specific flow
```

### Flow Definition

Flows are defined in YAML with the following structure:

```yaml
flow:
  name: "deployment_flow"
  description: "Deploy application to production"
  steps:
    - name: "clone_repo"
      plugin: "git"
      action: "clone"
      params:
        repo: "https://github.com/user/repo.git"
        branch: "main"
    
    - name: "setup_env"
      plugin: "shell"
      action: "exec"
      params:
        command: "make setup"
      depends_on:
        - step: "clone_repo"
          status: "success"
    
    - name: "deploy"
      plugin: "ansible"
      action: "playbook"
      params:
        playbook: "deploy.yml"
      depends_on:
        - step: "setup_env"
          status: "success"
```

### Dependency Management

Steps can have dependencies on other steps:

* `depends_on`: List of step dependencies
  * `step`: Name of the dependent step
  * `status`: Required status ("success" or "failure")

### Flow Chaining

Flows can be chained together:

```yaml
flow:
  name: "master_flow"
  description: "Complete deployment process"
  chain:
    - flow: "build_flow"
      on_success: "deploy_flow"
      on_failure: "notify_flow"
    - flow: "deploy_flow"
      on_success: "test_flow"
      on_failure: "rollback_flow"
```

### Plugin System

Core plugins:
* **Git**: Repository operations
* **Ansible**: Infrastructure automation
* **Shell**: Command execution

Remote plugins:
* Versioned plugins stored in repository
* Downloaded during plan phase if needed
* Plugin manifest for version management

## 📝 User Experience

### Init Phase

When a user runs `corynth init`:

1. Create a directory structure for the project
2. Generate a sample YAML file with common patterns
3. Set up plugin configuration
4. Create a .corynth directory for state

Example output:

```
$ corynth init my_project
✅ Created directory structure
✅ Generated sample flow definition
✅ Plugin configuration initialized
✨ Project initialized! Run 'corynth plan' to validate your flows
```

### Plan Phase

When a user runs `corynth plan`:

1. Parse all flow YAML files
2. Validate dependencies between steps
3. Check for required plugins
4. Download missing plugins if configured
5. Create an execution plan

Example output:

```
$ corynth plan
🔍 Validating flows...
✅ All flows validated successfully
🔌 Checking plugins...
⬇️ Downloading remote plugin: kubernetes@v1.2.0
✅ All plugins available
📋 Execution plan created:
  - deploy_flow (3 steps)
  - test_flow (2 steps)
Run 'corynth apply' to execute
```

### Apply Phase

When a user runs `corynth apply`:

1. Load the execution plan
2. Execute flows in the correct order
3. Manage dependencies between steps
4. Handle success/failure conditions
5. Generate detailed logs

Example output:

```
$ corynth apply deploy_flow
🚀 Executing flow: deploy_flow
  ✅ Step: clone_repo (2.3s)
  ✅ Step: setup_env (5.7s)
  ✅ Step: deploy (10.2s)
✨ Flow completed successfully!
```

## 📦 Plugin Specifications

### Built-in Plugin: Git

```yaml
steps:
  - name: "clone_repo"
    plugin: "git"
    action: "clone"
    params:
      repo: "https://github.com/user/repo.git"
      branch: "main"
      depth: 1
      directory: "./source"
```

Actions:
* `clone`: Clone a repository
* `checkout`: Switch branches
* `pull`: Update repository

### Built-in Plugin: Ansible

```yaml
steps:
  - name: "configure_servers"
    plugin: "ansible"
    action: "playbook"
    params:
      playbook: "configure.yml"
      inventory: "inventory.ini"
      extra_vars:
        env: "production"
```

Actions:
* `playbook`: Run an Ansible playbook
* `ad_hoc`: Run ad-hoc commands
* `inventory`: Generate dynamic inventory

### Built-in Plugin: Shell

```yaml
steps:
  - name: "build_app"
    plugin: "shell"
    action: "exec"
    params:
      command: "make build"
      working_dir: "./app"
      env:
        NODE_ENV: "production"
```

Actions:
* `exec`: Execute a command
* `script`: Run a script file
* `pipe`: Pipe output to another command

## 📚 Documentation

- [Getting Started](docs/getting_started.md)
- [Plugin Development](docs/plugin_development.md)
- [Advanced Flow Configuration](docs/advanced_flows.md)
- [State Management](docs/state_management.md)

## 🧪 Testing

```bash
go test ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.