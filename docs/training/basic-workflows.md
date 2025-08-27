# Basic Workflows Tutorial

Learn to create and execute Corynth workflows through hands-on examples.

## Prerequisites

- [Corynth installed](../user-guide/installation.md)
- Basic familiarity with command line
- Text editor for creating workflow files

## Tutorial Overview

This tutorial covers:
1. Creating your first workflow
2. Understanding workflow structure
3. Using parameters and variables
4. Working with multiple steps
5. Handling dependencies

## Lesson 1: Your First Workflow

Let's start with the simplest possible workflow:

### Create Project Directory

```bash
mkdir corynth-tutorial
cd corynth-tutorial
corynth init
```

### Create Hello World Workflow

Create `hello.hcl`:

```hcl
workflow "hello" {
  description = "My first workflow"
  version = "1.0.0"

  step "greet" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Hello, Corynth!'"
    }
  }
}
```

### Execute the Workflow

```bash
corynth apply hello.hcl
```

**Expected Output**:
```
• Loading workflow from hello.hcl...
Starting workflow execution...
Hello, Corynth!

✓ ✓ Workflow completed: hello
Duration: 0s
Steps: 1/1
```

**What happened?**:
1. Corynth parsed the HCL file
2. Loaded the built-in `shell` plugin
3. Executed the `exec` action with the command parameter
4. Displayed the command output
5. Saved execution state

### Check Execution History

```bash
corynth state list
```

You'll see your workflow execution recorded with a unique ID, status, and duration.

## Lesson 2: Variables and Flow Control

Variables make workflows reusable and configurable. Flow control enables conditional execution and loops.

Create `variables.hcl`:

```hcl
workflow "variables-demo" {
  description = "Demonstrating workflow variables and boolean logic"
  version = "1.0.0"

  variable "name" {
    type = string
    default = "World"
    description = "Name to greet"
  }

  variable "environment" {
    type = string
    default = "development"
    description = "Target environment"
  }

  variable "uppercase" {
    type = bool
    default = false
    description = "Convert to uppercase"
  }

  step "greet_person" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Hello, ${var.name}!'"
    }
  }

  step "environment_check" {
    plugin = "shell"
    action = "exec"
    condition = "equal(var.environment, \"production\")"
    depends_on = ["greet_person"]
    
    params = {
      command = "echo 'Running in production environment - extra caution required!'"
    }
  }

  step "maybe_uppercase" {
    plugin = "shell"
    action = "exec"
    depends_on = ["greet_person"]
    condition = "var.uppercase"
    
    params = {
      command = "echo 'HELLO, ${upper(var.name)}!'"
    }
  }

  step "complex_condition" {
    plugin = "shell"
    action = "exec"
    condition = "and(var.uppercase, not(equal(var.environment, \"development\")))"
    depends_on = ["greet_person"]
    
    params = {
      command = "echo 'Uppercase greeting for non-development environment'"
    }
  }
}
```

### Execute with Default Variables

```bash
corynth apply variables.hcl
```

### Execute with Custom Variables

```bash
# Test boolean logic and conditions
corynth apply variables.hcl -var name="Alice" -var uppercase=true -var environment="production"
```

**Expected Output**:
```
Starting workflow execution...
Hello, Alice!
Running in production environment - extra caution required!
HELLO, ALICE!
Uppercase greeting for non-development environment

✓ ✓ Workflow completed: variables-demo
Duration: 1s
Steps: 4/4
```

**What happened?**:
1. `greet_person` always executes
2. `environment_check` executed because environment equals "production"
3. `maybe_uppercase` executed because uppercase is true
4. `complex_condition` executed because uppercase is true AND environment is not "development"

## Lesson 3: Loop Processing

Learn how to iterate over collections and process multiple items.

Create `loops.hcl`:

