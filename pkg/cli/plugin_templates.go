package cli

import (
	"strings"
	"text/template"
)

// Template helper functions
func init() {
	template.Must(template.New("").Funcs(template.FuncMap{
		"ToUpper":     strings.ToUpper,
		"ToCamelCase": toCamelCase,
	}).Parse(""))
}

// Plugin template files for scaffolding

const pluginGoTemplate = `package main

import (
	"context"
	"fmt"{{range .Dependencies}}
	"{{.}}"{{end}}
	
	"github.com/corynth/corynth/pkg/plugin"
)

// {{.ClassName}} implements the Corynth plugin interface for {{.Description}}
type {{.ClassName}} struct {
	// Plugin configuration and state
	{{if .HasHTTP}}client *http.Client{{end}}
	{{if .HasDatabase}}db *sql.DB{{end}}
}

// Metadata returns plugin metadata
func (p *{{.ClassName}}) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "{{.Name}}",
		Version:     "1.0.0",
		Description: "{{.Description}}",
		Author:      "{{.Author}}",
		Tags:        []string{"{{.Type}}", "{{.Name}}", "automation"},
		License:     "{{.License}}",
	}
}

// Actions returns available plugin actions
func (p *{{.ClassName}}) Actions() []plugin.Action {
	return []plugin.Action{
		{{range .Actions}}{
			Name:        "{{.Name}}",
			Description: "{{.Description}}",
			Inputs:      map[string]plugin.InputSpec{
				{{range .Params}}"{{.Name}}": {
					Type:        "{{.Type}}",
					Description: "{{.Description}}",
					Required:    {{.Required}},{{if .Default}}
					Default:     {{.Default}},{{end}}
				},
				{{end}}
			},
			Outputs:     map[string]plugin.OutputSpec{
				{{range .Returns}}"{{.Name}}": {
					Type:        "{{.Type}}",
					Description: "{{.Description}}",
				},
				{{end}}
			},
		},
		{{end}}
	}
}

// Execute performs the plugin action
func (p *{{.ClassName}}) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	{{range .Actions}}case "{{.Name}}":
		return p.{{.Method}}(ctx, params)
	{{end}}default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Validate checks parameter validity
func (p *{{.ClassName}}) Validate(params map[string]interface{}) error {
	// TODO: Implement parameter validation
	return nil
}

{{range .Actions}}
// {{.Method}} implements the {{.Name}} action
func (p *{{$.ClassName}}) {{.Method}}(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	{{range .Params}}{{if .Required}}// Extract required parameter: {{.Name}}
	{{.Name}}, ok := params["{{.Name}}"]
	if !ok {
		return nil, fmt.Errorf("{{.Name}} parameter is required")
	}
	{{end}}{{end}}

	{{range .Params}}{{if not .Required}}// Extract optional parameter: {{.Name}}
	{{.Name}} := getParam(params, "{{.Name}}", {{if .Default}}{{.Default}}{{else}}nil{{end}})
	{{end}}{{end}}

	// TODO: Implement {{.Name}} logic here
	// This is a placeholder implementation
	
	{{if eq .Name "get"}}
	// HTTP GET implementation example
	{{if $.HasHTTP}}if p.client == nil {
		p.client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url.(string), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        string(body),
		"headers":     resp.Header,
	}, nil{{end}}
	{{else if eq .Name "query"}}
	// Database query implementation example
	{{if $.HasDatabase}}if p.db == nil {
		// TODO: Initialize database connection
		return nil, fmt.Errorf("database connection not initialized")
	}
	
	rows, err := p.db.QueryContext(ctx, query.(string))
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	
	// TODO: Process query results
	var results []map[string]interface{}
	
	return map[string]interface{}{
		"rows":  results,
		"count": len(results),
	}, nil{{end}}
	{{else if eq .Name "exec"}}
	// Command execution implementation example
	{{if $.HasExternal}}cmd := exec.CommandContext(ctx, command.(string))
	if workingDir != nil {
		cmd.Dir = workingDir.(string)
	}
	
	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}
	
	return map[string]interface{}{
		"output":    string(output),
		"exit_code": exitCode,
		"success":   exitCode == 0,
	}, nil{{end}}
	{{else}}
	// Custom implementation
	return map[string]interface{}{
		{{range .Returns}}"{{.Name}}": "placeholder_value", // TODO: Implement actual logic
		{{end}}
	}, nil
	{{end}}
}
{{end}}

// Helper functions

func getParam(params map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if value, ok := params[key]; ok {
		return value
	}
	return defaultValue
}

func getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, ok := params[key].(string); ok {
		return value
	}
	return defaultValue
}

func getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if value, ok := params[key].(int); ok {
		return value
	}
	if value, ok := params[key].(float64); ok {
		return int(value)
	}
	return defaultValue
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := params[key].(bool); ok {
		return value
	}
	return defaultValue
}

// Required: Export the plugin
var ExportedPlugin {{.ClassName}}
`

