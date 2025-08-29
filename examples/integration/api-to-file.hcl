workflow "api-to-file" {
  description = "Fetch data from API and save to file"
  version = "1.0.0"

  variable "api_url" {
    type = string
    default = "https://api.github.com/users/octocat"
    description = "API endpoint to fetch data from"
  }

  step "fetch_data" {
    plugin = "http"
    action = "get"
    
    params = {
      url = "{{.Variables.api_url}}"
      timeout = 30
    }
  }

  step "save_data" {
    plugin = "file"
    action = "write"
    depends_on = ["fetch_data"]
    
    params = {
      path = "/tmp/api-data.json"
      content = "{{.Steps.fetch_data.output.body}}"
    }
  }

  step "verify_save" {
    plugin = "shell"
    action = "exec"
    depends_on = ["save_data"]
    
    params = {
      command = "echo 'Data saved. File size:' && wc -c /tmp/api-data.json"
    }
  }

  step "cleanup" {
    plugin = "shell"
    action = "exec"
    depends_on = ["verify_save"]
    
    params = {
      command = "rm /tmp/api-data.json && echo 'Cleanup completed'"
    }
  }
}