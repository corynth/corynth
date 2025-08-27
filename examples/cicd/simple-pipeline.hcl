workflow "simple-pipeline" {
  description = "Simple CI/CD pipeline example"
  version = "1.0.0"

  variable "repo_url" {
    type = string
    default = "https://github.com/corynth/corynthplugins.git"
    description = "Repository to build and test"
  }

  variable "notification_url" {
    type = string
    description = "Webhook URL for notifications"
    required = false
  }

  step "clone_source" {
    plugin = "git"
    action = "clone"
    
    params = {
      url = var.repo_url
      path = "/tmp/ci-build"
      branch = "main"
    }
  }

  step "validate_structure" {
    plugin = "shell"
    action = "exec"
    depends_on = ["clone_source"]
    
    params = {
      command = "cd /tmp/ci-build && ls -la && echo 'Repository structure validated'"
    }
  }

  step "run_tests" {
    plugin = "shell"
    action = "exec"
    depends_on = ["validate_structure"]
    timeout = "5m"
    
    params = {
      command = "cd /tmp/ci-build && echo 'Running tests...' && sleep 1 && echo 'Tests passed'"
    }
    
    retry {
      max_attempts = 2
      delay = "30s"
    }
  }

  step "build_project" {
    plugin = "shell"
    action = "exec"
    depends_on = ["run_tests"]
    condition = "${run_tests.exit_code == 0}"
    
    params = {
      command = "cd /tmp/ci-build && echo 'Building project...' && sleep 2 && echo 'Build completed'"
    }
  }

  step "create_release_notes" {
    plugin = "file"
    action = "write"
    depends_on = ["build_project"]
    
    params = {
      path = "/tmp/ci-build/RELEASE.md"
      content = "# Release Notes\n\nBuild completed successfully!\n\n- Commit: ${clone_source.commit}\n- Tests: Passed\n- Build: Success\n- Timestamp: $(date)"
    }
  }

  step "notify_success" {
    plugin = "http"
    action = "post"
    depends_on = ["create_release_notes"]
    condition = "${var.notification_url != \"\"}"
    
    params = {
      url = var.notification_url
      body = "{\"status\": \"success\", \"message\": \"Pipeline completed successfully\", \"commit\": \"${clone_source.commit}\"}"
      headers = {
        "Content-Type" = "application/json"
      }
    }
    
    continue_on {
      error = true  # Don't fail pipeline if notification fails
    }
  }

  step "cleanup" {
    plugin = "shell"
    action = "exec"
    depends_on = ["notify_success"]
    
    params = {
      command = "rm -rf /tmp/ci-build && echo 'Pipeline cleanup completed'"
    }
  }
}