const goModTemplate = `module {{.ModulePath}}

go 1.21

require (
	github.com/corynth/corynth v1.0.0
)
`

const readmeTemplate = `# {{.Name}} Plugin

{{.Description}}

## Installation

This plugin is automatically downloaded and compiled when first used in a Corynth workflow.

## Actions

{{range .Actions}}### {{.Name}}

{{.Description}}

**Parameters:**
{{range .Params}}- ` + "`{{.Name}}`" + ` ({{.Type}}{{if .Required}}, required{{else}}, optional{{if .Default}}, default: {{.Default}}{{end}}{{end}}): {{.Description}}
{{end}}

**Returns:**
{{range .Returns}}- ` + "`{{.Name}}`" + ` ({{.Type}}): {{.Description}}
{{end}}

**Example:**
` + "```hcl" + `
step "{{.Name}}_example" {
  plugin = "{{$.Name}}"
  action = "{{.Name}}"
  
  params = {
    {{range .Params}}{{.Name}} = {{if eq .Type "string"}}"example_value"{{else if eq .Type "number"}}42{{else if eq .Type "boolean"}}true{{else if eq .Type "array"}}["item1", "item2"]{{else}}{}{{end}}
    {{end}}
  }
}
` + "```" + `

{{end}}

## Configuration

### Environment Variables
- ` + "`{{.Name | ToUpper}}_API_KEY`" + `: API key for authentication (if applicable)
- ` + "`{{.Name | ToUpper}}_TIMEOUT`" + `: Default timeout in seconds

### Authentication
{{if .HasHTTP}}Configure API authentication using environment variables or parameters.{{else}}No authentication required for this plugin.{{end}}

## Error Handling

Common errors and solutions:

- **Error: "required parameter missing"**: Ensure all required parameters are provided
- **Error: "connection failed"**: Check network connectivity and credentials
- **Error: "timeout exceeded"**: Increase timeout value or check performance

## Examples

See the [samples/](samples/) directory for complete workflow examples:

- [basic-usage.hcl](samples/basic-usage.hcl) - Basic plugin usage
- [advanced-features.hcl](samples/advanced-features.hcl) - Advanced features and error handling

## Development

### Testing
` + "```bash" + `
# Run unit tests
go test -v

# Test with Corynth
corynth plugin test .
corynth apply samples/basic-usage.hcl
` + "```" + `

### Building
` + "```bash" + `
# Build plugin
make build

# Build and test
make test
` + "```" + `

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

{{.License}} License

---
`

