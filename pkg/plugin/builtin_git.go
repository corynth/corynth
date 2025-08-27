package plugin

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitPlugin implements Git version control operations
type GitPlugin struct{}

// NewGitPlugin creates a new Git plugin
func NewGitPlugin() *GitPlugin {
	return &GitPlugin{}
}

func (p *GitPlugin) Metadata() Metadata {
	return Metadata{
		Name:        "git",
		Version:     "1.0.0",
		Description: "Git version control operations",
		Author:      "Corynth Team",
		Tags:        []string{"git", "vcs", "version-control", "source-control"},
		License:     "Apache-2.0",
	}
}

func (p *GitPlugin) Actions() []Action {
	return []Action{
		{
			Name:        "clone",
			Description: "Clone a Git repository",
			Inputs: map[string]InputSpec{
				"url": {
					Type:        "string",
					Description: "Repository URL to clone",
					Required:    true,
				},
				"path": {
					Type:        "string",
					Description: "Local path to clone to",
					Required:    false,
				},
				"branch": {
					Type:        "string", 
					Description: "Branch to clone",
					Required:    false,
				},
			},
			Outputs: map[string]OutputSpec{
				"path": {
					Type:        "string",
					Description: "Local path of cloned repository",
				},
				"commit": {
					Type:        "string",
					Description: "Latest commit hash",
				},
			},
		},
		{
			Name:        "status",
			Description: "Get Git repository status",
			Inputs: map[string]InputSpec{
				"path": {
					Type:        "string",
					Description: "Repository path",
					Required:    false,
					Default:     ".",
				},
			},
			Outputs: map[string]OutputSpec{
				"clean": {
					Type:        "bool",
					Description: "Whether repository is clean",
				},
				"branch": {
					Type:        "string",
					Description: "Current branch",
				},
			},
		},
		{
			Name:        "commit",
			Description: "Commit changes to repository",
			Inputs: map[string]InputSpec{
				"message": {
					Type:        "string",
					Description: "Commit message",
					Required:    true,
				},
				"path": {
					Type:        "string",
					Description: "Repository path",
					Required:    false,
					Default:     ".",
				},
				"add_all": {
					Type:        "bool",
					Description: "Add all changes before commit",
					Required:    false,
					Default:     true,
				},
			},
			Outputs: map[string]OutputSpec{
				"commit": {
					Type:        "string",
					Description: "New commit hash",
				},
			},
		},
	}
}

func (p *GitPlugin) Validate(params map[string]interface{}) error {
	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git command not found in PATH")
	}
	return nil
}

func (p *GitPlugin) Execute(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	switch action {
	case "clone":
		return p.executeClone(ctx, params)
	case "status":
		return p.executeStatus(ctx, params)
	case "commit":
		return p.executeCommit(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (p *GitPlugin) executeClone(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	url, _ := params["url"].(string)
	if url == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	path, _ := params["path"].(string)
	branch, _ := params["branch"].(string)

	args := []string{"clone"}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, url)
	if path != "" {
		args = append(args, path)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git clone failed: %s", string(output))
	}

	// Get the cloned path and latest commit
	clonePath := path
	if clonePath == "" {
		// Extract repo name from URL
		parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
		clonePath = parts[len(parts)-1]
	}

	// Get latest commit hash
	commitCmd := exec.CommandContext(ctx, "git", "-C", clonePath, "rev-parse", "HEAD")
	commitOutput, err := commitCmd.Output()
	commit := ""
	if err == nil {
		commit = strings.TrimSpace(string(commitOutput))
	}

	return map[string]interface{}{
		"path":   clonePath,
		"commit": commit,
		"output": string(output),
	}, nil
}

func (p *GitPlugin) executeStatus(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	// Check if directory is clean
	statusCmd := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	clean := len(strings.TrimSpace(string(statusOutput))) == 0

	// Get current branch
	branchCmd := exec.CommandContext(ctx, "git", "-C", path, "branch", "--show-current")
	branchOutput, err := branchCmd.Output()
	branch := ""
	if err == nil {
		branch = strings.TrimSpace(string(branchOutput))
	}

	return map[string]interface{}{
		"clean":  clean,
		"branch": branch,
		"status": string(statusOutput),
	}, nil
}

func (p *GitPlugin) executeCommit(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	message, _ := params["message"].(string)
	if message == "" {
		return nil, fmt.Errorf("message parameter is required")
	}

	path := "."
	if p, ok := params["path"].(string); ok {
		path = p
	}

	addAll := true
	if a, ok := params["add_all"].(bool); ok {
		addAll = a
	}

	// Add all changes if requested
	if addAll {
		addCmd := exec.CommandContext(ctx, "git", "-C", path, "add", ".")
		if err := addCmd.Run(); err != nil {
			return nil, fmt.Errorf("git add failed: %w", err)
		}
	}

	// Commit changes
	commitCmd := exec.CommandContext(ctx, "git", "-C", path, "commit", "-m", message)
	commitOutput, err := commitCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git commit failed: %s", string(commitOutput))
	}

	// Get new commit hash
	hashCmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "HEAD")
	hashOutput, err := hashCmd.Output()
	commit := ""
	if err == nil {
		commit = strings.TrimSpace(string(hashOutput))
	}

	return map[string]interface{}{
		"commit": commit,
		"output": string(commitOutput),
	}, nil
}