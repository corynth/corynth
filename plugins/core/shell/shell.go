package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ShellPlugin is a plugin for shell operations
type ShellPlugin struct{}

// Result represents the result of a step execution
type Result struct {
	Status    string
	Output    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// NewShellPlugin creates a new ShellPlugin
func NewShellPlugin() *ShellPlugin {
	return &ShellPlugin{}
}

// Name returns the name of the plugin
func (p *ShellPlugin) Name() string {
	return "shell"
}

// Execute executes a shell action
func (p *ShellPlugin) Execute(action string, params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	switch action {
	case "exec":
		return p.exec(params)
	case "script":
		return p.script(params)
	case "pipe":
		return p.pipe(params)
	default:
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("unknown action: %s", action),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("unknown action: %s", action)
	}
}

// exec executes a shell command
func (p *ShellPlugin) exec(params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Get parameters
	command, ok := params["command"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "command parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("command parameter is required")
	}

	workingDir, _ := params["working_dir"].(string)
	if workingDir == "" {
		workingDir = "."
	}

	// Get environment variables
	env := os.Environ()
	if envVars, ok := params["env"].(map[string]interface{}); ok {
		for key, value := range envVars {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Change to working directory
	currentDir, err := os.Getwd()
	if err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error getting current directory: %s", err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error getting current directory: %w", err)
	}

	if err := os.Chdir(workingDir); err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error changing to directory %s: %s", workingDir, err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error changing to directory %s: %w", workingDir, err)
	}

	defer os.Chdir(currentDir)

	// Execute command
	var cmd *exec.Cmd
	if strings.Contains(command, " ") {
		parts := strings.Split(command, " ")
		cmd = exec.Command(parts[0], parts[1:]...)
	} else {
		cmd = exec.Command(command)
	}

	cmd.Env = env
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing command: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// script executes a script file
func (p *ShellPlugin) script(params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Get parameters
	script, ok := params["script"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "script parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("script parameter is required")
	}

	workingDir, _ := params["working_dir"].(string)
	if workingDir == "" {
		workingDir = "."
	}

	// Get environment variables
	env := os.Environ()
	if envVars, ok := params["env"].(map[string]interface{}); ok {
		for key, value := range envVars {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Change to working directory
	currentDir, err := os.Getwd()
	if err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error getting current directory: %s", err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error getting current directory: %w", err)
	}

	if err := os.Chdir(workingDir); err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error changing to directory %s: %s", workingDir, err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error changing to directory %s: %w", workingDir, err)
	}

	defer os.Chdir(currentDir)

	// Check if script exists
	if _, err := os.Stat(script); os.IsNotExist(err) {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("script file %s does not exist", script),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("script file %s does not exist", script)
	}

	// Execute script
	cmd := exec.Command("sh", script)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing script: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// pipe pipes the output of one command to another
func (p *ShellPlugin) pipe(params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Get parameters
	command1, ok := params["command1"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "command1 parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("command1 parameter is required")
	}

	command2, ok := params["command2"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "command2 parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("command2 parameter is required")
	}

	workingDir, _ := params["working_dir"].(string)
	if workingDir == "" {
		workingDir = "."
	}

	// Get environment variables
	env := os.Environ()
	if envVars, ok := params["env"].(map[string]interface{}); ok {
		for key, value := range envVars {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Change to working directory
	currentDir, err := os.Getwd()
	if err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error getting current directory: %s", err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error getting current directory: %w", err)
	}

	if err := os.Chdir(workingDir); err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error changing to directory %s: %s", workingDir, err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error changing to directory %s: %w", workingDir, err)
	}

	defer os.Chdir(currentDir)

	// Execute commands with pipe
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s | %s", command1, command2))
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	if err != nil {
		return Result{
			Status:    "error",
			Output:    string(output),
			Error:     err.Error(),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error executing piped commands: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}