const testTemplate = `package main

import (
	"context"
	"testing"
	"time"
)

func TestPluginMetadata(t *testing.T) {
	plugin := &{{.ClassName}}{}
	meta := plugin.Metadata()

	if meta.Name != "{{.Name}}" {
		t.Errorf("Expected plugin name '{{.Name}}', got '%s'", meta.Name)
	}

	if meta.Version == "" {
		t.Error("Plugin version cannot be empty")
	}

	if meta.Description == "" {
		t.Error("Plugin description cannot be empty")
	}

	if len(meta.Tags) == 0 {
		t.Error("Plugin should have at least one tag")
	}
}

func TestPluginActions(t *testing.T) {
	plugin := &{{.ClassName}}{}
	actions := plugin.Actions()

	if len(actions) == 0 {
		t.Error("Plugin should have at least one action")
	}

	expectedActions := []string{
		{{range .Actions}}"{{.Name}}",
		{{end}}
	}

	for _, expectedAction := range expectedActions {
		found := false
		for _, action := range actions {
			if action.Name == expectedAction {
				found = true
				if action.Description == "" {
					t.Errorf("Action '%s' should have a description", expectedAction)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected action '%s' not found", expectedAction)
		}
	}
}

{{range .Actions}}
func TestPlugin{{.Name | ToCamelCase}}(t *testing.T) {
	plugin := &{{$.ClassName}}{}
	ctx := context.Background()

	// Test with valid parameters
	params := map[string]interface{}{
		{{range .Params}}{{if .Required}}"{{.Name}}": {{if eq .Type "string"}}"test_value"{{else if eq .Type "number"}}42{{else if eq .Type "boolean"}}true{{else if eq .Type "array"}}[]string{"test"}{{else}}map[string]interface{}{"key": "value"}{{end}},
		{{end}}{{end}}
	}

	result, err := plugin.Execute(ctx, "{{.Name}}", params)
	if err != nil {
		t.Fatalf("Plugin execution failed: %v", err)
	}

	if result == nil {
		t.Error("Plugin should return a result")
	}

	// Verify expected return values
	{{range .Returns}}if _, ok := result["{{.Name}}"]; !ok {
		t.Error("Result should contain '{{.Name}}'")
	}
	{{end}}
}

func TestPlugin{{.Name | ToCamelCase}}InvalidParams(t *testing.T) {
	plugin := &{{$.ClassName}}{}
	ctx := context.Background()

	// Test with missing required parameters
	params := map[string]interface{}{
		"invalid_param": "test",
	}

	_, err := plugin.Execute(ctx, "{{.Name}}", params)
	if err == nil {
		t.Error("Plugin should fail with invalid parameters")
	}
}
{{end}}

func TestPluginValidation(t *testing.T) {
	plugin := &{{.ClassName}}{}

	// Test valid parameters
	validParams := map[string]interface{}{
		{{range .Actions}}{{range .Params}}{{if .Required}}"{{.Name}}": {{if eq .Type "string"}}"test"{{else if eq .Type "number"}}42{{else if eq .Type "boolean"}}true{{else}}{}{{end}},
		{{end}}{{end}}{{end}}
	}

	if err := plugin.Validate(validParams); err != nil {
		t.Errorf("Valid parameters should pass validation: %v", err)
	}
}

func TestPluginTimeout(t *testing.T) {
	plugin := &{{.ClassName}}{}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	params := map[string]interface{}{
		{{range .Actions}}{{range .Params}}{{if .Required}}"{{.Name}}": {{if eq .Type "string"}}"test"{{else if eq .Type "number"}}42{{else if eq .Type "boolean"}}true{{else}}{}{{end}},
		{{end}}{{end}}{{end}}
	}

	// This test may pass if the operation is fast
	// Adjust based on your plugin's behavior
	_, err := plugin.Execute(ctx, "{{(index .Actions 0).Name}}", params)
	
	// Check if context was cancelled (timeout occurred)
	if ctx.Err() == context.DeadlineExceeded {
		t.Log("Plugin correctly respects context timeout")
	} else if err != nil {
		t.Logf("Plugin execution failed: %v", err)
	}
}

func TestPluginUnknownAction(t *testing.T) {
	plugin := &{{.ClassName}}{}
	ctx := context.Background()

	params := map[string]interface{}{}

	_, err := plugin.Execute(ctx, "unknown_action", params)
	if err == nil {
		t.Error("Plugin should fail with unknown action")
	}
}
`

const basicWorkflowTemplate = `workflow "{{.Name}}-basic-usage" {
  description = "Basic usage example for {{.Name}} plugin"
  version     = "1.0.0"

  variable "{{.Name}}_config" {
    type        = string
    default     = "default_value"
    description = "Configuration for {{.Name}} plugin"
  }

  {{range $index, $action := .Actions}}{{if eq $index 0}}step "{{$action.Name}}_example" {
    plugin = "{{$.Name}}"
    action = "{{$action.Name}}"
    
    params = {
      {{range $action.Params}}{{.Name}} = {{if eq .Type "string"}}{{if .Default}}"{{.Default}}"{{else}}"example_value"{{end}}{{else if eq .Type "number"}}{{if .Default}}{{.Default}}{{else}}42{{end}}{{else if eq .Type "boolean"}}{{if .Default}}{{.Default}}{{else}}true{{end}}{{else if eq .Type "array"}}["item1", "item2"]{{else}}{{if .Default}}{{.Default}}{{else}}{}{{end}}{{end}}
      {{end}}
    }
  }

  step "use_result" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["{{$action.Name}}_example"]
    
    params = {
      command = "echo 'Operation completed successfully'"
    }
  }{{end}}{{end}}
}
`

