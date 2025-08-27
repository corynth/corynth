# Workflow Syntax Guide

Comprehensive guide to Corynth's HCL-based workflow syntax.

## Basic Workflow Structure

```hcl
workflow "workflow-name" {
  description = "What this workflow does"
  version     = "1.0.0"

  step "step-name" {
    plugin = "plugin-name"
    action = "action-name"
    
    params = {
      parameter1 = "value1"
      parameter2 = "value2"
    }
  }
}
```

## Workflow Block

### Required Fields

```hcl
workflow "my-workflow" {
  description = "Brief description of workflow purpose"
  version     = "1.0.0"
  
  # Steps defined here
}
```

### Optional Fields

```hcl
workflow "advanced-workflow" {
  description = "Advanced workflow with all options"
  version     = "1.0.0"
  
  # Metadata
  metadata = {
    author = "Your Name"
    team   = "DevOps"
    env    = "production"
  }
  
  # Global outputs
  outputs = {
    result = "final_step.output"
    status = "success"
  }
  
  # Steps defined here
}
```

## Variables

### Variable Declaration

```hcl
workflow "with-variables" {
  description = "Workflow using variables"
  version = "1.0.0"

  variable "api_endpoint" {
    type        = string
    default     = "https://api.example.com"
    description = "API endpoint URL"
  }

  variable "timeout" {
    type        = number
    default     = 30
    description = "Request timeout in seconds"
    required    = false
  }

  variable "api_key" {
    type        = string
    description = "API key for authentication"
    required    = true
    sensitive   = true
  }

  step "api_call" {
    plugin = "http"
    action = "get"
    
    params = {
      url     = var.api_endpoint
      timeout = var.timeout
      headers = {
        "Authorization" = "Bearer ${var.api_key}"
      }
    }
  }
}
```

### Variable Types

- **string** - Text values
- **number** - Numeric values (int or float)
- **bool** - Boolean true/false
- **list** - Array of values
- **map** - Key-value pairs

### Variable Usage

```hcl
# Reference variables with var.name
url = var.api_endpoint

# String interpolation
message = "Hello ${var.username}!"

# Complex expressions
timeout = var.base_timeout * 2
```

## Steps

### Basic Step Structure

```hcl
step "step-name" {
  plugin = "plugin-name"
  action = "action-name"
  
  params = {
    key = "value"
  }
}
```

### Step Dependencies

```hcl
step "first" {
  plugin = "shell"
  action = "exec"
  
  params = {
    command = "echo 'First step'"
  }
}

step "second" {
  plugin = "shell"
  action = "exec"
  depends_on = ["first"]
  
  params = {
    command = "echo 'Second step'"
  }
}

step "third" {
  plugin = "shell"
  action = "exec"
  depends_on = ["first", "second"]
  
  params = {
    command = "echo 'Third step - depends on both'"
  }
}
```

### Step Output References

```hcl
step "api_call" {
  plugin = "http"
  action = "get"
  
  params = {
    url = "https://api.github.com/users/octocat"
  }
}

step "process_response" {
  plugin = "shell"
  action = "exec"
  depends_on = ["api_call"]
  
  params = {
    command = "echo 'Status: ${api_call.status_code}'"
  }
}

step "save_data" {
  plugin = "file"
  action = "write"
  depends_on = ["api_call"]
  
  params = {
    path = "/tmp/user-data.json"
    content = "${api_call.body}"
  }
}
```

## Advanced Features

### Boolean Logic & Conditional Execution

Steps can be conditionally executed based on boolean expressions:

```hcl
step "production_deploy" {
  plugin = "shell"
  action = "exec"
  condition = "var.environment == \"production\""
  
  params = {
    command = "echo 'Deploying to production'"
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

#### Available Boolean Functions

- `and(a, b)` - Logical AND
- `or(a, b)` - Logical OR  
- `not(a)` - Logical NOT
- `if(condition, true_value, false_value)` - Conditional expression
- `equal(a, b)` - Equality comparison
- `notequal(a, b)` - Inequality comparison
- `lessthan(a, b)` - Less than comparison
- `lessequal(a, b)` - Less than or equal comparison
- `greaterthan(a, b)` - Greater than comparison
- `greaterequal(a, b)` - Greater than or equal comparison

#### Boolean Expression Examples

```hcl
# Simple equality
condition = "var.environment == \"staging\""

