workflow "hello-world-sample" {
  description = "A simple greeting workflow demonstrating basic Corynth functionality"
  version     = "1.0.0"

  variable "name" {
    type        = string
    default     = "World"
    description = "Name to greet"
  }

  step "say_hello" {
    plugin = "shell"
    action = "exec"
    
    params = {
      command = "echo 'Hello, {{.Variables.name}}! Welcome to Corynth.'"
    }
  }

  step "show_timestamp" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["say_hello"]
    
    params = {
      command = "echo \"Workflow executed at: $(date)\""
    }
  }

  step "environment_info" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["show_timestamp"]
    
    params = {
      command = "echo \"Running on: $(hostname) as $(whoami)\""
    }
  }
}