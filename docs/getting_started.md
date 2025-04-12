# Getting Started with Corynth

Corynth is a powerful automation tool that enables users to define, plan, and execute sequential workflows through declarative YAML configurations. Built with Go, Corynth provides a familiar experience to Terraform users by following the init/plan/apply pattern while offering advanced automation capabilities.

## Installation

### Prerequisites

- Go 1.18 or higher
- Git
- Ansible (optional, for Ansible plugin)

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/corynth/corynth.git
   cd corynth
   ```

2. Build the binary:
   ```bash
   go build -o corynth
   ```

3. Install the binary:
   ```bash
   sudo mv corynth /usr/local/bin/
   ```

### Installing from Releases

1. Download the latest release for your platform from the [releases page](https://github.com/corynth/corynth/releases).

2. Make the binary executable:
   ```bash
   chmod +x corynth
   ```

3. Move the binary to your PATH:
   ```bash
   sudo mv corynth /usr/local/bin/
   ```

3. Install the binary:
   ```bash
   sudo mv corynth /usr/local/bin/
   ```

## Basic Usage

### Initializing a Project

To initialize a new Corynth project:

```bash
corynth init my_project
```

This will create a directory structure for your project:

```
my_project/
├── .corynth/
│   └── state.json
├── flows/
│   └── sample_flow.yaml
└── plugins/
    └── config.yaml
```

### Defining Flows

Flows are defined in YAML files in the `flows` directory. Here's an example flow:

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

### Planning Execution

To validate your flows and create an execution plan:

```bash
corynth plan my_project
```

This will:
1. Parse all flow YAML files
2. Validate dependencies between steps
3. Check for required plugins
4. Create an execution plan

### Applying Flows

To execute your flows:

```bash
corynth apply my_project
```

To execute a specific flow:

```bash
corynth apply my_project deployment_flow
```

## Core Plugins

### Git Plugin

The Git plugin provides actions for working with Git repositories:

- `clone`: Clone a repository
  ```yaml
  - name: "clone_repo"
    plugin: "git"
    action: "clone"
    params:
      repo: "https://github.com/user/repo.git"
      branch: "main"
      depth: 1
      directory: "./source"
  ```

- `checkout`: Switch branches
  ```yaml
  - name: "switch_branch"
    plugin: "git"
    action: "checkout"
    params:
      branch: "feature/new-feature"
      directory: "./source"
  ```

- `pull`: Update repository
  ```yaml
  - name: "update_repo"
    plugin: "git"
    action: "pull"
    params:
      directory: "./source"
  ```

### Shell Plugin

The Shell plugin provides actions for executing shell commands:

- `exec`: Execute a command
  ```yaml
  - name: "build_app"
    plugin: "shell"
    action: "exec"
    params:
      command: "make build"
      working_dir: "./app"
      env:
        NODE_ENV: "production"
  ```

- `script`: Run a script file
  ```yaml
  - name: "run_script"
    plugin: "shell"
    action: "script"
    params:
      script: "./scripts/deploy.sh"
      working_dir: "."
      env:
        DEPLOY_ENV: "production"
  ```

- `pipe`: Pipe output to another command
  ```yaml
  - name: "filter_logs"
    plugin: "shell"
    action: "pipe"
    params:
      command1: "cat logs/app.log"
      command2: "grep ERROR"
      working_dir: "."
  ```

### Ansible Plugin

The Ansible plugin provides actions for working with Ansible:

- `playbook`: Run an Ansible playbook
  ```yaml
  - name: "configure_servers"
    plugin: "ansible"
    action: "playbook"
    params:
      playbook: "playbooks/configure.yml"
      inventory: "inventory.ini"
      extra_vars:
        env: "production"
  ```

- `ad_hoc`: Run ad-hoc commands
  ```yaml
  - name: "check_uptime"
    plugin: "ansible"
    action: "ad_hoc"
    params:
      hosts: "all"
      module: "command"
      args: "uptime"
      inventory: "inventory.ini"
  ```

## Flow Dependencies

Steps can have dependencies on other steps:

```yaml
- name: "deploy"
  plugin: "ansible"
  action: "playbook"
  params:
    playbook: "deploy.yml"
  depends_on:
    - step: "setup_env"
      status: "success"
```

## Flow Chaining

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

## Next Steps

- [Plugin Development](plugin_development.md)
- [Advanced Flow Configuration](advanced_flows.md)
- [State Management](state_management.md)