```hcl
workflow "loops-demo" {
  description = "Demonstrating loop processing"
  version = "1.0.0"

  variable "servers" {
    type = list(string)
    default = ["web1", "web2", "web3"]
    description = "List of servers to process"
  }

  step "process_servers" {
    plugin = "shell"
    action = "exec"
    
    loop {
      over = "var.servers"
      variable = "server"
    }
    
    params = {
      command = "echo 'Processing server: ${loop.server}'"
    }
  }

  step "deploy_environments" {
    plugin = "shell"
    action = "exec"
    depends_on = ["process_servers"]
    
    loop {
      over = "[\"dev\", \"staging\", \"prod\"]"
      variable = "env"
    }
    
    params = {
      command = "echo 'Deploying to environment: ${loop.env}'"
    }
  }

  step "conditional_loop" {
    plugin = "shell"
    action = "exec"
    condition = "equal(length(var.servers), 3)"
    depends_on = ["deploy_environments"]
    
    loop {
      over = "var.servers"
      variable = "server"
    }
    
    params = {
      command = "echo 'Conditional processing for server: ${loop.server}'"
    }
  }
}
```

### Execute Loop Workflow

```bash
corynth apply loops.hcl
```

**Expected Output**:
```
Starting workflow execution...
Processing server: web1
Processing server: web2
Processing server: web3
Deploying to environment: dev
Deploying to environment: staging
Deploying to environment: prod
Conditional processing for server: web1
Conditional processing for server: web2
Conditional processing for server: web3

✓ ✓ Workflow completed: loops-demo
Duration: 2s
Steps: 3/3 (9 iterations total)
```

**What happened?**:
1. `process_servers` looped over the servers variable, executing 3 times
2. `deploy_environments` looped over inline array, executing 3 times
3. `conditional_loop` executed because server count equals 3, then looped 3 times
4. Each loop iteration creates a separate execution with access to the loop variable

## Lesson 4: Multi-Step Workflow

Create a workflow with multiple dependent steps.

Create `multi-step.hcl`:

```hcl
workflow "multi-step-demo" {
  description = "Multi-step workflow with dependencies"
  version = "1.0.0"

  step "setup" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "mkdir -p /tmp/tutorial && echo 'Workspace created'"
    }
  }

  step "create_data" {
    plugin = "shell"
    action = "exec"
    depends_on = ["setup"]
    
    params = {
      command = "echo 'Step 1 completed at $(date)' > /tmp/tutorial/data.txt"
    }
  }

  step "process_data" {
    plugin = "shell"
    action = "exec"
    depends_on = ["create_data"]
    
    params = {
      command = "echo 'Processing...' && cat /tmp/tutorial/data.txt"
    }
  }

  step "cleanup" {
    plugin = "shell"
    action = "exec"
    depends_on = ["process_data"]
    
    params = {
      command = "rm -rf /tmp/tutorial && echo 'Cleanup completed'"
    }
  }
}
```

### Execute Multi-Step Workflow

```bash
corynth apply multi-step.hcl
```

**Expected Output**:
```
Starting workflow execution...
Workspace created
Processing...
Step 1 completed at Wed Aug 20 12:00:00 BST 2025
Cleanup completed

✓ ✓ Workflow completed: multi-step-demo
Duration: 1s
Steps: 4/4
```

## Lesson 5: Using Remote Plugins

Now let's use plugins that auto-install from repositories.

Create `remote-plugins.hcl`:

```hcl
workflow "remote-plugins-demo" {
  description = "Using auto-installed remote plugins"
  version = "1.0.0"

  step "fetch_data" {
    plugin = "http"  # Will auto-install
    action = "get"
    
    params = {
      url = "https://api.github.com/users/octocat"
      timeout = 30
    }
  }

  step "save_data" {
    plugin = "file"  # Will auto-install
    action = "write"
    depends_on = ["fetch_data"]
    
    params = {
      path = "/tmp/github-user.json"
      content = "${fetch_data.body}"
    }
  }

  step "verify_data" {
    plugin = "shell"
    action = "exec"
    depends_on = ["save_data"]
    
    params = {
      command = "echo 'File size:' && wc -c /tmp/github-user.json"
    }
  }

  step "cleanup" {
    plugin = "shell"
    action = "exec"
    depends_on = ["verify_data"]
    
    params = {
      command = "rm /tmp/github-user.json && echo 'File removed'"
    }
  }
}
```

### Execute Remote Plugin Workflow

```bash
corynth apply remote-plugins.hcl
```