# Complex boolean logic
condition = "and(var.deploy_enabled, or(equal(var.env, \"dev\"), equal(var.env, \"staging\")))"

# Negation
condition = "not(var.skip_tests)"

# Conditional expression in parameter
params = {
  timeout = "if(var.environment == \"production\", 60, 30)"
}
```

### Loop Processing

Iterate over lists, tuples, and objects using loop blocks:

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

step "deploy_environments" {
  plugin = "shell"
  action = "exec"
  
  loop {
    over = "[\"dev\", \"staging\", \"prod\"]"
    variable = "env"
  }
  
  params = {
    command = "echo 'Deploying to environment'"
  }
}
```

#### Loop Syntax

```hcl
loop {
  over = "expression_that_evaluates_to_list"
  variable = "loop_variable_name"
}
```

#### Loop with Variables

```hcl
variable "target_servers" {
  type = list(string)
  default = ["server1", "server2", "server3"]
}

step "server_maintenance" {
  plugin = "shell"
  action = "exec"
  
  loop {
    over = "var.target_servers"
    variable = "server"
  }
  
  params = {
    command = "echo 'Maintaining server'"
  }
}
```

#### Combined Conditions and Loops

```hcl
step "conditional_deployment" {
  plugin = "shell"
  action = "exec"
  condition = "var.deploy_enabled"
  
  loop {
    over = "[\"web1\", \"web2\", \"web3\"]"
    variable = "server"
  }
  
  params = {
    command = "echo 'Deploying to server (conditional)'"
  }
}
```

### Retry Policies

```hcl
step "unreliable_operation" {
  plugin = "http"
  action = "get"
  
  params = {
    url = "https://unreliable-api.example.com"
  }
  
  retry {
    max_attempts = 3
    delay        = "5s"
    backoff      = "exponential"
  }
}
```

### Error Handling

```hcl
step "risky_operation" {
  plugin = "shell"
  action = "exec"
  
  params = {
    command = "potentially-failing-command"
  }
  
  continue_on {
    error   = true
    failure = false
  }
}
```

### Timeouts

```hcl
step "long_operation" {
  plugin = "shell"
  action = "exec"
  timeout = "10m"
  
  params = {
    command = "long-running-process"
  }
}
```

### Step Templates

```hcl
workflow "with-templates" {
  description = "Using step templates"
  version = "1.0.0"

  template "api_call" {
    plugin = "http"
    action = "get"
    
    defaults = {
      timeout = "30"
      headers = {
        "User-Agent" = "Corynth/1.2.0"
      }
    }
  }

  step "call_api_1" {
    template = "api_call"
    
    params = {
      url = "https://api.example.com/endpoint1"
    }
  }

  step "call_api_2" {
    template = "api_call"
    
    params = {
      url = "https://api.example.com/endpoint2"
      timeout = "60"  # Override default
    }
  }
}
```

## Parallel Execution

Steps without dependencies run in parallel automatically:

```hcl
workflow "parallel-example" {
  description = "Parallel execution example"
  version = "1.0.0"

  # These three steps run in parallel
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
      content = "Task 3 result"
    }
  }

  # This step waits for all parallel steps
  step "combine_results" {
    plugin = "shell"
    action = "exec"
    depends_on = ["task_1", "task_2", "task_3"]
    
    params = {
      command = "echo 'All tasks completed: ${task_1.output}, ${task_2.status_code}, ${task_3.success}'"
    }
  }
}
```

## Complex Example

Here's a comprehensive example showcasing multiple features:

