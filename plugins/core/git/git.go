package git

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// GitPlugin is a plugin for Git operations
type GitPlugin struct{}

// Result represents the result of a step execution
type Result struct {
	Status    string
	Output    string
	Error     string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// NewGitPlugin creates a new GitPlugin
func NewGitPlugin() *GitPlugin {
	return &GitPlugin{}
}

// Name returns the name of the plugin
func (p *GitPlugin) Name() string {
	return "git"
}

// Execute executes a Git action
func (p *GitPlugin) Execute(action string, params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	switch action {
	case "clone":
		return p.clone(params)
	case "checkout":
		return p.checkout(params)
	case "pull":
		return p.pull(params)
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

// clone clones a Git repository
func (p *GitPlugin) clone(params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Get parameters
	repo, ok := params["repo"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "repo parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("repo parameter is required")
	}

	branch, _ := params["branch"].(string)
	depth, _ := params["depth"].(int)
	directory, _ := params["directory"].(string)

	// Build command
	args := []string{"clone"}

	if branch != "" {
		args = append(args, "--branch", branch)
	}

	if depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", depth))
	}

	args = append(args, repo)

	if directory != "" {
		args = append(args, directory)
	}

	// Execute command
	cmd := exec.Command("git", args...)
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
		}, fmt.Errorf("error executing git clone: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// checkout checks out a branch
func (p *GitPlugin) checkout(params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Get parameters
	branch, ok := params["branch"].(string)
	if !ok {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     "branch parameter is required",
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("branch parameter is required")
	}

	directory, _ := params["directory"].(string)
	if directory == "" {
		directory = "."
	}

	// Change to directory
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

	if err := os.Chdir(directory); err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error changing to directory %s: %s", directory, err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error changing to directory %s: %w", directory, err)
	}

	defer os.Chdir(currentDir)

	// Execute command
	cmd := exec.Command("git", "checkout", branch)
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
		}, fmt.Errorf("error executing git checkout: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// pull pulls changes from a remote repository
func (p *GitPlugin) pull(params map[string]interface{}) (Result, error) {
	startTime := time.Now()

	// Get parameters
	directory, _ := params["directory"].(string)
	if directory == "" {
		directory = "."
	}

	// Change to directory
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

	if err := os.Chdir(directory); err != nil {
		endTime := time.Now()
		return Result{
			Status:    "error",
			Error:     fmt.Sprintf("error changing to directory %s: %s", directory, err),
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, fmt.Errorf("error changing to directory %s: %w", directory, err)
	}

	defer os.Chdir(currentDir)

	// Execute command
	cmd := exec.Command("git", "pull")
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
		}, fmt.Errorf("error executing git pull: %w", err)
	}

	return Result{
		Status:    "success",
		Output:    string(output),
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}