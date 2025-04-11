# 💾 State Management in Corynth

## 🌟 Introduction

Corynth uses a state management system to track the execution of flows and steps. This guide explains how state is managed in Corynth and how you can use it effectively.

## 📊 State Structure

The Corynth state is stored in a JSON file in the `.corynth` directory of your project. The state file has the following structure:

```json
{
  "last_apply": "2023-04-11T13:45:30Z",
  "flows": {
    "flow1": {
      "status": "success",
      "start_time": "2023-04-11T13:45:00Z",
      "end_time": "2023-04-11T13:45:30Z",
      "duration": 30,
      "steps": {
        "step1": {
          "status": "success",
          "start_time": "2023-04-11T13:45:00Z",
          "end_time": "2023-04-11T13:45:10Z",
          "duration": 10,
          "output": "Step 1 output",
          "error": ""
        },
        "step2": {
          "status": "success",
          "start_time": "2023-04-11T13:45:10Z",
          "end_time": "2023-04-11T13:45:20Z",
          "duration": 10,
          "output": "Step 2 output",
          "error": ""
        },
        "step3": {
          "status": "success",
          "start_time": "2023-04-11T13:45:20Z",
          "end_time": "2023-04-11T13:45:30Z",
          "duration": 10,
          "output": "Step 3 output",
          "error": ""
        }
      }
    }
  }
}
```

## 🔄 State Lifecycle

### 1️⃣ Initialization

When you run `corynth init`, an empty state file is created:

```json
{}
```

### 2️⃣ Planning

When you run `corynth plan`, the state is read but not modified. The plan command validates your flows and creates an execution plan.

### 3️⃣ Applying

When you run `corynth apply`, the state is updated with the results of the execution:

1. Before execution, the state is loaded
2. During execution, step results are collected
3. After execution, the state is updated with the results
4. The updated state is saved back to the state file

## 📖 Reading State

You can read the state file to see the results of previous executions:

```bash
cat .corynth/state.json
```

You can also use the state in your flows with variable interpolation:

```yaml
flow:
  name: "state_flow"
  description: "Flow using state"
  steps:
    - name: "check_previous_run"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'Previous run status: ${state.flows.flow1.status}'"
```

## 🔍 State Inspection

### Command Line

You can inspect the state using the `corynth state` command:

```bash
# Show the entire state
corynth state

# Show the state of a specific flow
corynth state flow1

# Show the state of a specific step
corynth state flow1 step1
```

### Programmatic Access

You can access the state programmatically using the Corynth API:

```go
package main

import (
	"fmt"
	"github.com/corynth/corynth/internal/engine"
)

func main() {
	// Create a state manager
	stateManager := engine.NewStateManager()

	// Load state
	state, err := stateManager.LoadState("./my_project")
	if err != nil {
		fmt.Printf("Error loading state: %s\n", err)
		return
	}

	// Access state
	fmt.Printf("Last apply: %s\n", state.LastApply)
	for flowName, flowState := range state.Flows {
		fmt.Printf("Flow: %s, Status: %s\n", flowName, flowState.Status)
		for stepName, stepState := range flowState.Steps {
			fmt.Printf("  Step: %s, Status: %s\n", stepName, stepState.Status)
		}
	}
}
```

## 🔄 State Manipulation

### Resetting State

You can reset the state by deleting the state file:

```bash
rm .corynth/state.json
```

Or by using the `corynth reset` command:

```bash
# Reset the entire state
corynth reset

# Reset the state of a specific flow
corynth reset flow1

# Reset the state of a specific step
corynth reset flow1 step1
```

### Importing and Exporting State

You can import and export state:

```bash
# Export state to a file
corynth state export state.json

# Import state from a file
corynth state import state.json
```

## 🔒 State Locking

Corynth uses state locking to prevent concurrent modifications:

```bash
# Lock state
corynth state lock

# Unlock state
corynth state unlock
```

When you run `corynth apply`, the state is automatically locked before execution and unlocked after execution.

## 🌐 Remote State

### Configuring Remote State

You can configure Corynth to store state in a remote location:

```yaml
# .corynth/config.yaml
state:
  type: "remote"
  backend: "s3"
  config:
    bucket: "my-corynth-state"
    key: "my-project/state.json"
    region: "us-west-2"
```

Supported backends:

- **S3**: Amazon S3
- **GCS**: Google Cloud Storage
- **Azure**: Azure Blob Storage
- **HTTP**: HTTP/HTTPS endpoint
- **Consul**: HashiCorp Consul
- **Etcd**: Etcd

### Using Remote State

When remote state is configured, Corynth automatically uses it:

```bash
# Initialize with remote state
corynth init --remote-state

# Plan with remote state
corynth plan

# Apply with remote state
corynth apply
```

### State Locking with Remote State

Remote state backends support locking to prevent concurrent modifications:

```yaml
# .corynth/config.yaml
state:
  type: "remote"
  backend: "s3"
  config:
    bucket: "my-corynth-state"
    key: "my-project/state.json"
    region: "us-west-2"
    lock_table: "my-corynth-locks"
```

## 🔄 State Versioning

Corynth supports state versioning to track changes to the state:

```yaml
# .corynth/config.yaml
state:
  versioning: true
```

When versioning is enabled, Corynth creates a new version of the state file each time it's updated.

### Viewing State History

You can view the state history:

```bash
# Show state history
corynth state history

# Show a specific state version
corynth state show --version 1
```

### Rolling Back State

You can roll back to a previous state version:

```bash
# Roll back to a specific version
corynth state rollback --version 1
```

## 🔍 State Diffing

You can compare different state versions:

```bash
# Compare current state with previous version
corynth state diff

# Compare specific versions
corynth state diff --version 1 --version 2
```

## 📊 State Visualization

You can visualize the state as a graph:

```bash
# Generate a graph of the state
corynth state graph --output state.dot

# Convert to PNG
dot -Tpng state.dot -o state.png
```

## 🔄 State Migration

When you upgrade Corynth, the state format may change. Corynth automatically migrates the state to the new format:

```bash
# Migrate state to the latest format
corynth state migrate
```

## 🔍 Debugging State

You can enable debug logging for state operations:

```bash
# Enable debug logging
corynth --debug state
```

## 🔒 State Security

### Encryption

You can encrypt the state file:

```yaml
# .corynth/config.yaml
state:
  encryption:
    enabled: true
    key: "${STATE_ENCRYPTION_KEY}"
```

### Access Control

When using remote state, you can configure access control:

```yaml
# .corynth/config.yaml
state:
  type: "remote"
  backend: "s3"
  config:
    bucket: "my-corynth-state"
    key: "my-project/state.json"
    region: "us-west-2"
    acl: "private"
```

## 🔍 Best Practices

1. **Version Control**: Store your state configuration in version control, but not the state file itself
2. **Remote State**: Use remote state for team collaboration
3. **State Locking**: Enable state locking to prevent concurrent modifications
4. **Encryption**: Encrypt sensitive data in the state
5. **Backup**: Regularly backup your state
6. **Versioning**: Enable state versioning to track changes
7. **Access Control**: Restrict access to the state file
8. **Monitoring**: Monitor state operations for errors
9. **Documentation**: Document your state configuration
10. **Testing**: Test state operations in development before using them in production