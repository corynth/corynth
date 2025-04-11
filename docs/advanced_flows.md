# 🌊 Advanced Flow Configuration

## 🌟 Introduction

Corynth flows can be configured in various advanced ways to handle complex automation scenarios. This guide covers advanced flow configuration techniques.

## 📋 Flow Structure Review

Let's first review the basic flow structure:

```yaml
flow:
  name: "flow_name"
  description: "Flow description"
  steps:
    - name: "step_name"
      plugin: "plugin_name"
      action: "action_name"
      params:
        param1: "value1"
        param2: "value2"
```

## 🔄 Flow Chaining

### Basic Flow Chaining

Flow chaining allows you to execute flows in sequence based on the outcome of previous flows:

```yaml
flow:
  name: "master_flow"
  description: "Master flow that chains other flows"
  chain:
    - flow: "flow1"
      on_success: "flow2"
      on_failure: "error_flow"
    
    - flow: "flow2"
      on_success: "flow3"
      on_failure: "rollback_flow"
```

### Conditional Flow Chaining

You can use conditions to determine which flow to execute next:

```yaml
flow:
  name: "conditional_flow"
  description: "Flow with conditional chaining"
  chain:
    - flow: "check_environment"
      on_success: "deploy_prod"
      on_failure: "deploy_dev"
      condition: "${ENV} == 'production'"
```

## 🔀 Parallel Execution

### Parallel Steps

You can execute steps in parallel by setting the `parallel` flag:

```yaml
flow:
  name: "parallel_flow"
  description: "Flow with parallel steps"
  steps:
    - name: "step1"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'Step 1'"
      parallel: true
    
    - name: "step2"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'Step 2'"
      parallel: true
    
    - name: "step3"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'Step 3'"
      depends_on:
        - step: "step1"
          status: "success"
        - step: "step2"
          status: "success"
```

### Parallel Flows

You can also execute flows in parallel:

```yaml
flow:
  name: "parallel_master_flow"
  description: "Master flow with parallel execution"
  chain:
    - flow: "flow1"
      parallel: true
    
    - flow: "flow2"
      parallel: true
    
    - flow: "flow3"
      depends_on:
        - flow: "flow1"
          status: "success"
        - flow: "flow2"
          status: "success"
```

## 🔍 Conditional Execution

### Step Conditions

You can conditionally execute steps based on expressions:

```yaml
flow:
  name: "conditional_flow"
  description: "Flow with conditional steps"
  steps:
    - name: "step1"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'Step 1'"
    
    - name: "step2"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'Step 2'"
      condition: "${step1.output} contains 'success'"
```

### Flow Conditions

You can also conditionally execute entire flows:

```yaml
flow:
  name: "conditional_master_flow"
  description: "Master flow with conditional execution"
  chain:
    - flow: "flow1"
    
    - flow: "flow2"
      condition: "${flow1.status} == 'success' && ${ENV} == 'production'"
```

## 📊 Variable Interpolation

### Basic Variable Interpolation

You can use variables in your flow definitions:

```yaml
flow:
  name: "variable_flow"
  description: "Flow with variable interpolation"
  variables:
    app_name: "my-app"
    version: "1.0.0"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy ${app_name} ${version}"
```

### Environment Variables

You can use environment variables:

```yaml
flow:
  name: "env_variable_flow"
  description: "Flow with environment variables"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy ${ENV_APP_NAME} ${ENV_VERSION}"
```

### Step Output Variables

You can use the output of previous steps:

```yaml
flow:
  name: "step_output_flow"
  description: "Flow with step output variables"
  steps:
    - name: "get_version"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo '1.0.0'"
    
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy ${get_version.output}"
      depends_on:
        - step: "get_version"
          status: "success"
```

## 🔒 Secret Management

### Using Secrets

You can use secrets in your flow definitions:

```yaml
flow:
  name: "secret_flow"
  description: "Flow with secrets"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy --token ${SECRET_API_TOKEN}"
```

### Secret Files

You can also use secret files:

```yaml
flow:
  name: "secret_file_flow"
  description: "Flow with secret files"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy --key-file ${SECRET_FILE_SSH_KEY}"
```

## 🔄 Loops and Iterations

### Foreach Loops

You can iterate over a list of items:

```yaml
flow:
  name: "foreach_flow"
  description: "Flow with foreach loops"
  variables:
    servers:
      - "server1"
      - "server2"
      - "server3"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      foreach: "${servers}"
      params:
        command: "deploy --server ${item}"
```

### Matrix Execution

You can execute steps with a matrix of parameters:

```yaml
flow:
  name: "matrix_flow"
  description: "Flow with matrix execution"
  matrix:
    environment:
      - "dev"
      - "staging"
      - "prod"
    region:
      - "us-east-1"
      - "us-west-2"
      - "eu-west-1"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy --env ${matrix.environment} --region ${matrix.region}"
```

## 🛑 Error Handling

### Retry Logic

You can configure steps to retry on failure:

```yaml
flow:
  name: "retry_flow"
  description: "Flow with retry logic"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy"
      retry:
        max_attempts: 3
        delay: "5s"
        backoff: 2.0
        condition: "${error} contains 'timeout'"
```