```hcl
workflow "complex-pipeline" {
  description = "Complex CI/CD-like pipeline"
  version = "1.0.0"

  variable "repo_url" {
    type = string
    default = "https://github.com/example/project.git"
    description = "Repository to process"
  }

  variable "api_endpoint" {
    type = string
    default = "https://api.example.com"
    description = "Deployment API endpoint"
  }

  # Setup phase
  step "prepare_workspace" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "mkdir -p /tmp/pipeline/{src,build,dist}"
    }
  }

  # Source phase (parallel)
  step "clone_source" {
    plugin = "git"
    action = "clone"
    depends_on = ["prepare_workspace"]
    
    params = {
      url = var.repo_url
      path = "/tmp/pipeline/src"
      branch = "main"
    }
  }

  step "fetch_dependencies" {
    plugin = "http"
    action = "get"
    depends_on = ["prepare_workspace"]
    
    params = {
      url = "${var.api_endpoint}/dependencies"
      timeout = 60
    }
  }

  # Build phase
  step "compile" {
    plugin = "shell"
    action = "exec"
    depends_on = ["clone_source", "fetch_dependencies"]
    
    params = {
      command = "cd /tmp/pipeline/src && make build"
    }
    
    retry {
      max_attempts = 2
      delay = "10s"
    }
  }

  # Test phase
  step "run_tests" {
    plugin = "shell"
    action = "exec"
    depends_on = ["compile"]
    timeout = "5m"
    
    params = {
      command = "cd /tmp/pipeline/src && make test"
    }
  }

  # Package phase
  step "create_package" {
    plugin = "shell"
    action = "exec"
    depends_on = ["run_tests"]
    condition = "${run_tests.exit_code == 0}"
    
    params = {
      command = "cd /tmp/pipeline/src && make package && mv dist/* /tmp/pipeline/dist/"
    }
  }

  # Deploy phase
  step "deploy" {
    plugin = "http"
    action = "post"
    depends_on = ["create_package"]
    
    params = {
      url = "${var.api_endpoint}/deploy"
      body = "Package ready for deployment"
      headers = {
        "Content-Type" = "application/json"
      }
    }
  }

  # Cleanup
  step "cleanup" {
    plugin = "shell"
    action = "exec"
    depends_on = ["deploy"]
    
    params = {
      command = "rm -rf /tmp/pipeline"
    }
    
    continue_on {
      error = true  # Always cleanup, even if deploy fails
    }
  }
}
```

## Best Practices

### 1. Naming Conventions
- Use lowercase with hyphens for workflow names: `deploy-to-staging`
- Use descriptive step names: `clone_repository`, `run_unit_tests`
- Use semantic versioning: `1.0.0`, `1.1.0`, `2.0.0`

### 2. Dependencies
- Keep dependency chains simple and clear
- Use parallel execution for independent operations
- Add explicit dependencies even if they seem obvious

### 3. Error Handling
- Add retry policies for unreliable operations
- Use timeouts for all external calls
- Implement proper cleanup steps

### 4. Security
- Use variables for sensitive data
- Never hardcode secrets in workflow files
- Use environment variables for credentials

### 5. Performance
- Minimize step dependencies to enable parallelism
- Use appropriate timeouts
- Clean up temporary resources

## Syntax Reference

### Comments
```hcl
# This is a comment
workflow "example" {
  # Another comment
  description = "Example workflow"
}
```

### Data Types
```hcl
params = {
  string_value = "text"
  number_value = 42
  boolean_value = true
  list_value = ["item1", "item2", "item3"]
  object_value = {
    nested_key = "nested_value"
    nested_number = 100
  }
}
```

### String Interpolation
```hcl
params = {
  simple_interpolation = "${variable_name}"
  complex_interpolation = "Result: ${step_name.output_field}"
  conditional = "${condition ? 'yes' : 'no'}"
}
```

### Functions
```hcl
params = {
  timestamp = timestamp()
  upper_case = upper("hello world")
  math_result = max(10, 20, 5)
}
```

## Next Steps

- Learn about [Dependencies & Parallel Execution](dependencies.md)
- Explore [Plugin System](../plugins/overview.md)
- Try [Example Workflows](../examples/)
- Read [Configuration Guide](configuration.md)

---

**Master Corynth's workflow syntax to build powerful automation solutions!** 🎯