**First execution**:
- HTTP plugin will be auto-installed (takes ~15 seconds)
- File plugin will be auto-installed (takes ~8 seconds)
- Workflow executes successfully

**Subsequent executions**:
- Plugins are cached, execution is immediate

## Lesson 6: Parallel Execution

Create workflows where independent steps run concurrently.

Create `parallel.hcl`:

```hcl
workflow "parallel-demo" {
  description = "Parallel execution demonstration"
  version = "1.0.0"

  # These three steps run in parallel (no dependencies)
  step "task_1" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Task 1 starting' && sleep 2 && echo 'Task 1 completed'"
    }
  }

  step "task_2" {
    plugin = "http"
    action = "get"
    
    params = {
      url = "https://httpbin.org/delay/1"
      timeout = 10
    }
  }

  step "task_3" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Task 3 starting' && sleep 1 && echo 'Task 3 completed'"
    }
  }

  # This step waits for all parallel steps to complete
  step "summarize" {
    plugin = "shell"
    action = "exec"
    depends_on = ["task_1", "task_2", "task_3"]
    
    params = {
      command = "echo 'All parallel tasks completed! HTTP status: ${task_2.status_code}'"
    }
  }
}
```

### Execute Parallel Workflow

```bash
corynth apply parallel.hcl
```

**Expected Behavior**:
- Tasks 1, 2, and 3 start simultaneously
- Each task runs independently
- The summarize step waits for all to complete
- Total execution time ≈ longest individual task time

## Lesson 7: Error Handling

Learn how to handle failures gracefully.

Create `error-handling.hcl`:

```hcl
workflow "error-handling-demo" {
  description = "Error handling and recovery"
  version = "1.0.0"

  step "might_fail" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "exit 1"  # This will fail
    }
    
    retry {
      max_attempts = 3
      delay = "2s"
    }
    
    continue_on {
      error = true  # Continue even if this fails
    }
  }

  step "check_failure" {
    plugin = "shell"
    action = "exec"
    depends_on = ["might_fail"]
    
    params = {
      command = "echo 'Previous step failed as expected, continuing...'"
    }
  }

  step "guaranteed_success" {
    plugin = "shell"
    action = "exec"
    depends_on = ["check_failure"]
    
    params = {
      command = "echo 'This step always succeeds!'"
    }
  }
}
```

### Execute Error Handling Workflow

```bash
corynth apply error-handling.hcl
```

**Expected Behavior**:
- First step fails but retries 3 times
- Workflow continues due to `continue_on.error = true`
- Subsequent steps execute normally
- Overall workflow status depends on final steps

## Lesson 8: Real-World Example

Create a practical workflow that demonstrates real-world usage.

Create `backup-workflow.hcl`:

```hcl
workflow "backup-workflow" {
  description = "Backup important data and notify team"
  version = "1.0.0"

  variable "backup_dir" {
    type = string
    default = "/tmp/backup"
    description = "Directory to store backups"
  }

  variable "source_repo" {
    type = string
    default = "https://github.com/corynth/corynthplugins.git"
    description = "Repository to backup"
  }

  step "prepare_backup_dir" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "mkdir -p ${var.backup_dir} && echo 'Backup directory ready'"
    }
  }

  step "backup_repository" {
    plugin = "git"
    action = "clone"
    depends_on = ["prepare_backup_dir"]
    
    params = {
      url = var.source_repo
      path = "${var.backup_dir}/repo"
      branch = "main"
    }
  }

  step "create_manifest" {
    plugin = "file"
    action = "write"
    depends_on = ["backup_repository"]
    
    params = {
      path = "${var.backup_dir}/manifest.json"
      content = "{\"backup_date\": \"$(date -Iseconds)\", \"source\": \"${var.source_repo}\", \"commit\": \"${backup_repository.commit}\"}"
    }
  }

  step "verify_backup" {
    plugin = "shell"
    action = "exec"
    depends_on = ["create_manifest"]
    
    params = {
      command = "ls -la ${var.backup_dir} && echo 'Backup verification complete'"
    }
  }

  step "health_check" {
    plugin = "http"
    action = "get"
    depends_on = ["verify_backup"]
    
    params = {
      url = "https://httpbin.org/status/200"
      timeout = 10
    }
  }

  step "cleanup_old_backups" {
    plugin = "shell"
    action = "exec"
    depends_on = ["health_check"]
    
    params = {
      command = "echo 'Cleaning up old backups...' && rm -rf ${var.backup_dir}"
    }
  }
}
```

