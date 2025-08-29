workflow "boolean-logic-demo" {
  description = "Demonstrates boolean logic and conditional execution"
  version = "1.0.0"

  variable "environment" {
    type = string
    default = "development"
    description = "Target environment (development, staging, production)"
  }

  variable "deploy_enabled" {
    type = bool
    default = true
    description = "Enable deployment steps"
  }

  variable "run_tests" {
    type = bool
    default = true
    description = "Run test suite"
  }

  step "environment_check" {
    plugin = "shell"
    action = "exec"
    condition = "{{eq .Variables.environment \"development\"}}"
    
    params = {
      command = "echo 'Running in development environment'"
    }
  }

  step "test_suite" {
    plugin = "shell"
    action = "exec"
    condition = "{{.Variables.run_tests}}"
    
    params = {
      command = "echo 'Running test suite...'"
    }
  }

  step "staging_deployment" {
    plugin = "shell"
    action = "exec"
    condition = "{{and .Variables.deploy_enabled (eq .Variables.environment \"staging\")}}"
    depends_on = ["test_suite"]
    
    params = {
      command = "echo 'Deploying to staging environment'"
    }
  }

  step "production_deployment" {
    plugin = "shell"
    action = "exec"
    condition = "{{and .Variables.deploy_enabled (eq .Variables.environment \"production\")}}"
    depends_on = ["test_suite"]
    
    params = {
      command = "echo 'Deploying to production environment'"
    }
  }

  step "skip_deployment" {
    plugin = "shell"
    action = "exec"
    condition = "{{not .Variables.deploy_enabled}}"
    
    params = {
      command = "echo 'Deployment is disabled, skipping...'"
    }
  }

  step "complex_condition" {
    plugin = "shell"
    action = "exec"
    condition = "{{or (and .Variables.deploy_enabled (eq .Variables.environment \"production\")) (eq .Variables.environment \"staging\")}}"
    
    params = {
      command = "echo 'Complex condition met: production deploy OR staging environment'"
    }
  }

  step "conditional_timeout" {
    plugin = "shell"
    action = "exec"
    condition = "{{.Variables.deploy_enabled}}"
    
    params = {
      command = "echo 'Operation with conditional timeout'"
      timeout = "{{if (eq .Variables.environment \"production\")}}300{{else}}60{{end}}"
    }
  }
}