const advancedWorkflowTemplate = `workflow "{{.Name}}-advanced-features" {
  description = "Advanced features and error handling for {{.Name}} plugin"
  version     = "1.0.0"

  variable "environment" {
    type        = string
    default     = "development"
    description = "Environment name"
  }

  variable "{{.Name}}_endpoint" {
    type        = string
    default     = "{{if .HasHTTP}}https://api.example.com{{else}}default_value{{end}}"
    description = "{{.Name}} endpoint or configuration"
  }

  {{range .Actions}}step "{{.Name}}_with_error_handling" {
    plugin = "{{$.Name}}"
    action = "{{.Name}}"
    
    params = {
      {{range .Params}}{{.Name}} = {{if eq .Type "string"}}{{if eq .Name "url"}}var.{{$.Name}}_endpoint{{else if .Default}}"{{.Default}}"{{else}}"advanced_example"{{end}}{{else if eq .Type "number"}}{{if .Default}}{{.Default}}{{else}}60{{end}}{{else if eq .Type "boolean"}}{{if .Default}}{{.Default}}{{else}}true{{end}}{{else if eq .Type "array"}}["option1", "option2", "option3"]{{else if eq .Type "object"}}{
        environment = var.environment
        retry_count = 3
      }{{else}}{{if .Default}}{{.Default}}{{else}}"default"{{end}}{{end}}
      {{end}}
    }
  }

  step "validate_result_{{.Name}}" {
    plugin = "shell"
    action = "exec"
    
    depends_on = ["{{.Name}}_with_error_handling"]
    
    params = {
      command = "echo 'Validating {{.Name}} result: Success'"
    }
  }
  {{end}}

  step "cleanup" {
    plugin = "shell"
    action = "exec"
    
    depends_on = [{{range $index, $action := .Actions}}{{if gt $index 0}}, {{end}}"validate_result_{{$action.Name}}"{{end}}]
    
    params = {
      command = "echo 'Workflow completed successfully in ${var.environment} environment'"
    }
  }
}
`

const githubWorkflowTemplate = `name: Test {{.Name}} Plugin

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: RUNNER_OS-go-HASH_FILES
        restore-keys: |
          RUNNER_OS-go-
    
    - name: Download dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out
    
    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
        flags: unittests
        name: codecov-umbrella
    
    - name: Build plugin
      run: go build -buildmode=plugin -o {{.Name}}.so plugin.go
    
    - name: Set up Corynth (if available)
      run: |
        # TODO: Add Corynth installation
        echo "Corynth installation step - to be implemented"
    
    - name: Test with Corynth workflows
      run: |
        # TODO: Test with actual Corynth workflows
        echo "Testing workflows:"
        ls samples/*.hcl
        # corynth validate samples/basic-usage.hcl
        # corynth plan samples/basic-usage.hcl

  lint:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --timeout=5m
`

const makefileTemplate = `# {{.Name}} Plugin Makefile

PLUGIN_NAME={{.Name}}
GO_VERSION=1.21

.PHONY: all build test clean lint fmt vet mod-tidy help

## Build the plugin
build:
	@echo "Building {{.Name}} plugin..."
	go build -buildmode=plugin -o $(PLUGIN_NAME).so plugin.go
	@echo "✅ Plugin built: $(PLUGIN_NAME).so"

## Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out
	@echo "✅ Tests completed"

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## Run linter
lint:
	@echo "Running linter..."
	golangci-lint run --timeout=5m
	@echo "✅ Linting completed"

## Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	@echo "✅ Code formatted"

## Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✅ Vet completed"

## Update dependencies
mod-tidy:
	@echo "Updating dependencies..."
	go mod tidy
	@echo "✅ Dependencies updated"

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(PLUGIN_NAME).so
	rm -f coverage.out coverage.html
	@echo "✅ Clean completed"

## Run all checks
check: fmt vet lint test
	@echo "✅ All checks passed"

## Development cycle
dev: clean mod-tidy check build
	@echo "✅ Development cycle completed"

## Test with Corynth (requires Corynth installation)
test-corynth:
	@echo "Testing with Corynth..."
	corynth validate samples/basic-usage.hcl
	corynth plan samples/basic-usage.hcl
	@echo "✅ Corynth tests completed"

## Help
help:
	@echo "{{.Name}} Plugin Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed 's/## /  /'
	@echo ""
	@echo "Examples:"
	@echo "  make dev          # Full development cycle"
	@echo "  make build        # Build plugin only"
	@echo "  make test         # Run tests only"
	@echo "  make check        # Run all checks"
`