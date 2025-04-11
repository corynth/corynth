package ansible

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// AnsiblePlugin is a plugin for Ansible operations
type AnsiblePlugin struct{}

// Result represents the result of a step execution
type Result struct {
	Status    string
	Output    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// NewAnsiblePlugin creates a new AnsiblePlugin
func NewAnsiblePlugin() *AnsiblePlugin {
	return &AnsiblePlugin{}
}

// Name returns the name of the plugin
func (p *AnsiblePlugin) Name() string {
	return "ansible"
}

// Execute executes an Ansible action
func (p *AnsiblePlugin) Execute(action string, params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	switch action {
	case "playbook":
		return p.playbook(params)
	case "ad_hoc":
		return p.adHoc(params)
	case "inventory":
		return p.inventory(params)
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

// playbook runs an Ansible playbook
func (p *AnsiblePlugin) playbook(params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Get parameters
	playbook, ok := params["playbook"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "playbook parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("playbook parameter is required")
	}

	// Build command
	args := []string{"playbook", playbook}

	// Add inventory if provided
	if inventory, ok := params["inventory"].(string); ok {
		args = append(args, "-i", inventory)
	}

	// Add extra vars if provided
	if extraVars, ok := params["extra_vars"].(map[string]interface{}); ok {
		for key, value := range extraVars {
			args = append(args, "-e", fmt.Sprintf("%s=%v", key, value))
		}
	}

	// Execute command
	cmd := exec.Command("ansible", args...)
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
		}, fmt.Errorf("error executing ansible playbook: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// adHoc runs an Ansible ad-hoc command
func (p *AnsiblePlugin) adHoc(params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Get parameters
	hosts, ok := params["hosts"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "hosts parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("hosts parameter is required")
	}

	module, ok := params["module"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "module parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("module parameter is required")
	}

	args, _ := params["args"].(string)

	// Build command
	cmdArgs := []string{hosts, "-m", module}

	if args != "" {
		cmdArgs = append(cmdArgs, "-a", args)
	}

	// Add inventory if provided
	if inventory, ok := params["inventory"].(string); ok {
		cmdArgs = append(cmdArgs, "-i", inventory)
	}

	// Execute command
	cmd := exec.Command("ansible", cmdArgs...)
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
		}, fmt.Errorf("error executing ansible ad-hoc command: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// inventory generates an Ansible dynamic inventory
func (p *AnsiblePlugin) inventory(params map[string]interface{}) (Result, error) {
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

	// Check if script exists
	if _, err := os.Stat(script); os.IsNotExist(err) {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("inventory script %s does not exist", script),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("inventory script %s does not exist", script)
	}

	// Execute script
	cmd := exec.Command(script)
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
		}, fmt.Errorf("error executing inventory script: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}