package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ShellPlugin implements basic shell command execution
type ShellPlugin struct {
	metadata Metadata
}

// NewShellPlugin creates a new shell plugin instance
func NewShellPlugin() *ShellPlugin {
	return &ShellPlugin{
		metadata: Metadata{
			Name:        "shell",
			Version:     "1.0.0",
			Description: "Execute shell commands and scripts",
			Author:      "Corynth Team",
			Tags:        []string{"system", "execution"},
		},
	}
}

// Metadata returns plugin metadata
func (p *ShellPlugin) Metadata() Metadata {
	return p.metadata
}

// Actions returns available actions
func (p *ShellPlugin) Actions() []Action {
	return []Action{
		{
			Name:        "exec",
			Description: "Execute a shell command",
			Inputs: map[string]InputSpec{
				"command": {
					Type:        "string",
					Description: "Command to execute",
					Required:    true,
				},
				"timeout": {
					Type:        "number",
					Description: "Timeout in seconds (default: 300)",
					Required:    false,
					Default:     300,
				},
				"working_directory": {
					Type:        "string",
					Description: "Working directory for command execution",
					Required:    false,
				},
				"shell": {
					Type:        "string",
					Description: "Shell to use (default: /bin/sh)",
					Required:    false,
					Default:     "/bin/sh",
				},
			},
			Outputs: map[string]OutputSpec{
				"stdout": {
					Type:        "string",
					Description: "Command standard output",
				},
				"stderr": {
					Type:        "string",
					Description: "Command standard error",
				},
				"exit_code": {
					Type:        "number",
					Description: "Command exit code",
				},
				"success": {
					Type:        "boolean",
					Description: "Whether command executed successfully",
				},
			},
			Examples: []Example{
				{
					Name:        "Simple command",
					Description: "Execute a simple echo command",
					Input: map[string]interface{}{
						"command": "echo 'Hello, World!'",
					},
				},
				{
					Name:        "Command with timeout",
					Description: "Execute command with custom timeout",
					Input: map[string]interface{}{
						"command": "sleep 10",
						"timeout": 5,
					},
				},
			},
		},
	}
}

// Execute runs the plugin action
func (p *ShellPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "exec":
		return p.executeCommand(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// Validate checks if the plugin configuration is valid
func (p *ShellPlugin) Validate(params map[string]interface{}) error {
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return fmt.Errorf("command parameter is required and must be a non-empty string")
	}

	if timeout, exists := params["timeout"]; exists {
		if _, ok := timeout.(float64); !ok {
			if _, ok := timeout.(int); !ok {
				return fmt.Errorf("timeout parameter must be a number")
			}
		}
	}

	if workingDir, exists := params["working_directory"]; exists {
		if _, ok := workingDir.(string); !ok {
			return fmt.Errorf("working_directory parameter must be a string")
		}
	}

	if shell, exists := params["shell"]; exists {
		if _, ok := shell.(string); !ok {
			return fmt.Errorf("shell parameter must be a string")
		}
	}

	return nil
}

// executeCommand executes a shell command
func (p *ShellPlugin) executeCommand(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	command := params["command"].(string)
	
	// Get shell (default to /bin/sh)
	shell := "/bin/sh"
	if s, exists := params["shell"].(string); exists && s != "" {
		shell = s
	}

	// Create command
	cmd := exec.CommandContext(ctx, shell, "-c", command)

	// Set working directory if specified
	if workingDir, exists := params["working_directory"].(string); exists && workingDir != "" {
		if _, err := os.Stat(workingDir); err != nil {
			return nil, fmt.Errorf("working directory does not exist: %s", workingDir)
		}
		cmd.Dir = workingDir
	}

	// Execute command and capture output
	stdout, err := cmd.Output()
	
	result := map[string]interface{}{
		"stdout":    strings.TrimSpace(string(stdout)),
		"stderr":    "",
		"exit_code": 0,
		"success":   true,
	}

	if err != nil {
		result["success"] = false
		
		if exitError, ok := err.(*exec.ExitError); ok {
			result["exit_code"] = exitError.ExitCode()
			result["stderr"] = strings.TrimSpace(string(exitError.Stderr))
		} else {
			result["exit_code"] = -1
			result["stderr"] = err.Error()
		}
		
		// For non-zero exit codes, we still return the result (not an error)
		// The caller can check the success field or exit_code
		return result, nil
	}

	return result, nil
}