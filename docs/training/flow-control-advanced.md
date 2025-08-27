# Advanced Flow Control Tutorial

Master complex flow control patterns in Corynth workflows.

## Prerequisites

- [Basic Workflows Tutorial](basic-workflows.md) completed
- Understanding of boolean logic concepts
- Familiarity with programming loops and conditions

## Tutorial Overview

This advanced tutorial covers:
1. Complex boolean expressions
2. Nested conditional logic
3. Advanced loop patterns
4. Dynamic flow control
5. Performance optimization
6. Real-world use cases

## Lesson 1: Complex Boolean Expressions

Learn to build sophisticated conditional logic using multiple boolean functions.

Create `complex-boolean.hcl`:

```hcl
workflow "complex-boolean-demo" {
  description = "Advanced boolean logic patterns"
  version = "1.0.0"

  variable "environment" {
    type = string
    default = "staging"
    description = "Target environment"
  }

  variable "branch" {
    type = string
    default = "main"
    description = "Git branch name"
  }

  variable "tests_passing" {
    type = bool
    default = true
    description = "Test suite status"
  }

  variable "maintenance_mode" {
    type = bool
    default = false
    description = "System maintenance status"
  }

  # Simple condition
  step "environment_check" {
    plugin = "shell"
    action = "exec"
    condition = "equal(var.environment, \"production\")"
    
    params = {
      command = "echo 'Production environment detected'"
    }
  }

  # Complex AND condition
  step "safe_deployment" {
    plugin = "shell"
    action = "exec"
    condition = "and(var.tests_passing, not(var.maintenance_mode), or(equal(var.branch, \"main\"), equal(var.branch, \"release\")))"
    
    params = {
      command = "echo 'Safe to deploy: tests passing, no maintenance, and branch is main or release'"
    }
  }

  # Multiple nested conditions
  step "complex_check" {
    plugin = "shell"
    action = "exec"
    condition = "or(and(equal(var.environment, \"development\"), var.tests_passing), and(equal(var.environment, \"production\"), var.tests_passing, not(var.maintenance_mode)))"
    
    params = {
      command = "echo 'Complex condition met: (dev AND tests pass) OR (prod AND tests pass AND no maintenance)'"
    }
  }

  # Conditional parameter values
  step "dynamic_timeout" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Using timeout based on environment'"
      timeout = "if(equal(var.environment, \"production\"), \"300\", \"60\")"
    }
  }
}
```

### Execute Complex Boolean Workflow

```bash
# Test with production settings
corynth apply complex-boolean.hcl -var environment="production" -var branch="main" -var tests_passing=true -var maintenance_mode=false
```

**Expected Output**:
```
Starting workflow execution...
Production environment detected
Safe to deploy: tests passing, no maintenance, and branch is main or release
Complex condition met: (dev AND tests pass) OR (prod AND tests pass AND no maintenance)
Using timeout based on environment

✓ ✓ Workflow completed: complex-boolean-demo
Duration: 2s
Steps: 4/4
```

## Lesson 2: Advanced Loop Patterns

Explore sophisticated iteration patterns and loop combinations.

Create `advanced-loops.hcl`:

```hcl
workflow "advanced-loops-demo" {
  description = "Advanced loop processing patterns"
  version = "1.0.0"

  variable "environments" {
    type = list(string)
    default = ["dev", "staging", "prod"]
    description = "Target environments"
  }

  variable "services" {
    type = list(string)
    default = ["api", "web", "worker"]
    description = "Services to deploy"
  }

  variable "enable_monitoring" {
    type = bool
    default = true
    description = "Enable monitoring setup"
  }

  # Simple loop with conditions
  step "prepare_environments" {
    plugin = "shell"
    action = "exec"
    
    loop {
      over = "var.environments"
      variable = "env"
    }
    
    params = {
      command = "echo 'Preparing environment: ${loop.env}'"
    }
  }

  # Conditional loop execution
  step "setup_monitoring" {
    plugin = "shell"
    action = "exec"
    condition = "var.enable_monitoring"
    depends_on = ["prepare_environments"]
    
    loop {
      over = "var.environments"
      variable = "env"
    }
    
    params = {
      command = "echo 'Setting up monitoring for: ${loop.env}'"
    }
  }

  # Loop with complex conditions
  step "production_deployment" {
    plugin = "shell"
    action = "exec"
    depends_on = ["setup_monitoring"]
    
    loop {
      over = "var.services"
      variable = "service"
    }
    
    # Only deploy to production if service is not 'worker'
    condition = "and(contains(var.environments, \"prod\"), not(equal(loop.service, \"worker\")))"
    
    params = {
      command = "echo 'Deploying service ${loop.service} to production (workers excluded)'"
    }
  }

  # Nested-style processing (separate steps)
  step "deploy_api_services" {
    plugin = "shell"
    action = "exec"
    depends_on = ["production_deployment"]
    
    loop {
      over = "[\"api-gateway\", \"user-api\", \"order-api\"]"
      variable = "api"
    }
    
    params = {
      command = "echo 'Deploying API service: ${loop.api}'"
    }
  }

  # Dynamic loop with computed values
  step "scale_services" {
    plugin = "shell"
    action = "exec"
    depends_on = ["deploy_api_services"]
    
    loop {
      over = "[1, 2, 3, 5]"
      variable = "replicas"
    }
    
    params = {
      command = "echo 'Scaling service to ${loop.replicas} replicas'"
    }
  }
}
```

### Execute Advanced Loop Workflow

```bash
corynth apply advanced-loops.hcl -var enable_monitoring=true
```

**Expected Behavior**:
- Prepares 3 environments sequentially
- Sets up monitoring for each environment (conditional)
- Deploys 2 services to production (excludes worker)
- Deploys 3 API services
- Tests 4 different scaling configurations

## Lesson 3: Dynamic Flow Control

Create workflows that adapt their execution based on runtime conditions.

Create `dynamic-flow.hcl`:

```hcl
workflow "dynamic-flow-demo" {
  description = "Dynamic flow control based on runtime conditions"
  version = "1.0.0"

  variable "deployment_strategy" {
    type = string
    default = "blue-green"
    description = "Deployment strategy (blue-green, rolling, canary)"
  }

  variable "health_check_url" {
    type = string
    default = "https://httpbin.org/status/200"
    description = "Health check endpoint"
  }

  # Initial health check
  step "health_check" {
    plugin = "http"
    action = "get"
    
    params = {
      url = var.health_check_url
      timeout = 10
    }
  }

  # Blue-green deployment path
  step "blue_green_deploy" {
    plugin = "shell"
    action = "exec"
    condition = "and(equal(var.deployment_strategy, \"blue-green\"), equal(health_check.status_code, 200))"
    depends_on = ["health_check"]
    
    params = {
      command = "echo 'Executing blue-green deployment strategy'"
    }
  }

  # Rolling deployment path
  step "rolling_deploy" {
    plugin = "shell"
    action = "exec"
    condition = "and(equal(var.deployment_strategy, \"rolling\"), equal(health_check.status_code, 200))"
    depends_on = ["health_check"]
    
    params = {
      command = "echo 'Executing rolling deployment strategy'"
    }
  }

  # Canary deployment path
  step "canary_deploy" {
    plugin = "shell"
    action = "exec"
    condition = "and(equal(var.deployment_strategy, \"canary\"), equal(health_check.status_code, 200))"
    depends_on = ["health_check"]
    
    loop {
      over = "[5, 25, 50, 100]"
      variable = "percentage"
    }
    
    params = {
      command = "echo 'Canary deployment: ${loop.percentage}% traffic'"
    }
  }

  # Fallback for unhealthy systems
  step "health_check_failed" {
    plugin = "shell"
    action = "exec"
    condition = "not(equal(health_check.status_code, 200))"
    depends_on = ["health_check"]
    
    params = {
      command = "echo 'Health check failed, aborting deployment'"
    }
  }

  # Post-deployment verification (runs for any successful deployment)
  step "verify_deployment" {
    plugin = "http"
    action = "get"
    condition = "equal(health_check.status_code, 200)"
    depends_on = ["blue_green_deploy", "rolling_deploy", "canary_deploy"]
    
    params = {
      url = var.health_check_url
      timeout = 15
    }
  }
}
```