### Execute Backup Workflow

```bash
corynth apply backup-workflow.hcl
```

## Practice Exercises

### Exercise 1: File Processing Pipeline

Create a workflow that:
1. Downloads a file from a URL
2. Processes the file contents
3. Saves the processed results
4. Sends a completion notification

### Exercise 2: API Integration

Create a workflow that:
1. Fetches data from multiple APIs in parallel
2. Combines the results
3. Stores the combined data
4. Performs validation checks

### Exercise 3: Development Pipeline

Create a workflow that:
1. Clones a Git repository
2. Runs tests
3. Builds the project
4. Deploys if tests pass
5. Cleans up temporary files

## Common Patterns

### 1. Setup → Process → Cleanup

```hcl
step "setup" { /* prepare environment */ }
step "process" { depends_on = ["setup"] /* main work */ }
step "cleanup" { depends_on = ["process"] /* clean up */ }
```

### 2. Parallel Processing → Combine

```hcl
step "task_1" { /* independent task */ }
step "task_2" { /* independent task */ }
step "task_3" { /* independent task */ }
step "combine" { depends_on = ["task_1", "task_2", "task_3"] }
```

### 3. Conditional Execution

```hcl
step "check" { /* verification step */ }
step "action" { 
  depends_on = ["check"]
  condition = "check.success"
}
```

### 4. Loop Processing

```hcl
step "batch_process" {
  loop {
    over = "var.items"
    variable = "item"
  }
  params = {
    target = "${loop.item}"
  }
}
```

### 5. Combined Flow Control

```hcl
step "conditional_loop" {
  condition = "var.enabled"
  loop {
    over = "var.servers"
    variable = "server"
  }
  params = {
    command = "process ${loop.server}"
  }
}
```

### 6. Error Recovery

```hcl
step "risky" {
  retry { max_attempts = 3 }
  continue_on { error = true }
}
step "fallback" { depends_on = ["risky"] }
```

## Best Practices from Lessons

### 1. Workflow Design
- Start simple and add complexity gradually
- Use descriptive names for workflows and steps
- Include meaningful descriptions
- Version your workflows

### 2. Dependencies
- Make dependencies explicit
- Use parallel execution for independent tasks
- Keep dependency chains simple

### 3. Error Handling
- Add retry policies for unreliable operations
- Use continue_on for non-critical steps
- Always include cleanup steps

### 4. Testing
- Validate workflows before execution
- Use plan mode to preview changes
- Test with different variable values

## Next Steps

- [Advanced Workflows](advanced-workflows.md) - Complex patterns and techniques
- [API Integration Tutorial](api-integration.md) - Working with external APIs
- [Plugin Development](../plugins/development.md) - Creating custom plugins
- [Real-World Examples](../examples/) - Production workflow examples

## Summary

You've learned:
- ✅ Basic workflow structure
- ✅ Using variables and parameters
- ✅ Boolean logic and conditional execution
- ✅ Loop processing and iteration
- ✅ Creating multi-step workflows
- ✅ Understanding dependencies
- ✅ Working with different plugins
- ✅ Handling errors and retries
- ✅ Parallel execution patterns

**Ready to build more complex workflows!** 🎯

## Quick Reference

### Essential Commands
```bash
corynth init                    # Initialize project
corynth validate workflow.hcl   # Check syntax
corynth plan workflow.hcl       # Preview execution
corynth apply workflow.hcl      # Execute workflow
corynth state list              # View history
```

### Basic Workflow Template
```hcl
workflow "name" {
  description = "What this does"
  version = "1.0.0"

  step "step_name" {
    plugin = "plugin_name"
    action = "action_name"
    
    params = {
      key = "value"
    }
  }
}
```

---

**Congratulations! You've mastered Corynth workflow basics.** 🎉