### Error Recovery

You can define recovery steps that execute when a step fails:

```yaml
flow:
  name: "recovery_flow"
  description: "Flow with error recovery"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy"
      on_failure:
        - name: "cleanup"
          plugin: "shell"
          action: "exec"
          params:
            command: "cleanup"
```

## 🔄 Flow Templates

### Defining Templates

You can define flow templates that can be reused:

```yaml
template:
  name: "deployment_template"
  description: "Template for deployments"
  parameters:
    - name: "app_name"
      required: true
    - name: "version"
      default: "latest"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy ${app_name} ${version}"
```

### Using Templates

You can use templates in your flows:

```yaml
flow:
  name: "template_flow"
  description: "Flow using a template"
  use_template: "deployment_template"
  parameters:
    app_name: "my-app"
    version: "1.0.0"
```

## 📝 Flow Composition

### Including Flows

You can include other flows in your flow:

```yaml
flow:
  name: "composite_flow"
  description: "Flow composed of other flows"
  include:
    - flow: "setup_flow"
    - flow: "deploy_flow"
    - flow: "test_flow"
```

### Overriding Steps

You can override steps from included flows:

```yaml
flow:
  name: "override_flow"
  description: "Flow with overridden steps"
  include:
    - flow: "deploy_flow"
      override:
        - name: "deploy"
          plugin: "shell"
          action: "exec"
          params:
            command: "custom_deploy"
```

## 🔍 Flow Validation

### Schema Validation

You can validate your flows against a schema:

```yaml
flow:
  name: "validated_flow"
  description: "Flow with schema validation"
  schema: "https://example.com/flow-schema.json"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy"
```

### Custom Validation

You can define custom validation rules:

```yaml
flow:
  name: "custom_validated_flow"
  description: "Flow with custom validation"
  validate:
    - rule: "${step1.output} contains 'success'"
      message: "Step 1 must output 'success'"
    - rule: "${ENV} in ['dev', 'staging', 'prod']"
      message: "ENV must be one of 'dev', 'staging', 'prod'"
  steps:
    - name: "step1"
      plugin: "shell"
      action: "exec"
      params:
        command: "echo 'success'"
```

## 🌐 Remote Flows

### Importing Remote Flows

You can import flows from remote sources:

```yaml
flow:
  name: "remote_flow"
  description: "Flow importing remote flows"
  import:
    - source: "https://example.com/flows/deploy_flow.yaml"
      as: "remote_deploy_flow"
    - source: "git://github.com/user/repo//flows/test_flow.yaml"
      as: "remote_test_flow"
  chain:
    - flow: "remote_deploy_flow"
      on_success: "remote_test_flow"
```

### Flow Registries

You can use flow registries to manage your flows:

```yaml
registry:
  name: "my_registry"
  url: "https://registry.example.com"
  auth:
    username: "${REGISTRY_USERNAME}"
    password: "${REGISTRY_PASSWORD}"

flow:
  name: "registry_flow"
  description: "Flow using registry flows"
  import:
    - registry: "my_registry"
      flow: "deploy_flow"
      version: "1.0.0"
      as: "registry_deploy_flow"
  chain:
    - flow: "registry_deploy_flow"
```

## 📊 Flow Metrics

### Collecting Metrics

You can collect metrics from your flows:

```yaml
flow:
  name: "metrics_flow"
  description: "Flow with metrics collection"
  metrics:
    namespace: "my_app"
    subsystem: "deployment"
    collect:
      - name: "duration"
        type: "gauge"
        help: "Flow duration in seconds"
        value: "${flow.duration}"
      - name: "success"
        type: "counter"
        help: "Flow success count"
        value: 1
        condition: "${flow.status} == 'success'"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy"
```

### Exporting Metrics

You can export metrics to various systems:

```yaml
flow:
  name: "metrics_export_flow"
  description: "Flow with metrics export"
  metrics:
    namespace: "my_app"
    subsystem: "deployment"
    collect:
      - name: "duration"
        type: "gauge"
        help: "Flow duration in seconds"
        value: "${flow.duration}"
    export:
      - type: "prometheus"
        endpoint: "http://prometheus:9090/metrics"
      - type: "statsd"
        endpoint: "statsd:8125"
        prefix: "corynth"
  steps:
    - name: "deploy"
      plugin: "shell"
      action: "exec"
      params:
        command: "deploy"
```

## 🔍 Best Practices

1. **Keep Flows Simple**: Break complex flows into smaller, reusable flows
2. **Use Descriptive Names**: Use clear, descriptive names for flows and steps
3. **Document Your Flows**: Add descriptions to flows and steps
4. **Use Variables**: Use variables to make flows more flexible
5. **Handle Errors**: Always include error handling in your flows
6. **Test Your Flows**: Test flows in development before using them in production
7. **Use Version Control**: Store your flows in version control
8. **Use Templates**: Use templates for common patterns
9. **Validate Flows**: Validate flows before execution
10. **Monitor Flows**: Collect and monitor flow metrics