### Test Dynamic Flow Control

```bash
# Test blue-green deployment
corynth apply dynamic-flow.hcl -var deployment_strategy="blue-green"

# Test canary deployment (with loops)
corynth apply dynamic-flow.hcl -var deployment_strategy="canary"

# Test with failing health check
corynth apply dynamic-flow.hcl -var health_check_url="https://httpbin.org/status/500"
```

## Lesson 4: Performance Optimization Patterns

Learn techniques to optimize flow control performance.

Create `optimized-flow.hcl`:

```hcl
workflow "optimized-flow-demo" {
  description = "Performance-optimized flow control patterns"
  version = "1.0.0"

  variable "parallel_tasks" {
    type = list(string)
    default = ["task-a", "task-b", "task-c", "task-d"]
    description = "Independent tasks for parallel processing"
  }

  variable "batch_size" {
    type = number
    default = 2
    description = "Batch processing size"
  }

  # Parallel independent checks
  step "parallel_health_checks" {
    plugin = "http"
    action = "get"
    
    loop {
      over = "[\"https://httpbin.org/status/200\", \"https://httpbin.org/delay/1\", \"https://httpbin.org/json\"]"
      variable = "endpoint"
    }
    
    params = {
      url = "${loop.endpoint}"
      timeout = 5
    }
  }

  # Batch processing pattern
  step "batch_process_items" {
    plugin = "shell"
    action = "exec"
    depends_on = ["parallel_health_checks"]
    
    loop {
      over = "[\"batch-1\", \"batch-2\"]"
      variable = "batch"
    }
    
    params = {
      command = "echo 'Processing batch: ${loop.batch} with size ${var.batch_size}'"
    }
  }

  # Conditional early termination
  step "early_termination_check" {
    plugin = "shell"
    action = "exec"
    depends_on = ["batch_process_items"]
    condition = "lessthan(var.batch_size, 5)"
    
    params = {
      command = "echo 'Batch size is small, enabling fast processing mode'"
    }
  }

  # Conditional expensive operation
  step "expensive_operation" {
    plugin = "shell"
    action = "exec"
    depends_on = ["batch_process_items"]
    condition = "greaterequal(var.batch_size, 5)"
    
    params = {
      command = "echo 'Batch size is large, running expensive deep analysis'"
    }
  }

  # Final summary (always runs)
  step "summary" {
    plugin = "shell"
    action = "exec"
    depends_on = ["early_termination_check", "expensive_operation"]
    
    params = {
      command = "echo 'Workflow completed with optimized execution path'"
    }
  }
}
```

## Lesson 5: Real-World CI/CD Pipeline

Apply advanced flow control in a realistic continuous integration scenario.

Create `cicd-pipeline.hcl`:

```hcl
workflow "cicd-pipeline" {
  description = "Production CI/CD pipeline with advanced flow control"
  version = "1.0.0"

  variable "git_branch" {
    type = string
    default = "main"
    description = "Git branch to deploy"
  }

  variable "environment" {
    type = string
    default = "staging"
    description = "Target environment"
  }

  variable "run_security_scan" {
    type = bool
    default = true
    description = "Enable security scanning"
  }

  variable "services" {
    type = list(string)
    default = ["frontend", "backend", "database"]
    description = "Services to deploy"
  }

  # Source code checkout
  step "checkout" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Checking out branch: ${var.git_branch}'"
    }
  }

  # Parallel testing phase
  step "unit_tests" {
    plugin = "shell"
    action = "exec"
    depends_on = ["checkout"]
    
    params = {
      command = "echo 'Running unit tests for branch: ${var.git_branch}'"
    }
  }

  step "integration_tests" {
    plugin = "shell"
    action = "exec"
    depends_on = ["checkout"]
    
    params = {
      command = "echo 'Running integration tests'"
    }
  }

  # Security scan (conditional)
  step "security_scan" {
    plugin = "shell"
    action = "exec"
    condition = "var.run_security_scan"
    depends_on = ["checkout"]
    
    params = {
      command = "echo 'Running security vulnerability scan'"
    }
  }

  # Build phase (waits for all tests)
  step "build" {
    plugin = "shell"
    action = "exec"
    depends_on = ["unit_tests", "integration_tests", "security_scan"]
    
    params = {
      command = "echo 'Building application artifacts'"
    }
  }

  # Production-specific checks
  step "production_checks" {
    plugin = "shell"
    action = "exec"
    condition = "equal(var.environment, \"production\")"
    depends_on = ["build"]
    
    params = {
      command = "echo 'Running production-specific validations'"
    }
  }

  # Service deployment loop
  step "deploy_services" {
    plugin = "shell"
    action = "exec"
    depends_on = ["build", "production_checks"]
    
    loop {
      over = "var.services"
      variable = "service"
    }
    
    params = {
      command = "echo 'Deploying ${loop.service} to ${var.environment}'"
    }
  }

  # Environment-specific post-deployment
  step "staging_smoke_tests" {
    plugin = "shell"
    action = "exec"
    condition = "equal(var.environment, \"staging\")"
    depends_on = ["deploy_services"]
    
    params = {
      command = "echo 'Running staging smoke tests'"
    }
  }

  step "production_monitoring" {
    plugin = "shell"
    action = "exec"
    condition = "equal(var.environment, \"production\")"
    depends_on = ["deploy_services"]
    
    loop {
      over = "[\"metrics\", \"logging\", \"alerting\"]"
      variable = "monitor"
    }
    
    params = {
      command = "echo 'Configuring ${loop.monitor} for production'"
    }
  }

  # Notification (always runs)
  step "notify_team" {
    plugin = "shell"
    action = "exec"
    depends_on = ["staging_smoke_tests", "production_monitoring"]
    
    params = {
      command = "echo 'Deployment completed for ${var.environment} environment'"
    }
  }
}
```

### Execute CI/CD Pipeline

```bash
# Staging deployment
corynth apply cicd-pipeline.hcl -var environment="staging" -var git_branch="feature/new-api"

# Production deployment
corynth apply cicd-pipeline.hcl -var environment="production" -var git_branch="main" -var run_security_scan=true
```

## Best Practices for Advanced Flow Control

### 1. Boolean Expression Optimization
- Use parentheses for clarity in complex expressions
- Break extremely complex conditions into multiple steps
- Prefer readable conditions over ultra-compact ones

### 2. Loop Performance
- Minimize loop iterations where possible
- Use parallel execution for independent loop iterations
- Consider batching for large datasets

### 3. Conditional Logic Design
- Design for the common path first
- Use early termination to skip unnecessary work
- Provide meaningful step names that indicate conditions

### 4. Error Handling in Flow Control
```hcl
step "conditional_with_retry" {
  condition = "var.enabled"
  retry {
    max_attempts = 3
    delay = "5s"
  }
  continue_on {
    error = true
  }
}
```

### 5. Dynamic Parameter Values
```hcl
params = {
  timeout = "if(equal(var.environment, \"production\"), \"300\", \"60\")"
  replicas = "if(var.high_availability, \"5\", \"2\")"
}
```

## Advanced Patterns Summary

### Flow Control Combinations
- **Conditional Loops**: Execute loops only when conditions are met
- **Dynamic Branching**: Choose execution paths based on runtime data
- **Parallel Conditionals**: Run multiple conditional checks simultaneously
- **Nested Logic**: Combine multiple boolean functions for complex decisions

### Performance Considerations
- **Early Termination**: Skip unnecessary steps with conditions
- **Batch Processing**: Group similar operations for efficiency
- **Parallel Execution**: Maximize concurrency for independent operations
- **Resource Optimization**: Use conditional resource allocation

### Real-World Applications
- **CI/CD Pipelines**: Environment-specific deployment logic
- **Infrastructure Management**: Conditional resource provisioning
- **Data Processing**: Dynamic workflow adaptation
- **Monitoring & Alerting**: Conditional notification logic

## Next Steps

- Practice combining multiple flow control patterns
- Build your own complex workflows using these techniques
- Explore [Plugin Development](../plugins/development.md) for custom logic
- Study [Production Examples](../examples/) for real-world patterns

---

**Master advanced flow control to build sophisticated, efficient workflows!** 🚀