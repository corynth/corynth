workflow "test-remote-plugins" {
  description = "Test remote plugin loading"
  version     = "1.0.0"

  step "test_shell" {
    plugin = "shell"
    action = "exec"
    params = {
      command = "echo 'Shell plugin (built-in) works!'"
    }
  }

  step "test_http" {
    plugin = "http"
    action = "get"
    depends_on = ["test_shell"]
    params = {
      url = "https://api.github.com/users/github"
    }
  }
}